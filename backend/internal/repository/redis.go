package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-gateway/internal/config"
	"ai-gateway/internal/model"

	"github.com/redis/go-redis/v9"
)

// RedisClient 全局Redis客户端
var RedisClient *redis.Client
var ctx = context.Background()

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	RedisClient = client
	return client, nil
}

// GetRedis 获取Redis客户端
func GetRedis() *redis.Client {
	if RedisClient == nil {
		panic("redis not initialized")
	}
	return RedisClient
}

// ===== 用户相关缓存 =====

// CacheUser 缓存用户信息
func CacheUser(user *model.User, expiration time.Duration) error {
	key := fmt.Sprintf("user:id:%d", user.ID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, data, expiration).Err()
}

// GetCachedUser 获取缓存的用户信息
func GetCachedUser(userID uint) (*model.User, error) {
	key := fmt.Sprintf("user:id:%d", userID)
	data, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// CacheUserByAPIKey 通过API Key缓存用户
func CacheUserByAPIKey(apiKey string, user *model.User, expiration time.Duration) error {
	key := fmt.Sprintf("user:apikey:%s", apiKey)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, data, expiration).Err()
}

// GetCachedUserByAPIKey 通过API Key获取缓存用户
func GetCachedUserByAPIKey(apiKey string) (*model.User, error) {
	key := fmt.Sprintf("user:apikey:%s", apiKey)
	data, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUserCache 删除用户缓存
func DeleteUserCache(userID uint, apiKey string) error {
	pipe := RedisClient.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("user:id:%d", userID))
	pipe.Del(ctx, fmt.Sprintf("user:apikey:%s", apiKey))
	_, err := pipe.Exec(ctx)
	return err
}

// ===== Token配额相关缓存 =====

// CacheUserQuota 缓存用户配额
func CacheUserQuota(quota *model.UserQuota, expiration time.Duration) error {
	key := fmt.Sprintf("quota:user:%d", quota.UserID)
	data, err := json.Marshal(quota)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, data, expiration).Err()
}

// GetCachedUserQuota 获取缓存的用户配额
func GetCachedUserQuota(userID uint) (*model.UserQuota, error) {
	key := fmt.Sprintf("quota:user:%d", userID)
	data, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var quota model.UserQuota
	if err := json.Unmarshal([]byte(data), &quota); err != nil {
		return nil, err
	}
	return &quota, nil
}

// IncrementTokenUsage 增加Token使用量
func IncrementTokenUsage(userID uint, tokens int64, period string) error {
	key := fmt.Sprintf("quota:%s:user:%d", period, userID)
	return RedisClient.IncrBy(ctx, key, tokens).Err()
}

// GetTokenUsage 获取Token使用量
func GetTokenUsage(userID uint, period string) (int64, error) {
	key := fmt.Sprintf("quota:%s:user:%d", period, userID)
	val, err := RedisClient.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// SetTokenUsageWithExpire 设置Token使用量并设置过期时间
func SetTokenUsageWithExpire(userID uint, tokens int64, period string, expiration time.Duration) error {
	key := fmt.Sprintf("quota:%s:user:%d", period, userID)
	return RedisClient.Set(ctx, key, tokens, expiration).Err()
}

// ===== 速率限制相关 =====

// CheckRateLimit 检查速率限制
func CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	// 使用Redis的滑动窗口限流
	pipe := RedisClient.Pipeline()
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// 移除窗口外的请求记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	// 获取当前窗口内的请求数
	pipe.ZCard(ctx, key)
	// 添加当前请求
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	// 设置key的过期时间
	pipe.Expire(ctx, key, window)

	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// results[1] 是 ZCard 的结果
	count := results[1].(*redis.IntCmd).Val()
	return count <= int64(limit), nil
}

// ===== 模型缓存 =====

// CacheModel 缓存模型信息
func CacheModel(model *model.AIModel, expiration time.Duration) error {
	key := fmt.Sprintf("model:id:%d", model.ID)
	data, err := json.Marshal(model)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, data, expiration).Err()
}

// GetCachedModel 获取缓存的模型信息
func GetCachedModel(modelID uint) (*model.AIModel, error) {
	key := fmt.Sprintf("model:id:%d", modelID)
	data, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var m model.AIModel
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// CacheDefaultModel 缓存默认模型
func CacheDefaultModel(model *model.AIModel, expiration time.Duration) error {
	key := "model:default"
	data, err := json.Marshal(model)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, data, expiration).Err()
}

// GetCachedDefaultModel 获取缓存的默认模型
func GetCachedDefaultModel() (*model.AIModel, error) {
	key := "model:default"
	data, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var m model.AIModel
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, err
	}
	return &m, nil
}
