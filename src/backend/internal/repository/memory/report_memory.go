// Package memory 提供基于内存的 Repository 实现，用于测试和开发阶段。
// 不依赖数据库，所有数据存储在 sync.Map 中。
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
)

// ReportRepository 内存版报告数据访问实现
type ReportRepository struct {
	mu      sync.RWMutex
	reports map[int64]*domain.Report
	nextID  int64
}

// NewReportRepository 创建内存版 ReportRepository
func NewReportRepository() *ReportRepository {
	return &ReportRepository{
		reports: make(map[int64]*domain.Report),
		nextID:  1,
	}
}

// GetByID 根据 ID 获取报告
func (r *ReportRepository) GetByID(_ context.Context, id int64) (*domain.Report, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	report, ok := r.reports[id]
	if !ok {
		return nil, fmt.Errorf("report %d not found", id)
	}
	// 返回副本，避免外部修改影响内部状态
	return copyReport(report), nil
}

// ListByUser 获取用户的报告列表
func (r *ReportRepository) ListByUser(_ context.Context, userID int64, limit, offset int) ([]*domain.Report, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Report
	for _, report := range r.reports {
		if report.CreatedBy == userID {
			result = append(result, copyReport(report))
		}
	}

	// 简单的分页
	if offset > len(result) {
		return []*domain.Report{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

// Create 创建报告
func (r *ReportRepository) Create(_ context.Context, report *domain.Report) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	report.ID = r.nextID
	r.nextID++
	report.CreatedAt = time.Now()
	report.UpdatedAt = report.CreatedAt
	report.Status = domain.ReportStatusPending

	r.reports[report.ID] = copyReport(report)
	return nil
}

// UpdateStatus 更新报告状态
func (r *ReportRepository) UpdateStatus(_ context.Context, id int64, status domain.ReportStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	report, ok := r.reports[id]
	if !ok {
		return fmt.Errorf("report %d not found", id)
	}
	report.Status = status
	report.UpdatedAt = time.Now()
	return nil
}

// copyReport 创建报告的深拷贝
func copyReport(r *domain.Report) *domain.Report {
	return &domain.Report{
		ID:          r.ID,
		Title:       r.Title,
		Topic:       r.Topic,
		Status:      r.Status,
		Content:     r.Content,
		Sources:     append([]string(nil), r.Sources...),
		CreatedBy:   r.CreatedBy,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		CompletedAt: r.CompletedAt,
	}
}
