package scanner

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Masterminds/semver/v3"
	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/repository"
	"github.com/jiin/stale/internal/service/github"
	"github.com/jiin/stale/internal/service/gitlab"
	"github.com/jiin/stale/internal/service/maven"
	"github.com/jiin/stale/internal/service/npm"
	"github.com/rs/zerolog/log"
)

// GitProvider is an interface for Git hosting providers (GitHub, GitLab)
type GitProvider interface {
	ListRepositories(ctx context.Context) ([]RepoInfo, error)
	GetFileContent(ctx context.Context, repoPath, filePath, ref string) ([]byte, error)
}

// RepoInfo contains common repository information
type RepoInfo struct {
	Name          string
	FullName      string
	DefaultBranch string
	HTMLURL       string
}

// GitHubAdapter adapts github.Client to GitProvider
type GitHubAdapter struct {
	client *github.Client
}

func (a *GitHubAdapter) ListRepositories(ctx context.Context) ([]RepoInfo, error) {
	repos, err := a.client.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]RepoInfo, len(repos))
	for i, r := range repos {
		result[i] = RepoInfo{
			Name:          r.Name,
			FullName:      r.FullName,
			DefaultBranch: r.DefaultBranch,
			HTMLURL:       r.HTMLURL,
		}
	}
	return result, nil
}

func (a *GitHubAdapter) GetFileContent(ctx context.Context, repoPath, filePath, ref string) ([]byte, error) {
	return a.client.GetFileContent(ctx, repoPath, filePath, ref)
}

// GitLabAdapter adapts gitlab.Client to GitProvider
type GitLabAdapter struct {
	client *gitlab.Client
}

func (a *GitLabAdapter) ListRepositories(ctx context.Context) ([]RepoInfo, error) {
	repos, err := a.client.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]RepoInfo, len(repos))
	for i, r := range repos {
		result[i] = RepoInfo{
			Name:          r.Name,
			FullName:      r.FullName,
			DefaultBranch: r.DefaultBranch,
			HTMLURL:       r.WebURL,
		}
	}
	return result, nil
}

func (a *GitLabAdapter) GetFileContent(ctx context.Context, repoPath, filePath, ref string) ([]byte, error) {
	return a.client.GetFileContent(ctx, repoPath, filePath, ref)
}

type Scanner struct {
	sourceRepo  *repository.SourceRepository
	repoRepo    *repository.RepoRepository
	depRepo     *repository.DependencyRepository
	scanRepo    *repository.ScanRepository
	npmClient   *npm.Client
	mavenClient *maven.Client
}

type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// PomXML represents a Maven pom.xml file
type PomXML struct {
	XMLName      xml.Name `xml:"project"`
	Dependencies struct {
		Dependency []struct {
			GroupID    string `xml:"groupId"`
			ArtifactID string `xml:"artifactId"`
			Version    string `xml:"version"`
			Scope      string `xml:"scope"`
		} `xml:"dependency"`
	} `xml:"dependencies"`
	DependencyManagement struct {
		Dependencies struct {
			Dependency []struct {
				GroupID    string `xml:"groupId"`
				ArtifactID string `xml:"artifactId"`
				Version    string `xml:"version"`
			} `xml:"dependency"`
		} `xml:"dependencies"`
	} `xml:"dependencyManagement"`
}

// GradleDependency represents a parsed Gradle dependency
type GradleDependency struct {
	Group    string
	Name     string
	Version  string
	IsPlugin bool
}

func New(
	sourceRepo *repository.SourceRepository,
	repoRepo *repository.RepoRepository,
	depRepo *repository.DependencyRepository,
	scanRepo *repository.ScanRepository,
) *Scanner {
	return &Scanner{
		sourceRepo:  sourceRepo,
		repoRepo:    repoRepo,
		depRepo:     depRepo,
		scanRepo:    scanRepo,
		npmClient:   npm.New(),
		mavenClient: maven.New(),
	}
}

func (s *Scanner) ScanAll(ctx context.Context, scanID int64) error {
	sources, err := s.sourceRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	var totalRepos, totalDeps int32

	for _, source := range sources {
		repos, deps, err := s.scanSource(ctx, source)
		if err != nil {
			log.Error().Err(err).Str("source", source.Name).Msg("failed to scan source")
			continue
		}
		atomic.AddInt32(&totalRepos, int32(repos))
		atomic.AddInt32(&totalDeps, int32(deps))

		_ = s.sourceRepo.UpdateLastScan(ctx, source.ID)
	}

	_ = s.scanRepo.UpdateStats(ctx, scanID, int(totalRepos), int(totalDeps))
	return nil
}

