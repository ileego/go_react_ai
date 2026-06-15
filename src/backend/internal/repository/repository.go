// Package repository 定义数据访问层接口。
// 接口放在这里，实现放在同包的子文件中。Service 层只依赖这些接口。
package repository

import (
	"context"
	"errors"

	"github.com/ileego/go_react_ai/internal/domain"
)

// ErrNotFound 是 Repository 层统一的"资源不存在"错误
// Service 层可以通过 errors.Is(err, repository.ErrNotFound) 判断
var ErrNotFound = errors.New("record not found")

// ErrDuplicate 是 Repository 层统一的"资源重复"错误
var ErrDuplicate = errors.New("record already exists")

// UserRepository 用户数据访问接口
type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByEmailWithPassword(ctx context.Context, email string) (*domain.User, string, error)
	GetByGithubID(ctx context.Context, githubID string) (*domain.User, error)
	CreateWithPassword(ctx context.Context, user *domain.User, passwordHash string) error
	CreateWithGithub(ctx context.Context, user *domain.User, githubID string) error
	Update(ctx context.Context, user *domain.User) error
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdateRole(ctx context.Context, id int64, role string) error
	UpdateGithubInfo(ctx context.Context, id int64, githubID, avatarURL string) error
}

// ReportRepository 报告数据访问接口
type ReportRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Report, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*domain.Report, error)
	Create(ctx context.Context, report *domain.Report) error
	UpdateStatus(ctx context.Context, id int64, status domain.ReportStatus) error
}

// AgentTaskRepository 智能体任务数据访问接口
type AgentTaskRepository interface {
	GetByReportID(ctx context.Context, reportID int64) ([]*domain.AgentTask, error)
	Create(ctx context.Context, task *domain.AgentTask) error
	UpdateResult(ctx context.Context, id int64, output string, costMs int64) error
}
