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
)

type DependencyHandler struct {
	repo *repository.DependencyRepository
}

func NewDependencyHandler(repo *repository.DependencyRepository) *DependencyHandler {
	return &DependencyHandler{repo: repo}
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

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	result, err := h.repo.GetPaginated(r.Context(), page, limit, upgradableOnly, repoFilter)
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
	stats, err := h.repo.GetStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(stats)
}

func (h *DependencyHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	repoFilter := r.URL.Query().Get("repo")

	// Support legacy outdated parameter
	if r.URL.Query().Get("outdated") == "true" && filter == "" {
		filter = "upgradable"
	}

	// Get all dependencies first
	deps, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if deps == nil {
		deps = []domain.DependencyWithRepo{}
	}

	// Apply filters
	var filtered []domain.DependencyWithRepo
	for _, dep := range deps {
		// Apply repository filter
		if repoFilter != "" && dep.RepoFullName != repoFilter {
			continue
		}

		// Apply status filter
		switch filter {
		case "upgradable":
			if !dep.IsOutdated {
				continue
			}
		case "uptodate":
			if dep.IsOutdated {
				continue
			}
		case "prod":
			if dep.Type != "dependency" {
				continue
			}
		case "dev":
			if dep.Type != "devDependency" {
				continue
			}
		}
		filtered = append(filtered, dep)
	}
	deps = filtered

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
