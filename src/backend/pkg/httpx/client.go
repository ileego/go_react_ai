// Package httpx 提供带超时、重试、降级能力的 HTTP 客户端封装。
package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"slices"
	"time"
)

// RetryConfig 定义重试策略。
type RetryConfig struct {
	MaxRetries           int
	InitialBackoff       time.Duration
	MaxBackoff           time.Duration
	RetryableStatusCodes []int
}

// DefaultRetryConfig 返回默认重试配置。
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:           3,
		InitialBackoff:       500 * time.Millisecond,
		MaxBackoff:           5 * time.Second,
		RetryableStatusCodes: []int{http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout},
	}
}

// FallbackFunc 定义所有重试失败后的降级函数。
// 参数为最后一次请求的错误与响应；返回降级后的响应体与错误。
type FallbackFunc func(err error, resp *http.Response) ([]byte, error)

// Client 是带重试与降级能力的 HTTP 客户端。
type Client struct {
	base     *http.Client
	retryCfg RetryConfig
	fallback FallbackFunc
}

// NewClient 创建 HTTP 客户端。
// timeout 为单次请求超时；retryCfg 为重试配置。
func NewClient(timeout time.Duration, retryCfg RetryConfig) *Client {
	return &Client{
		base: &http.Client{
			Timeout: timeout,
		},
		retryCfg: retryCfg,
	}
}

// NewClientWithFallback 创建带降级能力的 HTTP 客户端。
func NewClientWithFallback(timeout time.Duration, retryCfg RetryConfig, fallback FallbackFunc) *Client {
	c := NewClient(timeout, retryCfg)
	c.fallback = fallback
	return c
}

// Do 执行 HTTP 请求，带超时与指数退避重试。
// 如果配置了 fallback 且所有重试失败，会返回 fallback 结果而不是错误。
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var lastErr error
	var lastResp *http.Response

	body, err := readBody(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}

	for attempt := 0; attempt <= c.retryCfg.MaxRetries; attempt++ {
		if attempt > 0 {
			wait := backoff(attempt, c.retryCfg.InitialBackoff, c.retryCfg.MaxBackoff)
			time.Sleep(wait)
		}

		clonedReq := req.Clone(req.Context())
		if body != nil {
			clonedReq.Body = io.NopCloser(bytes.NewReader(body))
		}

		resp, err := c.base.Do(clonedReq)
		if err != nil {
			lastErr = err
			continue
		}

		if !isRetryableStatus(resp.StatusCode, c.retryCfg.RetryableStatusCodes) {
			return resp, nil
		}

		lastResp = resp
		lastErr = fmt.Errorf("retryable status %d", resp.StatusCode)
		_ = resp.Body.Close()
	}

	if c.fallback != nil {
		fallbackBody, err := c.fallback(lastErr, lastResp)
		if err != nil {
			return nil, fmt.Errorf("after %d retries, fallback failed: %w", c.retryCfg.MaxRetries, err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(fallbackBody)),
			Request:    req,
		}, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("after %d retries: %w", c.retryCfg.MaxRetries, lastErr)
	}
	return nil, fmt.Errorf("after %d retries, last status: %d", c.retryCfg.MaxRetries, lastResp.StatusCode)
}

// PostJSON 发送 POST JSON 请求。
func (c *Client) PostJSON(ctx context.Context, url string, body []byte) (*http.Response, error) {
	return c.PostJSONWithHeaders(ctx, url, body, nil)
}

// PostJSONWithHeaders 发送带自定义请求头的 POST JSON 请求。
func (c *Client) PostJSONWithHeaders(ctx context.Context, url string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.Do(req)
}

// DoStream 执行 HTTP 请求并返回响应，不读取响应体，不重试，不设置全局超时。
// 调用方必须通过 context 控制生命周期，并在读取完毕后关闭 resp.Body。
func (c *Client) DoStream(req *http.Request) (*http.Response, error) {
	// 流式请求使用无全局超时的客户端，由 context 控制；同时跳过重试，避免在读到一半时重试。
	streamClient := &http.Client{
		Transport: c.base.Transport,
	}
	return streamClient.Do(req)
}

// readBody 读取请求 Body 并在读取后还原，便于重试时复用。
func readBody(r io.ReadCloser) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	defer func() { _ = r.Close() }()
	return io.ReadAll(r)
}

// isRetryableStatus 判断状态码是否需要重试。
func isRetryableStatus(code int, retryable []int) bool {
	return slices.Contains(retryable, code)
}

// backoff 计算指数退避时间（带全抖动）。
func backoff(attempt int, initial, maxBackoff time.Duration) time.Duration {
	d := float64(initial) * math.Pow(2, float64(attempt-1))
	if d > float64(maxBackoff) {
		d = float64(maxBackoff)
	}
	// 全抖动：0 ~ d 之间的随机值
	d = d * rand.Float64()
	if d < float64(initial) {
		d = float64(initial)
	}
	return time.Duration(d)
}
