package handler

import (
	"ai-gateway/internal/model"
	"ai-gateway/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AIModelHandler AI模型处理器
type AIModelHandler struct {
	modelService *service.AIModelService
}

// NewAIModelHandler 创建AI模型处理器
func NewAIModelHandler() *AIModelHandler {
	return &AIModelHandler{
		modelService: service.NewAIModelService(),
	}
}

// CreateModelRequest 创建模型请求
type CreateModelRequest struct {
	Name            string  `json:"name" binding:"required"`
	DisplayName     string  `json:"display_name" binding:"required"`
	Provider        string  `json:"provider" binding:"required"`
	BaseURL         string  `json:"base_url" binding:"required,url"`
	APIKey          string  `json:"api_key" binding:"required"`
	ModelID         string  `json:"model_id" binding:"required"`
	MaxTokens       int     `json:"max_tokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
	Timeout         int     `json:"timeout,omitempty"`
	RateLimitRPM    int     `json:"rate_limit_rpm,omitempty"`
	RateLimitTPM    int     `json:"rate_limit_tpm,omitempty"`
	IsDefault       bool    `json:"is_default,omitempty"`
	Description     string  `json:"description,omitempty"`
}

// UpdateModelRequest 更新模型请求
type UpdateModelRequest struct {
	Name            string  `json:"name,omitempty"`
	DisplayName     string  `json:"display_name,omitempty"`
	Provider        string  `json:"provider,omitempty"`
	BaseURL         string  `json:"base_url,omitempty"`
	APIKey          string  `json:"api_key,omitempty"`
	ModelID         string  `json:"model_id,omitempty"`
	Status          string  `json:"status,omitempty"`
	MaxTokens       int     `json:"max_tokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
	Timeout         int     `json:"timeout,omitempty"`
	RateLimitRPM    int     `json:"rate_limit_rpm,omitempty"`
	RateLimitTPM    int     `json:"rate_limit_tpm,omitempty"`
	IsDefault       bool    `json:"is_default,omitempty"`
	Description     string  `json:"description,omitempty"`
}

// CreateModel 创建模型
func (h *AIModelHandler) CreateModel(c *gin.Context) {
	var req CreateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createReq := &service.CreateModelRequest{
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Provider:     model.AIModelProvider(req.Provider),
		BaseURL:      req.BaseURL,
		APIKey:       req.APIKey,
		ModelID:      req.ModelID,
		MaxTokens:    req.MaxTokens,
		Temperature:  req.Temperature,
		Timeout:      req.Timeout,
		RateLimitRPM: req.RateLimitRPM,
		RateLimitTPM: req.RateLimitTPM,
		IsDefault:    req.IsDefault,
		Description:  req.Description,
	}

	m, err := h.modelService.CreateModel(createReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, m)
}

// GetModel 获取模型详情
func (h *AIModelHandler) GetModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	m, err := h.modelService.GetModelByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m.ToPublic())
}

// UpdateModel 更新模型
func (h *AIModelHandler) UpdateModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var req UpdateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := &service.UpdateModelRequest{
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Provider:     model.AIModelProvider(req.Provider),
		BaseURL:      req.BaseURL,
		APIKey:       req.APIKey,
		ModelID:      req.ModelID,
		Status:       model.AIModelStatus(req.Status),
		MaxTokens:    req.MaxTokens,
		Temperature:  req.Temperature,
		Timeout:      req.Timeout,
		RateLimitRPM: req.RateLimitRPM,
		RateLimitTPM: req.RateLimitTPM,
		IsDefault:    req.IsDefault,
		Description:  req.Description,
	}

	m, err := h.modelService.UpdateModel(uint(id), updateReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m.ToPublic())
}

// DeleteModel 删除模型
func (h *AIModelHandler) DeleteModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	if err := h.modelService.DeleteModel(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "model deleted successfully"})
}

// ListModels 获取模型列表
func (h *AIModelHandler) ListModels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	provider := c.Query("provider")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	models, total, err := h.modelService.ListModels(page, pageSize, provider, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为公开信息
	var publicModels []map[string]interface{}
	for _, m := range models {
		publicModels = append(publicModels, m.ToPublic())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      publicModels,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetActiveModels 获取活跃模型列表（公开API）
func (h *AIModelHandler) GetActiveModels(c *gin.Context) {
	models, err := h.modelService.GetActiveModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []map[string]interface{}
	for _, m := range models {
		result = append(result, m.ToPublic())
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}
