package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/auth"
	"github.com/ileego/go_react_ai/internal/security"
	"github.com/redis/go-redis/v9"
)

func setupTestRouter(secret string, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(secret, nil))
	r.GET("/test", handler)
	return r
}

func TestJWTAuth_Success(t *testing.T) {
	secret := "test-secret"
	pair, _ := auth.GenerateTokenPair(auth.Config{
		Secret:         secret,
		AccessTokenTTL: time.Hour,
	}, 42, "user")

	r := setupTestRouter(secret, func(c *gin.Context) {
		if GetUserID(c) != 42 {
			t.Errorf("want userID 42, got %d", GetUserID(c))
		}
		if GetUserRole(c) != "user" {
			t.Errorf("want role user, got %s", GetUserRole(c))
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d", w.Code)
	}
}

func TestJWTAuth_MissingToken(t *testing.T) {
	secret := "test-secret"
	r := setupTestRouter(secret, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want status 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	secret := "test-secret"
	r := setupTestRouter(secret, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want status 401, got %d", w.Code)
	}
}

func TestJWTAuth_WrongTokenType(t *testing.T) {
	secret := "test-secret"
	pair, _ := auth.GenerateTokenPair(auth.Config{
		Secret:          secret,
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour,
	}, 42, "user")

	r := setupTestRouter(secret, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.RefreshToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want status 401, got %d", w.Code)
	}
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(userIDContextKey, int64(1))
		c.Set(userRoleContextKey, "admin")
		c.Next()
	})
	r.GET("/admin", RequireRole("admin", "system"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d", w.Code)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(userIDContextKey, int64(1))
		c.Set(userRoleContextKey, "user")
		c.Next()
	})
	r.GET("/admin", RequireRole("admin", "system"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want status 403, got %d", w.Code)
	}
}

func TestLoginRateLimit_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := redis.NewClient(&redis.Options{Addr: miniredis.RunT(t).Addr()})
	defer func() { _ = client.Close() }()

	limiter := security.NewRedisRateLimiter(client, security.IPLimitConfig{}, security.LoginLimitConfig{
		MaxAttempts: 5,
		Window:      time.Minute,
		Lockout:     time.Minute,
	})

	r := gin.New()
	r.POST("/login", LoginRateLimit(limiter), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	body := `{"email":"test@example.com","password":"Password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d", w.Code)
	}
}

func TestLoginRateLimit_Blocked(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := redis.NewClient(&redis.Options{Addr: miniredis.RunT(t).Addr()})
	defer func() { _ = client.Close() }()

	limiter := security.NewRedisRateLimiter(client, security.IPLimitConfig{}, security.LoginLimitConfig{
		MaxAttempts: 1,
		Window:      time.Minute,
		Lockout:     time.Minute,
	})

	r := gin.New()
	r.POST("/login", LoginRateLimit(limiter), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	body := `{"email":"test@example.com","password":"Password123"}`

	// 第一次记录一次失败
	_, _, _ = limiter.RecordLoginFailure(context.Background(), "email:test@example.com")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want status 403, got %d", w.Code)
	}
}
