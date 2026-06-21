// Package ai 封装真实 AI Provider 的调用。
// 提供统一的 AIProvider 接口、提示词模板与模型管理能力。
package ai

import (
	"context"
	"fmt"
)

// Role 定义对话消息角色。
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message 表示一条对话消息。
type Message struct {
	Role    Role
	Content string
}

// CompletionRequest 是 Provider 无关的补全请求。
type CompletionRequest struct {
	Model        string
	SystemPrompt string
	Messages     []Message
	Temperature  float64
	MaxTokens    int
}

// Validate 校验请求参数。
func (r CompletionRequest) Validate() error {
	if r.Model == "" {
		return fmt.Errorf("model is required")
	}
	if len(r.Messages) == 0 {
		return fmt.Errorf("messages is required")
	}
	if r.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}
	return nil
}

// CompletionResponse 是 Provider 无关的补全响应。
type CompletionResponse struct {
	Content          string
	Model            string
	PromptTokens     int
	CompletionTokens int
}

// StreamChunk 表示流式响应中的一个数据块。
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// AIProvider 抽象不同 AI Provider 的调用能力。
type AIProvider interface {
	// Name 返回 Provider 名称，如 openai、anthropic。
	Name() string
	// DefaultModel 返回该 Provider 的默认模型。
	DefaultModel() string
	// ValidateModel 校验模型是否被该 Provider 支持。
	ValidateModel(model string) error
	// Complete 发起非流式补全请求。
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	// Stream 发起流式补全请求，通过 channel 返回数据块。
	Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
}

// normalizeModel 如果传入空模型，返回 Provider 默认模型。
func normalizeModel(p AIProvider, model string) string {
	if model == "" {
		return p.DefaultModel()
	}
	return model
}

// applyDefaults 给请求填充默认值。
func applyDefaults(req CompletionRequest, p AIProvider) CompletionRequest {
	req.Model = normalizeModel(p, req.Model)
	if req.Temperature == 0 && req.MaxTokens == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = 4000
	}
	return req
}
