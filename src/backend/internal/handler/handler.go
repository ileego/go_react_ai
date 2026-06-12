// Package handler 定义 HTTP 接口层。
// 负责：参数绑定、请求校验、调用 Service、构造响应。
// 不包含任何业务逻辑。
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/service"
)

// Handlers 汇总所有 HTTP 处理器，方便路由注册时统一注入依赖
type Handlers struct {
	Report *ReportHandler
	Agent  *AgentHandler
	Health *HealthHandler
}

// NewHandlers 创建处理器实例，注入 Service 依赖
func NewHandlers(
	reportSvc service.ReportService,
	agentSvc service.AgentService,
) *Handlers {
	return &Handlers{
		Report: NewReportHandler(reportSvc),
		Agent:  NewAgentHandler(agentSvc),
		Health: NewHealthHandler(),
	}
}

// RegisterRoutes 将所有路由注册到 Gin Engine
func (h *Handlers) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/health", h.Health.Check)

		reports := api.Group("/reports")
		{
			reports.POST("", h.Report.Create)
			reports.GET("", h.Report.List)
			reports.GET("/:id", h.Report.Get)
			reports.POST("/:id/cancel", h.Report.Cancel)
			reports.POST("/:report_id/dispatch", h.Agent.Dispatch)
		}
	}
}
