package memory

import (
	"context"
	"sync"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
)

// FileRepository 内存版文件元数据访问实现。
type FileRepository struct {
	mu     sync.RWMutex
	files  map[int64]*domain.File
	nextID int64
}

// NewFileRepository 创建内存版 FileRepository。
func NewFileRepository() *FileRepository {
	return &FileRepository{
		files:  make(map[int64]*domain.File),
		nextID: 1,
	}
}

// GetByID 根据 ID 获取文件。
func (r *FileRepository) GetByID(_ context.Context, id int64) (*domain.File, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	file, ok := r.files[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return copyFile(file), nil
}

// Create 创建文件元数据。
func (r *FileRepository) Create(_ context.Context, file *domain.File) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file.ID = r.nextID
	r.nextID++
	file.CreatedAt = time.Now()
	r.files[file.ID] = copyFile(file)
	return nil
}

// Delete 删除文件元数据。
func (r *FileRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.files[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.files, id)
	return nil
}

// ListByUser 获取用户的文件列表。
func (r *FileRepository) ListByUser(_ context.Context, userID int64, limit, offset int) ([]*domain.File, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.File
	for _, file := range r.files {
		if file.CreatedBy == userID {
			result = append(result, copyFile(file))
		}
	}

	if offset > len(result) {
		return []*domain.File{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func copyFile(f *domain.File) *domain.File {
	return &domain.File{
		ID:          f.ID,
		Name:        f.Name,
		StorageKey:  f.StorageKey,
		ContentType: f.ContentType,
		Size:        f.Size,
		Bucket:      f.Bucket,
		CreatedBy:   f.CreatedBy,
		CreatedAt:   f.CreatedAt,
	}
}
