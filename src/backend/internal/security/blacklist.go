// Package security 提供基于 Redis 的安全相关组件：限流、登录锁定与 Token 黑名单。
package security

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklist Token 黑名单接口
type TokenBlacklist interface {
	// Add 将 JTI 加入黑名单，ttl 为令牌剩余有效期
	Add(ctx context.Context, jti string, ttl time.Duration) error
	// IsBlacklisted 检查 JTI 是否在黑名单中
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}

// RedisTokenBlacklist 基于 Redis 的 Token 黑名单实现
type RedisTokenBlacklist struct {
	client    *redis.Client
	keyPrefix string
}

// NewRedisTokenBlacklist 创建 Redis Token 黑名单
func NewRedisTokenBlacklist(client *redis.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{
		client:    client,
		keyPrefix: "goai:blacklist:jti",
	}
}

// Add 将 JTI 加入黑名单，使用 SETEX 让 Redis 自动清理过期条目
func (b *RedisTokenBlacklist) Add(ctx context.Context, jti string, ttl time.Duration) error {
	if jti == "" {
		return fmt.Errorf("jti is empty")
	}
	key := fmt.Sprintf("%s:%s", b.keyPrefix, jti)
	return b.client.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted 检查 JTI 是否在黑名单中
func (b *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	key := fmt.Sprintf("%s:%s", b.keyPrefix, jti)
	n, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}
	return n > 0, nil
}
