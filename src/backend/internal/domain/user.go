// Package domain 定义核心业务实体，不依赖任何外部框架或库。
// 这是 Clean Architecture 中最内层的“实体层”。
package domain

import (
	"errors"
	"time"
)

// UserRole 用户角色
type UserRole string

const (
	UserRoleSystem UserRole = "system" // 系统管理员
	UserRoleAdmin  UserRole = "admin"  // 普通管理员
	UserRoleUser   UserRole = "user"   // 普通用户
)

// User 表示平台用户
type User struct {
	ID        int64
	Email     string
	Nickname  string
	AvatarURL string
	Role      UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate 校验用户字段合法性
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if u.Nickname == "" {
		return errors.New("昵称不能为空")
	}
	if u.Role == "" {
		u.Role = UserRoleUser
	}
	return nil
}
