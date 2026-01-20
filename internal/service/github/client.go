package github

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/jiin/stale/internal/service/httputil"
	"golang.org/x/oauth2"
)

type Client struct {
	client    *github.Client
	org       string
	ownerOnly bool
}

func New(token, org string, ownerOnly bool) *Client {
	// Use custom transport with connection pooling and retry logic
	transport := httputil.DefaultTransportWithRetry()

	// Wrap with OAuth2 transport
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauth2Transport := &oauth2.Transport{
		Base:   transport,
		Source: ts,
	}

	httpClient := &http.Client{
		Transport: oauth2Transport,
		Timeout:   30 * time.Second,
	}

	return &Client{
		client:    github.NewClient(httpClient),
		org:       org,
		ownerOnly: ownerOnly,
	}
}

type Repository struct {
	Name          string
	FullName      string
	DefaultBranch string
	HTMLURL       string
}

func (c *Client) ListRepositories(ctx context.Context) ([]Repository, error) {
	var allRepos []Repository

	if c.org != "" {
		repos, err := c.listOrgRepos(ctx)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
	} else {
		repos, err := c.listUserRepos(ctx)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
	}

	return allRepos, nil
}

func (c *Client) listOrgRepos(ctx context.Context) ([]Repository, error) {
	var allRepos []Repository
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, c.org, opt)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			allRepos = append(allRepos, toRepository(repo))
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (c *Client) listUserRepos(ctx context.Context) ([]Repository, error) {
	var allRepos []Repository
	affiliation := "owner,collaborator,organization_member"
	if c.ownerOnly {
		affiliation = "owner"
	}
	opt := &github.RepositoryListByAuthenticatedUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Affiliation: affiliation,
	}

	for {
		repos, resp, err := c.client.Repositories.ListByAuthenticatedUser(ctx, opt)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			allRepos = append(allRepos, toRepository(repo))
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func toRepository(repo *github.Repository) Repository {
	defaultBranch := "main"
	if repo.DefaultBranch != nil {
		defaultBranch = *repo.DefaultBranch
	}

	return Repository{
		Name:          repo.GetName(),
		FullName:      repo.GetFullName(),
		DefaultBranch: defaultBranch,
		HTMLURL:       repo.GetHTMLURL(),
	}
}

func (c *Client) GetFileContent(ctx context.Context, fullName, path, branch string) ([]byte, error) {
	parts := strings.SplitN(fullName, "/", 2)
	owner := parts[0]
	repo := parts[1]

	fileContent, _, _, err := c.client.Repositories.GetContents(
		ctx, owner, repo, path,
		&github.RepositoryContentGetOptions{Ref: branch},
	)
	if err != nil {
		return nil, err
	}

	if fileContent.Content == nil {
		return nil, nil
	}

	content, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		decoded, decErr := base64.RawStdEncoding.DecodeString(
			strings.ReplaceAll(*fileContent.Content, "\n", ""),
		)
		if decErr != nil {
			return nil, err
		}
		return decoded, nil
	}

	return content, nil
}

func (c *Client) ValidateToken(ctx context.Context) error {
	_, _, err := c.client.Users.Get(ctx, "")
	return err
}
