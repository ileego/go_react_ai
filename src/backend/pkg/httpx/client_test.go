package httpx

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Do_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(5*time.Second, DefaultRetryConfig())
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_Do_RetryOn503(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(5*time.Second, RetryConfig{
		MaxRetries:           3,
		InitialBackoff:       10 * time.Millisecond,
		MaxBackoff:           50 * time.Millisecond,
		RetryableStatusCodes: []int{http.StatusServiceUnavailable},
	})
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_Do_MaxRetriesExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewClient(5*time.Second, RetryConfig{
		MaxRetries:           2,
		InitialBackoff:       10 * time.Millisecond,
		MaxBackoff:           50 * time.Millisecond,
		RetryableStatusCodes: []int{http.StatusServiceUnavailable},
	})
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	_, err := client.Do(req)
	if err == nil {
		t.Error("expected error after max retries")
	}
}

func TestClient_Do_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(50*time.Millisecond, RetryConfig{MaxRetries: 0})
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	_, err := client.Do(req)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestClient_Do_Fallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	fallbackCalled := false
	fallback := func(_ error, resp *http.Response) ([]byte, error) {
		fallbackCalled = true
		return []byte(`{"fallback":true}`), nil
	}

	client := NewClientWithFallback(5*time.Second, RetryConfig{
		MaxRetries:           1,
		InitialBackoff:       10 * time.Millisecond,
		MaxBackoff:           50 * time.Millisecond,
		RetryableStatusCodes: []int{http.StatusServiceUnavailable},
	}, fallback)

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error with fallback: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if !fallbackCalled {
		t.Error("fallback should be called")
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"fallback":true}` {
		t.Errorf("unexpected fallback body: %s", string(body))
	}
}

func TestClient_PostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"hello":"world"}` {
			t.Errorf("unexpected body: %s", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(5*time.Second, DefaultRetryConfig())
	resp, err := client.PostJSON(context.Background(), server.URL, []byte(`{"hello":"world"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
}

func TestBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		initial time.Duration
		max     time.Duration
	}{
		{1, 100 * time.Millisecond, 1 * time.Second},
		{2, 100 * time.Millisecond, 1 * time.Second},
		{3, 100 * time.Millisecond, 1 * time.Second},
	}

	for _, tt := range tests {
		got := backoff(tt.attempt, tt.initial, tt.max)
		if got < tt.initial || got > tt.max {
			t.Errorf("backoff out of range: %v", got)
		}
	}
}
