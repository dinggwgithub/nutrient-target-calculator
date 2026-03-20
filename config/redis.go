package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

// Redis TTL 常量
const (
	// CalculationHistoryTTL 计算历史记录过期时间：1天
	CalculationHistoryTTL = 24 * time.Hour
)

// Redis Key 前缀
const (
	// KeyPrefixCalculation 计算结果前缀
	KeyPrefixCalculation = "calc:"
	// KeyPrefixUserHistory 用户历史记录索引前缀
	KeyPrefixUserHistory = "history:"
)

// InitRedis 初始化 Redis 连接
func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // 无密码
		DB:       0,  // 默认数据库
		PoolSize: 10, // 连接池大小
	})

	// 测试连接
	if err := RedisClient.Ping(Ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %v", err)
	}

	log.Println("Redis connection established successfully")
	return nil
}

// GenerateCalculationKey 生成计算结果存储的 Key
// 格式: calc:{user_id}:{timestamp}:{uuid}
func GenerateCalculationKey(userID string, timestamp int64, uuid string) string {
	return fmt.Sprintf("%s%s:%d:%s", KeyPrefixCalculation, userID, timestamp, uuid)
}

// GenerateUserHistoryKey 生成用户历史记录索引的 Key
// 格式: history:{user_id}
func GenerateUserHistoryKey(userID string) string {
	return fmt.Sprintf("%s%s", KeyPrefixUserHistory, userID)
}

// SaveCalculationResult 保存计算结果到 Redis
func SaveCalculationResult(key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal calculation result: %v", err)
	}

	if err := RedisClient.Set(Ctx, key, jsonData, CalculationHistoryTTL).Err(); err != nil {
		return fmt.Errorf("failed to save to redis: %v", err)
	}

	return nil
}

// GetCalculationResult 从 Redis 获取计算结果
func GetCalculationResult(key string) (string, error) {
	result, err := RedisClient.Get(Ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("calculation result not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get from redis: %v", err)
	}

	return result, nil
}

// AddToUserHistory 添加计算记录到用户历史索引（Sorted Set，按时间排序）
func AddToUserHistory(userID string, calculationKey string, timestamp int64) error {
	historyKey := GenerateUserHistoryKey(userID)

	// 使用 ZAdd 添加到有序集合，score 为时间戳
	if err := RedisClient.ZAdd(Ctx, historyKey, redis.Z{
		Score:  float64(timestamp),
		Member: calculationKey,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add to user history: %v", err)
	}

	// 设置历史索引的过期时间
	RedisClient.Expire(Ctx, historyKey, CalculationHistoryTTL)

	return nil
}

// GetUserHistory 获取用户历史记录列表
// limit: 返回数量限制，0 表示不限制
func GetUserHistory(userID string, limit int64) ([]string, error) {
	historyKey := GenerateUserHistoryKey(userID)

	var results []string
	var err error

	if limit > 0 {
		results, err = RedisClient.ZRevRange(Ctx, historyKey, 0, limit-1).Result()
	} else {
		results, err = RedisClient.ZRevRange(Ctx, historyKey, 0, -1).Result()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user history: %v", err)
	}

	return results, nil
}

// DeleteCalculationResult 删除计算结果
func DeleteCalculationResult(key string) error {
	if err := RedisClient.Del(Ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from redis: %v", err)
	}
	return nil
}

// GetCalculationResultWithTTL 获取计算结果及其剩余 TTL
func GetCalculationResultWithTTL(key string) (string, time.Duration, error) {
	pipe := RedisClient.Pipeline()
	getCmd := pipe.Get(Ctx, key)
	ttlCmd := pipe.TTL(Ctx, key)

	_, err := pipe.Exec(Ctx)
	if err != nil {
		if err == redis.Nil {
			return "", 0, fmt.Errorf("calculation result not found")
		}
		return "", 0, fmt.Errorf("failed to get from redis: %v", err)
	}

	result, err := getCmd.Result()
	if err != nil {
		return "", 0, err
	}

	ttl, _ := ttlCmd.Result()
	return result, ttl, nil
}
