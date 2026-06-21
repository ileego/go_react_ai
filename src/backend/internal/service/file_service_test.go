package service

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/ileego/go_react_ai/internal/repository/memory"
	"github.com/ileego/go_react_ai/internal/storage"
)

func newTestFileService(t *testing.T) (FileService, *memory.FileRepository) {
	t.Helper()
	repo := memory.NewFileRepository()
	store, err := storage.NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}
	svc := NewFileService(repo, store, "test-bucket", 0, nil)
	return svc, repo
}

func TestFileService_Upload(t *testing.T) {
	svc, _ := newTestFileService(t)
	ctx := context.Background()

	file, err := svc.Upload(ctx, 1, "report.pdf", "application/pdf", []byte("pdf content"))
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if file.ID == 0 {
		t.Error("file ID should be assigned")
	}
	if file.Size != int64(len("pdf content")) {
		t.Errorf("size = %d, want %d", file.Size, len("pdf content"))
	}
	if file.Bucket != "test-bucket" {
		t.Errorf("bucket = %s, want test-bucket", file.Bucket)
	}
}

func TestFileService_Upload_UnsupportedType(t *testing.T) {
	svc, _ := newTestFileService(t)
	ctx := context.Background()

	_, err := svc.Upload(ctx, 1, "file.exe", "application/x-msdownload", []byte("exe"))
	if err == nil {
		t.Error("expected error for unsupported content type")
	}
}

func TestFileService_Upload_TooLarge(t *testing.T) {
	svc, _ := newTestFileService(t)
	ctx := context.Background()

	data := make([]byte, 11*1024*1024)
	_, err := svc.Upload(ctx, 1, "large.pdf", "application/pdf", data)
	if err == nil {
		t.Error("expected error for too large file")
	}
}

func TestFileService_GetByID(t *testing.T) {
	svc, _ := newTestFileService(t)
	ctx := context.Background()

	uploaded, _ := svc.Upload(ctx, 1, "report.pdf", "application/pdf", []byte("content"))
	file, err := svc.GetByID(ctx, uploaded.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if file.ID != uploaded.ID {
		t.Errorf("ID mismatch")
	}

	_, err = svc.GetByID(ctx, 99999)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestFileService_GetDownloadURL(t *testing.T) {
	svc, _ := newTestFileService(t)
	ctx := context.Background()

	uploaded, _ := svc.Upload(ctx, 1, "report.pdf", "application/pdf", []byte("content"))
	url, err := svc.GetDownloadURL(ctx, uploaded.ID)
	if err != nil {
		t.Fatalf("get download url failed: %v", err)
	}
	if url == "" {
		t.Error("download url should not be empty")
	}
}

func TestFileService_Delete(t *testing.T) {
	svc, repo := newTestFileService(t)
	ctx := context.Background()

	uploaded, _ := svc.Upload(ctx, 1, "report.pdf", "application/pdf", []byte("content"))
	if err := svc.Delete(ctx, uploaded.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := repo.GetByID(ctx, uploaded.ID); err == nil {
		t.Error("file metadata should be deleted")
	}
}

func TestFileService_ListByUser(t *testing.T) {
	svc, _ := newTestFileService(t)
	ctx := context.Background()

	_, _ = svc.Upload(ctx, 1, "a.pdf", "application/pdf", []byte("a"))
	_, _ = svc.Upload(ctx, 1, "b.pdf", "application/pdf", []byte("b"))
	_, _ = svc.Upload(ctx, 2, "c.pdf", "application/pdf", []byte("c"))

	files, total, err := svc.ListByUser(ctx, 1, 1, 10)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(files) != 2 {
		t.Errorf("len = %d, want 2", len(files))
	}
}

func TestFileService_PresignedUploadURL(t *testing.T) {
	repo := memory.NewFileRepository()
	store := &mockPresignedStorage{local: mustLocalStorage(t)}
	svc := NewFileService(repo, store, "test-bucket", 0, nil)
	ctx := context.Background()

	url, file, err := svc.PresignedUploadURL(ctx, 1, "report.pdf", "application/pdf")
	if err != nil {
		t.Fatalf("presigned upload url failed: %v", err)
	}
	if url == "" {
		t.Error("upload url should not be empty")
	}
	if file.ID == 0 {
		t.Error("file metadata should be created")
	}
}

// mockPresignedStorage 包装本地存储，使 PresignedPutURL 返回固定 URL。
type mockPresignedStorage struct {
	local storage.FileStorage
}

func (m *mockPresignedStorage) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	return m.local.Put(ctx, key, reader, size, contentType)
}

func (m *mockPresignedStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return m.local.Get(ctx, key)
}

func (m *mockPresignedStorage) Delete(ctx context.Context, key string) error {
	return m.local.Delete(ctx, key)
}

func (m *mockPresignedStorage) PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return m.local.PresignedGetURL(ctx, key, expiry)
}

func (m *mockPresignedStorage) PresignedPutURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "http://localhost:9000/test-bucket/" + key + "?X-Amz-Algorithm=AWS4-HMAC-SHA256", nil
}

func mustLocalStorage(t *testing.T) storage.FileStorage {
	t.Helper()
	store, err := storage.NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}
	return store
}
