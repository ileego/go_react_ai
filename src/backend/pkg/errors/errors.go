// Package errors 提供统一的错误类型，用于跨层传递业务错误。
// 替代标准库的 errors.New 和 fmt.Errorf，让调用方能做类型断言。
package errors

import "fmt"

// Kind 表示错误类型
type Kind string

const (
	KindValidation   Kind = "validation"   // 参数校验失败
	KindNotFound     Kind = "not_found"    // 资源不存在
	KindDuplicate    Kind = "duplicate"    // 重复数据
	KindInternal     Kind = "internal"     // 内部错误
	KindUnauthorized Kind = "unauthorized" // 未授权
	KindForbidden    Kind = "forbidden"    // 禁止访问
)

// Error 统一业务错误
type Error struct {
	Kind    Kind
	Code    string // 前端稳定错误码，如 REPORT_NOT_FOUND
	Field   string // 校验错误时对应的字段
	Message string
	Cause   error // 原始错误，用于日志追溯
}

func (e *Error) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("[%s] field=%s: %s", e.Kind, e.Field, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Kind, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// WithCode 为错误设置前端稳定错误码，支持链式调用
func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

// IsKind 判断错误类型
func IsKind(err error, kind Kind) bool {
	if e, ok := err.(*Error); ok {
		return e.Kind == kind
	}
	return false
}

// NewValidation 创建校验错误
func NewValidation(field, message string) *Error {
	return &Error{Kind: KindValidation, Field: field, Message: message}
}

// NewNotFound 创建资源不存在错误
func NewNotFound(resource string, id any) *Error {
	return &Error{Kind: KindNotFound, Message: fmt.Sprintf("%s (id=%v) 不存在", resource, id)}
}

// NewInternal 创建内部错误
func NewInternal(message string, cause error) *Error {
	return &Error{Kind: KindInternal, Message: message, Cause: cause}
}

// NewDuplicate 创建资源重复错误
func NewDuplicate(resource string, id any) *Error {
	return &Error{Kind: KindDuplicate, Message: fmt.Sprintf("%s (id=%v) 已存在", resource, id)}
}

// NewUnauthorized 创建未授权错误
func NewUnauthorized(message string) *Error {
	return &Error{Kind: KindUnauthorized, Message: message}
}

// NewForbidden 创建禁止访问错误
func NewForbidden(message string) *Error {
	return &Error{Kind: KindForbidden, Message: message}
}
