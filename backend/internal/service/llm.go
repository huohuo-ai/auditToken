package service

import (
	"ai-gateway/internal/model"
	"ai-gateway/internal/repository"
	"ai-gateway/pkg/llm"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// LLMService LLM服务
type LLMService struct {
	modelService *AIModelService
}

// NewLLMService 创建LLM服务
func NewLLMService() *LLMService {
	return &LLMService{
		modelService: NewAIModelService(),
	}
}

// ChatCompletionRequest 对话请求
type ChatCompletionRequest struct {
	Model       string    `json:"model" binding:"required"`
	Messages    []Message `json:"messages" binding:"required"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletion 处理对话请求
func (s *LLMService) ChatCompletion(c *gin.Context, userID uint, req *ChatCompletionRequest) {
	// 获取模型
	var m *model.AIModel
	var err error
	
	if req.Model == "" {
		m, err = s.modelService.GetDefaultModel()
	} else {
		// 尝试从缓存/数据库获取模型
		client, err := llm.GetClientByModelName(req.Model)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 获取模型信息
		db := repository.GetDB()
		db.Where("name = ? OR model_id = ?", req.Model, req.Model).First(&m)
		if m != nil {
			// 使用缓存的client
			_ = client
		}
	}
	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model not found: " + err.Error()})
		return
	}
	
	if m == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model not found"})
		return
	}
	
	// 创建LLM客户端
	client := llm.NewClient(m)
	
	// 构建LLM请求
	llmReq := &model.LLMRequest{
		Model:       m.ModelID,
		Messages:    convertMessages(req.Messages),
		Stream:      req.Stream,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
	}
	
	// 如果未设置温度，使用模型默认温度
	if llmReq.Temperature == 0 {
		llmReq.Temperature = m.Temperature
	}
	if llmReq.MaxTokens == 0 {
		llmReq.MaxTokens = m.MaxTokens
	}
	
	// 检查配额（更精确的预估）
	estimatedTokens := s.estimateTokens(req.Messages, req.MaxTokens)
	if allowed, reason := s.checkQuota(userID, estimatedTokens); !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": gin.H{
				"message": "Quota exceeded: " + reason,
				"type":    "quota_exceeded",
			},
		})
		return
	}
	
	// 设置请求上下文
	ctx := context.Background()
	requestID := uuid.New().String()
	c.Set("requestID", requestID)
	c.Set("isStream", req.Stream)
	
	if req.Stream {
		// 流式响应
		s.handleStreamResponse(c, client, llmReq, m, userID, requestID)
	} else {
		// 普通响应
		s.handleNormalResponse(c, client, llmReq, m, userID, requestID)
	}
}

// handleNormalResponse 处理普通响应
func (s *LLMService) handleNormalResponse(c *gin.Context, client *llm.Client, req *model.LLMRequest, m *model.AIModel, userID uint, requestID string) {
	startTime := time.Now()
	
	resp, err := client.ChatCompletion(c.Request.Context(), req)
	if err != nil {
		logrus.WithError(err).Error("LLM request failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "llm_error",
			},
		})
		return
	}
	
	latency := time.Since(startTime).Milliseconds()
	
	// 更新配额使用
	s.updateQuotaUsage(userID, resp.Usage.TotalTokens)
	
	// 记录使用日志
	s.recordUsage(userID, requestID, m.Name, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens, latency, "")
	
	// 设置响应头
	c.Header("X-Request-ID", requestID)
	c.Header("X-Response-Time", fmt.Sprintf("%dms", latency))
	
	c.JSON(http.StatusOK, resp)
}

// handleStreamResponse 处理流式响应
func (s *LLMService) handleStreamResponse(c *gin.Context, client *llm.Client, req *model.LLMRequest, m *model.AIModel, userID uint, requestID string) {
	startTime := time.Now()
	
	streamReader, err := client.ChatCompletionStream(c.Request.Context(), req)
	if err != nil {
		logrus.WithError(err).Error("LLM stream request failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "llm_error",
			},
		})
		return
	}
	defer streamReader.Close()
	
	// 设置流式响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Request-ID", requestID)
	
	var fullContent strings.Builder
	var promptTokens, completionTokens, totalTokens int64
	
	// 使用bufio.Scanner读取流式数据
	scanner := bufio.NewScanner(streamReader)
	scanner.Split(scanLines)
	
	c.Stream(func(w io.Writer) bool {
		if !scanner.Scan() {
			return false
		}
		
		line := scanner.Text()
		if line == "" {
			return true
		}
		
		// 处理data: 前缀
		if !strings.HasPrefix(line, "data: ") {
			return true
		}
		
		data := strings.TrimPrefix(line, "data: ")
		data = strings.TrimSpace(data)
		
		// 检查结束标记
		if data == "[DONE]" {
			// 发送结束标记
			fmt.Fprintf(w, "data: [DONE]\n\n")
			
			// 记录使用日志
			latency := time.Since(startTime).Milliseconds()
			s.updateQuotaUsage(userID, totalTokens)
			s.recordUsage(userID, requestID, m.Name, promptTokens, completionTokens, totalTokens, latency, "")
			
			return false
		}
		
		// 解析流式响应
		var streamResp model.LLMStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			return true
		}
		
		// 累积内容
		if len(streamResp.Choices) > 0 {
			fullContent.WriteString(streamResp.Choices[0].Delta.Content)
			completionTokens = int64(len(streamResp.Choices[0].Delta.Content)) / 4 // 粗略估计
		}
		
		// 转发数据
		fmt.Fprintf(w, "data: %s\n\n", data)
		return true
	})
}

// scanLines 自定义行分割函数
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
		return i + 2, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// convertMessages 转换消息格式
func convertMessages(messages []Message) []model.Message {
	result := make([]model.Message, len(messages))
	for i, m := range messages {
		result[i] = model.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return result
}

// estimateTokens 预估token数量
func (s *LLMService) estimateTokens(messages []Message, maxTokens int) int64 {
	// 简单的估算：每个字符约0.25个token
	var total int
	for _, m := range messages {
		total += len(m.Content)
	}
	estimated := int64(total / 4)
	if maxTokens > 0 {
		estimated += int64(maxTokens)
	} else {
		estimated += 2000 // 默认输出
	}
	return estimated
}

// checkQuota 检查配额
func (s *LLMService) checkQuota(userID uint, estimatedTokens int64) (bool, string) {
	// 获取用户配额
	db := repository.GetDB()
	var quota model.UserQuota
	if err := db.Where("user_id = ?", userID).First(&quota).Error; err != nil {
		// 如果没有配额记录，创建默认配额
		quota = model.UserQuota{
			UserID:       userID,
			DailyLimit:   100000,
			WeeklyLimit:  500000,
			MonthlyLimit: 2000000,
		}
		db.Create(&quota)
	}
	
	// 检查是否需要重置配额
	now := time.Now()
	if now.Sub(quota.LastResetDaily) >= 24*time.Hour {
		quota.DailyUsed = 0
		quota.LastResetDaily = now
		db.Model(&quota).Updates(map[string]interface{}{
			"daily_used":       0,
			"last_reset_daily": now,
		})
	}
	
	return quota.CheckQuota(estimatedTokens)
}

// updateQuotaUsage 更新配额使用
func (s *LLMService) updateQuotaUsage(userID uint, tokens int) {
	db := repository.GetDB()
	db.Model(&model.UserQuota{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"daily_used":   gorm.Expr("daily_used + ?", tokens),
		"weekly_used":  gorm.Expr("weekly_used + ?", tokens),
		"monthly_used": gorm.Expr("monthly_used + ?", tokens),
	})
}

// recordUsage 记录使用日志
func (s *LLMService) recordUsage(userID uint, requestID, modelName string, promptTokens, completionTokens, totalTokens int, latency int64, errorMsg string) {
	log := &model.UsageLog{
		UserID:           userID,
		RequestID:        requestID,
		ModelName:        modelName,
		PromptTokens:     int64(promptTokens),
		CompletionTokens: int64(completionTokens),
		TotalTokens:      int64(totalTokens),
		RequestTime:      time.Now().Add(-time.Duration(latency) * time.Millisecond),
		ResponseTime:     time.Now(),
		LatencyMs:        latency,
		Status:           "success",
		ErrorMessage:     errorMsg,
	}
	
	if errorMsg != "" {
		log.Status = "failed"
	}
	
	db := repository.GetDB()
	if err := db.Create(log).Error; err != nil {
		logrus.WithError(err).Error("Failed to record usage log")
	}
}

// ListModels 获取可用模型列表
func (s *LLMService) ListModels() ([]map[string]interface{}, error) {
	models, err := s.modelService.GetActiveModels()
	if err != nil {
		return nil, err
	}
	
	result := make([]map[string]interface{}, 0, len(models))
	for _, m := range models {
		result = append(result, m.ToPublic())
	}
	
	return result, nil
}
