package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go_react_ai/internal/config"
)

func main() {
	cfg := config.Load()

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	if cfg.Server.Mode == "debug" {
		r.Use(gin.Logger())
	}

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.Server.AllowOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"mode":   cfg.Server.Mode,
		})
	})

	addr := ":" + cfg.Server.Port
	log.Printf("Server starting on %s (mode=%s)", addr, cfg.Server.Mode)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
