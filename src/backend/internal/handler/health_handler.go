package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/middleware"
	"github.com/ileego/go_react_ai/pkg/response"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	dbHealth func(context.Context) error
}

// NewHealthHandler 创建 HealthHandler
// dbHealth 用于就绪探针检查数据库等依赖，传 nil 表示不检查依赖。
func NewHealthHandler(dbHealth func(context.Context) error) *HealthHandler {
	return &HealthHandler{dbHealth: dbHealth}
}

// Check 存活探针
// GET /api/health
func (h *HealthHandler) Check(c *gin.Context) {
	response.Data(c, gin.H{
		"status": "ok",
		"time":   time.Now().Unix(),
	})
}

// Ready 就绪探针
// GET /api/ready
func (h *HealthHandler) Ready(c *gin.Context) {
	if h.dbHealth == nil {
		response.Data(c, gin.H{
			"status": "ready",
			"time":   time.Now().Unix(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := h.dbHealth(ctx); err != nil {
		logger := middleware.GetLogger(c)
		logger.Error("readiness check failed", "error", err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"reason": "database_unavailable",
		})
		return
	}

	response.Data(c, gin.H{
		"status": "ready",
		"time":   time.Now().Unix(),
	})
}
