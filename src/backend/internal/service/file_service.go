package service

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
	"github.com/ileego/go_react_ai/internal/storage"
	apperrors "github.com/ileego/go_react_ai/pkg/errors"
)

// fileService 实现 FileService 接口。
type fileService struct {
	repo         repository.FileRepository
	storage      storage.FileStorage
	bucket       string
	maxSize      int64
	allowedTypes []string
}

// NewFileService 创建 FileService 实例。
func NewFileService(
	repo repository.FileRepository,
	store storage.FileStorage,
	bucket string,
	maxSize int64,
	allowedTypes []string,
) FileService {
	if maxSize <= 0 {
		maxSize = storage.DefaultMaxFileSize
	}
	if len(allowedTypes) == 0 {
		allowedTypes = storage.DefaultAllowedContentTypes
	}
	return &fileService{
		repo:         repo,
		storage:      store,
		bucket:       bucket,
		maxSize:      maxSize,
		allowedTypes: allowedTypes,
	}
}

// Upload 上传文件并保存元数据。
func (s *fileService) Upload(ctx context.Context, userID int64, name, contentType string, data []byte) (*domain.File, error) {
	if err := s.validateUpload(name, contentType, int64(len(data))); err != nil {
		return nil, err
	}

	file := &domain.File{
		Name:        name,
		StorageKey:  s.generateKey(name),
		ContentType: contentType,
		Size:        int64(len(data)),
		Bucket:      s.bucket,
		CreatedBy:   userID,
	}
	if err := file.Validate(); err != nil {
		return nil, apperrors.NewValidation("file", err.Error())
	}

	if err := s.storage.Put(ctx, file.StorageKey, bytes.NewReader(data), file.Size, contentType); err != nil {
		return nil, apperrors.NewInternal("上传文件失败", err)
	}

	if err := s.repo.Create(ctx, file); err != nil {
		// 元数据写入失败时，尝试清理已上传的文件
		_ = s.storage.Delete(ctx, file.StorageKey)
		return nil, apperrors.NewInternal("保存文件元数据失败", err)
	}
	return file, nil
}

// GetByID 获取文件元数据。
func (s *fileService) GetByID(ctx context.Context, id int64) (*domain.File, error) {
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, apperrors.NewNotFound("file", id)
		}
		return nil, apperrors.NewInternal("查询文件失败", err)
	}
	return file, nil
}

// GetDownloadURL 返回临时下载 URL。
func (s *fileService) GetDownloadURL(ctx context.Context, id int64) (string, error) {
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return "", apperrors.NewNotFound("file", id)
		}
		return "", apperrors.NewInternal("查询文件失败", err)
	}

	url, err := s.storage.PresignedGetURL(ctx, file.StorageKey, 15*time.Minute)
	if err != nil {
		return "", apperrors.NewInternal("生成下载链接失败", err)
	}
	return url, nil
}

// Delete 删除文件及其元数据。
func (s *fileService) Delete(ctx context.Context, id int64) error {
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return apperrors.NewNotFound("file", id)
		}
		return apperrors.NewInternal("查询文件失败", err)
	}

	if err := s.storage.Delete(ctx, file.StorageKey); err != nil {
		return apperrors.NewInternal("删除文件失败", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.NewInternal("删除文件元数据失败", err)
	}
	return nil
}

// ListByUser 获取用户的文件列表。
func (s *fileService) ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]*domain.File, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	limit := pageSize
	offset := (page - 1) * pageSize
	files, err := s.repo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, apperrors.NewInternal("查询文件列表失败", err)
	}
	return files, len(files), nil
}

// PresignedUploadURL 获取客户端直传的预签名上传 URL，同时预创建文件元数据。
func (s *fileService) PresignedUploadURL(ctx context.Context, userID int64, name, contentType string) (string, *domain.File, error) {
	if err := s.validateUpload(name, contentType, 0); err != nil {
		return "", nil, err
	}

	file := &domain.File{
		Name:        name,
		StorageKey:  s.generateKey(name),
		ContentType: contentType,
		Bucket:      s.bucket,
		CreatedBy:   userID,
	}
	if err := file.Validate(); err != nil {
		return "", nil, apperrors.NewValidation("file", err.Error())
	}

	url, err := s.storage.PresignedPutURL(ctx, file.StorageKey, 15*time.Minute)
	if err != nil {
		return "", nil, apperrors.NewInternal("生成上传链接失败", err)
	}

	if err := s.repo.Create(ctx, file); err != nil {
		return "", nil, apperrors.NewInternal("保存文件元数据失败", err)
	}
	return url, file, nil
}

func (s *fileService) validateUpload(name, contentType string, size int64) error {
	if name == "" {
		return apperrors.NewValidation("name", "文件名不能为空")
	}
	if contentType == "" {
		return apperrors.NewValidation("content_type", "Content-Type 不能为空")
	}
	if !storage.IsAllowedContentType(contentType, s.allowedTypes) {
		return apperrors.NewValidation("content_type", "不支持的文件类型").WithCode("UNSUPPORTED_CONTENT_TYPE")
	}
	if size > s.maxSize {
		return apperrors.NewValidation("size", fmt.Sprintf("文件大小超过限制 %d MB", s.maxSize/(1024*1024))).WithCode("FILE_TOO_LARGE")
	}
	return nil
}

func (s *fileService) generateKey(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	return fmt.Sprintf("%s/%s%s", uuid.New().String(), uuid.New().String(), ext)
}
