package llm

import (
	"ai-gateway/internal/model"
	"ai-gateway/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client LLM客户端
type Client struct {
	model  *model.AIModel
	httpClient *http.Client
}

// NewClient 创建新的LLM客户端
func NewClient(m *model.AIModel) *Client {
	timeout := time.Duration(m.Timeout) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	
	return &Client{
		model: m,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ChatCompletion 普通对话请求
func (c *Client) ChatCompletion(ctx context.Context, req *model.LLMRequest) (*model.LLMResponse, error) {
	// 根据提供商构建请求
	var body []byte
	var err error
	
	switch c.model.Provider {
	case model.ProviderOpenAI, model.ProviderAzure:
		body, err = json.Marshal(req)
	default:
		body, err = json.Marshal(req)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// 构建请求URL
	url := c.model.BaseURL
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += "chat/completions"
	
	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.model.APIKey)
	
	// 发送请求
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()
	
	// 读取响应
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// 检查状态码
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", httpResp.StatusCode, string(respBody))
	}
	
	// 解析响应
	var resp model.LLMResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &resp, nil
}

// ChatCompletionStream 流式对话请求
func (c *Client) ChatCompletionStream(ctx context.Context, req *model.LLMRequest) (io.ReadCloser, error) {
	// 设置stream为true
	req.Stream = true
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// 构建请求URL
	url := c.model.BaseURL
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += "chat/completions"
	
	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.model.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")
	
	// 发送请求
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	
	// 检查状态码
	if httpResp.StatusCode != http.StatusOK {
		httpResp.Body.Close()
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", httpResp.StatusCode, string(respBody))
	}
	
	return httpResp.Body, nil
}

// ParseStreamData 解析流式数据
func ParseStreamData(data []byte) (*model.LLMStreamResponse, error) {
	// 去掉 "data: " 前缀
	str := string(data)
	str = strings.TrimPrefix(str, "data: ")
	str = strings.TrimSpace(str)
	
	// 检查是否是结束标记
	if str == "[DONE]" {
		return nil, io.EOF
	}
	
	var resp model.LLMStreamResponse
	if err := json.Unmarshal([]byte(str), &resp); err != nil {
		return nil, err
	}
	
	return &resp, nil
}

// GetClientByModelName 根据模型名称获取客户端
func GetClientByModelName(modelName string) (*Client, error) {
	// 先从缓存获取
	var m *model.AIModel
	
	// 尝试作为ID获取
	var modelID uint
	if _, err := fmt.Sscanf(modelName, "%d", &modelID); err == nil {
		cached, err := repository.GetCachedModel(modelID)
		if err == nil {
			m = cached
		}
	}
	
	// 如果缓存中没有，从数据库获取
	if m == nil {
		db := repository.GetDB()
		if err := db.Where("name = ? OR model_id = ?", modelName, modelName).First(&m).Error; err != nil {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		// 缓存模型信息
		repository.CacheModel(m, 5*time.Minute)
	}
	
	if m.Status != model.ModelStatusActive {
		return nil, fmt.Errorf("model is not active: %s", modelName)
	}
	
	return NewClient(m), nil
}

// GetDefaultClient 获取默认客户端
func GetDefaultClient() (*Client, error) {
	// 先从缓存获取
	m, err := repository.GetCachedDefaultModel()
	if err == nil {
		return NewClient(m), nil
	}
	
	// 从数据库获取
	db := repository.GetDB()
	if err := db.Where("is_default = ? AND status = ?", true, model.ModelStatusActive).First(&m).Error; err != nil {
		return nil, fmt.Errorf("no default model found")
	}
	
	// 缓存
	repository.CacheDefaultModel(m, 5*time.Minute)
	
	return NewClient(m), nil
}
