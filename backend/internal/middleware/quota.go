package middleware

import (
	"ai-gateway/internal/model"
	"ai-gateway/internal/repository"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// QuotaCheck 配额检查中间件
func QuotaCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}
		
		uid := userID.(uint)
		
		// 获取用户配额
		quota, err := getUserQuota(uid)
		if err != nil {
			logrus.WithError(err).Error("failed to get user quota")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check quota"})
			c.Abort()
			return
		}
		
		// 重置配额（如果需要）
		if err := resetQuotaIfNeeded(quota); err != nil {
			logrus.WithError(err).Error("failed to reset quota")
		}
		
		// 检查配额是否超限
		// 预估请求token数（可以根据实际需求调整）
		estimatedTokens := int64(4000) // 默认预估
		
		if allowed, reason := quota.CheckQuota(estimatedTokens); !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":  "quota exceeded",
				"reason": reason,
			})
			c.Abort()
			return
		}
		
		// 将配额信息存入上下文
		c.Set("userQuota", quota)
		
		c.Next()
	}
}

// getUserQuota 获取用户配额
func getUserQuota(userID uint) (*model.UserQuota, error) {
	// 先从缓存获取
	quota, err := repository.GetCachedUserQuota(userID)
	if err == nil {
		return quota, nil
	}
	
	// 从数据库获取
	db := repository.GetDB()
	if err := db.Where("user_id = ?", userID).First(&quota).Error; err != nil {
		// 如果没有配额记录，创建一个默认的
		quota = &model.UserQuota{
			UserID:       userID,
			DailyLimit:   100000,  // 默认10万/天
			WeeklyLimit:  500000,  // 默认50万/周
			MonthlyLimit: 2000000, // 默认200万/月
		}
		if err := db.Create(quota).Error; err != nil {
			return nil, err
		}
	}
	
	// 缓存配额信息
	repository.CacheUserQuota(quota, 5*time.Minute)
	
	return quota, nil
}

// resetQuotaIfNeeded 根据需要重置配额
func resetQuotaIfNeeded(quota *model.UserQuota) error {
	now := time.Now()
	needUpdate := false
	
	// 检查日配额是否需要重置
	if now.Sub(quota.LastResetDaily) >= 24*time.Hour {
		quota.DailyUsed = 0
		quota.LastResetDaily = now
		needUpdate = true
		
		// 更新Redis缓存
		repository.SetTokenUsageWithExpire(quota.UserID, 0, "daily", 24*time.Hour)
	}
	
	// 检查周配额是否需要重置
	if now.Sub(quota.LastResetWeekly) >= 7*24*time.Hour {
		quota.WeeklyUsed = 0
		quota.LastResetWeekly = now
		needUpdate = true
		
		repository.SetTokenUsageWithExpire(quota.UserID, 0, "weekly", 7*24*time.Hour)
	}
	
	// 检查月配额是否需要重置
	if now.Day() == 1 && now.Sub(quota.LastResetMonthly) >= 24*time.Hour {
		quota.MonthlyUsed = 0
		quota.LastResetMonthly = now
		needUpdate = true
		
		repository.SetTokenUsageWithExpire(quota.UserID, 0, "monthly", 30*24*time.Hour)
	}
	
	if needUpdate {
		db := repository.GetDB()
		if err := db.Save(quota).Error; err != nil {
			return err
		}
		// 更新缓存
		repository.CacheUserQuota(quota, 5*time.Minute)
	}
	
	return nil
}

// UpdateQuotaUsage 更新配额使用量
func UpdateQuotaUsage(userID uint, tokens int64) error {
	// 更新数据库
	db := repository.GetDB()
	if err := db.Model(&model.UserQuota{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"daily_used":   gorm.Expr("daily_used + ?", tokens),
		"weekly_used":  gorm.Expr("weekly_used + ?", tokens),
		"monthly_used": gorm.Expr("monthly_used + ?", tokens),
	}).Error; err != nil {
		return err
	}
	
	// 更新Redis
	repository.IncrementTokenUsage(userID, tokens, "daily")
	repository.IncrementTokenUsage(userID, tokens, "weekly")
	repository.IncrementTokenUsage(userID, tokens, "monthly")
	
	// 删除缓存让下次重新加载
	repository.GetRedis().Del(context.Background(), fmt.Sprintf("quota:user:%d", userID))
	
	return nil
}

// RateLimit 速率限制中间件
func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}
		
		key := fmt.Sprintf("ratelimit:user:%d", userID.(uint))
		allowed, err := repository.CheckRateLimit(key, requestsPerMinute, time.Minute)
		if err != nil {
			logrus.WithError(err).Error("rate limit check failed")
			c.Next()
			return
		}
		
		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}
