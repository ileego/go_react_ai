package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/docs"
	"github.com/ileego/go_react_ai/internal/auth"
	"github.com/ileego/go_react_ai/internal/cache"
	"github.com/ileego/go_react_ai/internal/config"
	"github.com/ileego/go_react_ai/internal/handler"
	"github.com/ileego/go_react_ai/internal/middleware"
	"github.com/ileego/go_react_ai/internal/repository/postgres"
	"github.com/ileego/go_react_ai/internal/repository/redis"
	"github.com/ileego/go_react_ai/internal/security"
	"github.com/ileego/go_react_ai/internal/service"
	"github.com/ileego/go_react_ai/internal/storage"
	"github.com/ileego/go_react_ai/pkg/httpx"
	"github.com/ileego/go_react_ai/pkg/worker"
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
	defer func() { _ = db.Close() }()

	if err := postgres.MigrateUp(cfg.Database); err != nil {
		slog.Error("failed to migrate database", slog.Any("error", err))
		os.Exit(1)
	}

	// 初始化 Redis
	redisClient := redis.New(cfg.Redis)
	defer func() { _ = redisClient.Close() }()

	// 缓存层
	cacheMgr := cache.NewRedisCache(redisClient.Client, cache.Config{
		DefaultTTL: cfg.Cache.DefaultTTL(),
		Prefix:     cfg.Cache.Prefix,
	})

	// 安全组件
	rateLimiter := security.NewRedisRateLimiter(redisClient.Client, security.IPLimitConfig{
		Limit:  100,
		Window: time.Minute,
	}, security.LoginLimitConfig{
		MaxAttempts: 5,
		Window:      15 * time.Minute,
		Lockout:     15 * time.Minute,
	})
	blacklist := security.NewRedisTokenBlacklist(redisClient.Client)

	// GitHub OAuth2
	oauthCfg := service.NewGithubOAuthConfig(
		cfg.Auth.GithubClientID,
		cfg.Auth.GithubClientSecret,
		cfg.Auth.GithubRedirectURL,
	)
	var oauthState handler.OAuthStateStore
	if oauthCfg != nil {
		oauthState = handler.NewRedisOAuthStateStore(redisClient.Client)
	}

	// JWT 配置
	jwtCfg := auth.Config{
		Secret:          cfg.Auth.JWTSecret,
		AccessTokenTTL:  cfg.Auth.AccessTokenTTL(),
		RefreshTokenTTL: cfg.Auth.RefreshTokenTTL(),
		Issuer:          "goai",
	}

	// Worker Pool：有界并发与背压处理
	workerPool := worker.NewPool(cfg.WorkerPool.Workers, cfg.WorkerPool.QueueSize)
	workerPool.Start()

	// HTTP 客户端：带超时、重试与降级
	retryCfg := httpx.DefaultRetryConfig()
	retryCfg.MaxRetries = cfg.AI.RetryCount()
	aiHTTPClient := httpx.NewClientWithFallback(
		cfg.AI.APITimeout(),
		retryCfg,
		aiFallback(cfg.AI.Provider),
	)

	// 文件存储：优先 MinIO，连接失败时降级到本地文件系统
	fileStorage := newFileStorage(cfg.MinIO)

	// 依赖注入
	userRepo := postgres.NewUserRepository(db.DB)
	reportRepo := postgres.NewReportRepository(db.DB)
	taskRepo := postgres.NewAgentTaskRepository(db.DB)
	searchRepo := postgres.NewSearchRepository(db.DB)
	fileRepo := postgres.NewFileRepository(db.DB)

	reportSvc := service.NewReportService(reportRepo, cacheMgr)
	agentSvc := service.NewAgentService(reportSvc, taskRepo, workerPool, aiHTTPClient, cfg.AI.BaseURL+"/api/mock-ai")
	authSvc := service.NewAuthService(userRepo, jwtCfg, oauthCfg, rateLimiter, blacklist)
	searchSvc := service.NewSearchService(searchRepo)
	fileSvc := service.NewFileService(fileRepo, fileStorage, cfg.MinIO.Bucket, 0, nil)

	handlers := handler.NewHandlers(
		authSvc,
		reportSvc,
		agentSvc,
		searchSvc,
		fileSvc,
		oauthCfg,
		oauthState,
		cfg.Auth.JWTSecret,
		blacklist,
		rateLimiter,
		db.HealthCheck,
	)

	// 创建 Gin 引擎
	r := gin.New()

	// 中间件顺序很重要：Recovery 必须在最外层，才能捕获后续中间件的 panic
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.Server.AllowOrigins))
	r.Use(middleware.RateLimit(rateLimiter, 100, time.Minute))

	// 注册业务路由
	handlers.RegisterRoutes(r)

	// Swagger API 文档
	docs.RegisterRoutes(r)

	// HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	// 优雅启停：监听系统信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	slog.Info("server starting", slog.String("addr", srv.Addr), slog.String("mode", cfg.Server.Mode))

	// 等待退出信号
	<-ctx.Done()
	slog.Info("server shutting down")

	// 优雅关闭顺序：先停 Worker Pool，再关 HTTP server，最后释放基础资源
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workerPool.Stop()
	_ = srv.Shutdown(shutdownCtx)

	slog.Info("server stopped")
}

// newFileStorage 根据配置创建 MinIO 或本地存储。
func newFileStorage(cfg config.MinIOConfig) storage.FileStorage {
	if cfg.Endpoint != "" {
		store, err := storage.NewMinIOStorage(cfg.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.Bucket, cfg.UseSSL)
		if err == nil {
			slog.Info("using MinIO storage", "endpoint", cfg.Endpoint, "bucket", cfg.Bucket)
			return store
		}
		slog.Warn("failed to connect MinIO, falling back to local storage", "error", err)
	}
	store, err := storage.NewLocalStorage("data/uploads")
	if err != nil {
		slog.Error("failed to create local storage", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("using local file storage")
	return store
}

// aiFallback 返回 AI 调用失败时的兜底内容。
// 第 21 章会替换为真实 Provider 的降级策略。
func aiFallback(provider string) httpx.FallbackFunc {
	return func(err error, resp *http.Response) ([]byte, error) {
		slog.Warn("AI call failed, using fallback content", "error", err)
		content := "## 研究报告（降级内容）\n\n由于 AI 服务暂时不可用，本次生成使用兜底内容。请稍后重试或检查 API 配置。\n\n### 提供商\n" + provider
		result := map[string]string{"output": content}
		return json.Marshal(result)
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
