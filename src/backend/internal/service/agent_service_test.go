package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository/memory"
	"github.com/ileego/go_react_ai/pkg/httpx"
	"github.com/ileego/go_react_ai/pkg/worker"
)

func fallback() func(error, *http.Response) ([]byte, error) {
	return func(err error, resp *http.Response) ([]byte, error) {
		return []byte(`{"output":"降级生成的报告内容"}`), nil
	}
}

func newTestAgentService(endpoint string, client *httpx.Client) (AgentService, *worker.Pool, *memory.ReportRepository, *memory.AgentTaskRepository) {
	reportRepo := memory.NewReportRepository()
	taskRepo := memory.NewAgentTaskRepository()
	pool := worker.NewPool(2, 10)
	pool.Start()

	reportSvc := NewReportService(reportRepo, nil)
	svc := NewAgentService(reportSvc, taskRepo, pool, client, endpoint)
	return svc, pool, reportRepo, taskRepo
}

func TestAgentService_Dispatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"output":"AI 生成的报告内容"}`))
	}))
	defer server.Close()

	client := httpx.NewClientWithFallback(5*time.Second, httpx.DefaultRetryConfig(), fallback())
	svc, pool, reportRepo, taskRepo := newTestAgentService(server.URL+"/api/mock-ai", client)
	defer pool.Stop()

	ctx := context.Background()
	report := &domain.Report{
		Title:     "测试报告",
		Topic:     "测试主题",
		CreatedBy: 1,
	}
	if err := reportRepo.Create(ctx, report); err != nil {
		t.Fatalf("create report failed: %v", err)
	}

	if err := svc.Dispatch(ctx, report.ID); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	updated, _ := reportRepo.GetByID(ctx, report.ID)
	if updated.Status != domain.ReportStatusRunning {
		t.Errorf("status should be running, got %s", updated.Status)
	}

	// 等待异步任务完成
	time.Sleep(200 * time.Millisecond)

	tasks, _ := taskRepo.GetByReportID(ctx, report.ID)
	if len(tasks) == 0 {
		t.Fatal("task should be created")
	}
	if tasks[0].Status != "completed" {
		t.Errorf("task status should be completed, got %s", tasks[0].Status)
	}

	updated, _ = reportRepo.GetByID(ctx, report.ID)
	if updated.Status != domain.ReportStatusCompleted {
		t.Errorf("report status should be completed, got %s", updated.Status)
	}
}

func TestAgentService_Dispatch_NotFound(t *testing.T) {
	client := httpx.NewClient(5*time.Second, httpx.DefaultRetryConfig())
	svc, pool, _, _ := newTestAgentService("", client)
	defer pool.Stop()

	ctx := context.Background()
	err := svc.Dispatch(ctx, 99999)
	if err == nil {
		t.Error("should return error for non-existent report")
	}
}

func TestAgentService_Dispatch_NotPending(t *testing.T) {
	client := httpx.NewClient(5*time.Second, httpx.DefaultRetryConfig())
	svc, pool, reportRepo, _ := newTestAgentService("", client)
	defer pool.Stop()

	ctx := context.Background()
	report := &domain.Report{Title: "测试", Topic: "主题", CreatedBy: 1}
	if err := reportRepo.Create(ctx, report); err != nil {
		t.Fatalf("create report failed: %v", err)
	}
	_ = reportRepo.UpdateStatus(ctx, report.ID, domain.ReportStatusCompleted)

	err := svc.Dispatch(ctx, report.ID)
	if err == nil {
		t.Error("should return error for non-pending report")
	}
}

func TestAgentService_Dispatch_Fallback(t *testing.T) {
	// 不提供 mock server，让请求立即失败并触发 fallback
	client := httpx.NewClientWithFallback(
		5*time.Second,
		httpx.RetryConfig{MaxRetries: 0},
		fallback(),
	)
	svc, pool, reportRepo, taskRepo := newTestAgentService("http://localhost:1/api/mock-ai", client)
	defer pool.Stop()

	ctx := context.Background()
	report := &domain.Report{Title: "测试报告", Topic: "测试主题", CreatedBy: 1}
	if err := reportRepo.Create(ctx, report); err != nil {
		t.Fatalf("create report failed: %v", err)
	}

	if err := svc.Dispatch(ctx, report.ID); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	tasks, _ := taskRepo.GetByReportID(ctx, report.ID)
	if len(tasks) == 0 {
		t.Fatal("task should be created")
	}
	if tasks[0].Output != "降级生成的报告内容" {
		t.Errorf("expected fallback output, got %s", tasks[0].Output)
	}

	updated, _ := reportRepo.GetByID(ctx, report.ID)
	if updated.Status != domain.ReportStatusCompleted {
		t.Errorf("report status should be completed, got %s", updated.Status)
	}
}
