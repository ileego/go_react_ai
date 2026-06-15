package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

const loggerContextKey = "logger"

// GetLogger 从 gin.Context 获取注入的 logger；若不存在则返回默认 logger。
// Handler 和 Service 可通过它获得已经携带 request_id 等上下文信息的 logger。
func GetLogger(c *gin.Context) *slog.Logger {
	if v, ok := c.Get(loggerContextKey); ok {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	return slog.Default()
}
