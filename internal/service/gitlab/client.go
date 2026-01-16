package gitlab

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	httpClient   *http.Client
	token        string
	baseURL      string
	groupPath    string // Optional: for group-level operations
}

type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"path_with_namespace"`
	DefaultBranch string `json:"default_branch"`
	WebURL        string `json:"web_url"`
}

type FileContent struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func New(token, baseURL, groupPath string) *Client {
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token:     token,
		baseURL:   baseURL,
		groupPath: groupPath,
	}
}

func (c *Client) ValidateToken(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s/api/v4/user", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s/api/v4/user: %d", c.baseURL, resp.StatusCode)
	}

	return nil
}

func (c *Client) ListRepositories(ctx context.Context) ([]Repository, error) {
	var allRepos []Repository
	page := 1
	perPage := 100

	for {
		var endpoint string
		if c.groupPath != "" {
			// List projects in a group
			endpoint = fmt.Sprintf("%s/api/v4/groups/%s/projects?page=%d&per_page=%d&include_subgroups=true",
				c.baseURL, url.PathEscape(c.groupPath), page, perPage)
		} else {
			// List projects accessible to the user
			endpoint = fmt.Sprintf("%s/api/v4/projects?membership=true&page=%d&per_page=%d",
				c.baseURL, page, perPage)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("PRIVATE-TOKEN", c.token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("gitlab API returned status %d", resp.StatusCode)
		}

		var repos []Repository
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		if len(repos) == 0 {
			break
		}

		allRepos = append(allRepos, repos...)
		page++

		// Safety limit
		if page > 50 {
			break
		}
	}

	return allRepos, nil
}

func (c *Client) GetFileContent(ctx context.Context, projectPath, filePath, ref string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/api/v4/projects/%s/repository/files/%s?ref=%s",
		c.baseURL,
		url.PathEscape(projectPath),
		url.PathEscape(filePath),
		url.QueryEscape(ref),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gitlab API returned status %d", resp.StatusCode)
	}

	var file FileContent
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, err
	}

	if file.Encoding == "base64" {
		return base64.StdEncoding.DecodeString(file.Content)
	}

	return []byte(file.Content), nil
}
