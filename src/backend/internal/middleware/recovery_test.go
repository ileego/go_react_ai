package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecovery_Panic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))

	r := gin.New()
	r.Use(Recovery())
	r.Use(RequestID())
	r.Use(Logger())
	r.GET("/panic", func(c *gin.Context) {
		panic("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set(HeaderXRequestID, "req-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["code"] != float64(http.StatusInternalServerError) {
		t.Errorf("code = %v, want %d", body["code"], http.StatusInternalServerError)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "panic recovered") {
		t.Errorf("log should contain 'panic recovered', got %q", logOutput)
	}
	if !strings.Contains(logOutput, "req-123") {
		t.Errorf("log should contain request_id req-123, got %q", logOutput)
	}
}
