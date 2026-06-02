// Package docs 提供 Swagger API 文档服务。
// OpenAPI 规范文件放在此包中，通过 go:embed 嵌入到二进制。
package docs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var openapiYAML []byte

//go:embed swagger-ui.html
var swaggerUIHTML []byte

// RegisterRoutes 注册 Swagger 文档路由到 Gin Engine
func RegisterRoutes(r *gin.Engine) {
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
