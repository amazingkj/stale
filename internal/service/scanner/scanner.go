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
	"github.com/jiin/stale/internal/service/golang"
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
	goClient    *golang.Client
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
		goClient:    golang.New(),
	}
}

func (s *Scanner) ScanAll(ctx context.Context, scanID int64) error {
	sources, err := s.sourceRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	var totalRepos, totalDeps int32

	for _, source := range sources {
		err := s.scanSource(ctx, source, scanID, &totalRepos, &totalDeps)
		if err != nil {
			log.Error().Err(err).Str("source", source.Name).Msg("failed to scan source")
			continue
		}
		_ = s.sourceRepo.UpdateLastScan(ctx, source.ID)
	}

	return nil
}

func (s *Scanner) ScanSource(ctx context.Context, sourceID, scanID int64) error {
	source, err := s.sourceRepo.GetByID(ctx, sourceID)
	if err != nil {
		return err
	}

	var totalRepos, totalDeps int32
	err = s.scanSource(ctx, *source, scanID, &totalRepos, &totalDeps)
	if err != nil {
		return err
	}

	_ = s.sourceRepo.UpdateLastScan(ctx, sourceID)
	return nil
}

func (s *Scanner) scanSource(ctx context.Context, source domain.Source, scanID int64, totalRepos, totalDeps *int32) error {
	var provider GitProvider

	switch source.Type {
	case "gitlab":
		glClient := gitlab.New(source.Token, source.URL, source.Organization, source.InsecureSkipVerify, source.MembershipOnly)
		provider = &GitLabAdapter{client: glClient}
	default: // github
		ghClient := github.New(source.Token, source.Organization, source.OwnerOnly)
		provider = &GitHubAdapter{client: ghClient}
	}

	repos, err := provider.ListRepositories(ctx)
	if err != nil {
		return err
	}

	log.Info().Int("total_repos", len(repos)).Str("source", source.Name).Msg("fetched repositories from source")

	// Filter repos if specific repositories are configured
	if source.Repositories != "" {
		beforeFilter := len(repos)
		repos = filterRepositories(repos, source.Repositories)
		log.Info().Int("before", beforeFilter).Int("after", len(repos)).Str("filter", source.Repositories).Msg("filtered repositories")
	}

	if len(repos) == 0 {
		log.Warn().Str("source", source.Name).Msg("no repositories to scan")
		return nil
	}

	for _, repo := range repos {
		// Use source.ScanBranch if set, otherwise use repo's default branch
		scanBranch := repo.DefaultBranch
		if source.ScanBranch != "" {
			scanBranch = source.ScanBranch
		}

		log.Info().Str("repo", repo.FullName).Str("branch", scanBranch).Msg("scanning repository")
		repoEntity := domain.Repository{
			SourceID:      source.ID,
			Name:          repo.Name,
			FullName:      repo.FullName,
			DefaultBranch: scanBranch,
			HTMLURL:       repo.HTMLURL,
		}

		var foundManifest bool
		var repoDeps int32

		// Fetch all manifest files in parallel for better performance
		type manifestResult struct {
			name    string
			content []byte
		}

		manifestFiles := []string{"package.json", "pom.xml", "build.gradle", "build.gradle.kts", "go.mod"}
		results := make(chan manifestResult, len(manifestFiles))

		for _, file := range manifestFiles {
			go func(f string) {
				content, err := provider.GetFileContent(ctx, repo.FullName, f, scanBranch)
				if err != nil {
					results <- manifestResult{f, nil}
				} else {
					results <- manifestResult{f, content}
				}
			}(file)
		}

		// Collect results
		var packageJSON, pomXML, buildGradle, buildGradleKts, goMod []byte
		for i := 0; i < len(manifestFiles); i++ {
			result := <-results
			if result.content != nil {
				log.Info().Str("repo", repo.FullName).Str("file", result.name).Msg("found manifest")
				foundManifest = true
				switch result.name {
				case "package.json":
					packageJSON = result.content
					repoEntity.HasPackageJSON = true
				case "pom.xml":
					pomXML = result.content
					repoEntity.HasPomXML = true
				case "build.gradle":
					buildGradle = result.content
					repoEntity.HasBuildGradle = true
				case "build.gradle.kts":
					buildGradleKts = result.content
					repoEntity.HasBuildGradle = true
				case "go.mod":
					goMod = result.content
					repoEntity.HasGoMod = true
				}
			}
		}

		// Skip if no manifest found
		if !foundManifest {
			log.Info().Str("repo", repo.FullName).Msg("no supported manifest file found (package.json, pom.xml, build.gradle, go.mod)")
			continue
		}

		// Upsert repository and delete old dependencies
		repoID, err := s.repoRepo.Upsert(ctx, repoEntity)
		if err != nil {
			log.Error().Err(err).Str("repo", repo.FullName).Msg("failed to upsert repository")
			continue
		}

		// Delete old dependencies before adding new ones
		if err := s.depRepo.DeleteByRepoID(ctx, repoID); err != nil {
			log.Error().Err(err).Str("repo", repo.FullName).Msg("failed to delete old dependencies")
		}

		// Process each manifest type
		if packageJSON != nil {
			var pkg PackageJSON
			if err := json.Unmarshal(packageJSON, &pkg); err == nil {
				deps := s.processNpmDependencies(ctx, repoID, pkg.Dependencies, "dependency")
				deps += s.processNpmDependencies(ctx, repoID, pkg.DevDependencies, "devDependency")
				atomic.AddInt32(&repoDeps, int32(deps))
			}
		}

		if pomXML != nil {
			var pom PomXML
			if err := xml.Unmarshal(pomXML, &pom); err == nil {
				deps := s.processMavenDependencies(ctx, repoID, pom)
				atomic.AddInt32(&repoDeps, int32(deps))
			}
		}

		if buildGradle != nil {
			deps := s.processGradleDependencies(ctx, repoID, string(buildGradle))
			atomic.AddInt32(&repoDeps, int32(deps))
		}

		if buildGradleKts != nil {
			deps := s.processGradleDependencies(ctx, repoID, string(buildGradleKts))
			atomic.AddInt32(&repoDeps, int32(deps))
		}

		if goMod != nil {
			deps := s.processGoDependencies(ctx, repoID, string(goMod))
			atomic.AddInt32(&repoDeps, int32(deps))
		}

		atomic.AddInt32(totalRepos, 1)
		atomic.AddInt32(totalDeps, repoDeps)
		log.Info().Str("repo", repo.FullName).Int32("deps", repoDeps).Msg("repository scanned successfully")

		// Update stats in real-time after each repository
		_ = s.scanRepo.UpdateStats(ctx, scanID, int(atomic.LoadInt32(totalRepos)), int(atomic.LoadInt32(totalDeps)))
	}

	return nil
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
	var skipped int32

	// Process regular dependencies
	for _, dep := range pom.Dependencies.Dependency {
		if dep.Version == "" || strings.HasPrefix(dep.Version, "${") {
			atomic.AddInt32(&skipped, 1)
			log.Debug().
				Str("groupId", dep.GroupID).
				Str("artifactId", dep.ArtifactID).
				Str("version", dep.Version).
				Msg("skipping Maven dependency with property reference or empty version")
			continue
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

	if skipped > 0 {
		log.Info().Int32("skipped", skipped).Int32("processed", count).Msg("Maven dependencies with property references were skipped")
	}

	return int(count)
}

func (s *Scanner) processGradleDependencies(ctx context.Context, repoID int64, content string) int {
	deps, skipped := parseGradleDependencies(content)

	if len(skipped) > 0 {
		for _, dep := range skipped {
			log.Debug().Str("dependency", dep).Msg("skipping Gradle dependency with property reference")
		}
		log.Info().Int("skipped", len(skipped)).Msg("Gradle dependencies with property references were skipped")
	}

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
// Returns parsed dependencies and a list of skipped dependency names (those with property references)
func parseGradleDependencies(content string) ([]GradleDependency, []string) {
	var deps []GradleDependency
	var skipped []string

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
					skipped = append(skipped, match[1]+":"+match[2])
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

	return deps, skipped
}

// GoModDependency represents a parsed Go module dependency
type GoModDependency struct {
	Path    string
	Version string
}

func (s *Scanner) processGoDependencies(ctx context.Context, repoID int64, content string) int {
	deps := parseGoMod(content)
	if len(deps) == 0 {
		return 0
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	var count int32

	for _, dep := range deps {
		wg.Add(1)
		go func(d GoModDependency) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			latest, err := s.goClient.GetLatestVersion(ctx, d.Path)
			if err != nil {
				latest = ""
			}

			depEntity := domain.Dependency{
				RepositoryID:   repoID,
				Name:           d.Path,
				CurrentVersion: d.Version,
				LatestVersion:  latest,
				Type:           "dependency",
				Ecosystem:      "go",
				IsOutdated:     isOutdated(d.Version, latest),
			}

			if err := s.depRepo.Upsert(ctx, depEntity); err != nil {
				log.Error().Err(err).Str("dep", depEntity.Name).Msg("failed to upsert go dependency")
				return
			}

			atomic.AddInt32(&count, 1)
		}(dep)
	}

	wg.Wait()
	return int(count)
}

// parseGoMod parses go.mod content and extracts dependencies
func parseGoMod(content string) []GoModDependency {
	var deps []GoModDependency

	lines := strings.Split(content, "\n")
	inRequireBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(line, "//") {
			continue
		}

		// Check for require block start
		if strings.HasPrefix(line, "require (") || line == "require(" {
			inRequireBlock = true
			continue
		}

		// Check for block end
		if line == ")" && inRequireBlock {
			inRequireBlock = false
			continue
		}

		// Parse single-line require
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				deps = append(deps, GoModDependency{
					Path:    parts[1],
					Version: parts[2],
				})
			}
			continue
		}

		// Parse dependencies inside require block
		if inRequireBlock && line != "" {
			// Remove // indirect comment
			if idx := strings.Index(line, "//"); idx != -1 {
				line = strings.TrimSpace(line[:idx])
			}

			parts := strings.Fields(line)
			if len(parts) >= 2 {
				deps = append(deps, GoModDependency{
					Path:    parts[0],
					Version: parts[1],
				})
			}
		}
	}

	return deps
}

// filterRepositories filters repos based on comma-separated list of repo names
func filterRepositories(repos []RepoInfo, filter string) []RepoInfo {
	if filter == "" {
		return repos
	}

	// Parse filter list (supports both "repo" and "owner/repo" formats)
	allowedRepos := make(map[string]bool)
	for _, r := range strings.Split(filter, ",") {
		name := strings.TrimSpace(r)
		if name != "" {
			allowedRepos[strings.ToLower(name)] = true
		}
	}

	var filtered []RepoInfo
	for _, repo := range repos {
		// Check both full name (owner/repo) and just repo name
		if allowedRepos[strings.ToLower(repo.FullName)] || allowedRepos[strings.ToLower(repo.Name)] {
			filtered = append(filtered, repo)
		}
	}

	return filtered
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
