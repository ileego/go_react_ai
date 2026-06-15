package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
)

// ReportRepository PostgreSQL 版报告数据访问实现
type ReportRepository struct {
	db *sql.DB
}

// NewReportRepository 创建 PostgreSQL 版 ReportRepository
func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// GetByID 根据 ID 获取报告
func (r *ReportRepository) GetByID(ctx context.Context, id int64) (*domain.Report, error) {
	row := r.db.QueryRowContext(ctx, queryReportByID, id)
	report, err := scanReport(row)
	if err != nil {
		return nil, err
	}
	return report, nil
}

// ListByUser 获取用户的报告列表
func (r *ReportRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*domain.Report, error) {
	rows, err := r.db.QueryContext(ctx, queryReportsByUser, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	var reports []*domain.Report
	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports: %w", err)
	}
	return reports, nil
}

// Create 创建报告
func (r *ReportRepository) Create(ctx context.Context, report *domain.Report) error {
	now := time.Now()
	report.CreatedAt = now
	report.UpdatedAt = now
	report.Status = domain.ReportStatusPending

	return r.db.QueryRowContext(ctx, queryInsertReport,
		report.Title,
		report.Topic,
		report.Status,
		report.Content,
		formatPGArray(report.Sources),
		report.CreatedBy,
		report.CreatedAt,
		report.UpdatedAt,
		report.CompletedAt,
	).Scan(&report.ID)
}

// UpdateStatus 更新报告状态
func (r *ReportRepository) UpdateStatus(ctx context.Context, id int64, status domain.ReportStatus) error {
	result, err := r.db.ExecContext(ctx, queryUpdateReportStatus, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update report status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// reportScanner 统一扫描接口，兼容 *sql.Row 和 *sql.Rows
type reportScanner interface {
	Scan(dest ...any) error
}

func scanReport(scanner reportScanner) (*domain.Report, error) {
	var report domain.Report
	var completedAt sql.NullTime
	var sources string

	err := scanner.Scan(
		&report.ID,
		&report.Title,
		&report.Topic,
		&report.Status,
		&report.Content,
		&sources,
		&report.CreatedBy,
		&report.CreatedAt,
		&report.UpdatedAt,
		&completedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("scan report: %w", err)
	}

	report.Sources = parsePGArray(sources)
	if completedAt.Valid {
		report.CompletedAt = &completedAt.Time
	}
	return &report, nil
}

// parsePGArray 解析 PostgreSQL 数组字符串格式：{a,b,c}
func parsePGArray(s string) []string {
	if s == "" || s == "{}" {
		return nil
	}
	// 去掉首尾大括号
	s = strings.Trim(s, "{}")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.Trim(parts[i], "\"")
	}
	return parts
}

// formatPGArray 将 Go 字符串切片格式化为 PostgreSQL 数组字符串
func formatPGArray(items []string) string {
	if len(items) == 0 {
		return "{}"
	}
	var quoted []string
	for _, item := range items {
		quoted = append(quoted, fmt.Sprintf("\"%s\"", strings.ReplaceAll(item, "\"", "\\\"")))
	}
	return "{" + strings.Join(quoted, ",") + "}"
}

// SQL 常量：集中管理，避免在业务代码里拼接 SQL
const (
	queryReportByID = `
		SELECT id, title, topic, status, content, sources, created_by, created_at, updated_at, completed_at
		FROM reports
		WHERE id = $1
	`

	queryReportsByUser = `
		SELECT id, title, topic, status, content, sources, created_by, created_at, updated_at, completed_at
		FROM reports
		WHERE created_by = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	queryInsertReport = `
		INSERT INTO reports (title, topic, status, content, sources, created_by, created_at, updated_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	queryUpdateReportStatus = `
		UPDATE reports
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
)
