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
