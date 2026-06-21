package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorage 是本地文件系统存储实现，主要用于开发和测试。
type LocalStorage struct {
	basePath string
}

// NewLocalStorage 创建本地存储实例，并确保基础目录存在。
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &LocalStorage{basePath: basePath}, nil
}

// Put 保存文件到本地磁盘。
func (s *LocalStorage) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	path := s.path(key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// Get 读取本地文件。
func (s *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	f, err := os.Open(s.path(key))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("open file: %w", err)
	}
	return f, nil
}

// Delete 删除本地文件。
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	if err := os.Remove(s.path(key)); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("remove file: %w", err)
	}
	return nil
}

// PresignedGetURL 本地存储不支持真正的预签名 URL，返回相对路径仅供开发调试。
func (s *LocalStorage) PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return fmt.Sprintf("/local-files/%s", key), nil
}

// PresignedPutURL 本地存储不支持客户端直传。
func (s *LocalStorage) PresignedPutURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "", fmt.Errorf("presigned put not supported in local storage")
}

func (s *LocalStorage) path(key string) string {
	return filepath.Join(s.basePath, key)
}
