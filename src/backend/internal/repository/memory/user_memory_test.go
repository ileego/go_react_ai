package memory

import (
	"context"
	"testing"

	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/repository"
)

func TestUserRepository_CreateWithPassword(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	user := &domain.User{
		Email:    "test@example.com",
		Nickname: "Tester",
		Role:     domain.UserRoleUser,
	}
	err := repo.CreateWithPassword(ctx, user, "hashed-password")
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("user id should not be zero")
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	user := &domain.User{
		Email:    "test@example.com",
		Nickname: "Tester",
	}
	_ = repo.CreateWithPassword(ctx, user, "hashed-password")

	found, err := repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("get by email failed: %v", err)
	}
	if found.Email != "test@example.com" {
		t.Errorf("want email test@example.com, got %s", found.Email)
	}
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "missing@example.com")
	if !isNotFound(err) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestUserRepository_GetByEmailWithPassword(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	user := &domain.User{
		Email:    "test@example.com",
		Nickname: "Tester",
	}
	_ = repo.CreateWithPassword(ctx, user, "hashed-password")

	found, hash, err := repo.GetByEmailWithPassword(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("get by email with password failed: %v", err)
	}
	if hash != "hashed-password" {
		t.Errorf("want hash hashed-password, got %s", hash)
	}
	if found.ID != user.ID {
		t.Errorf("want user id %d, got %d", user.ID, found.ID)
	}
}

func TestUserRepository_CreateWithGithub(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	user := &domain.User{
		Email:    "github@example.com",
		Nickname: "GithubUser",
	}
	err := repo.CreateWithGithub(ctx, user, "123456")
	if err != nil {
		t.Fatalf("create github user failed: %v", err)
	}

	found, err := repo.GetByGithubID(ctx, "123456")
	if err != nil {
		t.Fatalf("get by github id failed: %v", err)
	}
	if found.Email != "github@example.com" {
		t.Errorf("want email github@example.com, got %s", found.Email)
	}
}

func TestUserRepository_UpdateGithubInfo(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	user := &domain.User{
		Email:    "test@example.com",
		Nickname: "Tester",
	}
	_ = repo.CreateWithPassword(ctx, user, "hashed-password")

	err := repo.UpdateGithubInfo(ctx, user.ID, "789", "https://avatar.url")
	if err != nil {
		t.Fatalf("update github info failed: %v", err)
	}

	found, err := repo.GetByGithubID(ctx, "789")
	if err != nil {
		t.Fatalf("get by github id failed: %v", err)
	}
	if found.AvatarURL != "https://avatar.url" {
		t.Errorf("want avatar url https://avatar.url, got %s", found.AvatarURL)
	}
}

func TestUserRepository_UpdateRole(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	user := &domain.User{
		Email:    "test@example.com",
		Nickname: "Tester",
	}
	_ = repo.CreateWithPassword(ctx, user, "hashed-password")

	err := repo.UpdateRole(ctx, user.ID, "admin")
	if err != nil {
		t.Fatalf("update role failed: %v", err)
	}

	found, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get by id failed: %v", err)
	}
	if found.Role != domain.UserRoleAdmin {
		t.Errorf("want role admin, got %s", found.Role)
	}
}

func isNotFound(err error) bool {
	return err == repository.ErrNotFound
}
