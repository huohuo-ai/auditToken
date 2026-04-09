package repository

import (
	"fmt"
	"ai-gateway/internal/config"
	"ai-gateway/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接
var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port, cfg.SSLMode)
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	return db, nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.UserQuota{},
		&model.UsageLog{},
		&model.AIModel{},
		&model.ModelAccess{},
		&model.PromptPattern{},
	)
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	if DB == nil {
		panic("database not initialized")
	}
	return DB
}

// CreateDefaultAdmin 创建默认管理员账号
func CreateDefaultAdmin(db *gorm.DB) error {
	var count int64
	db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&count)
	if count > 0 {
		logrus.Info("Admin user already exists, skipping creation")
		return nil
	}

	logrus.Info("Creating default admin user...")

	// 默认管理员账号：admin / admin123
	admin := &model.User{
		Username: "admin",
		Email:    "admin@company.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // bcrypt hash of "admin123"
		Role:     model.RoleAdmin,
		Status:   model.UserStatusActive,
	}

	if err := db.Create(admin).Error; err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	logrus.Info("Default admin user created successfully")

	// 创建默认配额（无限制）
	quota := &model.UserQuota{
		UserID:       admin.ID,
		DailyLimit:   0,
		WeeklyLimit:  0,
		MonthlyLimit: 0,
	}

	if err := db.Create(quota).Error; err != nil {
		return fmt.Errorf("failed to create default quota: %w", err)
	}

	return nil
}
