package handler

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed docs/openapi.yaml
var openapiYAML []byte

//go:embed docs/swagger-ui.html
var swaggerUIHTML []byte

// SwaggerHandler Swagger 文档处理器
type SwaggerHandler struct{}

// NewSwaggerHandler 创建 SwaggerHandler
func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

// RegisterRoutes 注册 Swagger 相关路由
func (h *SwaggerHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	r.GET("/swagger/index.html", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerUIHTML)
	})
	r.GET("/swagger/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", openapiYAML)
	})
}
