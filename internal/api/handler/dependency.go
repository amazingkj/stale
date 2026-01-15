package handler

import (
	"encoding/json"
	"net/http"

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
		deps, err := h.repo.GetOutdated(r.Context())
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

func (h *DependencyHandler) GetOutdated(w http.ResponseWriter, r *http.Request) {
	deps, err := h.repo.GetOutdated(r.Context())
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
