package scanner

import (
	"testing"
)

func TestCleanVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple version", "1.2.3", "1.2.3"},
		{"caret prefix", "^1.2.3", "1.2.3"},
		{"tilde prefix", "~1.2.3", "1.2.3"},
		{"greater than or equal", ">=1.2.3", "1.2.3"},
		{"less than or equal", "<=1.2.3", "1.2.3"},
		{"greater than", ">1.2.3", "1.2.3"},
		{"less than", "<1.2.3", "1.2.3"},
		{"equal sign", "=1.2.3", "1.2.3"},
		{"with space range", "1.2.3 2.0.0", "1.2.3"},
		{"with or operator", "^1.0.0 || ^2.0.0", "1.0.0"},
		{"empty string", "", ""},
		{"whitespace", "  1.2.3  ", "1.2.3"},
		{"multiple prefixes", "^~1.2.3", "1.2.3"},
		{"prerelease", "1.2.3-beta.1", "1.2.3-beta.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanVersion(tt.input)
			if result != tt.expected {
				t.Errorf("cleanVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsOutdated(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		latest   string
		expected bool
	}{
		{"same version", "1.2.3", "1.2.3", false},
		{"patch update", "1.2.3", "1.2.4", true},
		{"minor update", "1.2.3", "1.3.0", true},
		{"major update", "1.2.3", "2.0.0", true},
		{"current newer", "2.0.0", "1.0.0", false},
		{"empty current", "", "1.0.0", false},
		{"empty latest", "1.0.0", "", false},
		{"both empty", "", "", false},
		{"invalid current", "invalid", "1.0.0", false},
		{"invalid latest", "1.0.0", "invalid", false},
		{"prerelease current", "1.0.0-beta.1", "1.0.0", true},
		{"prerelease latest", "1.0.0", "1.0.1-beta.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOutdated(tt.current, tt.latest)
			if result != tt.expected {
				t.Errorf("isOutdated(%q, %q) = %v, want %v", tt.current, tt.latest, result, tt.expected)
			}
		})
	}
}

func TestParseGoMod(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []GoModDependency
	}{
		{
			name: "single line require",
			content: `module example.com/mymodule

go 1.21

require github.com/example/pkg v1.2.3
`,
			expected: []GoModDependency{
				{Path: "github.com/example/pkg", Version: "v1.2.3"},
			},
		},
		{
			name: "require block",
			content: `module example.com/mymodule

go 1.21

require (
	github.com/example/pkg1 v1.0.0
	github.com/example/pkg2 v2.0.0
)
`,
			expected: []GoModDependency{
				{Path: "github.com/example/pkg1", Version: "v1.0.0"},
				{Path: "github.com/example/pkg2", Version: "v2.0.0"},
			},
		},
		{
			name: "with indirect comments",
			content: `module example.com/mymodule

require (
	github.com/direct/pkg v1.0.0
	github.com/indirect/pkg v2.0.0 // indirect
)
`,
			expected: []GoModDependency{
				{Path: "github.com/direct/pkg", Version: "v1.0.0"},
				{Path: "github.com/indirect/pkg", Version: "v2.0.0"},
			},
		},
		{
			name:     "empty content",
			content:  "",
			expected: nil,
		},
		{
			name: "comments only",
			content: `// This is a comment
// Another comment
`,
			expected: nil,
		},
		{
			name: "mixed single and block require",
			content: `module example.com/mymodule

require github.com/single/pkg v1.0.0

require (
	github.com/block/pkg v2.0.0
)
`,
			expected: []GoModDependency{
				{Path: "github.com/single/pkg", Version: "v1.0.0"},
				{Path: "github.com/block/pkg", Version: "v2.0.0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGoMod(tt.content)

			if len(result) != len(tt.expected) {
				t.Errorf("parseGoMod() returned %d deps, want %d", len(result), len(tt.expected))
				return
			}

			for i, dep := range result {
				if dep.Path != tt.expected[i].Path {
					t.Errorf("parseGoMod()[%d].Path = %q, want %q", i, dep.Path, tt.expected[i].Path)
				}
				if dep.Version != tt.expected[i].Version {
					t.Errorf("parseGoMod()[%d].Version = %q, want %q", i, dep.Version, tt.expected[i].Version)
				}
			}
		})
	}
}

