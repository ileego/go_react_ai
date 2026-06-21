package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type testValue struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func TestRedisCache_GetSet(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	c := NewRedisCache(client, DefaultConfig())
	ctx := context.Background()

	// 未命中
	var got testValue
	if err := c.Get(ctx, "key", &got); err != ErrCacheMiss {
		t.Fatalf("expected cache miss, got %v", err)
	}

	// 写入并读取
	want := testValue{ID: 1, Name: "test"}
	if err := c.Set(ctx, "key", want, time.Minute); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if err := c.Get(ctx, "key", &got); err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}

	// 删除后未命中
	if err := c.Delete(ctx, "key"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := c.Get(ctx, "key", &got); err != ErrCacheMiss {
		t.Fatalf("expected cache miss after delete, got %v", err)
	}
}

func TestRedisCache_DeletePattern(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	c := NewRedisCache(client, Config{Prefix: "test"})
	ctx := context.Background()

	_ = c.Set(ctx, "reports:user:1:page:1:size:10", "a", time.Minute)
	_ = c.Set(ctx, "reports:user:1:page:2:size:10", "b", time.Minute)
	_ = c.Set(ctx, "reports:user:2:page:1:size:10", "c", time.Minute)

	if err := c.DeletePattern(ctx, "reports:user:1:*"); err != nil {
		t.Fatalf("delete pattern failed: %v", err)
	}

	var v string
	if err := c.Get(ctx, "reports:user:1:page:1:size:10", &v); err != ErrCacheMiss {
		t.Errorf("expected user:1 page:1 deleted, got %v", err)
	}
	if err := c.Get(ctx, "reports:user:1:page:2:size:10", &v); err != ErrCacheMiss {
		t.Errorf("expected user:1 page:2 deleted, got %v", err)
	}
	if err := c.Get(ctx, "reports:user:2:page:1:size:10", &v); err != nil {
		t.Errorf("expected user:2 retained, got %v", err)
	}
}

func TestMemoryCache_GetSet(t *testing.T) {
	c := NewMemoryCache(DefaultConfig())
	ctx := context.Background()

	var got testValue
	if err := c.Get(ctx, "key", &got); err != ErrCacheMiss {
		t.Fatalf("expected cache miss, got %v", err)
	}

	want := testValue{ID: 2, Name: "memory"}
	if err := c.Set(ctx, "key", want, time.Minute); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if err := c.Get(ctx, "key", &got); err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestMemoryCache_TTL(t *testing.T) {
	c := NewMemoryCache(DefaultConfig())
	ctx := context.Background()

	_ = c.Set(ctx, "key", testValue{ID: 1}, 50*time.Millisecond)
	time.Sleep(60 * time.Millisecond)

	var got testValue
	if err := c.Get(ctx, "key", &got); err != ErrCacheMiss {
		t.Fatalf("expected expired cache miss, got %v", err)
	}
}

func TestMemoryCache_DeletePattern(t *testing.T) {
	c := NewMemoryCache(Config{Prefix: "test"})
	ctx := context.Background()

	_ = c.Set(ctx, "reports:user:1:page:1:size:10", "a", time.Minute)
	_ = c.Set(ctx, "reports:user:1:page:2:size:10", "b", time.Minute)
	_ = c.Set(ctx, "reports:user:2:page:1:size:10", "c", time.Minute)

	_ = c.DeletePattern(ctx, "reports:user:1:*")

	var v string
	if err := c.Get(ctx, "reports:user:1:page:1:size:10", &v); err != ErrCacheMiss {
		t.Errorf("expected user:1 page:1 deleted, got %v", err)
	}
	if err := c.Get(ctx, "reports:user:2:page:1:size:10", &v); err != nil {
		t.Errorf("expected user:2 retained, got %v", err)
	}
}

func TestCacheKeys(t *testing.T) {
	if got := ReportKey(1); got != "report:1" {
		t.Errorf("ReportKey = %q", got)
	}
	if got := ReportListKey(1, 2, 10); got != "reports:user:1:page:2:size:10" {
		t.Errorf("ReportListKey = %q", got)
	}
	if got := WithPrefix("goai", "report:1"); got != "goai:report:1" {
		t.Errorf("WithPrefix = %q", got)
	}
}
