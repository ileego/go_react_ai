// Package cache 提供应用级缓存抽象。
// 目前主要用于报告详情、列表等读多写少场景的 Cache-Aside 缓存。
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrCacheMiss 表示缓存未命中。
var ErrCacheMiss = errors.New("cache miss")

// Manager 定义缓存管理接口。
type Manager interface {
	// Get 从缓存中获取 key 对应的值，并反序列化到 dest。
	// 未命中时返回 ErrCacheMiss。
	Get(ctx context.Context, key string, dest any) error
	// Set 将 value 序列化后写入缓存，ttl 为过期时间。
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	// Delete 删除指定 key。
	Delete(ctx context.Context, key string) error
	// DeletePattern 按模式删除缓存（具体语义由实现决定，Redis 实现使用 SCAN + DEL）。
	DeletePattern(ctx context.Context, pattern string) error
}

// ReportKey 返回报告详情缓存 key。
func ReportKey(id int64) string {
	return fmt.Sprintf("report:%d", id)
}

// ReportListKey 返回报告列表缓存 key。
func ReportListKey(userID int64, page, pageSize int) string {
	return fmt.Sprintf("reports:user:%d:page:%d:size:%d", userID, page, pageSize)
}

// serialize 将对象序列化为 JSON 字节。
func serialize(v any) ([]byte, error) {
	return json.Marshal(v)
}

// deserialize 将 JSON 字节反序列化到 dest。
func deserialize(data []byte, dest any) error {
	return json.Unmarshal(data, dest)
}

// Config 缓存配置。
type Config struct {
	// DefaultTTL 默认缓存过期时间。
	DefaultTTL time.Duration
	// Prefix 键前缀，用于多环境隔离。
	Prefix string
}

// DefaultConfig 返回默认缓存配置。
func DefaultConfig() Config {
	return Config{
		DefaultTTL: 5 * time.Minute,
		Prefix:     "goai",
	}
}

// WithPrefix 返回带前缀的 key。
func WithPrefix(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + ":" + key
}
