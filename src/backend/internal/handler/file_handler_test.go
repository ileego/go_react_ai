package handler

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/repository/memory"
	"github.com/ileego/go_react_ai/internal/service"
	"github.com/ileego/go_react_ai/internal/storage"
)

func newTestFileHandler(t *testing.T) (*FileHandler, *gin.Engine) {
	t.Helper()
	repo := memory.NewFileRepository()
	store, err := storage.NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}
	svc := service.NewFileService(repo, store, "test-bucket", 0, nil)
	h := NewFileHandler(svc)

	r := newTestRouter()
	r.POST("/api/files", h.Upload)
	r.GET("/api/files", h.List)
	r.GET("/api/files/:id", h.Get)
	r.GET("/api/files/:id/download", h.Download)
	r.DELETE("/api/files/:id", h.Delete)
	return h, r
}

func TestFileHandler_Upload(t *testing.T) {
	repo := memory.NewFileRepository()
	store, _ := storage.NewLocalStorage(t.TempDir())
	svc := service.NewFileService(repo, store, "test-bucket", 0, nil)
	h := NewFileHandler(svc)
	r := newTestRouter()
	r.POST("/api/files", h.Upload)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	// 显式设置 Content-Type，避免跨平台 MIME 识别差异
	part, _ := writer.CreatePart(map[string][]string{
		"Content-Disposition": {"form-data; name=\"file\"; filename=\"report.txt\""},
		"Content-Type":        {"text/plain"},
	})
	_, _ = part.Write([]byte("text content"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/files", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestFileHandler_Upload_TooLarge(t *testing.T) {
	repo := memory.NewFileRepository()
	store, _ := storage.NewLocalStorage(t.TempDir())
	svc := service.NewFileService(repo, store, "test-bucket", 1024, nil)
	h := NewFileHandler(svc)
	r := newTestRouter()
	r.POST("/api/files", h.Upload)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreatePart(map[string][]string{
		"Content-Disposition": {"form-data; name=\"file\"; filename=\"report.txt\""},
		"Content-Type":        {"text/plain"},
	})
	_, _ = part.Write(make([]byte, 2048))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/files", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestFileHandler_List(t *testing.T) {
	_, r := newTestFileHandler(t)
	ctx := t.Context()

	repo := memory.NewFileRepository()
	store, _ := storage.NewLocalStorage(t.TempDir())
	svc := service.NewFileService(repo, store, "test-bucket", 0, nil)
	_, _ = svc.Upload(ctx, testUserID, "a.pdf", "application/pdf", []byte("a"))

	w := doRequest(t, r, http.MethodGet, "/api/files", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestFileHandler_Get(t *testing.T) {
	ctx := t.Context()
	repo := memory.NewFileRepository()
	store, _ := storage.NewLocalStorage(t.TempDir())
	svc := service.NewFileService(repo, store, "test-bucket", 0, nil)
	file, _ := svc.Upload(ctx, testUserID, "report.pdf", "application/pdf", []byte("content"))

	h := NewFileHandler(svc)
	r := newTestRouter()
	r.GET("/api/files/:id", h.Get)

	w := doRequest(t, r, http.MethodGet, fmt.Sprintf("/api/files/%d", file.ID), "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestFileHandler_Download(t *testing.T) {
	ctx := t.Context()
	repo := memory.NewFileRepository()
	store, _ := storage.NewLocalStorage(t.TempDir())
	svc := service.NewFileService(repo, store, "test-bucket", 0, nil)
	file, _ := svc.Upload(ctx, testUserID, "report.pdf", "application/pdf", []byte("content"))

	h := NewFileHandler(svc)
	r := newTestRouter()
	r.GET("/api/files/:id/download", h.Download)

	w := doRequest(t, r, http.MethodGet, fmt.Sprintf("/api/files/%d/download", file.ID), "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestFileHandler_Delete(t *testing.T) {
	ctx := t.Context()
	repo := memory.NewFileRepository()
	store, _ := storage.NewLocalStorage(t.TempDir())
	svc := service.NewFileService(repo, store, "test-bucket", 0, nil)
	file, _ := svc.Upload(ctx, testUserID, "report.pdf", "application/pdf", []byte("content"))

	h := NewFileHandler(svc)
	r := newTestRouter()
	r.DELETE("/api/files/:id", h.Delete)

	w := doRequest(t, r, http.MethodDelete, fmt.Sprintf("/api/files/%d", file.ID), "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}
}
