package memory

import (
	"context"
	"strings"

	"github.com/ileego/go_react_ai/internal/domain"
)

// SearchRepository 内存版搜索实现，主要用于单元测试。
type SearchRepository struct {
	reportRepo *ReportRepository
}

// NewSearchRepository 创建内存版 SearchRepository。
func NewSearchRepository(reportRepo *ReportRepository) *SearchRepository {
	return &SearchRepository{reportRepo: reportRepo}
}

// SearchReports 在内存中按标题、主题、内容关键字搜索报告。
func (r *SearchRepository) SearchReports(ctx context.Context, userID int64, query string, limit, offset int) ([]*domain.Report, int, error) {
	if query == "" {
		return nil, 0, nil
	}

	reports := r.reportRepo.reports
	var matched []*domain.Report
	keywords := strings.Fields(strings.ToLower(query))

	for _, report := range reports {
		if report.CreatedBy != userID {
			continue
		}
		text := strings.ToLower(report.Title + " " + report.Topic + " " + report.Content)
		if containsAll(text, keywords) {
			matched = append(matched, report)
		}
	}

	total := len(matched)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return matched[offset:end], total, nil
}

func containsAll(text string, keywords []string) bool {
	for _, kw := range keywords {
		if !strings.Contains(text, kw) {
			return false
		}
	}
	return true
}
