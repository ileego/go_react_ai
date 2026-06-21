package handler

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// testUserID 测试用户 ID。
const testUserID int64 = 1

// testAuthMiddleware 测试用认证中间件，直接设置 user_id 和 role。
func testAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Set("user_role", "user")
		c.Next()
	}
}

// newTestRouter 创建带测试中间件的 Gin 引擎。
func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(testAuthMiddleware())
	return r
}

// doRequest 发送测试请求并返回响应。
func doRequest(t *testing.T, r *gin.Engine, method, path, token string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
