package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
)

// UserRepository PostgreSQL 版用户数据访问实现
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建 PostgreSQL 版 UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, queryUserByID, id)
	user, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail 根据邮箱获取用户（不返回密码）
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, queryUserByEmail, strings.ToLower(email))
	user, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmailWithPassword 根据邮箱获取用户并返回密码哈希
func (r *UserRepository) GetByEmailWithPassword(ctx context.Context, email string) (*domain.User, string, error) {
	var user domain.User
	var passwordHash string
	row := r.db.QueryRowContext(ctx, queryUserByEmailWithPassword, strings.ToLower(email))
	err := row.Scan(
		&user.ID,
		&user.Email,
		&passwordHash,
		&user.Nickname,
		&user.AvatarURL,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", repository.ErrNotFound
		}
		return nil, "", fmt.Errorf("scan user with password: %w", err)
	}
	return &user, passwordHash, nil
}

// GetByGithubID 根据 GitHub ID 获取用户
func (r *UserRepository) GetByGithubID(ctx context.Context, githubID string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, queryUserByGithubID, githubID)
	user, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateWithPassword 创建普通用户
func (r *UserRepository) CreateWithPassword(ctx context.Context, user *domain.User, passwordHash string) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.Role == "" {
		user.Role = domain.UserRoleUser
	}

	return r.db.QueryRowContext(ctx, queryInsertUserWithPassword,
		strings.ToLower(user.Email),
		passwordHash,
		user.Nickname,
		user.AvatarURL,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
}

// CreateWithGithub 创建通过 GitHub 登录的用户
func (r *UserRepository) CreateWithGithub(ctx context.Context, user *domain.User, githubID string) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.Role == "" {
		user.Role = domain.UserRoleUser
	}

	return r.db.QueryRowContext(ctx, queryInsertUserWithGithub,
		strings.ToLower(user.Email),
		user.Nickname,
		user.AvatarURL,
		user.Role,
		githubID,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
}

// Update 更新用户基本信息
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, queryUpdateUser,
		user.Nickname,
		user.AvatarURL,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
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

// UpdatePassword 更新用户密码
func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	result, err := r.db.ExecContext(ctx, queryUpdateUserPassword, passwordHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
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

// UpdateRole 更新用户角色
func (r *UserRepository) UpdateRole(ctx context.Context, id int64, role string) error {
	result, err := r.db.ExecContext(ctx, queryUpdateUserRole, role, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
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

// UpdateGithubInfo 更新 GitHub 绑定信息
func (r *UserRepository) UpdateGithubInfo(ctx context.Context, id int64, githubID, avatarURL string) error {
	result, err := r.db.ExecContext(ctx, queryUpdateUserGithubInfo, githubID, avatarURL, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update user github info: %w", err)
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

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner userScanner) (*domain.User, error) {
	var user domain.User
	err := scanner.Scan(
		&user.ID,
		&user.Email,
		&user.Nickname,
		&user.AvatarURL,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &user, nil
}

const (
	userBaseColumns = `
		SELECT id, email, nickname, avatar_url, role, created_at, updated_at
		FROM users
	`

	queryUserByID = userBaseColumns + ` WHERE id = $1 `

	queryUserByEmail = userBaseColumns + ` WHERE email = LOWER($1) `

	queryUserByEmailWithPassword = `
		SELECT id, email, password_hash, nickname, avatar_url, role, created_at, updated_at
		FROM users
		WHERE email = LOWER($1)
	`

	queryUserByGithubID = userBaseColumns + ` WHERE github_id = $1 `

	queryInsertUserWithPassword = `
		INSERT INTO users (email, password_hash, nickname, avatar_url, role, created_at, updated_at)
		VALUES (LOWER($1), $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	queryInsertUserWithGithub = `
		INSERT INTO users (email, password_hash, nickname, avatar_url, role, github_id, created_at, updated_at)
		VALUES (LOWER($1), '', $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	queryUpdateUser = `
		UPDATE users
		SET nickname = $1, avatar_url = $2, updated_at = $3
		WHERE id = $4
	`

	queryUpdateUserPassword = `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`

	queryUpdateUserRole = `
		UPDATE users
		SET role = $1, updated_at = $2
		WHERE id = $3
	`

	queryUpdateUserGithubInfo = `
		UPDATE users
		SET github_id = $1, avatar_url = $2, updated_at = $3
		WHERE id = $4
	`
)
