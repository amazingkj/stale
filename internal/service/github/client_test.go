package github

import (
	"testing"

	"github.com/google/go-github/v68/github"
)

func TestToRepository(t *testing.T) {
	tests := []struct {
		name     string
		repo     *github.Repository
		expected Repository
	}{
		{
			name: "with all fields",
			repo: &github.Repository{
				Name:          github.Ptr("test-repo"),
				FullName:      github.Ptr("owner/test-repo"),
				DefaultBranch: github.Ptr("main"),
				HTMLURL:       github.Ptr("https://github.com/owner/test-repo"),
			},
			expected: Repository{
				Name:          "test-repo",
				FullName:      "owner/test-repo",
				DefaultBranch: "main",
				HTMLURL:       "https://github.com/owner/test-repo",
			},
		},
		{
			name: "with master branch",
			repo: &github.Repository{
				Name:          github.Ptr("legacy-repo"),
				FullName:      github.Ptr("owner/legacy-repo"),
				DefaultBranch: github.Ptr("master"),
				HTMLURL:       github.Ptr("https://github.com/owner/legacy-repo"),
			},
			expected: Repository{
				Name:          "legacy-repo",
				FullName:      "owner/legacy-repo",
				DefaultBranch: "master",
				HTMLURL:       "https://github.com/owner/legacy-repo",
			},
		},
		{
			name: "with nil default branch",
			repo: &github.Repository{
				Name:          github.Ptr("no-branch-repo"),
				FullName:      github.Ptr("owner/no-branch-repo"),
				DefaultBranch: nil,
				HTMLURL:       github.Ptr("https://github.com/owner/no-branch-repo"),
			},
			expected: Repository{
				Name:          "no-branch-repo",
				FullName:      "owner/no-branch-repo",
				DefaultBranch: "main", // Default fallback
				HTMLURL:       "https://github.com/owner/no-branch-repo",
			},
		},
		{
			name: "with custom branch",
			repo: &github.Repository{
				Name:          github.Ptr("custom-repo"),
				FullName:      github.Ptr("owner/custom-repo"),
				DefaultBranch: github.Ptr("develop"),
				HTMLURL:       github.Ptr("https://github.com/owner/custom-repo"),
			},
			expected: Repository{
				Name:          "custom-repo",
				FullName:      "owner/custom-repo",
				DefaultBranch: "develop",
				HTMLURL:       "https://github.com/owner/custom-repo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toRepository(tt.repo)

			if result.Name != tt.expected.Name {
				t.Errorf("Name = %q, want %q", result.Name, tt.expected.Name)
			}
			if result.FullName != tt.expected.FullName {
				t.Errorf("FullName = %q, want %q", result.FullName, tt.expected.FullName)
			}
			if result.DefaultBranch != tt.expected.DefaultBranch {
				t.Errorf("DefaultBranch = %q, want %q", result.DefaultBranch, tt.expected.DefaultBranch)
			}
			if result.HTMLURL != tt.expected.HTMLURL {
				t.Errorf("HTMLURL = %q, want %q", result.HTMLURL, tt.expected.HTMLURL)
			}
		})
	}
}

func TestNew(t *testing.T) {
	client := New("test-token", "", false)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.client == nil {
		t.Error("github client should not be nil")
	}
	if client.org != "" {
		t.Errorf("org = %q, want empty string", client.org)
	}
	if client.ownerOnly {
		t.Error("ownerOnly should be false")
	}
}

func TestNew_WithOrg(t *testing.T) {
	client := New("test-token", "my-org", false)

	if client.org != "my-org" {
		t.Errorf("org = %q, want %q", client.org, "my-org")
	}
}

func TestNew_WithOwnerOnly(t *testing.T) {
	client := New("test-token", "", true)

	if !client.ownerOnly {
		t.Error("ownerOnly should be true")
	}
}

func TestRepository_Fields(t *testing.T) {
	repo := Repository{
		Name:          "test",
		FullName:      "owner/test",
		DefaultBranch: "main",
		HTMLURL:       "https://github.com/owner/test",
	}

	if repo.Name != "test" {
		t.Errorf("Name = %q, want %q", repo.Name, "test")
	}
	if repo.FullName != "owner/test" {
		t.Errorf("FullName = %q, want %q", repo.FullName, "owner/test")
	}
	if repo.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want %q", repo.DefaultBranch, "main")
	}
	if repo.HTMLURL != "https://github.com/owner/test" {
		t.Errorf("HTMLURL = %q, want %q", repo.HTMLURL, "https://github.com/owner/test")
	}
}
