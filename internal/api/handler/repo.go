package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/repository"
)

type RepoHandler struct {
	repo    *repository.RepoRepository
	depRepo *repository.DependencyRepository
}

func NewRepoHandler(repo *repository.RepoRepository, depRepo *repository.DependencyRepository) *RepoHandler {
	return &RepoHandler{repo: repo, depRepo: depRepo}
}

func (h *RepoHandler) List(w http.ResponseWriter, r *http.Request) {
	sourceIDStr := r.URL.Query().Get("source_id")

	if sourceIDStr != "" {
		sourceID, err := strconv.ParseInt(sourceIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid source_id", http.StatusBadRequest)
			return
		}
		repos, err := h.repo.GetBySourceID(r.Context(), sourceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if repos == nil {
			repos = []domain.Repository{}
		}
		json.NewEncoder(w).Encode(repos)
		return
	}

	repos, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if repos == nil {
		repos = []domain.Repository{}
	}
	json.NewEncoder(w).Encode(repos)
}

func (h *RepoHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	repo, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(repo)
}

func (h *RepoHandler) GetDependencies(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	deps, err := h.depRepo.GetByRepoID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if deps == nil {
		deps = []domain.Dependency{}
	}
	json.NewEncoder(w).Encode(deps)
}
