package main

import (
	"log"
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

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化结构化日志
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 连接 PostgreSQL 并自动执行迁移
	db, err := postgres.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()

	if err := postgres.MigrateUp(cfg.Database); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 依赖注入（报告使用 PostgreSQL，其余模块暂时用内存实现）
	reportRepo := postgres.NewReportRepository(db.DB)
	reportSvc := service.NewReportService(reportRepo)
	// TODO: taskRepo 在实现 AgentTaskRepository 后注入真实实现
	agentSvc := service.NewAgentService(reportRepo, nil)

	handlers := handler.NewHandlers(reportSvc, agentSvc)

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
	log.Printf("Server starting on %s (mode=%s)", addr, cfg.Server.Mode)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
