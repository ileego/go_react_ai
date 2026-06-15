// Package redis 提供 Redis 客户端初始化与连接管理。
package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ileego/go_react_ai/internal/config"
)

// Client 封装 go-redis 客户端
type Client struct {
	*redis.Client
}

// New 使用配置创建 Redis 连接
func New(cfg config.RedisConfig) *Client {
	addr := cfg.Host + ":" + cfg.Port
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // 开发环境无密码
		DB:       0,
	})

	return &Client{Client: rdb}
}

// HealthCheck 检查 Redis 连接是否正常
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.Client.Ping(ctx).Err()
}

// Close 关闭 Redis 连接
func (c *Client) Close() error {
	return c.Client.Close()
}
