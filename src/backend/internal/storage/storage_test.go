package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalStorage_PutGetDelete(t *testing.T) {
	dir := t.TempDir()
	s, err := NewLocalStorage(dir)
	if err != nil {
		t.Fatalf("new local storage: %v", err)
	}

	ctx := context.Background()
	key := "reports/1/attachment.txt"
	content := []byte("hello storage")

	if err := s.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/plain"); err != nil {
		t.Fatalf("put failed: %v", err)
	}

	rc, err := s.Get(ctx, key)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	got, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("content mismatch")
	}

	if err := s.Delete(ctx, key); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := s.Get(ctx, key); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalStorage_PresignedGetURL(t *testing.T) {
	dir := t.TempDir()
	s, err := NewLocalStorage(dir)
	if err != nil {
		t.Fatalf("new local storage: %v", err)
	}

	url, err := s.PresignedGetURL(context.Background(), "reports/1/file.pdf", 15*time.Minute)
	if err != nil {
		t.Fatalf("presigned get: %v", err)
	}
	if !strings.HasPrefix(url, "/local-files/") {
		t.Errorf("unexpected url: %s", url)
	}
}

func TestLocalStorage_PresignedPutURL(t *testing.T) {
	dir := t.TempDir()
	s, err := NewLocalStorage(dir)
	if err != nil {
		t.Fatalf("new local storage: %v", err)
	}

	_, err = s.PresignedPutURL(context.Background(), "reports/1/file.pdf", 15*time.Minute)
	if err == nil {
		t.Error("expected error for presigned put in local storage")
	}
}

func TestIsAllowedContentType(t *testing.T) {
	allowed := []string{"image/png", "application/pdf"}
	if !IsAllowedContentType("image/png", allowed) {
		t.Error("expected allowed")
	}
	if IsAllowedContentType("application/zip", allowed) {
		t.Error("expected not allowed")
	}
}

func TestLocalStorage_NestedKey(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewLocalStorage(dir)
	ctx := context.Background()
	key := "a/b/c/d.txt"
	content := []byte("nested")

	_ = s.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/plain")
	fullPath := filepath.Join(dir, key)
	if _, err := os.Stat(fullPath); err != nil {
		t.Errorf("file not created at %s: %v", fullPath, err)
	}
}