func TestParseGradleDependencies(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedDeps    []GradleDependency
		expectedSkipped int
	}{
		{
			name:    "single quotes",
			content: `implementation 'com.example:library:1.0.0'`,
			expectedDeps: []GradleDependency{
				{Group: "com.example", Name: "library", Version: "1.0.0"},
			},
			expectedSkipped: 0,
		},
		{
			name:    "double quotes",
			content: `implementation "com.example:library:1.0.0"`,
			expectedDeps: []GradleDependency{
				{Group: "com.example", Name: "library", Version: "1.0.0"},
			},
			expectedSkipped: 0,
		},
		{
			name:    "parentheses syntax",
			content: `implementation("com.example:library:1.0.0")`,
			expectedDeps: []GradleDependency{
				{Group: "com.example", Name: "library", Version: "1.0.0"},
			},
			expectedSkipped: 0,
		},
		{
			name:    "map syntax",
			content: `implementation group: 'com.example', name: 'library', version: '1.0.0'`,
			expectedDeps: []GradleDependency{
				{Group: "com.example", Name: "library", Version: "1.0.0"},
			},
			expectedSkipped: 0,
		},
		{
			name: "multiple configurations",
			content: `
dependencies {
    implementation 'com.example:impl:1.0.0'
    api 'com.example:api:2.0.0'
    testImplementation 'com.example:test:3.0.0'
    runtimeOnly 'com.example:runtime:4.0.0'
}
`,
			expectedDeps: []GradleDependency{
				{Group: "com.example", Name: "impl", Version: "1.0.0"},
				{Group: "com.example", Name: "api", Version: "2.0.0"},
				{Group: "com.example", Name: "test", Version: "3.0.0"},
				{Group: "com.example", Name: "runtime", Version: "4.0.0"},
			},
			expectedSkipped: 0,
		},
		{
			name:            "property reference skipped",
			content:         `implementation "com.example:library:$libraryVersion"`,
			expectedDeps:    nil,
			expectedSkipped: 1,
		},
		{
			name: "mixed with property refs",
			content: `
implementation 'com.example:fixed:1.0.0'
implementation "com.example:variable:$version"
`,
			expectedDeps: []GradleDependency{
				{Group: "com.example", Name: "fixed", Version: "1.0.0"},
			},
			expectedSkipped: 1,
		},
		{
			name:            "empty content",
			content:         "",
			expectedDeps:    nil,
			expectedSkipped: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, skipped := parseGradleDependencies(tt.content)

			if len(deps) != len(tt.expectedDeps) {
				t.Errorf("parseGradleDependencies() returned %d deps, want %d", len(deps), len(tt.expectedDeps))
				return
			}

			if len(skipped) != tt.expectedSkipped {
				t.Errorf("parseGradleDependencies() skipped %d, want %d", len(skipped), tt.expectedSkipped)
			}

			for i, dep := range deps {
				if dep.Group != tt.expectedDeps[i].Group {
					t.Errorf("parseGradleDependencies()[%d].Group = %q, want %q", i, dep.Group, tt.expectedDeps[i].Group)
				}
				if dep.Name != tt.expectedDeps[i].Name {
					t.Errorf("parseGradleDependencies()[%d].Name = %q, want %q", i, dep.Name, tt.expectedDeps[i].Name)
				}
				if dep.Version != tt.expectedDeps[i].Version {
					t.Errorf("parseGradleDependencies()[%d].Version = %q, want %q", i, dep.Version, tt.expectedDeps[i].Version)
				}
			}
		})
	}
}

func TestFilterRepositories(t *testing.T) {
	repos := []RepoInfo{
		{Name: "repo1", FullName: "owner/repo1"},
		{Name: "repo2", FullName: "owner/repo2"},
		{Name: "repo3", FullName: "other/repo3"},
		{Name: "myproject", FullName: "org/myproject"},
	}

	tests := []struct {
		name     string
		filter   string
		expected []string
	}{
		{
			name:     "empty filter returns all",
			filter:   "",
			expected: []string{"owner/repo1", "owner/repo2", "other/repo3", "org/myproject"},
		},
		{
			name:     "filter by name",
			filter:   "repo1",
			expected: []string{"owner/repo1"},
		},
		{
			name:     "filter by full name",
			filter:   "owner/repo2",
			expected: []string{"owner/repo2"},
		},
		{
			name:     "multiple filters",
			filter:   "repo1, repo3",
			expected: []string{"owner/repo1", "other/repo3"},
		},
		{
			name:     "case insensitive",
			filter:   "REPO1, Owner/Repo2",
			expected: []string{"owner/repo1", "owner/repo2"},
		},
		{
			name:     "no matches",
			filter:   "nonexistent",
			expected: nil,
		},
		{
			name:     "with whitespace",
			filter:   "  repo1  ,  repo2  ",
			expected: []string{"owner/repo1", "owner/repo2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterRepositories(repos, tt.filter)

			if len(result) != len(tt.expected) {
				t.Errorf("filterRepositories() returned %d repos, want %d", len(result), len(tt.expected))
				return
			}

			for i, repo := range result {
				if repo.FullName != tt.expected[i] {
					t.Errorf("filterRepositories()[%d] = %q, want %q", i, repo.FullName, tt.expected[i])
				}
			}
		})
	}
}
