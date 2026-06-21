package domain

import (
	"errors"
	"time"
)

// File 表示上传的文件元数据。
type File struct {
	ID          int64
	Name        string
	StorageKey  string
	ContentType string
	Size        int64
	Bucket      string
	CreatedBy   int64
	CreatedAt   time.Time
}

// Validate 校验文件元数据。
func (f *File) Validate() error {
	if f.Name == "" {
		return errors.New("文件名不能为空")
	}
	if f.StorageKey == "" {
		return errors.New("存储键不能为空")
	}
	if f.ContentType == "" {
		return errors.New("Content-Type 不能为空")
	}
	if f.Size < 0 {
		return errors.New("文件大小不能为负数")
	}
	return nil
}

// IsImage 判断是否为图片类型。
func (f *File) IsImage() bool {
	switch f.ContentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	}
	return false
}
