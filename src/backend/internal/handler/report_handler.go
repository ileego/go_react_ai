package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go_react_ai/internal/domain"
	"github.com/yourname/go_react_ai/internal/service"
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
	Title string `json:"title" binding:"required"`
	Topic string `json:"topic" binding:"required"`
}

// Create 创建报告
func (h *ReportHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 从 JWT 中获取真实用户ID
	var userID int64 = 1

	report, err := h.svc.Create(c.Request.Context(), userID, req.Title, req.Topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// Get 获取报告详情
func (h *ReportHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	report, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// List 获取报告列表
func (h *ReportHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// TODO: 从 JWT 中获取真实用户ID
	var userID int64 = 1

	reports, total, err := h.svc.ListByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reports,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// Cancel 取消报告
func (h *ReportHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.Cancel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cancelled"})
}
