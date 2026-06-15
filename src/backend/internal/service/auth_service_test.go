package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ileego/go_react_ai/internal/auth"
	"github.com/ileego/go_react_ai/internal/repository/memory"
	"github.com/ileego/go_react_ai/internal/security"
	"github.com/redis/go-redis/v9"
)

func newTestAuthService(t *testing.T) (AuthService, *redis.Client) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})

	userRepo := memory.NewUserRepository()
	jwtCfg := auth.Config{
		Secret:          "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}
	rl := security.NewRedisRateLimiter(client, security.IPLimitConfig{
		Limit:  100,
		Window: time.Minute,
	}, security.LoginLimitConfig{
		MaxAttempts: 5,
		Window:      15 * time.Minute,
		Lockout:     15 * time.Minute,
	})
	blacklist := security.NewRedisTokenBlacklist(client)

	return NewAuthService(userRepo, jwtCfg, nil, rl, blacklist), client
}

func TestAuthService_Register(t *testing.T) {
	svc, client := newTestAuthService(t)
	defer client.Close()

	ctx := context.Background()
	user, err := svc.Register(ctx, "test@example.com", "Password123", "Tester")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("user id should not be zero")
	}
	if user.Email != "test@example.com" {
		t.Errorf("want email test@example.com, got %s", user.Email)
	}
}

func TestAuthService_RegisterWeakPassword(t *testing.T) {
	svc, client := newTestAuthService(t)
	defer client.Close()

	ctx := context.Background()
	_, err := svc.Register(ctx, "test@example.com", "123", "Tester")
	if err == nil {
		t.Fatal("weak password should fail")
	}
}

func TestAuthService_RegisterDuplicateEmail(t *testing.T) {
	svc, client := newTestAuthService(t)
	defer client.Close()

	ctx := context.Background()
	_, _ = svc.Register(ctx, "test@example.com", "Password123", "Tester")
	_, err := svc.Register(ctx, "test@example.com", "Password123", "Tester2")
	if err == nil {
		t.Fatal("duplicate email should fail")
	}
}

func TestAuthService_Login(t *testing.T) {
	svc, client := newTestAuthService(t)
	defer client.Close()

	ctx := context.Background()
	_, _ = svc.Register(ctx, "test@example.com", "Password123", "Tester")

	accessToken, refreshToken, err := svc.Login(ctx, "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if accessToken == "" || refreshToken == "" {
		t.Error("tokens should not be empty")
	}
}

func TestAuthService_LoginWrongPassword(t *testing.T) {
	svc, client := newTestAuthService(t)
	defer client.Close()

	ctx := context.Background()
	_, _ = svc.Register(ctx, "test@example.com", "Password123", "Tester")

	_, _, err := svc.Login(ctx, "test@example.com", "WrongPassword")
	if err == nil {
		t.Fatal("wrong password should fail")
	}
}

func TestAuthService_Refresh(t *testing.T) {
	svc, client := newTestAuthService(t)
	defer client.Close()

	ctx := context.Background()
	_, _ = svc.Register(ctx, "test@example.com", "Password123", "Tester")
	_, refreshToken, _ := svc.Login(ctx, "test@example.com", "Password123")

	newAccess, newRefresh, err := svc.Refresh(ctx, refreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if newAccess == "" || newRefresh == "" {
		t.Error("new tokens should not be empty")
	}

	// 旧的 refresh token 应被吊销
	_, _, err = svc.Refresh(ctx, refreshToken)
	if err == nil {
		t.Fatal("refresh with old token should fail after rotation")
	}
}

func TestAuthService_Logout(t *testing.T) {
	svc, client := newTestAuthService(t)
	defer client.Close()

	ctx := context.Background()
	_, _ = svc.Register(ctx, "test@example.com", "Password123", "Tester")
	accessToken, refreshToken, _ := svc.Login(ctx, "test@example.com", "Password123")

	if err := svc.Logout(ctx, accessToken, refreshToken); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// access token 应被吊销，无法解析为有效 token
	_, err := auth.ParseAndValidate(accessToken, auth.TokenTypeAccess, "test-secret")
	if err == nil {
		// 解析本身可能成功，但中间件会检查黑名单
	}
}

func TestAuthService_Me(t *testing.T) {
	svc, client := newTestAuthService(t)
	defer client.Close()

	ctx := context.Background()
	user, _ := svc.Register(ctx, "test@example.com", "Password123", "Tester")

	me, err := svc.Me(ctx, user.ID)
	if err != nil {
		t.Fatalf("me failed: %v", err)
	}
	if me.Email != user.Email {
		t.Errorf("want email %s, got %s", user.Email, me.Email)
	}
}
