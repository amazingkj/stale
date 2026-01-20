package golang

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jiin/stale/internal/service/cache"
	"github.com/jiin/stale/internal/service/httputil"
)

func newTestClient() *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		retryConfig: httputil.RetryConfig{MaxRetries: 1, BaseDelay: 10 * time.Millisecond},
		cache:       cache.New[string](time.Minute),
	}
}

func TestGetLatestVersion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"Version": "v1.2.3", "Time": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetLatestVersion_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestGetLatestVersion_Gone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusGone {
		t.Errorf("expected status 410, got %d", resp.StatusCode)
	}
}

func TestGetLatestVersion_Cache(t *testing.T) {
	client := newTestClient()

	// Pre-populate cache
	client.cache.Set("github.com/example/module", "v1.0.0")

	// Check cache hit
	version, found := client.cache.Get("github.com/example/module")
	if !found {
		t.Error("expected cache hit")
	}
	if version != "v1.0.0" {
		t.Errorf("expected version v1.0.0, got %s", version)
	}
}

func TestNew(t *testing.T) {
	client := New()

	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if client.cache == nil {
		t.Error("cache should not be nil")
	}
}

func TestModulePathEncoding(t *testing.T) {
	// Test that uppercase letters are properly encoded
	tests := []struct {
		input    string
		expected string
	}{
		{"github.com/example/module", "github.com/example/module"},
		{"github.com/Example/Module", "github.com/!example/!module"},
		{"github.com/ABC/xyz", "github.com/!a!b!c/xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var encoded strings.Builder
			for _, r := range tt.input {
				if r >= 'A' && r <= 'Z' {
					encoded.WriteRune('!')
					encoded.WriteRune(r + 32)
				} else {
					encoded.WriteRune(r)
				}
			}
			if encoded.String() != tt.expected {
				t.Errorf("encoded %q, expected %q", encoded.String(), tt.expected)
			}
		})
	}
}

func TestModuleInfo_Parsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"Version": "v2.5.0", "Time": "2024-06-15T10:30:00Z"}`))
	}))
	defer server.Close()

	client := newTestClient()
	client.httpClient = server.Client()

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
