// Package handler 定义 HTTP 接口层。
// 负责：参数绑定、请求校验、调用 Service、构造响应。
// 不包含任何业务逻辑。
package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/middleware"
	"github.com/ileego/go_react_ai/internal/security"
	"github.com/ileego/go_react_ai/internal/service"
	"golang.org/x/oauth2"
)

// Handlers 汇总所有 HTTP 处理器，方便路由注册时统一注入依赖
type Handlers struct {
	Auth      *AuthHandler
	Report    *ReportHandler
	Agent     *AgentHandler
	Health    *HealthHandler
	jwtSecret string
	blacklist security.TokenBlacklist
	rl        security.RateLimiter
}

// NewHandlers 创建处理器实例，注入 Service 依赖
func NewHandlers(
	authSvc service.AuthService,
	reportSvc service.ReportService,
	agentSvc service.AgentService,
	oauthCfg *oauth2.Config,
	oauthState OAuthStateStore,
	jwtSecret string,
	blacklist security.TokenBlacklist,
	rl security.RateLimiter,
	dbHealth func(context.Context) error,
) *Handlers {
	return &Handlers{
		Auth:      NewAuthHandler(authSvc, oauthCfg, oauthState, rl),
		Report:    NewReportHandler(reportSvc),
		Agent:     NewAgentHandler(agentSvc),
		Health:    NewHealthHandler(dbHealth),
		jwtSecret: jwtSecret,
		blacklist: blacklist,
		rl:        rl,
	}
}

// RegisterRoutes 将所有路由注册到 Gin Engine
func (h *Handlers) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/health", h.Health.Check)
		api.GET("/ready", h.Health.Ready)

		// 公开认证接口
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", h.Auth.Register)
			authGroup.POST("/login", middleware.LoginRateLimit(h.rl), h.Auth.Login)
			authGroup.POST("/refresh", h.Auth.Refresh)
			authGroup.POST("/logout", h.Auth.Logout)
			authGroup.GET("/me", middleware.JWTAuth(h.jwtSecret, h.blacklist), h.Auth.Me)
			authGroup.GET("/github/login", h.Auth.GithubLogin)
			authGroup.GET("/github/callback", h.Auth.GithubCallback)
		}

		// 受保护接口
		reports := api.Group("/reports")
		reports.Use(middleware.JWTAuth(h.jwtSecret, h.blacklist))
		reports.Use(middleware.RateLimit(h.rl, 100, time.Minute))
		{
			reports.POST("", h.Report.Create)
			reports.GET("", h.Report.List)
			reports.GET("/:id", h.Report.Get)
			reports.POST("/:id/cancel", h.Report.Cancel)
			reports.POST("/:report_id/dispatch", middleware.RequireRole("admin", "system"), h.Agent.Dispatch)
		}
	}
}
