package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/middleware"
	"github.com/ileego/go_react_ai/internal/service"
	"github.com/ileego/go_react_ai/pkg/response"
)

// SearchHandler 搜索相关 HTTP 接口。
type SearchHandler struct {
	svc service.SearchService
}

// NewSearchHandler 创建 SearchHandler。
func NewSearchHandler(svc service.SearchService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// Reports 搜索报告。
// GET /api/search/reports?q=keyword&page=1&page_size=20
func (h *SearchHandler) Reports(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.BadRequest(c, "搜索关键词不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userID := middleware.GetUserID(c)
	reports, total, err := h.svc.SearchReports(c.Request.Context(), userID, query, page, pageSize)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.List(c, reports, total, page, pageSize)
}
