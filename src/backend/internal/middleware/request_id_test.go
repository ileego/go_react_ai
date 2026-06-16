package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestID_FromHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, GetRequestID(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(HeaderXRequestID, "req-abc")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Body.String(); got != "req-abc" {
		t.Errorf("body = %q, want req-abc", got)
	}
	if got := w.Header().Get(HeaderXRequestID); got != "req-abc" {
		t.Errorf("response header %s = %q, want req-abc", HeaderXRequestID, got)
	}
}

func TestRequestID_Generated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, GetRequestID(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Body.String(); got == "" || got == "unknown" {
		t.Errorf("generated request_id = %q, want non-empty", got)
	}
}

func TestRequestID_InContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var fromCtx string
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		fromCtx = GetRequestIDFromContext(c.Request.Context())
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(HeaderXRequestID, "req-ctx")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if fromCtx != "req-ctx" {
		t.Errorf("request_id from context = %q, want req-ctx", fromCtx)
	}
}

func TestGetRequestIDFromContext_Empty(t *testing.T) {
	if got := GetRequestIDFromContext(context.Background()); got != "" {
		t.Errorf("empty context request_id = %q, want empty", got)
	}
}
