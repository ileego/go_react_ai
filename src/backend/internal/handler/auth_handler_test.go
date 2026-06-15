package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/auth"
	"github.com/ileego/go_react_ai/internal/repository/memory"
	"github.com/ileego/go_react_ai/internal/security"
	"github.com/ileego/go_react_ai/internal/service"
	"github.com/redis/go-redis/v9"
)

func newTestAuthHandler(t *testing.T) (*AuthHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)

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
	authSvc := service.NewAuthService(userRepo, jwtCfg, nil, rl, blacklist)

	h := NewAuthHandler(authSvc, nil, nil, rl)
	r := gin.New()
	api := r.Group("/api/auth")
	{
		api.POST("/register", h.Register)
		api.POST("/login", h.Login)
		api.POST("/refresh", h.Refresh)
	}

	return h, r
}

func TestAuthHandler_Register(t *testing.T) {
	_, r := newTestAuthHandler(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "Password123",
		"nickname": "Tester",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want status 201, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestAuthHandler_RegisterWeakPassword(t *testing.T) {
	_, r := newTestAuthHandler(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "123",
		"nickname": "Tester",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want status 400, got %d", w.Code)
	}
}

func TestAuthHandler_Login(t *testing.T) {
	_, r := newTestAuthHandler(t)

	// 先注册
	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "Password123",
		"nickname": "Tester",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 再登录
	body, _ = json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "Password123",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Data TokenResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if resp.Data.AccessToken == "" {
		t.Error("access token is empty")
	}
}

func TestAuthHandler_LoginWrongPassword(t *testing.T) {
	_, r := newTestAuthHandler(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "Password123",
		"nickname": "Tester",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body, _ = json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "WrongPassword",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want status 401, got %d", w.Code)
	}
}
