package maven

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jiin/stale/internal/service/httputil"
)

type Client struct {
	httpClient  *http.Client
	retryConfig httputil.RetryConfig
}

type searchResponse struct {
	Response struct {
		Docs []struct {
			LatestVersion string `json:"latestVersion"`
		} `json:"docs"`
	} `json:"response"`
}

func New() *Client {
	return &Client{
		httpClient:  httputil.NewClient(10 * time.Second),
		retryConfig: httputil.DefaultRetryConfig(),
	}
}

// GetLatestVersion fetches the latest version from Maven Central
// groupID: e.g., "org.springframework.boot"
// artifactID: e.g., "spring-boot-starter-web"
func (c *Client) GetLatestVersion(ctx context.Context, groupID, artifactID string) (string, error) {
	url := fmt.Sprintf(
		"https://search.maven.org/solrsearch/select?q=g:%%22%s%%22+AND+a:%%22%s%%22&rows=1&wt=json",
		groupID, artifactID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := httputil.DoWithRetry(ctx, c.httpClient, req, c.retryConfig)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("maven central returned status %d", resp.StatusCode)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Response.Docs) == 0 {
		return "", fmt.Errorf("artifact not found: %s:%s", groupID, artifactID)
	}

	return result.Response.Docs[0].LatestVersion, nil
}