func (s *Scanner) ScanSource(ctx context.Context, sourceID, scanID int64) error {
	source, err := s.sourceRepo.GetByID(ctx, sourceID)
	if err != nil {
		return err
	}

	repos, deps, err := s.scanSource(ctx, *source)
	if err != nil {
		return err
	}

	_ = s.sourceRepo.UpdateLastScan(ctx, sourceID)
	_ = s.scanRepo.UpdateStats(ctx, scanID, repos, deps)
	return nil
}

func (s *Scanner) scanSource(ctx context.Context, source domain.Source) (int, int, error) {
	var provider GitProvider

	switch source.Type {
	case "gitlab":
		glClient := gitlab.New(source.Token, source.URL, source.Organization)
		provider = &GitLabAdapter{client: glClient}
	default: // github
		ghClient := github.New(source.Token, source.Organization)
		provider = &GitHubAdapter{client: ghClient}
	}

	repos, err := provider.ListRepositories(ctx)
	if err != nil {
		return 0, 0, err
	}

	var repoCount, depCount int32

	for _, repo := range repos {
		repoEntity := domain.Repository{
			SourceID:      source.ID,
			Name:          repo.Name,
			FullName:      repo.FullName,
			DefaultBranch: repo.DefaultBranch,
			HTMLURL:       repo.HTMLURL,
		}

		var foundManifest bool
		var totalDeps int32

		// Check for package.json (npm)
		if content, err := provider.GetFileContent(ctx, repo.FullName, "package.json", repo.DefaultBranch); err == nil {
			var pkg PackageJSON
			if err := json.Unmarshal(content, &pkg); err == nil {
				repoEntity.HasPackageJSON = true
				foundManifest = true

				repoID, err := s.repoRepo.Upsert(ctx, repoEntity)
				if err != nil {
					log.Error().Err(err).Str("repo", repo.FullName).Msg("failed to upsert repository")
					continue
				}

				deps := s.processNpmDependencies(ctx, repoID, pkg.Dependencies, "dependency")
				deps += s.processNpmDependencies(ctx, repoID, pkg.DevDependencies, "devDependency")
				atomic.AddInt32(&totalDeps, int32(deps))
			}
		}

		// Check for pom.xml (Maven)
		if content, err := provider.GetFileContent(ctx, repo.FullName, "pom.xml", repo.DefaultBranch); err == nil {
			var pom PomXML
			if err := xml.Unmarshal(content, &pom); err == nil {
				repoEntity.HasPomXML = true
				foundManifest = true

				repoID, err := s.repoRepo.Upsert(ctx, repoEntity)
				if err != nil {
					log.Error().Err(err).Str("repo", repo.FullName).Msg("failed to upsert repository")
					continue
				}

				deps := s.processMavenDependencies(ctx, repoID, pom)
				atomic.AddInt32(&totalDeps, int32(deps))
			}
		}

		// Check for build.gradle (Gradle)
		if content, err := provider.GetFileContent(ctx, repo.FullName, "build.gradle", repo.DefaultBranch); err == nil {
			repoEntity.HasBuildGradle = true
			foundManifest = true

			repoID, err := s.repoRepo.Upsert(ctx, repoEntity)
			if err != nil {
				log.Error().Err(err).Str("repo", repo.FullName).Msg("failed to upsert repository")
				continue
			}

			deps := s.processGradleDependencies(ctx, repoID, string(content))
			atomic.AddInt32(&totalDeps, int32(deps))
		}

		// Also check for build.gradle.kts (Kotlin DSL)
		if content, err := provider.GetFileContent(ctx, repo.FullName, "build.gradle.kts", repo.DefaultBranch); err == nil {
			repoEntity.HasBuildGradle = true
			foundManifest = true

			repoID, err := s.repoRepo.Upsert(ctx, repoEntity)
			if err != nil {
				log.Error().Err(err).Str("repo", repo.FullName).Msg("failed to upsert repository")
				continue
			}

			deps := s.processGradleDependencies(ctx, repoID, string(content))
			atomic.AddInt32(&totalDeps, int32(deps))
		}

		if foundManifest {
			atomic.AddInt32(&repoCount, 1)
			atomic.AddInt32(&depCount, totalDeps)
		}
	}

	return int(repoCount), int(depCount), nil
}

