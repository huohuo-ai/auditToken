package handler

import (
	"ai-gateway/internal/middleware"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 全局中间件
	r.Use(middleware.AuditLog())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API v1
	apiV1 := r.Group("/api/v1")
	{
		// 认证相关（公开）
		authHandler := NewAuthHandler()
		apiV1.POST("/auth/login", authHandler.Login)
		apiV1.POST("/auth/register", authHandler.Register)

		// 需要认证的路由
		authorized := apiV1.Group("")
		authorized.Use(middleware.JWTAuth())
		{
			// 用户相关
			authorized.GET("/auth/profile", authHandler.GetProfile)
			authorized.POST("/auth/change-password", authHandler.ChangePassword)
			authorized.POST("/auth/regenerate-apikey", authHandler.RegenerateAPIKey)

			// LLM API（兼容OpenAI格式，使用API Key认证）
			llmHandler := NewLLMHandler()
			authorized.GET("/models", llmHandler.ListModels)
			authorized.POST("/chat/completions", llmHandler.ChatCompletion)
		}

		// 管理员路由
		admin := apiV1.Group("/admin")
		admin.Use(middleware.JWTAuth())
		admin.Use(middleware.AdminRequired())
		{
			// 用户管理
			userHandler := NewUserHandler()
			admin.GET("/users", userHandler.ListUsers)
			admin.POST("/users", userHandler.CreateUser)
			admin.GET("/users/:id", userHandler.GetUser)
			admin.PUT("/users/:id", userHandler.UpdateUser)
			admin.DELETE("/users/:id", userHandler.DeleteUser)
			admin.POST("/users/:id/reset-password", userHandler.ResetPassword)
			admin.GET("/users/:id/quota", userHandler.GetUserQuota)
			admin.PUT("/users/:id/quota", userHandler.UpdateUserQuota)

			// 模型管理
			modelHandler := NewAIModelHandler()
			admin.GET("/models", modelHandler.ListModels)
			admin.POST("/models", modelHandler.CreateModel)
			admin.GET("/models/:id", modelHandler.GetModel)
			admin.PUT("/models/:id", modelHandler.UpdateModel)
			admin.DELETE("/models/:id", modelHandler.DeleteModel)

			// 审计管理
			auditHandler := NewAuditHandler()
			admin.GET("/audit/logs", auditHandler.QueryAuditLogs)
			admin.GET("/audit/risk-events", auditHandler.GetRiskEvents)
			admin.POST("/audit/risk-events/:event_id/resolve", auditHandler.ResolveRiskEvent)
			admin.GET("/audit/users/:user_id/statistics", auditHandler.GetUserStatistics)
			admin.GET("/audit/dashboard", auditHandler.GetDashboardStats)
		}
	}

	// LLM API 网关路由（API Key认证，兼容OpenAI API格式）
	llmHandler := NewLLMHandler()
	v1 := r.Group("/v1")
	v1.Use(middleware.APIKeyAuth())
	v1.Use(middleware.QuotaCheck())
	v1.Use(middleware.RateLimit(60))
	{
		v1.GET("/models", llmHandler.ListModels)
		v1.POST("/chat/completions", llmHandler.ChatCompletion)
	}

	return r
}
