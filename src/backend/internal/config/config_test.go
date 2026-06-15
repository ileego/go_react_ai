package config

import (
	"os"
	"strings"
	"testing"
)

func TestDatabaseConfig_DSN(t *testing.T) {
	d := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "goai",
		Password: "secret",
		Name:     "goai",
	}
	want := "postgres://goai:secret@localhost:5432/goai?sslmode=disable"
	if got := d.DSN(); got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestLoad_Defaults(t *testing.T) {
	// 清理环境变量，避免外部配置干扰测试
	for _, key := range []string{
		"GOAI_SERVER_PORT", "GOAI_SERVER_MODE",
		"GOAI_LOG_LEVEL", "GOAI_LOG_FORMAT",
		"GOAI_DB_HOST", "GOAI_DB_PORT",
		"GOAI_REDIS_HOST", "GOAI_REDIS_PORT",
		"GOAI_AI_PROVIDER",
	} {
		os.Unsetenv(key)
	}

	cfg := Load()

	if cfg.Server.Port != "8080" {
		t.Errorf("default port = %q, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("default mode = %q, want debug", cfg.Server.Mode)
	}
	if cfg.Server.LogLevel != "info" {
		t.Errorf("default log level = %q, want info", cfg.Server.LogLevel)
	}
	if cfg.Server.LogFormat != "json" {
		t.Errorf("default log format = %q, want json", cfg.Server.LogFormat)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("default db host = %q, want localhost", cfg.Database.Host)
	}
	if cfg.Redis.Port != "6379" {
		t.Errorf("default redis port = %q, want 6379", cfg.Redis.Port)
	}
	if cfg.AI.Provider != "openai" {
		t.Errorf("default ai provider = %q, want openai", cfg.AI.Provider)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("GOAI_SERVER_PORT", "9090")
	os.Setenv("GOAI_SERVER_MODE", "release")
	os.Setenv("GOAI_LOG_LEVEL", "debug")
	os.Setenv("GOAI_LOG_FORMAT", "text")
	os.Setenv("GOAI_DB_HOST", "db.example.com")
	os.Setenv("GOAI_AI_PROVIDER", "deepseek")
	defer func() {
		os.Unsetenv("GOAI_SERVER_PORT")
		os.Unsetenv("GOAI_SERVER_MODE")
		os.Unsetenv("GOAI_LOG_LEVEL")
		os.Unsetenv("GOAI_LOG_FORMAT")
		os.Unsetenv("GOAI_DB_HOST")
		os.Unsetenv("GOAI_AI_PROVIDER")
	}()

	cfg := Load()

	if cfg.Server.Port != "9090" {
		t.Errorf("port = %q, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("mode = %q, want release", cfg.Server.Mode)
	}
	if cfg.Server.LogLevel != "debug" {
		t.Errorf("log level = %q, want debug", cfg.Server.LogLevel)
	}
	if cfg.Server.LogFormat != "text" {
		t.Errorf("log format = %q, want text", cfg.Server.LogFormat)
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("db host = %q, want db.example.com", cfg.Database.Host)
	}
	if cfg.AI.Provider != "deepseek" {
		t.Errorf("ai provider = %q, want deepseek", cfg.AI.Provider)
	}
}

func TestLoad_DSNWithSpecialChars(t *testing.T) {
	os.Setenv("GOAI_DB_USER", "user")
	os.Setenv("GOAI_DB_PASSWORD", "p@ss:w0rd!")
	os.Setenv("GOAI_DB_HOST", "host")
	os.Setenv("GOAI_DB_PORT", "5432")
	os.Setenv("GOAI_DB_NAME", "db")
	defer func() {
		os.Unsetenv("GOAI_DB_USER")
		os.Unsetenv("GOAI_DB_PASSWORD")
		os.Unsetenv("GOAI_DB_HOST")
		os.Unsetenv("GOAI_DB_PORT")
		os.Unsetenv("GOAI_DB_NAME")
	}()

	cfg := Load()
	dsn := cfg.Database.DSN()
	if !strings.Contains(dsn, "p@ss:w0rd!") {
		t.Errorf("DSN should contain raw password; got %q", dsn)
	}
}
