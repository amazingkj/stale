package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/repository"
	"github.com/jiin/stale/internal/service/cache"
)

type DependencyHandler struct {
	repo       *repository.DependencyRepository
	statsCache *cache.Cache[*domain.DependencyStats]
	reposCache *cache.Cache[[]string]
}

func NewDependencyHandler(repo *repository.DependencyRepository) *DependencyHandler {
	return &DependencyHandler{
		repo:       repo,
		statsCache: cache.New[*domain.DependencyStats](2 * time.Minute),
		reposCache: cache.New[[]string](5 * time.Minute),
	}
}

func (h *DependencyHandler) List(w http.ResponseWriter, r *http.Request) {
	outdated := r.URL.Query().Get("outdated")

	if outdated == "true" {
		deps, err := h.repo.GetUpgradable(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if deps == nil {
			deps = []domain.DependencyWithRepo{}
		}
		json.NewEncoder(w).Encode(deps)
		return
	}

	deps, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if deps == nil {
		deps = []domain.DependencyWithRepo{}
	}
	json.NewEncoder(w).Encode(deps)
}

func (h *DependencyHandler) ListPaginated(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	upgradableOnly := r.URL.Query().Get("upgradable") == "true"
	repoFilter := r.URL.Query().Get("repo")
	ecosystemFilter := r.URL.Query().Get("ecosystem")
	search := r.URL.Query().Get("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	result, err := h.repo.GetPaginated(r.Context(), page, limit, upgradableOnly, repoFilter, ecosystemFilter, search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.Data == nil {
		result.Data = []domain.DependencyWithRepo{}
	}
	json.NewEncoder(w).Encode(result)
}

func (h *DependencyHandler) GetUpgradable(w http.ResponseWriter, r *http.Request) {
	deps, err := h.repo.GetUpgradable(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if deps == nil {
		deps = []domain.DependencyWithRepo{}
	}
	json.NewEncoder(w).Encode(deps)
}

func (h *DependencyHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	// Check cache first
	if stats, found := h.statsCache.Get("stats"); found {
		json.NewEncoder(w).Encode(stats)
		return
	}

	stats, err := h.repo.GetStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache the result
	h.statsCache.Set("stats", stats)
	json.NewEncoder(w).Encode(stats)
}

func (h *DependencyHandler) GetRepositoryNames(w http.ResponseWriter, r *http.Request) {
	// Check cache first
	if names, found := h.reposCache.Get("repos"); found {
		json.NewEncoder(w).Encode(names)
		return
	}

	names, err := h.repo.GetRepositoryNames(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if names == nil {
		names = []string{}
	}

	// Cache the result
	h.reposCache.Set("repos", names)
	json.NewEncoder(w).Encode(names)
}

func (h *DependencyHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	repoFilter := r.URL.Query().Get("repo")

	// Support legacy outdated parameter
	if r.URL.Query().Get("outdated") == "true" && filter == "" {
		filter = "upgradable"
	}

	// Get filtered dependencies directly from database for better performance
	deps, err := h.repo.GetFiltered(r.Context(), filter, repoFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if deps == nil {
		deps = []domain.DependencyWithRepo{}
	}

	// Build filename
	var filenameParts []string

	// Add repository name if filtered
	if repoFilter != "" {
		// Replace / with _ for valid filename
		repoName := repoFilter
		for _, char := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"} {
			repoName = strings.ReplaceAll(repoName, char, "_")
		}
		filenameParts = append(filenameParts, repoName)
	}

	// Add filter type
	switch filter {
	case "upgradable":
		filenameParts = append(filenameParts, "upgradable")
	case "uptodate":
		filenameParts = append(filenameParts, "uptodate")
	case "prod":
		filenameParts = append(filenameParts, "production")
	case "dev":
		filenameParts = append(filenameParts, "development")
	}

	filenameParts = append(filenameParts, "dependencies")
	filenameParts = append(filenameParts, time.Now().Format("2006-01-02"))

	filename := strings.Join(filenameParts, "_") + ".csv"

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header row
	header := []string{"Repository", "Source", "Dependency", "Ecosystem", "Type", "Current Version", "Latest Version", "Upgradable"}
	writer.Write(header)

	// Write data rows
	for _, dep := range deps {
		upgradable := "No"
		if dep.IsOutdated {
			upgradable = "Yes"
		}

		row := []string{
			dep.RepoFullName,
			dep.SourceName,
			dep.Name,
			dep.Ecosystem,
			dep.Type,
			dep.CurrentVersion,
			dep.LatestVersion,
			upgradable,
		}
		writer.Write(row)
	}
}
