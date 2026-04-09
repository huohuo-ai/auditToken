package main

import (
	"ai-gateway/internal/config"
	"ai-gateway/internal/handler"
	"ai-gateway/internal/repository"

	"github.com/sirupsen/logrus"
)

func init() {
	// 设置日志级别为 Info，否则 Info 级别的日志不会输出
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logrus.WithError(err).Warn("Failed to load config file, using default values and continuing")
	}

	// 初始化数据库
	db, err := repository.InitDatabase(&cfg.Database)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize database")
	}

	// 自动迁移
	if err := repository.AutoMigrate(db); err != nil {
		logrus.WithError(err).Fatal("Failed to auto migrate database")
	}

	// 创建默认管理员
	if err := repository.CreateDefaultAdmin(db); err != nil {
		logrus.WithError(err).Warn("Failed to create default admin")
	}

	// 初始化Redis
	if _, err := repository.InitRedis(&cfg.Redis); err != nil {
		logrus.WithError(err).Fatal("Failed to initialize redis")
	}

	// 初始化ClickHouse
	if _, err := repository.InitClickHouse(&cfg.ClickHouse); err != nil {
		logrus.WithError(err).Fatal("Failed to initialize clickhouse")
	}

	// 创建ClickHouse表
	if err := repository.CreateClickHouseTables(); err != nil {
		logrus.WithError(err).Fatal("Failed to create clickhouse tables")
	}

	// 设置路由
	r := handler.SetupRouter()

	// 启动服务器
	addr := ":" + cfg.Server.Port
	logrus.Infof("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		logrus.WithError(err).Fatal("Failed to start server")
	}
}
