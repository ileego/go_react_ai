package service

import (
	"context"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
)

// searchService 实现 SearchService 接口。
type searchService struct {
	repo repository.SearchRepository
}

// NewSearchService 创建 SearchService 实例。
func NewSearchService(repo repository.SearchRepository) SearchService {
	return &searchService{repo: repo}
}

// SearchReports 搜索当前用户的报告。
func (s *searchService) SearchReports(ctx context.Context, userID int64, query string, page, pageSize int) ([]*domain.Report, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.repo.SearchReports(ctx, userID, query, limit, offset)
}
