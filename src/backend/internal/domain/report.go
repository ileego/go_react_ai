package domain

import (
	"errors"
	"time"
)

// ReportStatus 报告状态
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"   // 待处理
	ReportStatusRunning   ReportStatus = "running"   // 执行中
	ReportStatusCompleted ReportStatus = "completed" // 已完成
	ReportStatusFailed    ReportStatus = "failed"    // 失败
)

// Report 表示一份深度研究报告
type Report struct {
	ID          int64
	Title       string
	Topic       string // 研究主题
	Status      ReportStatus
	Content     string   // 最终报告内容（Markdown）
	Sources     []string // 引用来源
	CreatedBy   int64    // 创建者用户ID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// Validate 校验报告字段
func (r *Report) Validate() error {
	if r.Title == "" {
		return errors.New("标题不能为空")
	}
	if r.Topic == "" {
		return errors.New("主题不能为空")
	}
	return nil
}

// CanCancel 报告是否可以取消
func (r *Report) CanCancel() bool {
	return r.Status == ReportStatusPending || r.Status == ReportStatusRunning
}
