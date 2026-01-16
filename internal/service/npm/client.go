package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jiin/stale/internal/service/httputil"
)

const registryURL = "https://registry.npmjs.org"

type Client struct {
	httpClient  *http.Client
	retryConfig httputil.RetryConfig
}

type PackageInfo struct {
	DistTags map[string]string `json:"dist-tags"`
}

func New() *Client {
	return &Client{
		httpClient:  httputil.NewClient(10 * time.Second),
		retryConfig: httputil.DefaultRetryConfig(),
	}
}

func (c *Client) GetLatestVersion(ctx context.Context, packageName string) (string, error) {
	encodedName := url.PathEscape(packageName)
	reqURL := fmt.Sprintf("%s/%s", registryURL, encodedName)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.npm.install-v1+json")

	resp, err := httputil.DoWithRetry(ctx, c.httpClient, req, c.retryConfig)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("package %s not found", packageName)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("npm registry returned %d for %s", resp.StatusCode, packageName)
	}

	var info PackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}

	if latest, ok := info.DistTags["latest"]; ok {
		return latest, nil
	}

	return "", fmt.Errorf("no latest version found for %s", packageName)
}
