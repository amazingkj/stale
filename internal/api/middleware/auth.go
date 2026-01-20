package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	APIKey     string // Plain API key (for backward compatibility)
	APIKeyHash string // SHA-256 hash of API key (recommended for production)
	Enabled    bool
}

// DefaultAuthConfig returns authentication configuration from environment
// Supports both plain API key (STALE_API_KEY) and hashed API key (STALE_API_KEY_HASH)
// If STALE_API_KEY_HASH is set, it takes precedence over STALE_API_KEY
func DefaultAuthConfig() AuthConfig {
	apiKey := os.Getenv("STALE_API_KEY")
	apiKeyHash := os.Getenv("STALE_API_KEY_HASH")
	enabled := apiKey != "" || apiKeyHash != ""

	if apiKeyHash != "" {
		log.Info().Msg("API authentication enabled (using hashed key)")
	} else if apiKey != "" {
		log.Info().Msg("API authentication enabled (using plain key - consider using STALE_API_KEY_HASH for better security)")
	} else {
		log.Warn().Msg("API authentication disabled - set STALE_API_KEY or STALE_API_KEY_HASH to enable")
	}

	return AuthConfig{
		APIKey:     apiKey,
		APIKeyHash: apiKeyHash,
		Enabled:    enabled,
	}
}

// HashAPIKey generates a SHA-256 hash of the API key
// Use this to generate the hash for STALE_API_KEY_HASH environment variable
func HashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// Auth returns an authentication middleware handler
func Auth(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth if not enabled
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Skip auth for health check endpoint
			if r.URL.Path == "/api/v1/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip auth for frontend routes (non-API)
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}

			// Check Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Also check X-API-Key header
				authHeader = r.Header.Get("X-API-Key")
			}

			// Extract API key from Bearer token or direct key
			providedKey := ""
			if strings.HasPrefix(authHeader, "Bearer ") {
				providedKey = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				providedKey = authHeader
			}

			// Validate API key
			if !validateAPIKey(providedKey, config) {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error": "Unauthorized", "message": "Invalid or missing API key", "code": 401}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// validateAPIKey checks if the provided API key is valid
// Uses constant-time comparison to prevent timing attacks
func validateAPIKey(providedKey string, config AuthConfig) bool {
	if providedKey == "" {
		return false
	}

	// If hash is configured, compare using hash
	if config.APIKeyHash != "" {
		providedHash := HashAPIKey(providedKey)
		return subtle.ConstantTimeCompare([]byte(providedHash), []byte(config.APIKeyHash)) == 1
	}

	// Fall back to direct comparison (constant-time)
	if config.APIKey != "" {
		return subtle.ConstantTimeCompare([]byte(providedKey), []byte(config.APIKey)) == 1
	}

	return false
}
