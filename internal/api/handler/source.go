package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/repository"
	"github.com/jiin/stale/internal/service/github"
	"github.com/jiin/stale/internal/service/gitlab"
)

type SourceHandler struct {
	repo *repository.SourceRepository
}

func NewSourceHandler(repo *repository.SourceRepository) *SourceHandler {
	return &SourceHandler{repo: repo}
}

func (h *SourceHandler) List(w http.ResponseWriter, r *http.Request) {
	sources, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sources == nil {
		sources = []domain.Source{}
	}
	json.NewEncoder(w).Encode(sources)
}

func (h *SourceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	source, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(source)
}

func (h *SourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.SourceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.Token == "" {
		http.Error(w, "name and token are required", http.StatusBadRequest)
		return
	}

	if input.Type == "" {
		input.Type = "github"
	}

	// Validate token based on source type
	if input.Type == "gitlab" {
		glClient := gitlab.New(input.Token, input.URL, input.Organization)
		if err := glClient.ValidateToken(context.Background()); err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		ghClient := github.New(input.Token, input.Organization)
		if err := ghClient.ValidateToken(context.Background()); err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	source, err := h.repo.Create(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(source)
}

func (h *SourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func (h *SourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var input domain.SourceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.Token == "" {
		http.Error(w, "name and token are required", http.StatusBadRequest)
		return
	}

	if input.Type == "" {
		input.Type = "github"
	}

	// Validate token based on source type
	if input.Type == "gitlab" {
		glClient := gitlab.New(input.Token, input.URL, input.Organization)
		if err := glClient.ValidateToken(context.Background()); err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		ghClient := github.New(input.Token, input.Organization)
		if err := ghClient.ValidateToken(context.Background()); err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	source, err := h.repo.Update(r.Context(), id, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(source)
}
