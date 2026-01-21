package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/repository"
	"github.com/jiin/stale/internal/service/github"
	"github.com/jiin/stale/internal/service/gitlab"
)

type SourceHandler struct {
	repo    *repository.SourceRepository
	repoRep *repository.RepoRepository
	depRepo *repository.DependencyRepository
}

func NewSourceHandler(repo *repository.SourceRepository, repoRep *repository.RepoRepository, depRepo *repository.DependencyRepository) *SourceHandler {
	return &SourceHandler{repo: repo, repoRep: repoRep, depRepo: depRepo}
}

func (h *SourceHandler) List(w http.ResponseWriter, r *http.Request) {
	sources, err := h.repo.GetAll(r.Context())
	if err != nil {
		RespondInternalError(w, err)
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
		RespondBadRequest(w, "invalid id")
		return
	}

	source, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		RespondNotFound(w, "source not found")
		return
	}
	json.NewEncoder(w).Encode(source)
}

func (h *SourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	LimitBody(r)
	var input domain.SourceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondBadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" || input.Token == "" {
		RespondBadRequest(w, "name and token are required")
		return
	}

	// Validate and normalize type
	if input.Type == "" {
		input.Type = "github"
	}
	input.Type = strings.ToLower(input.Type)
	if input.Type != "github" && input.Type != "gitlab" {
		RespondBadRequest(w, "type must be 'github' or 'gitlab'")
		return
	}

	// Validate organization name (prevent injection)
	if input.Organization != "" && len(input.Organization) > 100 {
		RespondBadRequest(w, "organization name too long")
		return
	}

	// Validate GitLab URL if provided
	if input.Type == "gitlab" && input.URL != "" {
		parsedURL, err := url.Parse(input.URL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			RespondBadRequest(w, "invalid GitLab URL")
			return
		}
	}

	// Validate token based on source type (use request context for proper timeout)
	ctx := r.Context()
	if input.Type == "gitlab" {
		glClient := gitlab.New(input.Token, input.URL, input.Organization, input.InsecureSkipVerify, input.MembershipOnly)
		if err := glClient.ValidateToken(ctx); err != nil {
			RespondError(w, http.StatusBadRequest, "invalid token: unable to authenticate", err)
			return
		}
	} else {
		ghClient := github.New(input.Token, input.Organization, input.OwnerOnly)
		if err := ghClient.ValidateToken(ctx); err != nil {
			RespondError(w, http.StatusBadRequest, "invalid token: unable to authenticate", err)
			return
		}
	}

	source, err := h.repo.Create(ctx, input)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(source)
}

func (h *SourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		RespondBadRequest(w, "invalid id")
		return
	}

	ctx := r.Context()

	// Cascade delete: dependencies -> repositories -> source
	if err := h.depRepo.DeleteBySourceID(ctx, id); err != nil {
		RespondInternalError(w, err)
		return
	}

	if err := h.repoRep.DeleteBySourceID(ctx, id); err != nil {
		RespondInternalError(w, err)
		return
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		RespondInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		RespondBadRequest(w, "invalid id")
		return
	}

	LimitBody(r)
	var input domain.SourceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondBadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" || input.Token == "" {
		RespondBadRequest(w, "name and token are required")
		return
	}

	// Validate and normalize type
	if input.Type == "" {
		input.Type = "github"
	}
	input.Type = strings.ToLower(input.Type)
	if input.Type != "github" && input.Type != "gitlab" {
		RespondBadRequest(w, "type must be 'github' or 'gitlab'")
		return
	}

	// Validate organization name (prevent injection)
	if input.Organization != "" && len(input.Organization) > 100 {
		RespondBadRequest(w, "organization name too long")
		return
	}

	// Validate GitLab URL if provided
	if input.Type == "gitlab" && input.URL != "" {
		parsedURL, err := url.Parse(input.URL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			RespondBadRequest(w, "invalid GitLab URL")
			return
		}
	}

	// Validate token based on source type (use request context for proper timeout)
	ctx := r.Context()
	if input.Type == "gitlab" {
		glClient := gitlab.New(input.Token, input.URL, input.Organization, input.InsecureSkipVerify, input.MembershipOnly)
		if err := glClient.ValidateToken(ctx); err != nil {
			RespondError(w, http.StatusBadRequest, "invalid token: unable to authenticate", err)
			return
		}
	} else {
		ghClient := github.New(input.Token, input.Organization, input.OwnerOnly)
		if err := ghClient.ValidateToken(ctx); err != nil {
			RespondError(w, http.StatusBadRequest, "invalid token: unable to authenticate", err)
			return
		}
	}

	source, err := h.repo.Update(ctx, id, input)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	json.NewEncoder(w).Encode(source)
}
