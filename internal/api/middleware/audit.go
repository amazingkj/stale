package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// AuditLog logs security-relevant events
func AuditLog() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Only log security-relevant events
			if shouldAudit(r) {
				duration := time.Since(start)
				clientIP := getClientIP(r)

				logEvent := log.Info().
					Str("type", "audit").
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Int("status", wrapped.statusCode).
					Str("client_ip", clientIP).
					Dur("duration", duration)

				// Add action description
				action := getAuditAction(r)
				if action != "" {
					logEvent = logEvent.Str("action", action)
				}

				// Log if API key was used
				if r.Header.Get("X-API-Key") != "" || strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
					logEvent = logEvent.Bool("api_key_used", true)
				}

				// Log failed attempts with warning level
				if wrapped.statusCode == http.StatusUnauthorized || wrapped.statusCode == http.StatusForbidden {
					logEvent.Msg("security: unauthorized access attempt")
				} else {
					logEvent.Msg("audit: " + action)
				}
			}
		})
	}
}

// shouldAudit determines if the request should be logged for audit purposes
func shouldAudit(r *http.Request) bool {
	path := r.URL.Path

	// Always audit authentication failures (handled in the logger above)
	// Audit settings changes
	if strings.HasPrefix(path, "/api/v1/settings") && r.Method != http.MethodGet {
		return true
	}

	// Audit source management (tokens involved)
	if strings.HasPrefix(path, "/api/v1/sources") && r.Method != http.MethodGet {
		return true
	}

	// Audit scan triggers
	if strings.HasPrefix(path, "/api/v1/scans") && r.Method == http.MethodPost {
		return true
	}

	// Audit delete operations
	if r.Method == http.MethodDelete {
		return true
	}

	// Audit bulk operations
	if strings.Contains(path, "bulk") {
		return true
	}

	return false
}

// getAuditAction returns a human-readable action description
func getAuditAction(r *http.Request) string {
	path := r.URL.Path
	method := r.Method

	switch {
	case strings.HasPrefix(path, "/api/v1/settings"):
		if method == http.MethodPut {
			return "settings_updated"
		}
		if strings.HasSuffix(path, "/test-email") {
			return "test_email_sent"
		}
	case strings.HasPrefix(path, "/api/v1/sources"):
		switch method {
		case http.MethodPost:
			return "source_created"
		case http.MethodPut:
			return "source_updated"
		case http.MethodDelete:
			return "source_deleted"
		}
	case strings.HasPrefix(path, "/api/v1/repositories"):
		if method == http.MethodDelete {
			return "repository_deleted"
		}
		if strings.Contains(path, "bulk-delete") {
			return "repositories_bulk_deleted"
		}
	case strings.HasPrefix(path, "/api/v1/scans"):
		if method == http.MethodPost {
			if strings.Contains(path, "cancel") {
				return "scan_cancelled"
			}
			return "scan_triggered"
		}
	case strings.HasPrefix(path, "/api/v1/ignored"):
		switch {
		case strings.Contains(path, "bulk-delete"):
			return "ignored_bulk_removed"
		case strings.Contains(path, "bulk"):
			return "ignored_bulk_added"
		case method == http.MethodPost:
			return "ignored_added"
		case method == http.MethodDelete:
			return "ignored_removed"
		}
	}

	return method + " " + path
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
