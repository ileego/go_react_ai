// Package response 提供统一的 HTTP 响应封装。
// Handler 层统一使用这些函数构造响应，保证前后端契约一致。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/pkg/errors"
)

// Body 统一响应体
type Body[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

// ListBody 列表响应体
type ListBody[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
	Total   int    `json:"total,omitempty"`
	Page    int    `json:"page,omitempty"`
	Size    int    `json:"size,omitempty"`
}

// OK 成功响应（无数据）
func OK(c *gin.Context) {
	c.JSON(http.StatusOK, Body[any]{Code: 0})
}

// Data 成功响应（有数据）
func Data[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Body[T]{Code: 0, Data: data})
}

// List 列表响应
func List[T any](c *gin.Context, data T, total, page, size int) {
	c.JSON(http.StatusOK, ListBody[T]{
		Code:  0,
		Data:  data,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Created 创建成功响应
func Created[T any](c *gin.Context, data T) {
	c.JSON(http.StatusCreated, Body[T]{Code: 0, Data: data})
}

// Error 错误响应
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Body[any]{Code: status, Message: message})
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalServerError 内部错误
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// FromError 根据业务错误类型返回对应的 HTTP 响应
// 如果是 pkg/errors.Error，按 Kind 映射状态码；否则统一返回 500
func FromError(c *gin.Context, err error) {
	if e, ok := err.(*errors.Error); ok {
		switch e.Kind {
		case errors.KindValidation:
			Error(c, http.StatusBadRequest, e.Message)
		case errors.KindNotFound:
			Error(c, http.StatusNotFound, e.Message)
		case errors.KindDuplicate:
			Error(c, http.StatusConflict, e.Message)
		case errors.KindUnauthorized:
			Error(c, http.StatusUnauthorized, e.Message)
		case errors.KindForbidden:
			Error(c, http.StatusForbidden, e.Message)
		default:
			Error(c, http.StatusInternalServerError, e.Message)
		}
		return
	}
	InternalServerError(c, err.Error())
}
