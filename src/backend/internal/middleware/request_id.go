package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const HeaderXRequestID = "X-Request-ID"

// requestIDKey 用于把 request_id 存放到 context.Context 中。
// 使用自定义类型而非字符串，避免不同包之间的 key 冲突。
type requestIDKey struct{}

// RequestID 为每个请求注入唯一追踪 ID。
// 如果客户端已提供，则复用；否则生成新的随机 ID。
// 注入的 request_id 同时写入 gin.Context 和 context.Context，方便全链路透传。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderXRequestID)
		if rid == "" {
			rid = generateID()
		}

		// 1) 保留 gin.Context 中的读取方式，兼容现有 Handler/Middleware
		c.Set(HeaderXRequestID, rid)

		// 2) 同时写入 request.Context，使 Service/Repository/Worker 都能通过 ctx 获取
		ctx := context.WithValue(c.Request.Context(), requestIDKey{}, rid)
		c.Request = c.Request.WithContext(ctx)

		c.Writer.Header().Set(HeaderXRequestID, rid)
		c.Next()
	}
}

// GetRequestID 从 gin.Context 或对应的 request.Context 中获取请求 ID。
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(HeaderXRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return GetRequestIDFromContext(c.Request.Context())
}

// GetRequestIDFromContext 从 context.Context 中获取请求 ID。
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}

func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// 降级：使用时间戳（几乎不会发生）
		return "unknown"
	}
	return hex.EncodeToString(b)
}
