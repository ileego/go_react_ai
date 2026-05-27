package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go_react_ai/internal/service"
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
func (h *AgentHandler) Dispatch(c *gin.Context) {
	reportID, err := strconv.ParseInt(c.Param("report_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report_id"})
		return
	}

	if err := h.svc.Dispatch(c.Request.Context(), reportID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "dispatched"})
}
