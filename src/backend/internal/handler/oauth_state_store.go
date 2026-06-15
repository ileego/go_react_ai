package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOAuthStateStore 基于 Redis 的 OAuth state 存储实现
type RedisOAuthStateStore struct {
	client    *redis.Client
	keyPrefix string
}

// NewRedisOAuthStateStore 创建 Redis OAuth state 存储
func NewRedisOAuthStateStore(client *redis.Client) *RedisOAuthStateStore {
	return &RedisOAuthStateStore{
		client:    client,
		keyPrefix: "goai:oauth:state",
	}
}

// Save 保存 state
func (s *RedisOAuthStateStore) Save(ctx context.Context, state string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", s.keyPrefix, state)
	return s.client.Set(ctx, key, "1", ttl).Err()
}

// Verify 校验 state，校验成功后删除
func (s *RedisOAuthStateStore) Verify(ctx context.Context, state string) (bool, error) {
	key := fmt.Sprintf("%s:%s", s.keyPrefix, state)
	n, err := s.client.Del(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
