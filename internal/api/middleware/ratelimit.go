package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu         sync.Mutex
	tokens     map[string]*bucket
	rate       int           // tokens per interval
	interval   time.Duration // refill interval
	maxTokens  int           // max tokens per client
	cleanupInt time.Duration // cleanup interval for stale entries
}

type bucket struct {
	tokens   int
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per interval
// interval: time window for rate limiting
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:     make(map[string]*bucket),
		rate:       rate,
		interval:   interval,
		maxTokens:  rate * 2, // allow burst up to 2x rate
		cleanupInt: 5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes stale entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInt)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, b := range rl.tokens {
			if now.Sub(b.lastSeen) > rl.cleanupInt {
				delete(rl.tokens, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if the client is allowed to make a request
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.tokens[clientIP]
	if !exists {
		rl.tokens[clientIP] = &bucket{
			tokens:   rl.maxTokens - 1,
			lastSeen: now,
		}
		return true
	}

	// Refill tokens based on time passed
	elapsed := now.Sub(b.lastSeen)
	tokensToAdd := int(elapsed / rl.interval) * rl.rate
	b.tokens = min(b.tokens+tokensToAdd, rl.maxTokens)
	b.lastSeen = now

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// Handler returns a middleware that rate limits requests
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !rl.Allow(clientIP) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}
