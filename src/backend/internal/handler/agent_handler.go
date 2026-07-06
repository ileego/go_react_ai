package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/service"
	"github.com/ileego/go_react_ai/pkg/response"
)

// AgentHandler 智能体相关 HTTP 接口
type AgentHandler struct {
	svc service.AgentService
}

// NewAgentHandler 创建 AgentHandler
func NewAgentHandler(svc service.AgentService) *AgentHandler {
	return &AgentHandler{svc: svc}
}

// Dispatch 派发报告生成任务
// POST /api/reports/:id/dispatch
func (h *AgentHandler) Dispatch(c *gin.Context) {
	reportID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.svc.Dispatch(c.Request.Context(), reportID); err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c)
}
