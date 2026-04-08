package handler

import (
	"ai-gateway/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LLMHandler LLM处理器
type LLMHandler struct {
	llmService *service.LLMService
}

// NewLLMHandler 创建LLM处理器
func NewLLMHandler() *LLMHandler {
	return &LLMHandler{
		llmService: service.NewLLMService(),
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

// ChatCompletion 对话完成
func (h *LLMHandler) ChatCompletion(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req service.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.llmService.ChatCompletion(c, userID.(uint), &req)
}

// ListModels 获取可用模型列表
func (h *LLMHandler) ListModels(c *gin.Context) {
	models, err := h.llmService.ListModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   models,
		"object": "list",
	})
}
