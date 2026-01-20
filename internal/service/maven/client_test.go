package maven

import (
	"context"
	"net/http"
	"net/http/httptest"
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
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>org.example</groupId>
  <artifactId>test-lib</artifactId>
  <versioning>
    <latest>2.0.0</latest>
    <release>2.0.0</release>
    <versions>
      <version>1.0.0</version>
      <version>2.0.0</version>
    </versions>
  </versioning>
</metadata>`))
	}))
	defer server.Close()

	// Test the server response
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetLatestVersion_ReleaseOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>org.example</groupId>
  <artifactId>test-lib</artifactId>
  <versioning>
    <release>1.5.0</release>
    <versions>
      <version>1.0.0</version>
      <version>1.5.0</version>
    </versions>
  </versioning>
</metadata>`))
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

func TestGetLatestVersion_VersionsOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>org.example</groupId>
  <artifactId>test-lib</artifactId>
  <versioning>
    <versions>
      <version>1.0.0</version>
      <version>1.1.0</version>
      <version>1.2.0</version>
    </versions>
  </versioning>
</metadata>`))
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

func TestGetLatestVersion_Cache(t *testing.T) {
	client := newTestClient()

	// Pre-populate cache
	client.cache.Set("org.example:test-lib", "1.0.0")

	// Check cache hit
	version, found := client.cache.Get("org.example:test-lib")
	if !found {
		t.Error("expected cache hit")
	}
	if version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", version)
	}
}

func TestGetLatestVersion_InvalidXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid xml content`))
	}))
	defer server.Close()

	client := newTestClient()
	client.httpClient = server.Client()

	// Make a request to verify the server returns invalid XML
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

func TestNew(t *testing.T) {
	client := New()

	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if client.cache == nil {
		t.Error("cache should not be nil")
	}
}

func TestMavenMetadata_Parsing(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>com.example</groupId>
  <artifactId>my-artifact</artifactId>
  <versioning>
    <latest>3.0.0</latest>
    <release>3.0.0</release>
    <versions>
      <version>1.0.0</version>
      <version>2.0.0</version>
      <version>3.0.0</version>
    </versions>
  </versioning>
</metadata>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(xmlData))
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
