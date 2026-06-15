package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 结构化请求日志中间件
// 记录：请求方法、路径、状态码、耗时、客户端 IP、请求 ID
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 把带 request_id 的 logger 注入上下文，供后续 handler/service 使用
		logger := slog.With(slog.String("request_id", GetRequestID(c)))
		c.Set(loggerContextKey, logger)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("http request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", c.ClientIP()),
			slog.Int("errors", len(c.Errors)),
		)
	}
}
