package handler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/repository/memory"
	"github.com/ileego/go_react_ai/internal/service"
)

func newTestSearchHandler() (*SearchHandler, *gin.Engine) {
	reportRepo := memory.NewReportRepository()
	searchRepo := memory.NewSearchRepository(reportRepo)
	svc := service.NewSearchService(searchRepo)
	h := NewSearchHandler(svc)

	r := newTestRouter()
	r.GET("/api/search/reports", h.Reports)
	return h, r
}

func TestSearchHandler_Reports(t *testing.T) {
	_, r := newTestSearchHandler()
	ctx := t.Context()

	// 创建测试数据
	reportRepo := memory.NewReportRepository()
	svc := service.NewReportService(reportRepo, nil)
	_, _ = svc.Create(ctx, testUserID, "人工智能医疗", "AI 医疗")
	_, _ = svc.Create(ctx, testUserID, "机器学习基础", "ML 基础")

	w := doRequest(t, r, http.MethodGet, "/api/search/reports?q=人工智能", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestSearchHandler_Reports_MissingQuery(t *testing.T) {
	_, r := newTestSearchHandler()

	w := doRequest(t, r, http.MethodGet, "/api/search/reports", "", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
