package security

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *redis.Client {
	s := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: s.Addr()})
}

func TestRedisRateLimiter_AllowIP(t *testing.T) {
	client := newTestRedis(t)
	defer func() { _ = client.Close() }()

	limiter := NewRedisRateLimiter(client, IPLimitConfig{
		Limit:  2,
		Window: time.Minute,
	}, LoginLimitConfig{})

	ctx := context.Background()
	allowed, err := limiter.AllowIP(ctx, "192.168.1.1")
	if err != nil {
		t.Fatalf("allow ip failed: %v", err)
	}
	if !allowed {
		t.Fatal("first request should be allowed")
	}

	allowed, _ = limiter.AllowIP(ctx, "192.168.1.1")
	if !allowed {
		t.Fatal("second request should be allowed")
	}

	allowed, _ = limiter.AllowIP(ctx, "192.168.1.1")
	if allowed {
		t.Fatal("third request should be denied")
	}
}

func TestRedisRateLimiter_AllowLogin(t *testing.T) {
	client := newTestRedis(t)
	defer func() { _ = client.Close() }()

	limiter := NewRedisRateLimiter(client, IPLimitConfig{}, LoginLimitConfig{
		MaxAttempts: 2,
		Window:      time.Minute,
		Lockout:     time.Minute,
	})

	ctx := context.Background()
	key := "user@example.com"

	allowed, _, _ := limiter.AllowLogin(ctx, key)
	if !allowed {
		t.Fatal("login should be allowed initially")
	}

	_, _, _ = limiter.RecordLoginFailure(ctx, key)
	allowed, _, _ = limiter.AllowLogin(ctx, key)
	if !allowed {
		t.Fatal("login should still be allowed after one failure")
	}

	_, _, _ = limiter.RecordLoginFailure(ctx, key)
	allowed, lockout, _ := limiter.AllowLogin(ctx, key)
	if allowed {
		t.Fatal("login should be locked after max attempts")
	}
	if lockout <= 0 {
		t.Error("lockout duration should be positive")
	}
}

func TestRedisRateLimiter_ResetLoginFailures(t *testing.T) {
	client := newTestRedis(t)
	defer func() { _ = client.Close() }()

	limiter := NewRedisRateLimiter(client, IPLimitConfig{}, LoginLimitConfig{
		MaxAttempts: 2,
		Window:      time.Minute,
		Lockout:     time.Minute,
	})

	ctx := context.Background()
	key := "user@example.com"

	_, _, _ = limiter.RecordLoginFailure(ctx, key)
	_, _, _ = limiter.RecordLoginFailure(ctx, key)

	if err := limiter.ResetLoginFailures(ctx, key); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	allowed, _, _ := limiter.AllowLogin(ctx, key)
	if !allowed {
		t.Fatal("login should be allowed after reset")
	}
}

func TestRedisTokenBlacklist(t *testing.T) {
	client := newTestRedis(t)
	defer func() { _ = client.Close() }()

	blacklist := NewRedisTokenBlacklist(client)
	ctx := context.Background()

	blacklisted, err := blacklist.IsBlacklisted(ctx, "jti-1")
	if err != nil {
		t.Fatalf("check blacklist failed: %v", err)
	}
	if blacklisted {
		t.Fatal("jti-1 should not be blacklisted initially")
	}

	if err := blacklist.Add(ctx, "jti-1", time.Minute); err != nil {
		t.Fatalf("add to blacklist failed: %v", err)
	}

	blacklisted, err = blacklist.IsBlacklisted(ctx, "jti-1")
	if err != nil {
		t.Fatalf("check blacklist failed: %v", err)
	}
	if !blacklisted {
		t.Fatal("jti-1 should be blacklisted")
	}
}

func TestRedisTokenBlacklist_EmptyJTI(t *testing.T) {
	client := newTestRedis(t)
	defer func() { _ = client.Close() }()

	blacklist := NewRedisTokenBlacklist(client)
	ctx := context.Background()

	if err := blacklist.Add(ctx, "", time.Minute); err == nil {
		t.Fatal("adding empty jti should fail")
	}

	blacklisted, err := blacklist.IsBlacklisted(ctx, "")
	if err != nil {
		t.Fatalf("check empty jti failed: %v", err)
	}
	if blacklisted {
		t.Fatal("empty jti should not be blacklisted")
	}
}
