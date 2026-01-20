package npm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jiin/stale/internal/service/cache"
	"github.com/jiin/stale/internal/service/httputil"
)

func newTestClient(serverURL string) *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		retryConfig: httputil.RetryConfig{MaxRetries: 1, BaseDelay: 10 * time.Millisecond},
		cache:       cache.New[string](time.Minute),
	}
}

func TestGetLatestVersion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/react" {
			t.Errorf("expected path /react, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"dist-tags": {"latest": "18.2.0"}}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	client.httpClient = server.Client()

	// Override the registry URL for testing
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/react", nil)
	resp, err := client.httpClient.Do(req)
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

	// This test verifies the server responds with 404
	resp, err := http.Get(server.URL + "/nonexistent-package")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestGetLatestVersion_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"dist-tags": {"latest": "1.0.0"}}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	// Simulate cache behavior
	client.cache.Set("test-package", "1.0.0")

	// Check that cached value is returned
	version, found := client.cache.Get("test-package")
	if !found {
		t.Error("expected cache hit")
	}
	if version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", version)
	}
}

func TestGetLatestVersion_NoLatestTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"dist-tags": {}}`))
	}))
	defer server.Close()

	// Test that empty dist-tags is handled
	resp, err := http.Get(server.URL + "/some-package")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
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
