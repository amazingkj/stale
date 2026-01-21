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
	LimitBody(r)
	var req TriggerScanRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondBadRequest(w, "invalid request body")
			return
		}
	}

	scan, err := h.scheduler.TriggerScan(r.Context(), req.SourceID)
	if err != nil {
		if errors.Is(err, scheduler.ErrScanAlreadyRunning) {
			RespondError(w, http.StatusConflict, "a scan is already running", nil)
			return
		}
		RespondInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(scan)
}

func (h *ScanHandler) List(w http.ResponseWriter, r *http.Request) {
	scans, err := h.repo.GetAll(r.Context())
	if err != nil {
		RespondInternalError(w, err)
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
		RespondBadRequest(w, "invalid id")
		return
	}

	scan, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		RespondNotFound(w, "scan not found")
		return
	}
	json.NewEncoder(w).Encode(scan)
}

func (h *ScanHandler) GetRunning(w http.ResponseWriter, r *http.Request) {
	// First, cleanup any stale scans that have been running too long
	_, _ = h.repo.CleanupStaleScans(r.Context())

	scan, err := h.repo.GetLatestRunning(r.Context())
	if err != nil {
		// No running scan - return null
		w.Write([]byte("null"))
		return
	}
	json.NewEncoder(w).Encode(scan)
}

func (h *ScanHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		RespondBadRequest(w, "invalid id")
		return
	}

	// Get the scan first to check status
	scan, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		RespondNotFound(w, "scan not found")
		return
	}

	// Only cancel pending or running scans
	if scan.Status != domain.ScanStatusPending && scan.Status != domain.ScanStatusRunning {
		RespondBadRequest(w, "scan is already completed or failed")
		return
	}

	// Mark as failed with cancelled message
	if err := h.repo.UpdateStatus(r.Context(), id, domain.ScanStatusFailed, errors.New("cancelled by user")); err != nil {
		RespondInternalError(w, err)
		return
	}

	// Clear the scheduler's running job ID if this is the current scan
	h.scheduler.ClearRunningJob(id)

	w.WriteHeader(http.StatusNoContent)
}
