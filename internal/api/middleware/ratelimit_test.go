package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Create a rate limiter: 5 requests per second
	rl := NewRateLimiter(5, time.Second)

	clientIP := "192.168.1.1"

	// First 10 requests should be allowed (burst = 2x rate = 10)
	for i := 0; i < 10; i++ {
		if !rl.Allow(clientIP) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11th request should be denied
	if rl.Allow(clientIP) {
		t.Error("11th request should be denied")
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	// Create a rate limiter: 10 requests per 100ms
	rl := NewRateLimiter(10, 100*time.Millisecond)

	clientIP := "192.168.1.2"

	// Use all tokens (burst = 20)
	for i := 0; i < 20; i++ {
		rl.Allow(clientIP)
	}

	// Should be denied now
	if rl.Allow(clientIP) {
		t.Error("Request should be denied after using all tokens")
	}

	// Wait for token refill
	time.Sleep(150 * time.Millisecond)

	// Should be allowed now (10 tokens refilled)
	if !rl.Allow(clientIP) {
		t.Error("Request should be allowed after token refill")
	}
}

func TestRateLimiter_DifferentClients(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	client1 := "192.168.1.1"
	client2 := "192.168.1.2"

	// Both clients should have their own bucket
	for i := 0; i < 4; i++ {
		if !rl.Allow(client1) {
			t.Errorf("Client1 request %d should be allowed", i+1)
		}
		if !rl.Allow(client2) {
			t.Errorf("Client2 request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_Handler(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 4 requests should succeed (burst = 4)
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// 5th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(10, time.Second)

	// Stop should not panic
	rl.Stop()
}

func TestRateLimiter_RetryAfterHeader(t *testing.T) {
	rl := NewRateLimiter(1, time.Second)

	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use up all tokens
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Third request should be rate limited
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}

	// Check Retry-After header
	if retryAfter := w.Header().Get("Retry-After"); retryAfter != "1" {
		t.Errorf("Expected Retry-After header to be '1', got %q", retryAfter)
	}
}

func TestRateLimiter_MaxTokensCap(t *testing.T) {
	rl := NewRateLimiter(5, time.Second)

	clientIP := "192.168.1.100"

	// Use some tokens
	for i := 0; i < 5; i++ {
		rl.Allow(clientIP)
	}

	// Wait for refill beyond max
	time.Sleep(500 * time.Millisecond)

	// Check that tokens don't exceed maxTokens
	rl.mu.Lock()
	b := rl.tokens[clientIP]
	// After refill, tokens should be capped at maxTokens (10)
	if b.tokens > rl.maxTokens {
		t.Errorf("Tokens %d exceeded max %d", b.tokens, rl.maxTokens)
	}
	rl.mu.Unlock()
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		xff        string
		xri        string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For",
			xff:        "203.0.113.195",
			expected:   "203.0.113.195",
		},
		{
			name:       "X-Real-IP",
			xri:        "203.0.113.195",
			expected:   "203.0.113.195",
		},
		{
			name:       "RemoteAddr",
			remoteAddr: "192.168.1.1:12345",
			expected:   "192.168.1.1:12345",
		},
		{
			name:       "XFF takes priority",
			xff:        "203.0.113.195",
			xri:        "10.0.0.1",
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.195",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			got := getClientIP(req)
			if got != tt.expected {
				t.Errorf("getClientIP() = %v, want %v", got, tt.expected)
			}
		})
	}
}
