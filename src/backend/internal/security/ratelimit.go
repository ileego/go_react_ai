// Package security 提供基于 Redis 的安全相关组件：限流、登录锁定与 Token 黑名单。
package security

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// IPLimitConfig IP 限流配置
type IPLimitConfig struct {
	Limit  int
	Window time.Duration
}

// LoginLimitConfig 登录限流与锁定配置
type LoginLimitConfig struct {
	MaxAttempts int           // 窗口内允许的最大失败次数
	Window      time.Duration // 失败计数窗口
	Lockout     time.Duration // 超过最大次数后的锁定时长
}

// RateLimiter 基于 Redis 的限流器接口
type RateLimiter interface {
	// AllowIP 根据 key（通常是客户端 IP）判断是否允许通过
	AllowIP(ctx context.Context, key string) (bool, error)
	// AllowLogin 判断登录是否被锁定；返回是否允许、剩余锁定时间
	AllowLogin(ctx context.Context, key string) (bool, time.Duration, error)
	// RecordLoginFailure 记录一次登录失败，返回当前失败次数与剩余锁定时间
	RecordLoginFailure(ctx context.Context, key string) (int, time.Duration, error)
	// ResetLoginFailures 登录成功后重置失败计数
	ResetLoginFailures(ctx context.Context, key string) error
}

// RedisRateLimiter 基于 Redis 的限流器实现
type RedisRateLimiter struct {
	client    *redis.Client
	ipCfg     IPLimitConfig
	loginCfg  LoginLimitConfig
	keyPrefix string
}

// NewRedisRateLimiter 创建 Redis 限流器
func NewRedisRateLimiter(client *redis.Client, ipCfg IPLimitConfig, loginCfg LoginLimitConfig) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:    client,
		ipCfg:     ipCfg,
		loginCfg:  loginCfg,
		keyPrefix: "goai:security",
	}
}

// AllowIP 使用固定窗口计数实现 IP 限流
func (r *RedisRateLimiter) AllowIP(ctx context.Context, key string) (bool, error) {
	if r.ipCfg.Limit <= 0 {
		return true, nil
	}
	redisKey := fmt.Sprintf("%s:ip:%s", r.keyPrefix, key)
	return r.allow(ctx, redisKey, r.ipCfg.Limit, r.ipCfg.Window)
}

// AllowLogin 检查登录是否被锁定
func (r *RedisRateLimiter) AllowLogin(ctx context.Context, key string) (bool, time.Duration, error) {
	lockKey := fmt.Sprintf("%s:login:lock:%s", r.keyPrefix, key)
	lockTTL, err := r.client.TTL(ctx, lockKey).Result()
	if err != nil {
		return false, 0, fmt.Errorf("check login lock: %w", err)
	}
	if lockTTL > 0 {
		return false, lockTTL, nil
	}
	return true, 0, nil
}

// RecordLoginFailure 记录登录失败并触发锁定
func (r *RedisRateLimiter) RecordLoginFailure(ctx context.Context, key string) (int, time.Duration, error) {
	counterKey := fmt.Sprintf("%s:login:fail:%s", r.keyPrefix, key)
	lockKey := fmt.Sprintf("%s:login:lock:%s", r.keyPrefix, key)

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, counterKey)
	pipe.Expire(ctx, counterKey, r.loginCfg.Window)
	getLock := pipe.TTL(ctx, lockKey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("record login failure: %w", err)
	}

	count := int(incr.Val())
	lockTTL := getLock.Val()

	// 达到阈值后设置锁定键
	if count >= r.loginCfg.MaxAttempts && lockTTL <= 0 {
		err = r.client.Set(ctx, lockKey, "1", r.loginCfg.Lockout).Err()
		if err != nil {
			return count, 0, fmt.Errorf("set login lock: %w", err)
		}
		lockTTL = r.loginCfg.Lockout
	}

	if lockTTL > 0 {
		return count, lockTTL, nil
	}
	return count, 0, nil
}

// ResetLoginFailures 重置登录失败计数与锁定
func (r *RedisRateLimiter) ResetLoginFailures(ctx context.Context, key string) error {
	counterKey := fmt.Sprintf("%s:login:fail:%s", r.keyPrefix, key)
	lockKey := fmt.Sprintf("%s:login:lock:%s", r.keyPrefix, key)
	return r.client.Del(ctx, counterKey, lockKey).Err()
}

// allow 内部固定窗口限流实现
func (r *RedisRateLimiter) allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("rate limit increment: %w", err)
	}

	count := int(incr.Val())

	return count <= limit, nil
}
