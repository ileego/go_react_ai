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
	"github.com/ileego/go_react_ai/internal/repository/memory"
	"github.com/ileego/go_react_ai/internal/service"
)

func main() {
	cfg := config.Load()

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化结构化日志
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 依赖注入（目前使用内存实现，第7章替换为 PostgreSQL）
	reportRepo := memory.NewReportRepository()
	// TODO: taskRepo 在实现 AgentTaskRepository 内存版后注入
	// taskRepo := memory.NewAgentTaskRepository()

	reportSvc := service.NewReportService(reportRepo)
	// TODO: agentSvc 在 taskRepo 就绪后注入真实实现
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
