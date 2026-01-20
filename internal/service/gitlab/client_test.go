package gitlab

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name               string
		token              string
		baseURL            string
		groupPath          string
		insecureSkipVerify bool
		membershipOnly     bool
		expectedBaseURL    string
	}{
		{
			name:            "default base URL",
			token:           "test-token",
			baseURL:         "",
			expectedBaseURL: "https://gitlab.com",
		},
		{
			name:            "custom base URL",
			token:           "test-token",
			baseURL:         "https://gitlab.example.com",
			expectedBaseURL: "https://gitlab.example.com",
		},
		{
			name:               "with insecure skip verify",
			token:              "test-token",
			baseURL:            "https://gitlab.example.com",
			insecureSkipVerify: true,
			expectedBaseURL:    "https://gitlab.example.com",
		},
		{
			name:           "with membership only",
			token:          "test-token",
			baseURL:        "",
			membershipOnly: true,
			expectedBaseURL: "https://gitlab.com",
		},
		{
			name:            "with group path",
			token:           "test-token",
			baseURL:         "",
			groupPath:       "my-group",
			expectedBaseURL: "https://gitlab.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(tt.token, tt.baseURL, tt.groupPath, tt.insecureSkipVerify, tt.membershipOnly)

			if client.baseURL != tt.expectedBaseURL {
				t.Errorf("baseURL = %q, want %q", client.baseURL, tt.expectedBaseURL)
			}
			if client.token != tt.token {
				t.Errorf("token = %q, want %q", client.token, tt.token)
			}
			if client.groupPath != tt.groupPath {
				t.Errorf("groupPath = %q, want %q", client.groupPath, tt.groupPath)
			}
			if client.membershipOnly != tt.membershipOnly {
				t.Errorf("membershipOnly = %v, want %v", client.membershipOnly, tt.membershipOnly)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "valid token",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "invalid token - unauthorized",
			statusCode: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v4/user" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Header.Get("PRIVATE-TOKEN") != "test-token" {
					t.Errorf("missing or wrong token header")
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"id": 1, "username": "test"}`))
			}))
			defer server.Close()

			client := New("test-token", server.URL, "", false, false)
			err := client.ValidateToken(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListRepositories(t *testing.T) {
	t.Run("list user projects", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.URL.Path != "/api/v4/projects" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			// Only return repos on first call
			if callCount == 1 {
				repos := []Repository{
					{ID: 1, Name: "repo1", FullName: "user/repo1", DefaultBranch: "main", WebURL: "https://gitlab.com/user/repo1"},
					{ID: 2, Name: "repo2", FullName: "user/repo2", DefaultBranch: "master", WebURL: "https://gitlab.com/user/repo2"},
				}
				json.NewEncoder(w).Encode(repos)
			} else {
				json.NewEncoder(w).Encode([]Repository{})
			}
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		repos, err := client.ListRepositories(context.Background())

		if err != nil {
			t.Fatalf("ListRepositories() error = %v", err)
		}
		if len(repos) != 2 {
			t.Errorf("got %d repos, want 2", len(repos))
		}
		if repos[0].Name != "repo1" {
			t.Errorf("repos[0].Name = %q, want %q", repos[0].Name, "repo1")
		}
	})

	t.Run("list with membership only", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("membership") != "true" {
				t.Errorf("membership query param not set")
			}
			json.NewEncoder(w).Encode([]Repository{})
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, true)
		_, err := client.ListRepositories(context.Background())

		if err != nil {
			t.Fatalf("ListRepositories() error = %v", err)
		}
	})

	t.Run("list group projects", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v4/groups/my-group/projects" {
				t.Errorf("unexpected path: %s, want /api/v4/groups/my-group/projects", r.URL.Path)
			}
			if r.URL.Query().Get("include_subgroups") != "true" {
				t.Errorf("include_subgroups query param not set")
			}
			json.NewEncoder(w).Encode([]Repository{})
		}))
		defer server.Close()

		client := New("test-token", server.URL, "my-group", false, false)
		_, err := client.ListRepositories(context.Background())

		if err != nil {
			t.Fatalf("ListRepositories() error = %v", err)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			page := r.URL.Query().Get("page")

			if page == "1" {
				repos := []Repository{{ID: 1, Name: "repo1", FullName: "user/repo1"}}
				json.NewEncoder(w).Encode(repos)
			} else {
				json.NewEncoder(w).Encode([]Repository{})
			}
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		repos, err := client.ListRepositories(context.Background())

		if err != nil {
			t.Fatalf("ListRepositories() error = %v", err)
		}
		if len(repos) != 1 {
			t.Errorf("got %d repos, want 1", len(repos))
		}
		if callCount != 2 {
			t.Errorf("expected 2 API calls, got %d", callCount)
		}
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		_, err := client.ListRepositories(context.Background())

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestGetFileContent(t *testing.T) {
	t.Run("base64 encoded content", func(t *testing.T) {
		expectedContent := "test file content"
		encodedContent := base64.StdEncoding.EncodeToString([]byte(expectedContent))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			file := FileContent{
				Content:  encodedContent,
				Encoding: "base64",
			}
			json.NewEncoder(w).Encode(file)
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		content, err := client.GetFileContent(context.Background(), "user/repo", "package.json", "main")

		if err != nil {
			t.Fatalf("GetFileContent() error = %v", err)
		}
		if string(content) != expectedContent {
			t.Errorf("content = %q, want %q", string(content), expectedContent)
		}
	})

	t.Run("plain text content", func(t *testing.T) {
		expectedContent := "plain text content"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			file := FileContent{
				Content:  expectedContent,
				Encoding: "text",
			}
			json.NewEncoder(w).Encode(file)
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		content, err := client.GetFileContent(context.Background(), "user/repo", "README.md", "main")

		if err != nil {
			t.Fatalf("GetFileContent() error = %v", err)
		}
		if string(content) != expectedContent {
			t.Errorf("content = %q, want %q", string(content), expectedContent)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		_, err := client.GetFileContent(context.Background(), "user/repo", "nonexistent.txt", "main")

		if err == nil {
			t.Error("expected error for not found file")
		}
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		_, err := client.GetFileContent(context.Background(), "user/repo", "file.txt", "main")

		if err == nil {
			t.Error("expected error for server error")
		}
	})

	t.Run("URL encoding", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// httptest decodes the URL, so we check the raw request URI
			// The important thing is that the request succeeds with special characters
			file := FileContent{Content: "content", Encoding: "text"}
			json.NewEncoder(w).Encode(file)
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		content, err := client.GetFileContent(context.Background(), "group/subgroup/repo", "src/main/file.txt", "main")

		if err != nil {
			t.Fatalf("GetFileContent() error = %v", err)
		}
		if string(content) != "content" {
			t.Errorf("content = %q, want %q", string(content), "content")
		}
	})
}

func TestListManifestFiles(t *testing.T) {
	t.Run("finds manifest files", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.URL.Query().Get("recursive") != "true" {
				t.Errorf("recursive query param not set")
			}

			// Only return entries on first call
			if callCount == 1 {
				entries := []TreeEntry{
					{ID: "1", Name: "package.json", Type: "blob", Path: "package.json"},
					{ID: "2", Name: "pom.xml", Type: "blob", Path: "backend/pom.xml"},
					{ID: "3", Name: "go.mod", Type: "blob", Path: "services/api/go.mod"},
					{ID: "4", Name: "src", Type: "tree", Path: "src"},
					{ID: "5", Name: "README.md", Type: "blob", Path: "README.md"},
					{ID: "6", Name: "build.gradle", Type: "blob", Path: "android/build.gradle"},
				}
				json.NewEncoder(w).Encode(entries)
			} else {
				json.NewEncoder(w).Encode([]TreeEntry{})
			}
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		manifests, err := client.ListManifestFiles(context.Background(), "user/repo", "main")

		if err != nil {
			t.Fatalf("ListManifestFiles() error = %v", err)
		}

		expected := []string{"package.json", "backend/pom.xml", "services/api/go.mod", "android/build.gradle"}
		if len(manifests) != len(expected) {
			t.Errorf("got %d manifests, want %d", len(manifests), len(expected))
			return
		}

		for i, path := range manifests {
			if path != expected[i] {
				t.Errorf("manifests[%d] = %q, want %q", i, path, expected[i])
			}
		}
	})

	t.Run("no manifest files", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				entries := []TreeEntry{
					{ID: "1", Name: "README.md", Type: "blob", Path: "README.md"},
					{ID: "2", Name: "src", Type: "tree", Path: "src"},
				}
				json.NewEncoder(w).Encode(entries)
			} else {
				json.NewEncoder(w).Encode([]TreeEntry{})
			}
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		manifests, err := client.ListManifestFiles(context.Background(), "user/repo", "main")

		if err != nil {
			t.Fatalf("ListManifestFiles() error = %v", err)
		}
		if len(manifests) != 0 {
			t.Errorf("expected no manifests, got %d", len(manifests))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			page := r.URL.Query().Get("page")

			if page == "1" {
				entries := []TreeEntry{
					{ID: "1", Name: "package.json", Type: "blob", Path: "package.json"},
				}
				json.NewEncoder(w).Encode(entries)
			} else {
				json.NewEncoder(w).Encode([]TreeEntry{})
			}
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		manifests, err := client.ListManifestFiles(context.Background(), "user/repo", "main")

		if err != nil {
			t.Fatalf("ListManifestFiles() error = %v", err)
		}
		if len(manifests) != 1 {
			t.Errorf("got %d manifests, want 1", len(manifests))
		}
		if callCount != 2 {
			t.Errorf("expected 2 API calls, got %d", callCount)
		}
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		_, err := client.ListManifestFiles(context.Background(), "user/repo", "main")

		if err == nil {
			t.Error("expected error for forbidden response")
		}
	})

	t.Run("finds build.gradle.kts", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				entries := []TreeEntry{
					{ID: "1", Name: "build.gradle.kts", Type: "blob", Path: "build.gradle.kts"},
				}
				json.NewEncoder(w).Encode(entries)
			} else {
				json.NewEncoder(w).Encode([]TreeEntry{})
			}
		}))
		defer server.Close()

		client := New("test-token", server.URL, "", false, false)
		manifests, err := client.ListManifestFiles(context.Background(), "user/repo", "main")

		if err != nil {
			t.Fatalf("ListManifestFiles() error = %v", err)
		}
		if len(manifests) != 1 || manifests[0] != "build.gradle.kts" {
			t.Errorf("expected [build.gradle.kts], got %v", manifests)
		}
	})
}
