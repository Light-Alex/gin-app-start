package database

import (
	"context"
	"fmt"
	"time"

	"gin-app-start/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr         string // Redis地址，格式为"host:port"
	Password     string // Redis密码
	DB           int    // Redis数据库索引
	PoolSize     int    // 连接池大小
	MinIdleConns int    // 最小空闲连接数
	MaxRetries   int    // 最大重试次数
}

func NewRedisClient(config *RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 需5s内连接成功，否则报错
	_, err := client.Ping(timeoutCtx).Result()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to redis: %w", err)
	}

	logger.Info("connected to redis successfully")

	return client, nil
}
