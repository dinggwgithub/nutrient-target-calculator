package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}

	log.Println("Redis connection established successfully")
	return nil
}
