package domain

import "fmt"

// ValidationError 字段校验错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("字段 %s 校验失败: %s", e.Field, e.Message)
}

// NewValidationError 创建校验错误
func NewValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

// NotFoundError 资源不存在错误
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s (id=%v) 不存在", e.Resource, e.ID)
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(resource string, id interface{}) error {
	return &NotFoundError{Resource: resource, ID: id}
}
