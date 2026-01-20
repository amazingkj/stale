package httputil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultTransport(t *testing.T) {
	transport := DefaultTransport()

	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
	if transport.MaxIdleConns != 100 {
		t.Errorf("MaxIdleConns = %d, want 100", transport.MaxIdleConns)
	}
	if transport.MaxConnsPerHost != 10 {
		t.Errorf("MaxConnsPerHost = %d, want 10", transport.MaxConnsPerHost)
	}
	if transport.MaxIdleConnsPerHost != 10 {
		t.Errorf("MaxIdleConnsPerHost = %d, want 10", transport.MaxIdleConnsPerHost)
	}
	if transport.IdleConnTimeout != 90*time.Second {
		t.Errorf("IdleConnTimeout = %v, want 90s", transport.IdleConnTimeout)
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient(30 * time.Second)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", client.Timeout)
	}
	if client.Transport == nil {
		t.Error("Transport should not be nil")
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}
	if config.BaseDelay != 100*time.Millisecond {
		t.Errorf("BaseDelay = %v, want 100ms", config.BaseDelay)
	}
	if config.MaxDelay != 2*time.Second {
		t.Errorf("MaxDelay = %v, want 2s", config.MaxDelay)
	}
	if !config.RetryOn5xx {
		t.Error("RetryOn5xx should be true")
	}
	if !config.RetryOnTemp {
		t.Error("RetryOnTemp should be true")
	}
}

func TestDoWithRetry_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	config := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := DoWithRetry(context.Background(), client, req, config)

	if err != nil {
		t.Fatalf("DoWithRetry failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestDoWithRetry_Retry5xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	config := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		RetryOn5xx: true,
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := DoWithRetry(context.Background(), client, req, config)

	if err != nil {
		t.Fatalf("DoWithRetry failed: %v", err)
	}
	defer resp.Body.Close()

	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestDoWithRetry_MaxRetriesExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	config := RetryConfig{
		MaxRetries: 2,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		RetryOn5xx: true,
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	_, err := DoWithRetry(context.Background(), client, req, config)

	if err == nil {
		t.Error("expected error after max retries exceeded")
	}
}

func TestDoWithRetry_NoRetryOn4xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	config := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		RetryOn5xx: true,
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := DoWithRetry(context.Background(), client, req, config)

	if err != nil {
		t.Fatalf("DoWithRetry failed: %v", err)
	}
	defer resp.Body.Close()

	// Should not retry on 4xx
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (no retry on 4xx)", attempts)
	}
}

func TestDoWithRetry_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	config := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  50 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	_, err := DoWithRetry(ctx, client, req, config)

	if err == nil {
		t.Error("expected error when context is canceled")
	}
}

func TestDefaultTransportWithRetry(t *testing.T) {
	transport := DefaultTransportWithRetry()

	if transport == nil {
		t.Fatal("expected non-nil transport")
	}

	rt, ok := transport.(*RetryTransport)
	if !ok {
		t.Fatal("expected *RetryTransport type")
	}

	if rt.Base == nil {
		t.Error("Base transport should not be nil")
	}
	if rt.Config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", rt.Config.MaxRetries)
	}
}

func TestRetryTransport_RoundTrip_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &RetryTransport{
		Base: http.DefaultTransport,
		Config: RetryConfig{
			MaxRetries: 3,
			BaseDelay:  10 * time.Millisecond,
			MaxDelay:   100 * time.Millisecond,
		},
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestRetryTransport_RoundTrip_Retry429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &RetryTransport{
		Base: http.DefaultTransport,
		Config: RetryConfig{
			MaxRetries: 3,
			BaseDelay:  10 * time.Millisecond,
			MaxDelay:   100 * time.Millisecond,
			RetryOn5xx: true,
		},
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}
