// Package domain 定义核心业务实体，不依赖任何外部框架或库。
// 这是 Clean Architecture 中最内层的“实体层”。
package domain

import (
	"errors"
	"time"
)

// User 表示平台用户
type User struct {
	ID        int64
	Email     string
	Name      string
	AvatarURL string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate 校验用户字段合法性
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if u.Name == "" {
		return errors.New("姓名不能为空")
	}
	return nil
}
