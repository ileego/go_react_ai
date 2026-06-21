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
	// UpdateStatus 更新报告状态（内部使用，会失效缓存）
	UpdateStatus(ctx context.Context, id int64, status domain.ReportStatus) error
}

// AgentService 智能体调度业务逻辑接口
type AgentService interface {
	// Dispatch 派发报告生成任务给智能体流水线
	Dispatch(ctx context.Context, reportID int64) error
}

// SearchService 搜索业务逻辑接口
type SearchService interface {
	// SearchReports 搜索当前用户的报告
	SearchReports(ctx context.Context, userID int64, query string, page, pageSize int) ([]*domain.Report, int, error)
}

// FileService 文件存储业务逻辑接口
type FileService interface {
	// Upload 上传文件并保存元数据
	Upload(ctx context.Context, userID int64, name, contentType string, data []byte) (*domain.File, error)
	// GetByID 获取文件元数据
	GetByID(ctx context.Context, id int64) (*domain.File, error)
	// GetDownloadURL 获取临时下载 URL
	GetDownloadURL(ctx context.Context, id int64) (string, error)
	// Delete 删除文件及其元数据
	Delete(ctx context.Context, id int64) error
	// ListByUser 获取用户的文件列表
	ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]*domain.File, int, error)
	// PresignedUploadURL 获取客户端直传的预签名上传 URL
	PresignedUploadURL(ctx context.Context, userID int64, name, contentType string) (string, *domain.File, error)
}
