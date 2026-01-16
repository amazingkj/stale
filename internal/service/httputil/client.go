package httputil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// DefaultTransport returns an http.Transport with optimized connection pooling
func DefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxConnsPerHost:       10,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}
}

// NewClient creates an HTTP client with connection pooling
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: DefaultTransport(),
	}
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	RetryOn5xx  bool
	RetryOnTemp bool
}

// DefaultRetryConfig returns sensible retry defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		RetryOn5xx:  true,
		RetryOnTemp: true,
	}
}

// DoWithRetry performs an HTTP request with exponential backoff retry
func DoWithRetry(ctx context.Context, client *http.Client, req *http.Request, config RetryConfig) (*http.Response, error) {
	var lastErr error
	delay := config.BaseDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			// Exponential backoff
			delay *= 2
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}

		// Clone request for retry (body needs to be re-readable)
		reqCopy := req.Clone(ctx)

		resp, err := client.Do(reqCopy)
		if err != nil {
			// Check for temporary errors
			if config.RetryOnTemp && isTemporaryError(err) {
				lastErr = err
				continue
			}
			return nil, err
		}

		// Check for retryable status codes
		if config.RetryOn5xx && resp.StatusCode >= 500 && resp.StatusCode < 600 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
	}
	return nil, fmt.Errorf("max retries exceeded")
}

func isTemporaryError(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	return false
}
