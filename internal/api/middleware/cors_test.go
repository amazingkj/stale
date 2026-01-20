package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_AllowedOrigin(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expectHeader   bool
	}{
		{
			name:           "wildcard allows any origin",
			allowedOrigins: []string{"*"},
			requestOrigin:  "https://example.com",
			expectHeader:   true,
		},
		{
			name:           "specific origin allowed",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://example.com",
			expectHeader:   true,
		},
		{
			name:           "origin not in list",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://other.com",
			expectHeader:   false,
		},
		{
			name:           "no origin header",
			allowedOrigins: []string{"*"},
			requestOrigin:  "",
			expectHeader:   false,
		},
		{
			name:           "multiple allowed origins",
			allowedOrigins: []string{"https://example.com", "https://app.example.com"},
			requestOrigin:  "https://app.example.com",
			expectHeader:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CORSConfig{
				AllowedOrigins: tt.allowedOrigins,
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
				MaxAge:         86400,
			}

			handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			hasHeader := w.Header().Get("Access-Control-Allow-Origin") != ""
			if hasHeader != tt.expectHeader {
				t.Errorf("Access-Control-Allow-Origin header present = %v, want %v", hasHeader, tt.expectHeader)
			}
		})
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	config := CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:         86400,
	}

	handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should not reach here"))
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("preflight response status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Body should be empty for preflight
	if w.Body.Len() != 0 {
		t.Error("preflight response should have empty body")
	}
}

func TestCORS_HeadersSet(t *testing.T) {
	config := CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type", "X-Custom"},
		MaxAge:         86400,
	}

	handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "https://example.com",
		"Access-Control-Allow-Methods":     "GET, POST",
		"Access-Control-Allow-Headers":     "Content-Type, X-Custom",
		"Access-Control-Allow-Credentials": "true",
	}

	for header, expected := range expectedHeaders {
		if got := w.Header().Get(header); got != expected {
			t.Errorf("%s = %q, want %q", header, got, expected)
		}
	}
}

func TestCORS_NextHandlerCalled(t *testing.T) {
	config := CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}

	called := false
	handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("next handler was not called")
	}
}
