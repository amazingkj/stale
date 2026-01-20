package maven

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jiin/stale/internal/service/cache"
	"github.com/jiin/stale/internal/service/httputil"
)

// Cache TTL: 1 hour - maven versions don't change that frequently
const cacheTTL = 1 * time.Hour

type Client struct {
	httpClient  *http.Client
	retryConfig httputil.RetryConfig
	cache       *cache.Cache[string]
}

// mavenMetadata represents the maven-metadata.xml structure
type mavenMetadata struct {
	XMLName    xml.Name `xml:"metadata"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Versioning struct {
		Latest   string `xml:"latest"`
		Release  string `xml:"release"`
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
	} `xml:"versioning"`
}

func New() *Client {
	return &Client{
		httpClient:  httputil.NewClient(10 * time.Second),
		retryConfig: httputil.DefaultRetryConfig(),
		cache:       cache.New[string](cacheTTL),
	}
}

// GetLatestVersion fetches the latest version from Maven Central
// groupID: e.g., "org.springframework.boot"
// artifactID: e.g., "spring-boot-starter-web"
func (c *Client) GetLatestVersion(ctx context.Context, groupID, artifactID string) (string, error) {
	// Cache key is groupID:artifactID
	cacheKey := groupID + ":" + artifactID

	// Check cache first
	if version, found := c.cache.Get(cacheKey); found {
		return version, nil
	}

	// Use maven-metadata.xml from Maven Central repository (more accurate than search API)
	// Convert groupID dots to path separators: org.springframework.boot -> org/springframework/boot
	groupPath := strings.ReplaceAll(groupID, ".", "/")
	url := fmt.Sprintf(
		"https://repo1.maven.org/maven2/%s/%s/maven-metadata.xml",
		groupPath, artifactID,
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
		return "", fmt.Errorf("maven central returned status %d for %s:%s", resp.StatusCode, groupID, artifactID)
	}

	var metadata mavenMetadata
	if err := xml.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return "", fmt.Errorf("failed to parse maven-metadata.xml: %w", err)
	}

	// Prefer release version, fallback to latest, then last version in list
	version := metadata.Versioning.Release
	if version == "" {
		version = metadata.Versioning.Latest
	}
	if version == "" && len(metadata.Versioning.Versions.Version) > 0 {
		// Get the last version in the list (usually the newest)
		version = metadata.Versioning.Versions.Version[len(metadata.Versioning.Versions.Version)-1]
	}

	if version == "" {
		return "", fmt.Errorf("no version found for %s:%s", groupID, artifactID)
	}

	// Store in cache
	c.cache.Set(cacheKey, version)
	return version, nil
}
