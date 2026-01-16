package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/repository"
	"github.com/jiin/stale/internal/service/scheduler"
)

type ScanHandler struct {
	repo      *repository.ScanRepository
	scheduler *scheduler.Scheduler
}

func NewScanHandler(repo *repository.ScanRepository, scheduler *scheduler.Scheduler) *ScanHandler {
	return &ScanHandler{repo: repo, scheduler: scheduler}
}

type TriggerScanRequest struct {
	SourceID *int64 `json:"source_id,omitempty"`
}

func (h *ScanHandler) TriggerScan(w http.ResponseWriter, r *http.Request) {
	var req TriggerScanRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	scan, err := h.scheduler.TriggerScan(r.Context(), req.SourceID)
	if err != nil {
		if errors.Is(err, scheduler.ErrScanAlreadyRunning) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(scan)
}

func (h *ScanHandler) List(w http.ResponseWriter, r *http.Request) {
	scans, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if scans == nil {
		scans = []domain.ScanJob{}
	}
	json.NewEncoder(w).Encode(scans)
}

func (h *ScanHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	scan, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(scan)
}

func (h *ScanHandler) GetRunning(w http.ResponseWriter, r *http.Request) {
	scan, err := h.repo.GetLatestRunning(r.Context())
	if err != nil {
		// No running scan - return null
		w.Write([]byte("null"))
		return
	}
	json.NewEncoder(w).Encode(scan)
}
