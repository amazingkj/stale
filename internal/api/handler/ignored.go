package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/repository"
)

type IgnoredHandler struct {
	repo *repository.IgnoredRepository
}

func NewIgnoredHandler(repo *repository.IgnoredRepository) *IgnoredHandler {
	return &IgnoredHandler{repo: repo}
}

func (h *IgnoredHandler) List(w http.ResponseWriter, r *http.Request) {
	ignored, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ignored == nil {
		ignored = []domain.IgnoredDependency{}
	}
	json.NewEncoder(w).Encode(ignored)
}

func (h *IgnoredHandler) Create(w http.ResponseWriter, r *http.Request) {
	LimitBody(r)
	var input domain.IgnoredDependencyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	ignored, err := h.repo.Create(r.Context(), &input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ignored)
}

func (h *IgnoredHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// BulkCreate adds multiple dependencies to the ignore list
func (h *IgnoredHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	LimitBody(r)
	var input struct {
		Items []domain.IgnoredDependencyInput `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(input.Items) == 0 {
		http.Error(w, "items array is required", http.StatusBadRequest)
		return
	}

	var created []domain.IgnoredDependency
	var skipped int
	var duplicates int

	ctx := r.Context()
	for _, item := range input.Items {
		if item.Name == "" {
			skipped++
			continue
		}
		ignored, err := h.repo.Create(ctx, &item)
		if err != nil {
			// Count duplicates separately (UNIQUE constraint violation)
			duplicates++
			continue
		}
		created = append(created, *ignored)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"created":    len(created),
		"skipped":    skipped,
		"duplicates": duplicates,
		"items":      created,
	})
}

// BulkDelete removes multiple dependencies from the ignore list
func (h *IgnoredHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	LimitBody(r)
	var input struct {
		IDs []int64 `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(input.IDs) == 0 {
		http.Error(w, "ids array is required", http.StatusBadRequest)
		return
	}

	var deleted int
	var failed int
	ctx := r.Context()
	for _, id := range input.IDs {
		if err := h.repo.Delete(ctx, id); err == nil {
			deleted++
		} else {
			failed++
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": deleted,
		"failed":  failed,
	})
}
