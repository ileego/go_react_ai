package postgres

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ileego/go_react_ai/internal/config"
)

// MigrateUp 执行数据库迁移（升级到最新版本）
func MigrateUp(cfg config.DatabaseConfig) error {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/repository/migrations",
		"pgx", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// MigrateDown 回滚所有迁移（主要用于测试清理）
func MigrateDown(cfg config.DatabaseConfig) error {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/repository/migrations",
		"pgx", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}
