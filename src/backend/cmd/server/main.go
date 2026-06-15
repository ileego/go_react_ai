package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/docs"
	"github.com/ileego/go_react_ai/internal/config"
	"github.com/ileego/go_react_ai/internal/handler"
	"github.com/ileego/go_react_ai/internal/middleware"
	"github.com/ileego/go_react_ai/internal/repository/postgres"
	"github.com/ileego/go_react_ai/internal/service"
)

func main() {
	cfg := config.Load()

	// 初始化结构化日志：级别和格式均可通过环境变量配置
	initLogger(cfg.Server.LogLevel, cfg.Server.LogFormat)

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 连接 PostgreSQL 并自动执行迁移
	db, err := postgres.New(cfg.Database)
	if err != nil {
		slog.Error("failed to connect database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	if err := postgres.MigrateUp(cfg.Database); err != nil {
		slog.Error("failed to migrate database", slog.Any("error", err))
		os.Exit(1)
	}

	// 依赖注入（报告使用 PostgreSQL，其余模块暂时用内存实现）
	reportRepo := postgres.NewReportRepository(db.DB)
	reportSvc := service.NewReportService(reportRepo)
	// TODO: taskRepo 在实现 AgentTaskRepository 后注入真实实现
	agentSvc := service.NewAgentService(reportRepo, nil)

	handlers := handler.NewHandlers(reportSvc, agentSvc, db.HealthCheck)

	// 创建 Gin 引擎
	r := gin.New()

	// 中间件顺序很重要：Recovery 必须在最外层，才能捕获后续中间件的 panic
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.Server.AllowOrigins))

	// 注册业务路由
	handlers.RegisterRoutes(r)

	// Swagger API 文档
	docs.RegisterRoutes(r)

	addr := ":" + cfg.Server.Port
	slog.Info("server starting", slog.String("addr", addr), slog.String("mode", cfg.Server.Mode))
	if err := r.Run(addr); err != nil {
		slog.Error("server failed", slog.Any("error", err))
		os.Exit(1)
	}
}

// initLogger 根据配置初始化 slog，支持运行时调整日志级别
func initLogger(levelStr, format string) {
	var level slog.LevelVar
	if err := level.UnmarshalText([]byte(levelStr)); err != nil {
		level.Set(slog.LevelInfo)
	}

	opts := &slog.HandlerOptions{Level: &level}
	var handler slog.Handler
	switch format {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}
