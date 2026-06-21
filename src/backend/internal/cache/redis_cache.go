package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache 是基于 Redis 的缓存实现。
type RedisCache struct {
	client *redis.Client
	cfg    Config
}

// NewRedisCache 创建 Redis 缓存实例。
func NewRedisCache(client *redis.Client, cfg Config) *RedisCache {
	if cfg.DefaultTTL <= 0 {
		cfg.DefaultTTL = DefaultConfig().DefaultTTL
	}
	return &RedisCache{client: client, cfg: cfg}
}

// Get 从 Redis 获取值并反序列化。
func (c *RedisCache) Get(ctx context.Context, key string, dest any) error {
	data, err := c.client.Get(ctx, c.prefixed(key)).Bytes()
	if err == redis.Nil {
		return ErrCacheMiss
	}
	if err != nil {
		return fmt.Errorf("cache get: %w", err)
	}
	if err := deserialize(data, dest); err != nil {
		return fmt.Errorf("cache deserialize: %w", err)
	}
	return nil
}

// Set 将值序列化后写入 Redis。
func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := serialize(value)
	if err != nil {
		return fmt.Errorf("cache serialize: %w", err)
	}
	if ttl <= 0 {
		ttl = c.cfg.DefaultTTL
	}
	if err := c.client.Set(ctx, c.prefixed(key), data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}
	return nil
}

// Delete 删除指定 key。
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, c.prefixed(key)).Err(); err != nil {
		return fmt.Errorf("cache delete: %w", err)
	}
	return nil
}

// DeletePattern 按模式删除匹配的 key（使用 SCAN 避免阻塞）。
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := c.prefixed(pattern)
	iter := c.client.Scan(ctx, 0, fullPattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("cache delete pattern: %w", err)
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("cache scan: %w", err)
	}
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("cache delete pattern: %w", err)
		}
	}
	return nil
}

func (c *RedisCache) prefixed(key string) string {
	return WithPrefix(c.cfg.Prefix, key)
}
