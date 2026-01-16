package golang

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const proxyURL = "https://proxy.golang.org"

type Client struct {
	httpClient *http.Client
}

type ModuleInfo struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
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

	resp, err := c.httpClient.Do(req)
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
