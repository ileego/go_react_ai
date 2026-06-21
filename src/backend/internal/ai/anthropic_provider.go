package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ileego/go_react_ai/pkg/httpx"
)

// anthropicProvider 实现 Anthropic Messages API。
type anthropicProvider struct {
	baseURL string
	apiKey  string
	model   string
	client  *httpx.Client
}

// NewAnthropicProvider 创建 Anthropic Provider。
func NewAnthropicProvider(baseURL, apiKey, model string, client *httpx.Client) AIProvider {
	return &anthropicProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client:  client,
	}
}

func (p *anthropicProvider) Name() string { return "anthropic" }

func (p *anthropicProvider) DefaultModel() string { return DefaultModel("anthropic") }

func (p *anthropicProvider) ValidateModel(model string) error {
	return ValidateModel("anthropic", model)
}

func (p *anthropicProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	req = applyDefaults(req, p)
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := p.ValidateModel(req.Model); err != nil {
		return nil, err
	}

	body, err := p.buildRequestBody(req, false)
	if err != nil {
		return nil, err
	}

	url := p.baseURL + "/messages"
	resp, err := p.client.PostJSONWithHeaders(ctx, url, body, p.headers())
	if err != nil {
		return nil, fmt.Errorf("anthropic complete request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic complete status %d: %s", resp.StatusCode, string(body))
	}

	var result anthropicCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("anthropic decode response: %w", err)
	}
	content := p.extractContent(result.Content)
	return &CompletionResponse{
		Content:          content,
		Model:            result.Model,
		PromptTokens:     result.Usage.InputTokens,
		CompletionTokens: result.Usage.OutputTokens,
	}, nil
}

func (p *anthropicProvider) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	req = applyDefaults(req, p)
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := p.ValidateModel(req.Model); err != nil {
		return nil, err
	}
	if !SupportsStream("anthropic", req.Model) {
		return nil, fmt.Errorf("model %s does not support streaming", req.Model)
	}

	body, err := p.buildRequestBody(req, true)
	if err != nil {
		return nil, err
	}

	url := p.baseURL + "/messages"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.DoStream(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic stream request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic stream status %d: %s", resp.StatusCode, string(b))
	}

	ch := make(chan StreamChunk)
	go p.readSSE(resp.Body, ch)
	return ch, nil
}

func (p *anthropicProvider) headers() map[string]string {
	return map[string]string{
		"x-api-key":         p.apiKey,
		"anthropic-version": "2023-06-01",
	}
}

func (p *anthropicProvider) buildRequestBody(req CompletionRequest, stream bool) ([]byte, error) {
	messages := make([]anthropicMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == RoleSystem {
			continue // system prompt 在顶层字段
		}
		messages = append(messages, anthropicMessage{Role: string(m.Role), Content: m.Content})
	}
	body := anthropicChatRequest{
		Model:     req.Model,
		System:    req.SystemPrompt,
		Messages:  messages,
		MaxTokens: req.MaxTokens,
		Stream:    stream,
	}
	if req.Temperature > 0 {
		body.Temperature = req.Temperature
	}
	return json.Marshal(body)
}

func (p *anthropicProvider) extractContent(blocks []anthropicContentBlock) string {
	var sb strings.Builder
	for _, b := range blocks {
		if b.Type == "text" {
			sb.WriteString(b.Text)
		}
	}
	return sb.String()
}

func (p *anthropicProvider) readSSE(r io.ReadCloser, ch chan<- StreamChunk) {
	defer close(ch)
	defer func() { _ = r.Close() }()

	scanner := bufio.NewScanner(r)
	var event string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			event = ""
			continue
		}
		if strings.HasPrefix(line, "event: ") {
			event = strings.TrimPrefix(line, "event: ")
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		if event == "content_block_delta" {
			var delta anthropicContentDelta
			if err := json.Unmarshal([]byte(data), &delta); err != nil {
				ch <- StreamChunk{Error: fmt.Errorf("parse sse delta: %w", err)}
				return
			}
			if delta.Delta.Type == "text_delta" {
				ch <- StreamChunk{Content: delta.Delta.Text}
			}
		}
		if event == "message_stop" {
			ch <- StreamChunk{Done: true}
			return
		}
	}
	if err := scanner.Err(); err != nil {
		ch <- StreamChunk{Error: fmt.Errorf("read sse: %w", err)}
	}
}

type anthropicChatRequest struct {
	Model       string              `json:"model"`
	System      string              `json:"system,omitempty"`
	Messages    []anthropicMessage  `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float64             `json:"temperature,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicCompletionResponse struct {
	Model   string                     `json:"model"`
	Content []anthropicContentBlock    `json:"content"`
	Usage   anthropicUsage             `json:"usage"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicContentDelta struct {
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}
