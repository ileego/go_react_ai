package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const HeaderXRequestID = "X-Request-ID"

// RequestID 为每个请求注入唯一追踪 ID。
// 如果客户端已提供，则复用；否则生成新的随机 ID。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderXRequestID)
		if rid == "" {
			rid = generateID()
		}
		c.Set(HeaderXRequestID, rid)
		c.Writer.Header().Set(HeaderXRequestID, rid)
		c.Next()
	}
}

// GetRequestID 从上下文中获取请求 ID
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(HeaderXRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
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
