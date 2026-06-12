package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/service"
	"github.com/ileego/go_react_ai/pkg/response"
)

// ReportHandler 报告相关 HTTP 接口
type ReportHandler struct {
	svc service.ReportService
}

// NewReportHandler 创建 ReportHandler
func NewReportHandler(svc service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// CreateRequest 创建报告请求参数
type CreateRequest struct {
	Title string `json:"title" binding:"required,min=1,max=200"`
	Topic string `json:"topic" binding:"required,min=1,max=500"`
}

// Create 创建报告
// POST /api/reports
func (h *ReportHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: 从 JWT 中获取真实用户ID
	var userID int64 = 1

	report, err := h.svc.Create(c.Request.Context(), userID, req.Title, req.Topic)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, report)
}

// Get 获取报告详情
// GET /api/reports/:id
func (h *ReportHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	report, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Data(c, report)
}

// List 获取报告列表
// GET /api/reports?page=1&page_size=20
func (h *ReportHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// TODO: 从 JWT 中获取真实用户ID
	var userID int64 = 1

	reports, total, err := h.svc.ListByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.List(c, reports, total, page, pageSize)
}

// Cancel 取消报告
// POST /api/reports/:id/cancel
func (h *ReportHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.svc.Cancel(c.Request.Context(), id); err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c)
}
