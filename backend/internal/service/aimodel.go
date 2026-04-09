package service

import (
	"ai-gateway/internal/model"
	"ai-gateway/internal/repository"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// AIModelService AI模型服务
type AIModelService struct {
	db *gorm.DB
}

// NewAIModelService 创建AI模型服务
func NewAIModelService() *AIModelService {
	return &AIModelService{
		db: repository.GetDB(),
	}
}

// CreateModelRequest 创建模型请求
type CreateModelRequest struct {
	Name            string                `json:"name" binding:"required"`
	DisplayName     string                `json:"display_name" binding:"required"`
	Provider        model.AIModelProvider `json:"provider" binding:"required"`
	BaseURL         string                `json:"base_url" binding:"required,url"`
	APIKey          string                `json:"api_key" binding:"required"`
	ModelID         string                `json:"model_id" binding:"required"`
	MaxTokens       int                   `json:"max_tokens,omitempty"`
	Temperature     float64               `json:"temperature,omitempty"`
	Timeout         int                   `json:"timeout,omitempty"`
	RateLimitRPM    int                   `json:"rate_limit_rpm,omitempty"`
	RateLimitTPM    int                   `json:"rate_limit_tpm,omitempty"`
	IsDefault       bool                  `json:"is_default,omitempty"`
	SystemPrompt    string                `json:"system_prompt,omitempty"`
	Description     string                `json:"description,omitempty"`
}

// UpdateModelRequest 更新模型请求
type UpdateModelRequest struct {
	Name            string                `json:"name,omitempty"`
	DisplayName     string                `json:"display_name,omitempty"`
	Provider        model.AIModelProvider `json:"provider,omitempty"`
	BaseURL         string                `json:"base_url,omitempty"`
	APIKey          string                `json:"api_key,omitempty"`
	ModelID         string                `json:"model_id,omitempty"`
	Status          model.AIModelStatus   `json:"status,omitempty"`
	MaxTokens       int                   `json:"max_tokens,omitempty"`
	Temperature     float64               `json:"temperature,omitempty"`
	Timeout         int                   `json:"timeout,omitempty"`
	RateLimitRPM    int                   `json:"rate_limit_rpm,omitempty"`
	RateLimitTPM    int                   `json:"rate_limit_tpm,omitempty"`
	IsDefault       bool                  `json:"is_default,omitempty"`
	SystemPrompt    string                `json:"system_prompt,omitempty"`
	Description     string                `json:"description,omitempty"`
}

// CreateModel 创建模型
func (s *AIModelService) CreateModel(req *CreateModelRequest) (*model.AIModel, error) {
	// 检查名称是否已存在
	var count int64
	if err := s.db.Model(&model.AIModel{}).Where("name = ?", req.Name).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("model name already exists")
	}
	
	// 如果设置为默认模型，取消其他模型的默认状态
	if req.IsDefault {
		if err := s.db.Model(&model.AIModel{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}
	
	m := &model.AIModel{
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Provider:     req.Provider,
		BaseURL:      req.BaseURL,
		APIKey:       req.APIKey,
		ModelID:      req.ModelID,
		Status:       model.ModelStatusActive,
		MaxTokens:    req.MaxTokens,
		Temperature:  req.Temperature,
		Timeout:      req.Timeout,
		RateLimitRPM: req.RateLimitRPM,
		RateLimitTPM: req.RateLimitTPM,
		IsDefault:    req.IsDefault,
		SystemPrompt: req.SystemPrompt,
		Description:  req.Description,
	}
	
	if m.MaxTokens == 0 {
		m.MaxTokens = 4096
	}
	if m.Temperature == 0 {
		m.Temperature = 0.7
	}
	if m.Timeout == 0 {
		m.Timeout = 60
	}
	if m.RateLimitRPM == 0 {
		m.RateLimitRPM = 60
	}
	if m.RateLimitTPM == 0 {
		m.RateLimitTPM = 100000
	}
	
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	
	return m, nil
}

// GetModelByID 根据ID获取模型
func (s *AIModelService) GetModelByID(id uint) (*model.AIModel, error) {
	var m model.AIModel
	if err := s.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("model not found")
		}
		return nil, err
	}
	return &m, nil
}

