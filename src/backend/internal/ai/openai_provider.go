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

// openAIProvider 实现 OpenAI 兼容 API，包括 OpenAI、DeepSeek、Kimi。
type openAIProvider struct {
	name    string
	baseURL string
	apiKey  string
	model   string
	client  *httpx.Client
}

// NewOpenAIProvider 创建 OpenAI Provider。
func NewOpenAIProvider(baseURL, apiKey, model string, client *httpx.Client) AIProvider {
	return &openAIProvider{
		name:    "openai",
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client:  client,
	}
}

// NewDeepSeekProvider 创建 DeepSeek Provider。
func NewDeepSeekProvider(baseURL, apiKey, model string, client *httpx.Client) AIProvider {
	return &openAIProvider{
		name:    "deepseek",
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client:  client,
	}
}

// NewKimiProvider 创建 Kimi/Moonshot Provider。
func NewKimiProvider(baseURL, apiKey, model string, client *httpx.Client) AIProvider {
	return &openAIProvider{
		name:    "kimi",
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client:  client,
	}
}

func (p *openAIProvider) Name() string { return p.name }

func (p *openAIProvider) DefaultModel() string { return DefaultModel(p.name) }

func (p *openAIProvider) ValidateModel(model string) error {
	return ValidateModel(p.name, model)
}

func (p *openAIProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
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

	url := p.baseURL + "/chat/completions"
	resp, err := p.client.PostJSONWithHeaders(ctx, url, body, p.headers())
	if err != nil {
		return nil, fmt.Errorf("%s complete request: %w", p.name, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s complete status %d: %s", p.name, resp.StatusCode, string(body))
	}

	var result openAICompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("%s decode response: %w", p.name, err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("%s empty choices", p.name)
	}

	return &CompletionResponse{
		Content:          result.Choices[0].Message.Content,
		Model:            result.Model,
		PromptTokens:     result.Usage.PromptTokens,
		CompletionTokens: result.Usage.CompletionTokens,
	}, nil
}

func (p *openAIProvider) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	req = applyDefaults(req, p)
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := p.ValidateModel(req.Model); err != nil {
		return nil, err
	}
	if !SupportsStream(p.name, req.Model) {
		return nil, fmt.Errorf("model %s does not support streaming", req.Model)
	}

	body, err := p.buildRequestBody(req, true)
	if err != nil {
		return nil, err
	}

	url := p.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range p.headers() {
		httpReq.Header.Set(k, v)
	}

	resp, err := p.client.DoStream(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%s stream request: %w", p.name, err)
	}

	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s stream status %d: %s", p.name, resp.StatusCode, string(b))
	}

	ch := make(chan StreamChunk)
	go p.readSSE(resp.Body, ch)
	return ch, nil
}

func (p *openAIProvider) headers() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + p.apiKey,
	}
}

func (p *openAIProvider) buildRequestBody(req CompletionRequest, stream bool) ([]byte, error) {
	messages := make([]openAIMessage, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, openAIMessage{Role: "system", Content: req.SystemPrompt})
	}
	for _, m := range req.Messages {
		messages = append(messages, openAIMessage{Role: string(m.Role), Content: m.Content})
	}
	body := openAIChatRequest{
		Model:       req.Model,
		Messages:    messages,
		Stream:      stream,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
	return json.Marshal(body)
}

func (p *openAIProvider) readSSE(r io.ReadCloser, ch chan<- StreamChunk) {
	defer close(ch)
	defer func() { _ = r.Close() }()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			ch <- StreamChunk{Done: true}
			return
		}

		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("parse sse chunk: %w", err)}
			return
		}
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				ch <- StreamChunk{Content: content}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		ch <- StreamChunk{Error: fmt.Errorf("read sse: %w", err)}
	}
}

type openAIChatRequest struct {
	Model       string           `json:"model"`
	Messages    []openAIMessage  `json:"messages"`
	Stream      bool             `json:"stream"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAICompletionResponse struct {
	Model   string                   `json:"model"`
	Choices []openAIChoice           `json:"choices"`
	Usage   openAIUsage              `json:"usage"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}
