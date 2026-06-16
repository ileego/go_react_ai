package memory

import (
	"context"
	"sync"
	"time"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
)

// UserRepository 内存版用户数据访问实现（主要用于测试）
type UserRepository struct {
	mu        sync.RWMutex
	users     map[int64]*domain.User
	passwords map[int64]string
	githubIDs map[string]int64
	emails    map[string]int64
	nextID    int64
}

// NewUserRepository 创建内存版 UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:     make(map[int64]*domain.User),
		passwords: make(map[int64]string),
		githubIDs: make(map[string]int64),
		emails:    make(map[string]int64),
		nextID:    1,
	}
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return cloneUser(user), nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.emails[email]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return cloneUser(r.users[id]), nil
}

// GetByEmailWithPassword 根据邮箱获取用户并返回密码哈希
func (r *UserRepository) GetByEmailWithPassword(ctx context.Context, email string) (*domain.User, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.emails[email]
	if !ok {
		return nil, "", repository.ErrNotFound
	}
	return cloneUser(r.users[id]), r.passwords[id], nil
}

// GetByGithubID 根据 GitHub ID 获取用户
func (r *UserRepository) GetByGithubID(ctx context.Context, githubID string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.githubIDs[githubID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return cloneUser(r.users[id]), nil
}

// CreateWithPassword 创建普通用户
func (r *UserRepository) CreateWithPassword(ctx context.Context, user *domain.User, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.emails[user.Email]; exists {
		return repository.ErrDuplicate
	}

	if user.Role == "" {
		user.Role = domain.UserRoleUser
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.ID = r.nextID
	r.nextID++

	r.users[user.ID] = cloneUser(user)
	r.passwords[user.ID] = passwordHash
	r.emails[user.Email] = user.ID
	return nil
}

// CreateWithGithub 创建通过 GitHub 登录的用户
func (r *UserRepository) CreateWithGithub(ctx context.Context, user *domain.User, githubID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.emails[user.Email]; exists {
		return repository.ErrDuplicate
	}
	if _, exists := r.githubIDs[githubID]; exists {
		return repository.ErrDuplicate
	}

	if user.Role == "" {
		user.Role = domain.UserRoleUser
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.ID = r.nextID
	r.nextID++

	r.users[user.ID] = cloneUser(user)
	r.emails[user.Email] = user.ID
	r.githubIDs[githubID] = user.ID
	return nil
}

// Update 更新用户基本信息
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.users[user.ID]
	if !ok {
		return repository.ErrNotFound
	}

	existing.Nickname = user.Nickname
	existing.AvatarURL = user.AvatarURL
	existing.UpdatedAt = time.Now()
	return nil
}

// UpdatePassword 更新用户密码
func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return repository.ErrNotFound
	}
	r.passwords[id] = passwordHash
	r.users[id].UpdatedAt = time.Now()
	return nil
}

// UpdateRole 更新用户角色
func (r *UserRepository) UpdateRole(ctx context.Context, id int64, role string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[id]
	if !ok {
		return repository.ErrNotFound
	}
	user.Role = domain.UserRole(role)
	user.UpdatedAt = time.Now()
	return nil
}

// UpdateGithubInfo 更新 GitHub 绑定信息
func (r *UserRepository) UpdateGithubInfo(ctx context.Context, id int64, githubID, avatarURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[id]
	if !ok {
		return repository.ErrNotFound
	}
	if githubID != "" {
		r.githubIDs[githubID] = id
	}
	user.AvatarURL = avatarURL
	user.UpdatedAt = time.Now()
	return nil
}

func cloneUser(u *domain.User) *domain.User {
	if u == nil {
		return nil
	}
	cp := *u
	return &cp
}
