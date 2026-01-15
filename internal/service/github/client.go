package github

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	org    string
}

func New(token, org string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		client: github.NewClient(tc),
		org:    org,
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
	opt := &github.RepositoryListByAuthenticatedUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Affiliation: "owner,collaborator,organization_member",
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
