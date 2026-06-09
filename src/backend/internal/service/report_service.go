// Package service 实现业务逻辑层。
package service

import (
	"context"
	"fmt"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
	"github.com/ileego/go_react_ai/pkg/errors"
)

// reportService 实现 ReportService 接口
type reportService struct {
	repo repository.ReportRepository
}

// NewReportService 创建 ReportService 实例
func NewReportService(repo repository.ReportRepository) ReportService {
	return &reportService{repo: repo}
}

// Create 创建新的研究报告
func (s *reportService) Create(ctx context.Context, userID int64, title, topic string) (*domain.Report, error) {
	report := &domain.Report{
		Title:     title,
		Topic:     topic,
		CreatedBy: userID,
	}

	if err := report.Validate(); err != nil {
		return nil, errors.NewValidation("report", err.Error())
	}

	if err := s.repo.Create(ctx, report); err != nil {
		return nil, errors.NewInternal("创建报告失败", err)
	}

	return report, nil
}

// GetByID 获取报告详情
func (s *reportService) GetByID(ctx context.Context, id int64) (*domain.Report, error) {
	report, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFound("report", id)
	}
	return report, nil
}

// ListByUser 获取用户的报告列表
func (s *reportService) ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]*domain.Report, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 先获取总数（简化实现：先取全部再分页）
	all, err := s.repo.ListByUser(ctx, userID, 10000, 0)
	if err != nil {
		return nil, 0, errors.NewInternal("查询报告列表失败", err)
	}
	total := len(all)

	offset := (page - 1) * pageSize
	reports, err := s.repo.ListByUser(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, errors.NewInternal("查询报告列表失败", err)
	}

	return reports, total, nil
}

// Cancel 取消正在进行的报告
func (s *reportService) Cancel(ctx context.Context, id int64) error {
	report, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFound("report", id)
	}

	if !report.CanCancel() {
		return errors.NewValidation("status", fmt.Sprintf("当前状态 %s 不允许取消", report.Status))
	}

	if err := s.repo.UpdateStatus(ctx, id, domain.ReportStatusFailed); err != nil {
		return errors.NewInternal("取消报告失败", err)
	}

	return nil
}
