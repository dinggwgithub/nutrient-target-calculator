package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

// Redis配置
const (
	RedisAddr     = "127.0.0.1:6379"
	RedisPassword = ""
	RedisDB       = 0
	// Key前缀，用于区分不同业务
	KeyPrefixCalculation = "calc:result:"   // 计算结果存储前缀
	KeyPrefixHistory     = "calc:history:"  // 用户历史记录前缀
	KeyPrefixIndex       = "calc:index"     // 全局索引（按时间排序的所有计算记录）
	// TTL配置 - 1天过期
	CalculationTTL = 24 * time.Hour
)

// InitRedis 初始化Redis连接
func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     RedisAddr,
		Password: RedisPassword,
		DB:       RedisDB,
		// 连接池配置
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Println("Redis connection established successfully")
	return nil
}

// CloseRedis 关闭Redis连接
func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
	}
}
