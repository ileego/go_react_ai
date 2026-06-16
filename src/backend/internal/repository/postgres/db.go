// Package postgres 提供基于 PostgreSQL 的 Repository 实现和数据库连接管理。
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ileego/go_react_ai/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB 封装数据库连接，提供连接池和迁移支持
type DB struct {
	*sql.DB
}

// New 使用配置创建 PostgreSQL 连接
func New(cfg config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 连接池配置：从配置文件读取，生产环境根据负载调整
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)

	// 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Close 关闭数据库连接
func (d *DB) Close() error {
	return d.DB.Close()
}

// HealthCheck 检查数据库连接是否正常
func (d *DB) HealthCheck(ctx context.Context) error {
	return d.PingContext(ctx)
}
