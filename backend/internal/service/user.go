package service

import (
	"ai-gateway/internal/model"
	"ai-gateway/internal/repository"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		db: repository.GetDB(),
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string         `json:"username" binding:"required,min=3,max=50"`
	Email    string         `json:"email" binding:"required,email"`
	Password string         `json:"password" binding:"required,min=6"`
	Role     model.UserRole `json:"role" binding:"required"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string         `json:"username,omitempty"`
	Email    string         `json:"email,omitempty"`
	Role     model.UserRole `json:"role,omitempty"`
	Status   model.UserStatus `json:"status,omitempty"`
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req *CreateUserRequest) (*model.User, error) {
	// 检查用户名是否已存在
	var count int64
	if err := s.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("username already exists")
	}
	
	// 检查邮箱是否已存在
	if err := s.db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("email already exists")
	}
	
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	
	// 创建用户
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
		Status:   model.UserStatusActive,
	}
	
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	
	// 创建默认配额
	quota := &model.UserQuota{
		UserID:       user.ID,
		DailyLimit:   100000,
		WeeklyLimit:  500000,
		MonthlyLimit: 2000000,
	}
	if err := s.db.Create(quota).Error; err != nil {
		return nil, err
	}
	
	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(id uint, req *UpdateUserRequest) (*model.User, error) {
	user, err := s.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	
	updates := make(map[string]interface{})
	
	if req.Username != "" && req.Username != user.Username {
		// 检查用户名是否已存在
		var count int64
		s.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
		if count > 0 {
			return nil, errors.New("username already exists")
		}
		updates["username"] = req.Username
	}
	
	if req.Email != "" && req.Email != user.Email {
		// 检查邮箱是否已存在
		var count int64
		s.db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
		if count > 0 {
			return nil, errors.New("email already exists")
		}
		updates["email"] = req.Email
	}
	
	if req.Role != "" {
		updates["role"] = req.Role
	}
	
	if req.Status != "" {
		updates["status"] = req.Status
	}
	
	if len(updates) > 0 {
		if err := s.db.Model(user).Updates(updates).Error; err != nil {
			return nil, err
		}
		// 更新缓存
		repository.DeleteUserCache(user.ID, user.ApiKey)
	}
	
	return s.GetUserByID(id)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}
	
	// 删除缓存
	repository.DeleteUserCache(user.ID, user.ApiKey)
	
	return s.db.Delete(user).Error
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int, role string, status string) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	
	query := s.db.Model(&model.User{})
	
	if role != "" {
		query = query.Where("role = ?", role)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

// ResetPassword 重置密码
func (s *UserService) ResetPassword(id uint, newPassword string) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	if err := s.db.Model(user).Update("password", string(hashedPassword)).Error; err != nil {
		return err
	}
	
	// 删除缓存，强制重新登录
	repository.DeleteUserCache(user.ID, user.ApiKey)
	
	return nil
}

// RegenerateAPIKey 重新生成API Key
func (s *UserService) RegenerateAPIKey(id uint) (string, error) {
	user, err := s.GetUserByID(id)
	if err != nil {
		return "", err
	}
	
	// 删除旧缓存
	repository.DeleteUserCache(user.ID, user.ApiKey)
	
	// 生成新API Key
	newAPIKey := "ak-" + uuid.New().String()
	if err := s.db.Model(user).Update("api_key", newAPIKey).Error; err != nil {
		return "", err
	}
	
	return newAPIKey, nil
}

// UpdateUserQuota 更新用户配额
func (s *UserService) UpdateUserQuota(userID uint, daily, weekly, monthly int64) error {
	var quota model.UserQuota
	if err := s.db.Where("user_id = ?", userID).First(&quota).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新配额
			quota = model.UserQuota{
				UserID:       userID,
				DailyLimit:   daily,
				WeeklyLimit:  weekly,
				MonthlyLimit: monthly,
			}
			return s.db.Create(&quota).Error
		}
		return err
	}
	
	updates := map[string]interface{}{
		"daily_limit":   daily,
		"weekly_limit":  weekly,
		"monthly_limit": monthly,
	}
	
	if err := s.db.Model(&quota).Updates(updates).Error; err != nil {
		return err
	}
	
	// 更新缓存
	repository.CacheUserQuota(&quota, 5*time.Minute)
	
	return nil
}

// GetUserQuota 获取用户配额
func (s *UserService) GetUserQuota(userID uint) (*model.UserQuota, error) {
	var quota model.UserQuota
	if err := s.db.Where("user_id = ?", userID).First(&quota).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("quota not found")
		}
		return nil, err
	}
	return &quota, nil
}
