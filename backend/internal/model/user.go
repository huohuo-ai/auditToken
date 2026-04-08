package model

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole 用户角色
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
)

// UserStatus 用户状态
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBanned   UserStatus = "banned"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UUID      string         `json:"uuid" gorm:"uniqueIndex;size:36"`
	Username  string         `json:"username" gorm:"uniqueIndex;size:50"`
	Email     string         `json:"email" gorm:"uniqueIndex;size:100"`
	Password  string         `json:"-" gorm:"size:255"` // 不返回给前端
	Role      UserRole       `json:"role" gorm:"size:20;default:'user'"`
	Status    UserStatus     `json:"status" gorm:"size:20;default:'active'"`
	ApiKey    string         `json:"api_key" gorm:"uniqueIndex;size:64"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Quota     *UserQuota     `json:"quota,omitempty"`
	UsageLogs []UsageLog     `json:"-"`
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}
	if u.ApiKey == "" {
		u.ApiKey = "ak-" + uuid.New().String()
	}
	return nil
}

// IsAdmin 检查是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// UserQuota 用户配额模型
type UserQuota struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id" gorm:"uniqueIndex"`
	DailyLimit       int64     `json:"daily_limit" gorm:"default:0"`     // 0表示无限制
	WeeklyLimit      int64     `json:"weekly_limit" gorm:"default:0"`    // 0表示无限制
	MonthlyLimit     int64     `json:"monthly_limit" gorm:"default:0"`   // 0表示无限制
	DailyUsed        int64     `json:"daily_used" gorm:"default:0"`
	WeeklyUsed       int64     `json:"weekly_used" gorm:"default:0"`
	MonthlyUsed      int64     `json:"monthly_used" gorm:"default:0"`
	LastResetDaily   time.Time `json:"last_reset_daily"`
	LastResetWeekly  time.Time `json:"last_reset_weekly"`
	LastResetMonthly time.Time `json:"last_reset_monthly"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName 指定表名
func (UserQuota) TableName() string {
	return "user_quotas"
}

// CheckQuota 检查配额是否超出
func (q *UserQuota) CheckQuota(requestTokens int64) (bool, string) {
	// 检查日限制
	if q.DailyLimit > 0 && q.DailyUsed+requestTokens > q.DailyLimit {
		return false, "daily quota exceeded"
	}
	// 检查周限制
	if q.WeeklyLimit > 0 && q.WeeklyUsed+requestTokens > q.WeeklyLimit {
		return false, "weekly quota exceeded"
	}
	// 检查月限制
	if q.MonthlyLimit > 0 && q.MonthlyUsed+requestTokens > q.MonthlyLimit {
		return false, "monthly quota exceeded"
	}
	return true, ""
}

// UsageLog 使用日志（用于MySQL中的实时统计）
type UsageLog struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          uint      `json:"user_id" gorm:"index"`
	RequestID       string    `json:"request_id" gorm:"size:36;index"`
	ModelName       string    `json:"model_name" gorm:"size:50;index"`
	PromptTokens    int64     `json:"prompt_tokens"`
	CompletionTokens int64    `json:"completion_tokens"`
	TotalTokens     int64     `json:"total_tokens"`
	RequestTime     time.Time `json:"request_time"`
	ResponseTime    time.Time `json:"response_time"`
	LatencyMs       int64     `json:"latency_ms"`
	Status          string    `json:"status" gorm:"size:20"` // success, failed
	ErrorMessage    string    `json:"error_message" gorm:"size:500"`
	IP              string    `json:"ip" gorm:"size:50"`
	UserAgent       string    `json:"user_agent" gorm:"size:255"`
	CreatedAt       time.Time `json:"created_at"`
}

// UsageStatistics 使用统计
type UsageStatistics struct {
	Date           string `json:"date"`
	UserID         uint   `json:"user_id"`
	ModelName      string `json:"model_name"`
	TotalRequests  int64  `json:"total_requests"`
	TotalTokens    int64  `json:"total_tokens"`
	PromptTokens   int64  `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
}
