// Package service 定义业务逻辑层接口。
// Handler 层调用这些接口，不需要关心底层实现细节。
package service

import (
	"context"

	"github.com/ileego/go_react_ai/internal/domain"
)

// AuthService 认证与授权业务逻辑接口
type AuthService interface {
	// Register 用户注册
	Register(ctx context.Context, email, password, nickname string) (*domain.User, error)
	// Login 用户登录，返回 access token 与 refresh token
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	// Refresh 使用 refresh token 刷新令牌对
	Refresh(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error)
	// Logout 登出，将当前 access 与 refresh token 加入黑名单
	Logout(ctx context.Context, accessToken, refreshToken string) error
	// Me 获取当前用户信息
	Me(ctx context.Context, userID int64) (*domain.User, error)
	// GithubLogin 使用 GitHub OAuth2 code 登录或注册
	GithubLogin(ctx context.Context, code string) (accessToken, refreshToken string, err error)
}

// ReportService 报告业务逻辑接口
type ReportService interface {
	// Create 创建新的研究报告
	Create(ctx context.Context, userID int64, title, topic string) (*domain.Report, error)
	// GetByID 获取报告详情
	GetByID(ctx context.Context, id int64) (*domain.Report, error)
	// ListByUser 获取用户的报告列表
	ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]*domain.Report, int, error)
	// Cancel 取消正在进行的报告
	Cancel(ctx context.Context, id int64) error
}

// AgentService 智能体调度业务逻辑接口
type AgentService interface {
	// Dispatch 派发报告生成任务给智能体流水线
	Dispatch(ctx context.Context, reportID int64) error
}
