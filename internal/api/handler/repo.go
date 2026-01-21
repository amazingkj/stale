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
			RespondBadRequest(w, "invalid source_id")
			return
		}
		repos, err := h.repo.GetBySourceID(r.Context(), sourceID)
		if err != nil {
			RespondInternalError(w, err)
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
		RespondInternalError(w, err)
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
		RespondBadRequest(w, "invalid id")
		return
	}

	repo, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		RespondNotFound(w, "repository not found")
		return
	}
	json.NewEncoder(w).Encode(repo)
}

func (h *RepoHandler) GetDependencies(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		RespondBadRequest(w, "invalid id")
		return
	}

	deps, err := h.depRepo.GetByRepoID(r.Context(), id)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	if deps == nil {
		deps = []domain.Dependency{}
	}
	json.NewEncoder(w).Encode(deps)
}

func (h *RepoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		RespondBadRequest(w, "invalid id")
		return
	}

	// Delete dependencies first
	if err := h.depRepo.DeleteByRepoID(r.Context(), id); err != nil {
		RespondInternalError(w, err)
		return
	}

	// Delete repository
	if err := h.repo.Delete(r.Context(), id); err != nil {
		RespondInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type BulkDeleteRequest struct {
	IDs []int64 `json:"ids"`
}

type BulkDeleteError struct {
	ID    int64  `json:"id"`
	Error string `json:"error"`
}

type BulkDeleteResponse struct {
	Deleted int               `json:"deleted"`
	Failed  int               `json:"failed"`
	Errors  []BulkDeleteError `json:"errors,omitempty"`
}

func (h *RepoHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	LimitBody(r)
	var req BulkDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		RespondBadRequest(w, "no ids provided")
		return
	}

	ctx := r.Context()
	response := BulkDeleteResponse{
		Errors: []BulkDeleteError{},
	}

	for _, id := range req.IDs {
		// Delete dependencies first
		if err := h.depRepo.DeleteByRepoID(ctx, id); err != nil {
			response.Failed++
			response.Errors = append(response.Errors, BulkDeleteError{
				ID:    id,
				Error: "failed to delete dependencies: " + err.Error(),
			})
			continue
		}

		// Delete repository
		if err := h.repo.Delete(ctx, id); err != nil {
			response.Failed++
			response.Errors = append(response.Errors, BulkDeleteError{
				ID:    id,
				Error: "failed to delete repository: " + err.Error(),
			})
			continue
		}
		response.Deleted++
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
