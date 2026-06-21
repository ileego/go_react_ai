package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ileego/go_react_ai/internal/domain"
)

// SearchRepository PostgreSQL 全文搜索实现。
type SearchRepository struct {
	db *sql.DB
}

// NewSearchRepository 创建 PostgreSQL 版 SearchRepository。
func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// SearchReports 根据用户 ID 和查询词搜索报告。
func (r *SearchRepository) SearchReports(ctx context.Context, userID int64, query string, limit, offset int) ([]*domain.Report, int, error) {
	if query == "" {
		return nil, 0, nil
	}

	var total int
	if err := r.db.QueryRowContext(ctx, querySearchReportsCount, userID, query).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count search reports: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	rows, err := r.db.QueryContext(ctx, querySearchReports, userID, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("search reports: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var reports []*domain.Report
	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			return nil, 0, err
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate search reports: %w", err)
	}
	return reports, total, nil
}

const (
	querySearchReports = `
		SELECT id, title, topic, status, content, sources, created_by, created_at, updated_at, completed_at
		FROM reports
		WHERE created_by = $1
		  AND search_vector @@ plainto_tsquery('simple', $2)
		ORDER BY ts_rank(search_vector, plainto_tsquery('simple', $2)) DESC, created_at DESC
		LIMIT $3 OFFSET $4
	`

	querySearchReportsCount = `
		SELECT COUNT(*)
		FROM reports
		WHERE created_by = $1
		  AND search_vector @@ plainto_tsquery('simple', $2)
	`
)
