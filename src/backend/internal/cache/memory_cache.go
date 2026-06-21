package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryCache 是内存缓存实现，主要用于单元测试和本地开发。
type MemoryCache struct {
	mu      sync.RWMutex
	data    map[string][]byte
	expires map[string]time.Time
	cfg     Config
}

// NewMemoryCache 创建内存缓存实例。
func NewMemoryCache(cfg Config) *MemoryCache {
	if cfg.DefaultTTL <= 0 {
		cfg.DefaultTTL = DefaultConfig().DefaultTTL
	}
	return &MemoryCache{
		data:    make(map[string][]byte),
		expires: make(map[string]time.Time),
		cfg:     cfg,
	}
}

// Get 从内存中获取值。
func (c *MemoryCache) Get(ctx context.Context, key string, dest any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fullKey := c.prefixed(key)
	exp, ok := c.expires[fullKey]
	if !ok {
		return ErrCacheMiss
	}
	if time.Now().After(exp) {
		return ErrCacheMiss
	}
	data, ok := c.data[fullKey]
	if !ok {
		return ErrCacheMiss
	}
	if err := deserialize(data, dest); err != nil {
		return fmt.Errorf("cache deserialize: %w", err)
	}
	return nil
}

// Set 写入内存缓存。
func (c *MemoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := serialize(value)
	if err != nil {
		return fmt.Errorf("cache serialize: %w", err)
	}
	if ttl <= 0 {
		ttl = c.cfg.DefaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	fullKey := c.prefixed(key)
	c.data[fullKey] = data
	c.expires[fullKey] = time.Now().Add(ttl)
	return nil
}

// Delete 删除指定 key。
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullKey := c.prefixed(key)
	delete(c.data, fullKey)
	delete(c.expires, fullKey)
	return nil
}

// DeletePattern 按模式删除。内存实现支持以 * 结尾的前缀匹配。
func (c *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullPattern := c.prefixed(pattern)
	prefix := strings.TrimSuffix(fullPattern, "*")

	for k := range c.data {
		if matchMemoryPattern(fullPattern, k) {
			delete(c.data, k)
			delete(c.expires, k)
		}
	}
	_ = prefix
	return nil
}

func (c *MemoryCache) prefixed(key string) string {
	return WithPrefix(c.cfg.Prefix, key)
}

// matchMemoryPattern 仅支持 * 通配符：
// - "abc*" 匹配以 "abc" 开头的字符串；
// - "*abc" 匹配以 "abc" 结尾的字符串；
// - "*abc*" 匹配包含 "abc" 的字符串；
// - 不含 * 则精确匹配。
func matchMemoryPattern(pattern, s string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == s
	}
	parts := strings.Split(pattern, "*")
	start := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(s[start:], part)
		if idx < 0 {
			return false
		}
		if i == 0 && idx != 0 {
			return false
		}
		start += idx + len(part)
	}
	lastPart := parts[len(parts)-1]
	if lastPart != "" && !strings.HasSuffix(s, lastPart) {
		return false
	}
	return true
}
