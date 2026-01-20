package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHashAPIKey(t *testing.T) {
	// Test that HashAPIKey produces consistent hashes
	key := "my-secret-api-key"
	hash1 := HashAPIKey(key)
	hash2 := HashAPIKey(key)

	if hash1 != hash2 {
		t.Error("HashAPIKey should produce consistent hashes")
	}

	// Different keys should produce different hashes
	differentHash := HashAPIKey("different-key")
	if hash1 == differentHash {
		t.Error("Different keys should produce different hashes")
	}

	// Hash should be 64 characters (SHA-256 hex)
	if len(hash1) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash1))
	}
}

func TestAuth_Disabled(t *testing.T) {
	config := AuthConfig{
		Enabled: false,
	}

	called := false
	handler := Auth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/sources", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("next handler should be called when auth is disabled")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuth_HealthEndpointSkipped(t *testing.T) {
	config := AuthConfig{
		APIKey:  "secret",
		Enabled: true,
	}

	called := false
	handler := Auth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// Health endpoint should be accessible without auth
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("health endpoint should skip authentication")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuth_FrontendSkipped(t *testing.T) {
	config := AuthConfig{
		APIKey:  "secret",
		Enabled: true,
	}

	called := false
	handler := Auth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// Frontend routes should be accessible without auth
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("frontend routes should skip authentication")
	}
}

func TestAuth_ValidAPIKey_PlainKey(t *testing.T) {
	config := AuthConfig{
		APIKey:  "my-secret-key",
		Enabled: true,
	}

	called := false
	handler := Auth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/sources", nil)
	req.Header.Set("X-API-Key", "my-secret-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("handler should be called with valid API key")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuth_ValidAPIKey_BearerToken(t *testing.T) {
	config := AuthConfig{
		APIKey:  "my-secret-key",
		Enabled: true,
	}

	called := false
	handler := Auth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/sources", nil)
	req.Header.Set("Authorization", "Bearer my-secret-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("handler should be called with valid Bearer token")
	}
}

func TestAuth_ValidAPIKey_HashedKey(t *testing.T) {
	plainKey := "my-secret-key"
	hashedKey := HashAPIKey(plainKey)

	config := AuthConfig{
		APIKeyHash: hashedKey,
		Enabled:    true,
	}

	called := false
	handler := Auth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/sources", nil)
	req.Header.Set("X-API-Key", plainKey)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("handler should be called with valid hashed API key")
	}
}

func TestAuth_InvalidAPIKey(t *testing.T) {
	config := AuthConfig{
		APIKey:  "correct-key",
		Enabled: true,
	}

	handler := Auth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/sources", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_MissingAPIKey(t *testing.T) {
	config := AuthConfig{
		APIKey:  "secret",
		Enabled: true,
	}

	handler := Auth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/sources", nil)
	// No API key header
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestValidateAPIKey(t *testing.T) {
	hashedKey := HashAPIKey("test-key")

	tests := []struct {
		name        string
		providedKey string
		config      AuthConfig
		expected    bool
	}{
		{
			name:        "valid plain key",
			providedKey: "test-key",
			config:      AuthConfig{APIKey: "test-key"},
			expected:    true,
		},
		{
			name:        "invalid plain key",
			providedKey: "wrong-key",
			config:      AuthConfig{APIKey: "test-key"},
			expected:    false,
		},
		{
			name:        "valid hashed key",
			providedKey: "test-key",
			config:      AuthConfig{APIKeyHash: hashedKey},
			expected:    true,
		},
		{
			name:        "invalid hashed key",
			providedKey: "wrong-key",
			config:      AuthConfig{APIKeyHash: hashedKey},
			expected:    false,
		},
		{
			name:        "empty key",
			providedKey: "",
			config:      AuthConfig{APIKey: "test-key"},
			expected:    false,
		},
		{
			name:        "no config keys",
			providedKey: "any-key",
			config:      AuthConfig{},
			expected:    false,
		},
		{
			name:        "hash takes precedence over plain",
			providedKey: "test-key",
			config:      AuthConfig{APIKey: "plain-key", APIKeyHash: hashedKey},
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateAPIKey(tt.providedKey, tt.config)
			if result != tt.expected {
				t.Errorf("validateAPIKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}
