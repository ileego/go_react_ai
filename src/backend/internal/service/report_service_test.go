package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/yourname/go_react_ai/internal/domain"
	"github.com/yourname/go_react_ai/internal/repository/memory"
)

// TestReportService_Create 测试创建报告
func TestReportService_Create(t *testing.T) {
	// Arrange: 准备测试数据
	repo := memory.NewReportRepository()
	svc := NewReportService(repo)
	ctx := context.Background()

	// Act: 执行被测操作
	report, err := svc.Create(ctx, 1, "AI 医疗研究", "人工智能在医疗诊断中的应用")

	// Assert: 验证结果
	if err != nil {
		t.Fatalf("创建报告失败: %v", err)
	}
	if report.ID == 0 {
		t.Error("报告 ID 应该被自动分配")
	}
	if report.Status != domain.ReportStatusPending {
		t.Errorf("新建报告状态应为 pending, 得到 %s", report.Status)
	}
	if report.Title != "AI 医疗研究" {
		t.Errorf("标题不匹配: got %s, want %s", report.Title, "AI 医疗研究")
	}
}

// TestReportService_Create_Validation 测试创建报告时的参数校验
func TestReportService_Create_Validation(t *testing.T) {
	repo := memory.NewReportRepository()
	svc := NewReportService(repo)
	ctx := context.Background()

	tests := []struct {
		name    string
		title   string
		topic   string
		wantErr bool
	}{
		{"空标题", "", "有效主题", true},
		{"空主题", "有效标题", "", true},
		{"正常创建", "有效标题", "有效主题", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, 1, tt.title, tt.topic)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestReportService_GetByID 测试获取报告
func TestReportService_GetByID(t *testing.T) {
	repo := memory.NewReportRepository()
	svc := NewReportService(repo)
	ctx := context.Background()

	// 先创建一个报告
	created, _ := svc.Create(ctx, 1, "测试报告", "测试主题")

	// 获取已创建的报告
	report, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("获取报告失败: %v", err)
	}
	if report.ID != created.ID {
		t.Errorf("ID 不匹配: got %d, want %d", report.ID, created.ID)
	}

	// 获取不存在的报告
	_, err = svc.GetByID(ctx, 99999)
	if err == nil {
		t.Error("获取不存在的报告应该返回错误")
	}
}

// TestReportService_Cancel 测试取消报告
func TestReportService_Cancel(t *testing.T) {
	repo := memory.NewReportRepository()
	svc := NewReportService(repo)
	ctx := context.Background()

	// 创建并取消一个 pending 状态的报告
	report, _ := svc.Create(ctx, 1, "可取消报告", "主题")
	if err := svc.Cancel(ctx, report.ID); err != nil {
		t.Fatalf("取消报告失败: %v", err)
	}

	// 验证状态已变更
	updated, _ := svc.GetByID(ctx, report.ID)
	if updated.Status != domain.ReportStatusFailed {
		t.Errorf("取消后状态应为 failed, 得到 %s", updated.Status)
	}
}

// TestReportService_Cancel_AlreadyCompleted 测试已完成报告不能取消
func TestReportService_Cancel_AlreadyCompleted(t *testing.T) {
	repo := memory.NewReportRepository()
	svc := NewReportService(repo)
	ctx := context.Background()

	report, _ := svc.Create(ctx, 1, "已完成报告", "主题")
	// 手动更新状态为 completed（模拟报告已完成）
	repo.UpdateStatus(ctx, report.ID, domain.ReportStatusCompleted)

	// 尝试取消已完成的报告
	err := svc.Cancel(ctx, report.ID)
	if err == nil {
		t.Error("取消已完成的报告应该返回错误")
	}
}

// TestReportService_ListByUser 测试分页列表
func TestReportService_ListByUser(t *testing.T) {
	repo := memory.NewReportRepository()
	svc := NewReportService(repo)
	ctx := context.Background()

	// 创建 5 个报告
	for i := 0; i < 5; i++ {
		svc.Create(ctx, 1, fmt.Sprintf("报告%d", i+1), "主题")
	}

	// 分页查询
	reports, total, err := svc.ListByUser(ctx, 1, 1, 2)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if total != 5 {
		t.Errorf("总数应为 5, 得到 %d", total)
	}
	if len(reports) != 2 {
		t.Errorf("每页 2 条，应返回 2 条, 得到 %d", len(reports))
	}
}
