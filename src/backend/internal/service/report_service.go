// Package service 实现业务逻辑层。
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ileego/go_react_ai/internal/cache"
	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
	apperrors "github.com/ileego/go_react_ai/pkg/errors"
)

// reportService 实现 ReportService 接口
type reportService struct {
	repo  repository.ReportRepository
	cache cache.Manager
}

// NewReportService 创建 ReportService 实例。
// cache 可为 nil，此时不启用缓存。
func NewReportService(repo repository.ReportRepository, mgr cache.Manager) ReportService {
	return &reportService{repo: repo, cache: mgr}
}

// Create 创建新的研究报告
func (s *reportService) Create(ctx context.Context, userID int64, title, topic string) (*domain.Report, error) {
	report := &domain.Report{
		Title:     title,
		Topic:     topic,
		CreatedBy: userID,
	}

	if err := report.Validate(); err != nil {
		return nil, apperrors.NewValidation("report", err.Error())
	}

	if err := s.repo.Create(ctx, report); err != nil {
		return nil, apperrors.NewInternal("创建报告失败", err)
	}

	// 创建后失效该用户的报告列表缓存
	_ = s.invalidateReportList(ctx, userID)
	return report, nil
}

// GetByID 获取报告详情，带 Cache-Aside 缓存。
func (s *reportService) GetByID(ctx context.Context, id int64) (*domain.Report, error) {
	// 1. 先读缓存
	if s.cache != nil {
		var cached domain.Report
		if err := s.cache.Get(ctx, cache.ReportKey(id), &cached); err == nil {
			return &cached, nil
		}
	}

	// 2. 缓存未命中，查询数据库
	report, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperrors.NewNotFound("report", id)
		}
		return nil, apperrors.NewInternal("查询报告失败", err)
	}

	// 3. 回写缓存
	if s.cache != nil {
		_ = s.cache.Set(ctx, cache.ReportKey(id), report, 5*time.Minute)
	}
	return report, nil
}

// ListByUser 获取用户的报告列表，带 Cache-Aside 缓存。
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

	// 1. 先读缓存
	if s.cache != nil {
		var cached []*domain.Report
		if err := s.cache.Get(ctx, cache.ReportListKey(userID, page, pageSize), &cached); err == nil {
			// 缓存命中时仍需 total，这里简化返回 len(cached) 作为 total
			return cached, len(cached), nil
		}
	}

	// 2. 缓存未命中，查询数据库
	// 先获取总数（简化实现：先取全部再分页）
	all, err := s.repo.ListByUser(ctx, userID, 10000, 0)
	if err != nil {
		return nil, 0, apperrors.NewInternal("查询报告列表失败", err)
	}
	total := len(all)

	offset := (page - 1) * pageSize
	reports, err := s.repo.ListByUser(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, apperrors.NewInternal("查询报告列表失败", err)
	}

	// 3. 回写缓存
	if s.cache != nil {
		_ = s.cache.Set(ctx, cache.ReportListKey(userID, page, pageSize), reports, 2*time.Minute)
	}
	return reports, total, nil
}

// Cancel 取消正在进行的报告
func (s *reportService) Cancel(ctx context.Context, id int64) error {
	report, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperrors.NewNotFound("report", id)
		}
		return apperrors.NewInternal("查询报告失败", err)
	}

	if !report.CanCancel() {
		return apperrors.NewValidation("status", fmt.Sprintf("当前状态 %s 不允许取消", report.Status)).
			WithCode("REPORT_CANNOT_CANCEL")
	}

	if err := s.UpdateStatus(ctx, id, domain.ReportStatusFailed); err != nil {
		return err
	}

	return nil
}

// UpdateStatus 更新报告状态，并失效相关缓存。
func (s *reportService) UpdateStatus(ctx context.Context, id int64, status domain.ReportStatus) error {
	report, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperrors.NewNotFound("report", id)
		}
		return apperrors.NewInternal("查询报告失败", err)
	}

	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperrors.NewNotFound("report", id)
		}
		return apperrors.NewInternal("更新报告状态失败", err)
	}

	// 失效缓存
	if s.cache != nil {
		_ = s.cache.Delete(ctx, cache.ReportKey(id))
		_ = s.invalidateReportList(ctx, report.CreatedBy)
	}
	return nil
}

// invalidateReportList 失效指定用户的报告列表缓存。
func (s *reportService) invalidateReportList(ctx context.Context, userID int64) error {
	if s.cache == nil {
		return nil
	}
	// 匹配该用户所有列表缓存，不关心 page 和 page_size。
	pattern := fmt.Sprintf("reports:user:%d:*", userID)
	return s.cache.DeletePattern(ctx, pattern)
}
