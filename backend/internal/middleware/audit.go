package middleware

import (
	"ai-gateway/internal/model"
	"ai-gateway/internal/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AuditLog 审计日志中间件
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只记录API请求
		if c.Request.URL.Path == "/api/v1/auth/login" || 
		   c.Request.URL.Path == "/api/v1/auth/register" ||
		   c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		
		// 生成请求ID
		requestID := uuid.New().String()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		
		// 记录请求开始时间
		startTime := time.Now()
		c.Set("requestStartTime", startTime)
		
		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}
		c.Set("requestBody", string(requestBody))
		
		// 包装ResponseWriter以捕获响应
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		
		c.Next()
		
		// 记录审计日志
		go func() {
			if err := recordAuditLog(c, requestID, startTime, string(requestBody), blw.body.String()); err != nil {
				logrus.WithError(err).Error("failed to record audit log")
			}
		}()
	}
}

// bodyLogWriter 用于捕获响应体的writer
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// recordAuditLog 记录审计日志
func recordAuditLog(c *gin.Context, requestID string, startTime time.Time, requestBody, responseBody string) error {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	
	uid := uint64(0)
	uname := ""
	uemail := ""
	
	if userID != nil {
		uid = uint64(userID.(uint))
		if username != nil {
			uname = username.(string)
		}
		// 获取用户邮箱
		if user, err := GetCurrentUser(c); err == nil && user != nil {
			uemail = user.Email
		}
	}
	
	// 提取模型名称
	modelName := c.Param("model")
	if modelName == "" {
		// 从请求体中解析
		var reqBody map[string]interface{}
		if err := json.Unmarshal([]byte(requestBody), &reqBody); err == nil {
			if m, ok := reqBody["model"].(string); ok {
				modelName = m
			}
		}
	}
	
	// 解析token使用情况
	var promptTokens, completionTokens, totalTokens int64
	var responseStatus int = c.Writer.Status()
	var hasError bool
	var errorMessage string
	
	// 解析响应体获取token使用情况
	var respBody map[string]interface{}
	if err := json.Unmarshal([]byte(responseBody), &respBody); err == nil {
		if usage, ok := respBody["usage"].(map[string]interface{}); ok {
			if pt, ok := usage["prompt_tokens"].(float64); ok {
				promptTokens = int64(pt)
			}
			if ct, ok := usage["completion_tokens"].(float64); ok {
				completionTokens = int64(ct)
			}
			if tt, ok := usage["total_tokens"].(float64); ok {
				totalTokens = int64(tt)
			}
		}
		if _, ok := respBody["error"]; ok {
			hasError = true
			if errMsg, ok := respBody["error"].(map[string]interface{})["message"].(string); ok {
				errorMessage = errMsg
			}
		}
	}
	
	// 截断过长的内容
	maxLength := 10000
	if len(requestBody) > maxLength {
		requestBody = requestBody[:maxLength] + "..."
	}
	if len(responseBody) > maxLength {
		responseBody = responseBody[:maxLength] + "..."
	}
	
	// 构建审计日志
	auditLog := &model.AuditLog{
		Timestamp:        time.Now(),
		RequestID:        requestID,
		UserID:           uid,
		UserName:         uname,
		UserEmail:        uemail,
		RequestTime:      startTime,
		RequestMethod:    c.Request.Method,
		RequestPath:      c.Request.URL.Path,
		RequestIP:        c.ClientIP(),
		UserAgent:        c.Request.UserAgent(),
		RequestHeaders:   headersToJSON(c.Request.Header),
		RequestBody:      requestBody,
		ModelName:        modelName,
		ModelProvider:    "", // 可以从配置获取
		ResponseTime:     time.Now(),
		ResponseStatus:   responseStatus,
		ResponseBody:     responseBody,
		ResponseHeaders:  headersToJSON(c.Writer.Header()),
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		LatencyMs:        time.Since(startTime).Milliseconds(),
		IsStream:         c.GetBool("isStream"),
		HasError:         hasError || responseStatus >= 400,
		ErrorMessage:     errorMessage,
	}
	
	// 保存到ClickHouse
	if err := repository.InsertAuditLog(auditLog); err != nil {
		return err
	}
	
	// 同时保存到MySQL用于实时统计
	usageLog := &model.UsageLog{
		UserID:           uint(uid),
		RequestID:        requestID,
		ModelName:        modelName,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		RequestTime:      startTime,
		ResponseTime:     time.Now(),
		LatencyMs:        time.Since(startTime).Milliseconds(),
		Status:           getStatusString(responseStatus),
		ErrorMessage:     errorMessage,
		IP:               c.ClientIP(),
		UserAgent:        c.Request.UserAgent(),
	}
	
	if err := repository.GetDB().Create(usageLog).Error; err != nil {
		logrus.WithError(err).Error("failed to save usage log to MySQL")
	}
	
	return nil
}

// headersToJSON 将Header转换为JSON字符串
func headersToJSON(headers map[string][]string) string {
	data, _ := json.Marshal(headers)
	return string(data)
}

// getStatusString 获取状态字符串
func getStatusString(status int) string {
	if status >= 200 && status < 300 {
		return "success"
	}
	return "failed"
}
