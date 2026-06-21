package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
)

// FileRepository PostgreSQL 版文件元数据访问实现。
type FileRepository struct {
	db *sql.DB
}

// NewFileRepository 创建 PostgreSQL 版 FileRepository。
func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

// GetByID 根据 ID 获取文件元数据。
func (r *FileRepository) GetByID(ctx context.Context, id int64) (*domain.File, error) {
	row := r.db.QueryRowContext(ctx, queryFileByID, id)
	file, err := scanFile(row)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Create 创建文件元数据记录。
func (r *FileRepository) Create(ctx context.Context, file *domain.File) error {
	file.CreatedAt = time.Now()
	return r.db.QueryRowContext(ctx, queryInsertFile,
		file.Name,
		file.StorageKey,
		file.ContentType,
		file.Size,
		file.Bucket,
		file.CreatedBy,
		file.CreatedAt,
	).Scan(&file.ID)
}

// Delete 删除文件元数据记录。
func (r *FileRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, queryDeleteFile, id)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// ListByUser 获取用户的文件列表。
func (r *FileRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*domain.File, error) {
	rows, err := r.db.QueryContext(ctx, queryFilesByUser, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var files []*domain.File
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate files: %w", err)
	}
	return files, nil
}

type fileScanner interface {
	Scan(dest ...any) error
}

func scanFile(scanner fileScanner) (*domain.File, error) {
	var file domain.File
	err := scanner.Scan(
		&file.ID,
		&file.Name,
		&file.StorageKey,
		&file.ContentType,
		&file.Size,
		&file.Bucket,
		&file.CreatedBy,
		&file.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("scan file: %w", err)
	}
	return &file, nil
}

const (
	queryFileByID = `
		SELECT id, name, storage_key, content_type, size, bucket, created_by, created_at
		FROM files
		WHERE id = $1
	`

	queryInsertFile = `
		INSERT INTO files (name, storage_key, content_type, size, bucket, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	queryDeleteFile = `
		DELETE FROM files
		WHERE id = $1
	`

	queryFilesByUser = `
		SELECT id, name, storage_key, content_type, size, bucket, created_by, created_at
		FROM files
		WHERE created_by = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
)
