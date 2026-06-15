package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type healthResp struct {
	Code    int            `json:"code"`
	Data    map[string]any `json:"data"`
	Message string         `json:"message"`
}

func TestHealthHandler_Check(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHealthHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/health", nil)
	h.Check(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body healthResp
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Data["status"] != "ok" {
		t.Errorf("status = %v, want ok", body.Data["status"])
	}
}

func TestHealthHandler_Ready_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbHealth := func(ctx context.Context) error { return nil }
	h := NewHealthHandler(dbHealth)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/ready", nil)
	h.Ready(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body healthResp
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Data["status"] != "ready" {
		t.Errorf("status = %v, want ready", body.Data["status"])
	}
}

func TestHealthHandler_Ready_Fail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbHealth := func(ctx context.Context) error { return errors.New("connection refused") }
	h := NewHealthHandler(dbHealth)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/ready", nil)
	h.Ready(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["status"] != "not_ready" {
		t.Errorf("status = %v, want not_ready", body["status"])
	}
	if body["reason"] != "database_unavailable" {
		t.Errorf("reason = %v, want database_unavailable", body["reason"])
	}
}
