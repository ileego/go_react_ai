package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/pkg/response"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建 HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check 健康检查接口
// GET /api/health
func (h *HealthHandler) Check(c *gin.Context) {
	response.Data(c, gin.H{
		"status": "ok",
		"time":   time.Now().Unix(),
	})
}
