package model

import (
	"time"
)

// RiskLevel 风险等级
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// RiskType 风险类型
type RiskType string

const (
	RiskTypeTokenAbuse        RiskType = "token_abuse"        // Token滥用
	RiskTypeOffHoursAccess    RiskType = "off_hours_access"   // 非工作时间访问
	RiskTypeSensitiveInfo     RiskType = "sensitive_info"     // 尝试获取敏感信息
	RiskTypeAbnormalFrequency RiskType = "abnormal_frequency" // 异常请求频率
	RiskTypeAbnormalPattern   RiskType = "abnormal_pattern"   // 异常请求模式
	RiskTypeIPAnomaly         RiskType = "ip_anomaly"         // IP异常
	RiskTypeModelAbuse        RiskType = "model_abuse"        // 模型滥用
)

// AuditLog 审计日志（ClickHouse存储）
type AuditLog struct {
	// 基础信息
	Timestamp       time.Time `json:"timestamp" ch:"timestamp"`
	RequestID       string    `json:"request_id" ch:"request_id"`
	UserID          uint64    `json:"user_id" ch:"user_id"`
	UserName        string    `json:"user_name" ch:"user_name"`
	UserEmail       string    `json:"user_email" ch:"user_email"`
	
	// 请求信息
	RequestTime     time.Time `json:"request_time" ch:"request_time"`
	RequestMethod   string    `json:"request_method" ch:"request_method"`
	RequestPath     string    `json:"request_path" ch:"request_path"`
	RequestIP       string    `json:"request_ip" ch:"request_ip"`
	UserAgent       string    `json:"user_agent" ch:"user_agent"`
	RequestHeaders  string    `json:"request_headers" ch:"request_headers"`
	RequestBody     string    `json:"request_body" ch:"request_body"`
	
	// 模型信息
	ModelName       string    `json:"model_name" ch:"model_name"`
	ModelProvider   string    `json:"model_provider" ch:"model_provider"`
	
	// 响应信息
	ResponseTime    time.Time `json:"response_time" ch:"response_time"`
	ResponseStatus  int       `json:"response_status" ch:"response_status"`
	ResponseBody    string    `json:"response_body" ch:"response_body"`
	ResponseHeaders string    `json:"response_headers" ch:"response_headers"`
	
	// Token使用
	PromptTokens    int64     `json:"prompt_tokens" ch:"prompt_tokens"`
	CompletionTokens int64    `json:"completion_tokens" ch:"completion_tokens"`
	TotalTokens     int64     `json:"total_tokens" ch:"total_tokens"`
	
	// 性能指标
	LatencyMs       int64     `json:"latency_ms" ch:"latency_ms"`
	
	// 审计相关
	IsStream        bool      `json:"is_stream" ch:"is_stream"`
	HasError        bool      `json:"has_error" ch:"has_error"`
	ErrorMessage    string    `json:"error_message" ch:"error_message"`
}

// TableName ClickHouse表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// RiskEvent 风险事件
type RiskEvent struct {
	// 基础信息
	Timestamp    time.Time `json:"timestamp" ch:"timestamp"`
	EventID      string    `json:"event_id" ch:"event_id"`
	RequestID    string    `json:"request_id" ch:"request_id"`
	UserID       uint64    `json:"user_id" ch:"user_id"`
	UserName     string    `json:"user_name" ch:"user_name"`
	
	// 风险信息
	RiskLevel    string    `json:"risk_level" ch:"risk_level"`
	RiskType     string    `json:"risk_type" ch:"risk_type"`
	RiskScore    float64   `json:"risk_score" ch:"risk_score"`
	RiskReason   string    `json:"risk_reason" ch:"risk_reason"`
	
	// 详细信息
	Description  string    `json:"description" ch:"description"`
	Evidence     string    `json:"evidence" ch:"evidence"`
	RequestIP    string    `json:"request_ip" ch:"request_ip"`
	ModelName    string    `json:"model_name" ch:"model_name"`
	
	// 状态
	IsResolved   bool      `json:"is_resolved" ch:"is_resolved"`
	ResolvedBy   string    `json:"resolved_by" ch:"resolved_by"`
	ResolvedAt   time.Time `json:"resolved_at" ch:"resolved_at"`
	Note         string    `json:"note" ch:"note"`
}

// TableName ClickHouse表名
func (RiskEvent) TableName() string {
	return "risk_events"
}

// UserBehaviorSummary 用户行为汇总
type UserBehaviorSummary struct {
	Date              string  `json:"date" ch:"date"`
	UserID            uint64  `json:"user_id" ch:"user_id"`
	UserName          string  `json:"user_name" ch:"user_name"`
	
	// 使用统计
	TotalRequests     int64   `json:"total_requests" ch:"total_requests"`
	TotalTokens       int64   `json:"total_tokens" ch:"total_tokens"`
	AvgLatencyMs      float64 `json:"avg_latency_ms" ch:"avg_latency_ms"`
	MaxLatencyMs      int64   `json:"max_latency_ms" ch:"max_latency_ms"`
	
	// 时间分布
	OffHoursRequests  int64   `json:"off_hours_requests" ch:"off_hours_requests"`
	PeakHourRequests  int64   `json:"peak_hour_requests" ch:"peak_hour_requests"`
	
	// 风险统计
	HighRiskRequests  int64   `json:"high_risk_requests" ch:"high_risk_requests"`
	RiskEvents        int64   `json:"risk_events" ch:"risk_events"`
	
	// 模型使用
	ModelsUsed        string  `json:"models_used" ch:"models_used"` // 逗号分隔
}

// TableName ClickHouse表名
func (UserBehaviorSummary) TableName() string {
	return "user_behavior_summaries"
}

// PromptPattern 提示词模式（用于检测异常）
type PromptPattern struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Pattern       string    `json:"pattern" gorm:"size:500"`
	PatternType   string    `json:"pattern_type" gorm:"size:50"` // sensitive_info, injection, etc.
	RiskLevel     RiskLevel `json:"risk_level" gorm:"size:20"`
	Description   string    `json:"description" gorm:"size:500"`
	IsEnabled     bool      `json:"is_enabled" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AuditQueryRequest 审计查询请求
type AuditQueryRequest struct {
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	UserID      *uint64    `json:"user_id,omitempty"`
	ModelName   string     `json:"model_name,omitempty"`
	RiskLevel   RiskLevel  `json:"risk_level,omitempty"`
	RequestIP   string     `json:"request_ip,omitempty"`
	Page        int        `json:"page,omitempty"`
	PageSize    int        `json:"page_size,omitempty"`
}

// AuditQueryResponse 审计查询响应
type AuditQueryResponse struct {
	Data       []AuditLog `json:"data"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
}
