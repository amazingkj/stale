package golang

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jiin/stale/internal/service/httputil"
)

const proxyURL = "https://proxy.golang.org"

type Client struct {
	httpClient  *http.Client
	retryConfig httputil.RetryConfig
}

type ModuleInfo struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
}

func New() *Client {
	return &Client{
		httpClient:  httputil.NewClient(10 * time.Second),
		retryConfig: httputil.DefaultRetryConfig(),
	}
}

func (c *Client) GetLatestVersion(ctx context.Context, modulePath string) (string, error) {
	// Encode module path: replace / with encoded form
	encodedPath := strings.ReplaceAll(modulePath, "/", "/")
	// For case-insensitive paths, lowercase and add ! prefix to uppercase letters
	var encoded strings.Builder
	for _, r := range encodedPath {
		if r >= 'A' && r <= 'Z' {
			encoded.WriteRune('!')
			encoded.WriteRune(r + 32) // lowercase
		} else {
			encoded.WriteRune(r)
		}
	}

	reqURL := fmt.Sprintf("%s/%s/@latest", proxyURL, encoded.String())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := httputil.DoWithRetry(ctx, c.httpClient, req, c.retryConfig)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return "", fmt.Errorf("module %s not found", modulePath)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("go proxy returned %d for %s", resp.StatusCode, modulePath)
	}

	var info ModuleInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}

	return info.Version, nil
}
