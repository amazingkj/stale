package middleware

import (
	"net/http"
)

// SecurityHeaders adds security-related HTTP headers to responses
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking
			w.Header().Set("X-Frame-Options", "DENY")

			// XSS protection (legacy, but still useful for older browsers)
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer policy - don't leak referrer to external sites
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions policy - restrict browser features
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// Content Security Policy for API responses
			// More restrictive for API, allows scripts/styles for SPA
			if isAPIRequest(r) {
				w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			} else {
				// SPA needs inline scripts and styles from Vite build
				w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAPIRequest(r *http.Request) bool {
	return len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api"
}