// GetModelByName 根据名称获取模型
func (s *AIModelService) GetModelByName(name string) (*model.AIModel, error) {
	var m model.AIModel
	if err := s.db.Where("name = ?", name).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("model not found")
		}
		return nil, err
	}
	return &m, nil
}

// UpdateModel 更新模型
func (s *AIModelService) UpdateModel(id uint, req *UpdateModelRequest) (*model.AIModel, error) {
	m, err := s.GetModelByID(id)
	if err != nil {
		return nil, err
	}
	
	// 如果设置为默认模型，取消其他模型的默认状态
	if req.IsDefault && !m.IsDefault {
		if err := s.db.Model(&model.AIModel{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}
	
	updates := make(map[string]interface{})
	
	if req.Name != "" && req.Name != m.Name {
		var count int64
		s.db.Model(&model.AIModel{}).Where("name = ?", req.Name).Count(&count)
		if count > 0 {
			return nil, errors.New("model name already exists")
		}
		updates["name"] = req.Name
	}
	
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Provider != "" {
		updates["provider"] = req.Provider
	}
	if req.BaseURL != "" {
		updates["base_url"] = req.BaseURL
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.ModelID != "" {
		updates["model_id"] = req.ModelID
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.MaxTokens > 0 {
		updates["max_tokens"] = req.MaxTokens
	}
	if req.Temperature >= 0 {
		updates["temperature"] = req.Temperature
	}
	if req.Timeout > 0 {
		updates["timeout"] = req.Timeout
	}
	if req.RateLimitRPM > 0 {
		updates["rate_limit_rpm"] = req.RateLimitRPM
	}
	if req.RateLimitTPM > 0 {
		updates["rate_limit_tpm"] = req.RateLimitTPM
	}
	if req.IsDefault != m.IsDefault {
		updates["is_default"] = req.IsDefault
	}
	// SystemPrompt可以为空字符串，所以不检查是否为空
	updates["system_prompt"] = req.SystemPrompt
	if req.Description != "" {
		updates["description"] = req.Description
	}
	
	if len(updates) > 0 {
		if err := s.db.Model(m).Updates(updates).Error; err != nil {
			return nil, err
		}
		// 更新缓存
		repository.GetRedis().Del(context.Background(), fmt.Sprintf("model:id:%d", id))
		if m.IsDefault {
			repository.GetRedis().Del(context.Background(), "model:default")
		}
	}
	
	return s.GetModelByID(id)
}

// DeleteModel 删除模型
func (s *AIModelService) DeleteModel(id uint) error {
	m, err := s.GetModelByID(id)
	if err != nil {
		return err
	}
	
	// 删除缓存
	repository.GetRedis().Del(context.Background(), fmt.Sprintf("model:id:%d", id))
	if m.IsDefault {
		repository.GetRedis().Del(context.Background(), "model:default")
	}
	
	return s.db.Delete(m).Error
}

// ListModels 获取模型列表
func (s *AIModelService) ListModels(page, pageSize int, provider string, status string) ([]model.AIModel, int64, error) {
	var models []model.AIModel
	var total int64
	
	query := s.db.Model(&model.AIModel{})
	
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	
	return models, total, nil
}

// GetActiveModels 获取所有活跃模型
func (s *AIModelService) GetActiveModels() ([]model.AIModel, error) {
	var models []model.AIModel
	if err := s.db.Where("status = ?", model.ModelStatusActive).Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

// GetDefaultModel 获取默认模型
func (s *AIModelService) GetDefaultModel() (*model.AIModel, error) {
	var m model.AIModel
	if err := s.db.Where("is_default = ? AND status = ?", true, model.ModelStatusActive).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有默认模型，返回第一个活跃的模型
			if err := s.db.Where("status = ?", model.ModelStatusActive).First(&m).Error; err != nil {
				return nil, errors.New("no active model found")
			}
			return &m, nil
		}
		return nil, err
	}
	return &m, nil
}
