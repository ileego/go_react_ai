package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	MinIO    MinIOConfig
	AI       AIConfig
}

type ServerConfig struct {
	Port         string `mapstructure:"SERVER_PORT"`
	Mode         string `mapstructure:"SERVER_MODE"` // debug / release
	AllowOrigins string `mapstructure:"ALLOW_ORIGINS"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
}

type RedisConfig struct {
	Host string `mapstructure:"REDIS_HOST"`
	Port string `mapstructure:"REDIS_PORT"`
}

type MinIOConfig struct {
	Endpoint  string `mapstructure:"MINIO_ENDPOINT"`
	AccessKey string `mapstructure:"MINIO_ACCESS_KEY"`
	SecretKey string `mapstructure:"MINIO_SECRET_KEY"`
	Bucket    string `mapstructure:"MINIO_BUCKET"`
	UseSSL    bool   `mapstructure:"MINIO_USE_SSL"`
}

type AIConfig struct {
	Provider string `mapstructure:"AI_PROVIDER"` // openai / anthropic / deepseek / kimi
	APIKey   string `mapstructure:"AI_API_KEY"`
	BaseURL  string `mapstructure:"AI_BASE_URL"`
	Model    string `mapstructure:"AI_MODEL"`
}

func (d DatabaseConfig) DSN() string {
	return "postgres://" + d.User + ":" + d.Password +
		"@" + d.Host + ":" + d.Port + "/" + d.Name +
		"?sslmode=disable"
}

func Load() *Config {
	v := viper.New()

	// 自动读取环境变量
	v.AutomaticEnv()

	// 设置默认值
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("SERVER_MODE", "debug")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("MINIO_USE_SSL", false)
	v.SetDefault("AI_PROVIDER", "openai")

	// 环境变量前缀（可选，避免冲突）
	v.SetEnvPrefix("GOAI")
	// 让 viper 能读取带点的环境变量，比如 GOAI.DB_HOST
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("配置解析失败: %v", err)
	}

	return &cfg
}
