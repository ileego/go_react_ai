package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/pkg/response"
)

// Recovery 自定义 panic 恢复中间件
// 捕获未处理的 panic，返回 500 错误，并将堆栈信息写入日志
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger := GetLogger(c)
				logger.Error("panic recovered",
					slog.Any("error", err),
					slog.String("path", c.Request.URL.Path),
					slog.String("method", c.Request.Method),
				)
				response.InternalServerError(c, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
