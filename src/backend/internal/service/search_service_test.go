package service

import (
	"context"
	"testing"

	"github.com/ileego/go_react_ai/internal/repository/memory"
)

func TestSearchService_SearchReports(t *testing.T) {
	reportRepo := memory.NewReportRepository()
	searchRepo := memory.NewSearchRepository(reportRepo)
	svc := NewSearchService(searchRepo)
	ctx := context.Background()

	// 创建报告
	_, _ = NewReportService(reportRepo, nil).Create(ctx, 1, "人工智能医疗应用", "AI 在医疗诊断中的研究")
	_, _ = NewReportService(reportRepo, nil).Create(ctx, 1, "机器学习基础", "机器学习算法入门")
	_, _ = NewReportService(reportRepo, nil).Create(ctx, 1, "人工智能金融应用", "AI 在金融风控中的研究")
	_, _ = NewReportService(reportRepo, nil).Create(ctx, 2, "其他用户报告", "人工智能")

	reports, total, err := svc.SearchReports(ctx, 1, "人工智能", 1, 10)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(reports) != 2 {
		t.Errorf("len = %d, want 2", len(reports))
	}

	// 空查询
	_, total, err = svc.SearchReports(ctx, 1, "", 1, 10)
	if err != nil {
		t.Fatalf("empty search failed: %v", err)
	}
	if total != 0 {
		t.Errorf("empty query total = %d, want 0", total)
	}

	// 无结果
	_, total, err = svc.SearchReports(ctx, 1, "区块链", 1, 10)
	if err != nil {
		t.Fatalf("no result search failed: %v", err)
	}
	if total != 0 {
		t.Errorf("no result total = %d, want 0", total)
	}
}

func TestSearchService_Pagination(t *testing.T) {
	reportRepo := memory.NewReportRepository()
	searchRepo := memory.NewSearchRepository(reportRepo)
	svc := NewSearchService(searchRepo)
	ctx := context.Background()

	for range 5 {
		_, _ = NewReportService(reportRepo, nil).Create(ctx, 1, "人工智能研究", "主题")
	}

	reports, total, err := svc.SearchReports(ctx, 1, "人工智能", 1, 2)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(reports) != 2 {
		t.Errorf("page 1 len = %d, want 2", len(reports))
	}
}
