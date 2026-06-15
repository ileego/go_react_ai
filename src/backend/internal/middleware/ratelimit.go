package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/security"
	apperrors "github.com/ileego/go_react_ai/pkg/errors"
	"github.com/ileego/go_react_ai/pkg/response"
)

// RateLimit 基于客户端 IP 的全局限流中间件
func RateLimit(limiter security.RateLimiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, err := limiter.AllowIP(c.Request.Context(), c.ClientIP())
		if err != nil {
			response.FromError(c, apperrors.NewInternal("限流检查失败", err))
			c.Abort()
			return
		}
		if !allowed {
			c.Header("Retry-After", window.String())
			response.FromError(c, apperrors.NewForbidden("请求过于频繁，请稍后再试").WithCode("RATE_LIMITED"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// LoginRateLimit 登录接口专用限流与锁定中间件
// key 优先使用请求体中的 email，回退到客户端 IP
func LoginRateLimit(limiter security.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.PostForm("email")
		if key == "" {
			body, err := c.GetRawData()
			if err == nil {
				// 恢复请求体，供后续 Handler 再次绑定
				c.Request.Body = io.NopCloser(bytes.NewReader(body))
				var req struct {
					Email string `json:"email"`
				}
				if err := json.Unmarshal(body, &req); err == nil {
					key = req.Email
				}
			}
		}
		if key == "" {
			key = "ip:" + c.ClientIP()
		} else {
			key = "email:" + key
		}

		allowed, lockout, err := limiter.AllowLogin(c.Request.Context(), key)
		if err != nil {
			response.FromError(c, apperrors.NewInternal("登录限流检查失败", err))
			c.Abort()
			return
		}
		if !allowed {
			c.Header("Retry-After", lockout.String())
			response.FromError(c, apperrors.NewForbidden("登录失败次数过多，请稍后再试").WithCode("LOGIN_LOCKED"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitAfterLogin 登录失败后记录失败次数（供 Handler 调用）
func RateLimitAfterLogin(limiter security.RateLimiter, c *gin.Context, email string) {
	key := "email:" + email
	_, _, _ = limiter.RecordLoginFailure(c.Request.Context(), key)
}

// ResetLoginRateLimit 登录成功后重置失败计数（供 Handler 调用）
func ResetLoginRateLimit(limiter security.RateLimiter, c *gin.Context, email string) {
	key := "email:" + email
	_ = limiter.ResetLoginFailures(c.Request.Context(), key)
}
