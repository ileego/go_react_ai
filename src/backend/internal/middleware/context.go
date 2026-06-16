package middleware

import (
	"context"
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
	return GetLoggerFromContext(c.Request.Context())
}

// GetLoggerFromContext 从 context.Context 中构造一个携带 request_id 的 logger。
// 当 logger 没有通过 gin.Context 显式注入时，Service/Worker 可通过此方法复用请求追踪 ID。
func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if rid := GetRequestIDFromContext(ctx); rid != "" {
		return slog.With(slog.String("request_id", rid))
	}
	return slog.Default()
}