func (s *Scanner) processNpmDependencies(ctx context.Context, repoID int64, deps map[string]string, depType string) int {
	if len(deps) == 0 {
		return 0
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrent npm requests
	var count int32

	for name, version := range deps {
		wg.Add(1)
		go func(name, version string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			cleanedVersion := cleanVersion(version)
			latest, err := s.npmClient.GetLatestVersion(ctx, name)
			if err != nil {
				latest = ""
			}

			dep := domain.Dependency{
				RepositoryID:   repoID,
				Name:           name,
				CurrentVersion: cleanedVersion,
				LatestVersion:  latest,
				Type:           depType,
				Ecosystem:      "npm",
				IsOutdated:     isOutdated(cleanedVersion, latest),
			}

			if err := s.depRepo.Upsert(ctx, dep); err != nil {
				log.Error().Err(err).Str("dep", name).Msg("failed to upsert dependency")
				return
			}

			atomic.AddInt32(&count, 1)
		}(name, version)
	}

	wg.Wait()
	return int(count)
}

func (s *Scanner) processMavenDependencies(ctx context.Context, repoID int64, pom PomXML) int {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	var count int32

	// Process regular dependencies
	for _, dep := range pom.Dependencies.Dependency {
		if dep.Version == "" || strings.HasPrefix(dep.Version, "${") {
			continue // Skip dependencies with property references
		}

		wg.Add(1)
		go func(groupID, artifactID, version, scope string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			latest, err := s.mavenClient.GetLatestVersion(ctx, groupID, artifactID)
			if err != nil {
				latest = ""
			}

			depType := "dependency"
			if scope == "test" {
				depType = "devDependency"
			}

			d := domain.Dependency{
				RepositoryID:   repoID,
				Name:           groupID + ":" + artifactID,
				CurrentVersion: version,
				LatestVersion:  latest,
				Type:           depType,
				Ecosystem:      "maven",
				IsOutdated:     isOutdated(version, latest),
			}

			if err := s.depRepo.Upsert(ctx, d); err != nil {
				log.Error().Err(err).Str("dep", d.Name).Msg("failed to upsert maven dependency")
				return
			}

			atomic.AddInt32(&count, 1)
		}(dep.GroupID, dep.ArtifactID, dep.Version, dep.Scope)
	}

	wg.Wait()
	return int(count)
}

func (s *Scanner) processGradleDependencies(ctx context.Context, repoID int64, content string) int {
	deps := parseGradleDependencies(content)
	if len(deps) == 0 {
		return 0
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	var count int32

	for _, dep := range deps {
		wg.Add(1)
		go func(d GradleDependency) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			latest, err := s.mavenClient.GetLatestVersion(ctx, d.Group, d.Name)
			if err != nil {
				latest = ""
			}

			depEntity := domain.Dependency{
				RepositoryID:   repoID,
				Name:           d.Group + ":" + d.Name,
				CurrentVersion: d.Version,
				LatestVersion:  latest,
				Type:           "dependency",
				Ecosystem:      "gradle",
				IsOutdated:     isOutdated(d.Version, latest),
			}

			if err := s.depRepo.Upsert(ctx, depEntity); err != nil {
				log.Error().Err(err).Str("dep", depEntity.Name).Msg("failed to upsert gradle dependency")
				return
			}

			atomic.AddInt32(&count, 1)
		}(dep)
	}

	wg.Wait()
	return int(count)
}

// parseGradleDependencies extracts dependencies from build.gradle content
func parseGradleDependencies(content string) []GradleDependency {
	var deps []GradleDependency

	// Match patterns like: implementation 'group:name:version'
	// or implementation "group:name:version"
	// or implementation("group:name:version")
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:implementation|api|compile|testImplementation|testCompile|runtimeOnly|compileOnly)\s*\(?['"]([^:'"]+):([^:'"]+):([^'"]+)['"]\)?`),
		regexp.MustCompile(`(?:implementation|api|compile|testImplementation|testCompile|runtimeOnly|compileOnly)\s+group:\s*['"]([^'"]+)['"]\s*,\s*name:\s*['"]([^'"]+)['"]\s*,\s*version:\s*['"]([^'"]+)['"]`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 4 {
				// Skip property references like $version
				if strings.Contains(match[3], "$") {
					continue
				}
				deps = append(deps, GradleDependency{
					Group:   match[1],
					Name:    match[2],
					Version: match[3],
				})
			}
		}
	}

	return deps
}

func cleanVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "^")
	version = strings.TrimPrefix(version, "~")
	version = strings.TrimPrefix(version, ">=")
	version = strings.TrimPrefix(version, ">")
	version = strings.TrimPrefix(version, "<=")
	version = strings.TrimPrefix(version, "<")
	version = strings.TrimPrefix(version, "=")

	if strings.Contains(version, " ") {
		version = strings.Split(version, " ")[0]
	}
	if strings.Contains(version, "||") {
		version = strings.TrimSpace(strings.Split(version, "||")[0])
	}

	return version
}

func isOutdated(current, latest string) bool {
	if current == "" || latest == "" {
		return false
	}

	currentVer, err := semver.NewVersion(current)
	if err != nil {
		return false
	}

	latestVer, err := semver.NewVersion(latest)
	if err != nil {
		return false
	}

	return currentVer.LessThan(latestVer)
}
