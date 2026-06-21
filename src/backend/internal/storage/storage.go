// Package storage 提供文件存储抽象。
// 支持 MinIO（生产）与本地文件系统（开发/测试）两种实现。
package storage

import (
	"context"
	"errors"
	"io"
	"slices"
	"time"
)

// ErrNotFound 表示文件不存在。
var ErrNotFound = errors.New("file not found")

// FileStorage 定义文件存储接口。
type FileStorage interface {
	// Put 上传文件。key 为存储层唯一标识。
	Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	// Get 下载文件。
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete 删除文件。
	Delete(ctx context.Context, key string) error
	// PresignedGetURL 返回临时下载 URL。
	PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	// PresignedPutURL 返回临时上传 URL（客户端直传）。
	PresignedPutURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// IsAllowedContentType 判断 MIME 类型是否在白名单内。
func IsAllowedContentType(contentType string, allowed []string) bool {
	return slices.Contains(allowed, contentType)
}

// DefaultAllowedContentTypes 默认允许上传的文件类型。
var DefaultAllowedContentTypes = []string{
	"application/pdf",
	"application/msword",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"application/vnd.ms-excel",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"text/plain",
	"text/markdown",
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
}

// DefaultMaxFileSize 默认最大文件大小（10 MB）。
const DefaultMaxFileSize = 10 * 1024 * 1024
