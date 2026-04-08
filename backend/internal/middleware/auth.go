package middleware

import (
	"ai-gateway/internal/config"
	"ai-gateway/internal/model"
	"ai-gateway/internal/repository"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint            `json:"user_id"`
	Username string          `json:"username"`
	Email    string          `json:"email"`
	Role     model.UserRole  `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(user *model.User) (string, error) {
	cfg := config.GetConfig().JWT
	
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.ExpiresIn) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ai-gateway",
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}
		
		// 提取token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}
		
		tokenString := parts[1]
		cfg := config.GetConfig().JWT
		
		// 解析token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})
		
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		
		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}
		
		// 设置用户信息到上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("userRole", claims.Role)
		
		c.Next()
	}
}

// APIKeyAuth API Key认证中间件
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// 尝试从Authorization头获取
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "api key is required"})
			c.Abort()
			return
		}
		
		// 先从缓存获取用户
		user, err := repository.GetCachedUserByAPIKey(apiKey)
		if err != nil {
			// 从数据库获取
			db := repository.GetDB()
			if err := db.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
				c.Abort()
				return
			}
			// 缓存用户信息
			repository.CacheUserByAPIKey(apiKey, user, 5*time.Minute)
		}
		
		// 检查用户状态
		if user.Status != model.UserStatusActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "user is not active"})
			c.Abort()
			return
		}
		
		// 设置用户信息到上下文
		c.Set("userID", user.ID)
		c.Set("username", user.Username)
		c.Set("userRole", user.Role)
		c.Set("apiKey", apiKey)
		
		c.Next()
	}
}

// AdminRequired 管理员权限中间件
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}
		
		if role != model.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin permission required"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// GetCurrentUser 获取当前登录用户
func GetCurrentUser(c *gin.Context) (*model.User, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return nil, nil
	}
	
	// 先从缓存获取
	user, err := repository.GetCachedUser(userID.(uint))
	if err != nil {
		// 从数据库获取
		db := repository.GetDB()
		if err := db.First(&user, userID).Error; err != nil {
			return nil, err
		}
		// 缓存用户信息
		repository.CacheUser(user, 5*time.Minute)
	}
	
	return user, nil
}
