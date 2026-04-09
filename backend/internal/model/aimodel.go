package model

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AIModelStatus AI模型状态
type AIModelStatus string

const (
	ModelStatusActive   AIModelStatus = "active"
	ModelStatusInactive AIModelStatus = "inactive"
)

// AIModelProvider 模型提供商
type AIModelProvider string

const (
	ProviderOpenAI    AIModelProvider = "openai"
	ProviderAzure     AIModelProvider = "azure"
	ProviderAnthropic AIModelProvider = "anthropic"
	ProviderClaude    AIModelProvider = "claude"
	ProviderGemini    AIModelProvider = "gemini"
	ProviderCustom    AIModelProvider = "custom"
)

// AIModel AI模型配置
type AIModel struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	UUID            string          `json:"uuid" gorm:"uniqueIndex;size:36"`
	Name            string          `json:"name" gorm:"size:50;index"`
	DisplayName     string          `json:"display_name" gorm:"size:100"`
	Provider        AIModelProvider `json:"provider" gorm:"size:50"`
	BaseURL         string          `json:"base_url" gorm:"size:500"`
	APIKey          string          `json:"-" gorm:"size:500"` // 不返回给前端
	ModelID         string          `json:"model_id" gorm:"size:100"` // 实际调用时使用的模型ID
	Status          AIModelStatus   `json:"status" gorm:"size:20;default:'active'"`
	MaxTokens       int             `json:"max_tokens" gorm:"default:4096"`
	Temperature     float64         `json:"temperature" gorm:"default:0.7"`
	Timeout         int             `json:"timeout" gorm:"default:60"` // 秒
	RateLimitRPM    int             `json:"rate_limit_rpm" gorm:"default:60"` // 每分钟请求数限制
	RateLimitTPM    int             `json:"rate_limit_tpm" gorm:"default:100000"` // 每分钟token数限制
	IsDefault       bool            `json:"is_default" gorm:"default:false"`
	SystemPrompt    string          `json:"system_prompt" gorm:"type:text"` // 系统提示词
	Description     string          `json:"description" gorm:"size:500"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `json:"-" gorm:"index"`
}

// BeforeCreate 创建前钩子
func (m *AIModel) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == "" {
		m.UUID = uuid.New().String()
	}
	return nil
}

// MaskAPIKey 脱敏显示API Key
func (m *AIModel) MaskAPIKey() string {
	if len(m.APIKey) <= 8 {
		return "****"
	}
	return m.APIKey[:4] + "****" + m.APIKey[len(m.APIKey)-4:]
}

// ToPublic 转换为公开信息（隐藏敏感字段）
func (m *AIModel) ToPublic() map[string]interface{} {
	return map[string]interface{}{
		"id":            m.ID,
		"uuid":          m.UUID,
		"name":          m.Name,
		"display_name":  m.DisplayName,
		"provider":      m.Provider,
		"model_id":      m.ModelID,
		"status":        m.Status,
		"max_tokens":    m.MaxTokens,
		"temperature":   m.Temperature,
		"is_default":    m.IsDefault,
		"system_prompt": m.SystemPrompt,
		"description":   m.Description,
		"created_at":    m.CreatedAt,
	}
}

// ModelAccess 模型访问权限（可以控制哪些用户/部门可以访问哪些模型）
type ModelAccess struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ModelID   uint      `json:"model_id" gorm:"index"`
	UserID    *uint     `json:"user_id" gorm:"index"`     // 为空表示所有用户
	IsAllowed bool      `json:"is_allowed" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LLMRequest LLM请求结构
type LLMRequest struct {
	Model       string         `json:"model"`
	Messages    []Message      `json:"messages"`
	Stream      bool           `json:"stream,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse LLM响应结构
type LLMResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
		TotalTokens      int64 `json:"total_tokens"`
	} `json:"usage"`
}

// LLMStreamResponse LLM流式响应结构
type LLMStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}
