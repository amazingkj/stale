package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type HealthHandler struct {
	db *sqlx.DB
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Checks   map[string]string `json:"checks,omitempty"`
	Version  string            `json:"version,omitempty"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	checks := make(map[string]string)
	status := "ok"

	// Check database connection
	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := h.db.PingContext(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			status = "degraded"
		} else {
			checks["database"] = "healthy"
		}
	}

	response := HealthResponse{
		Status:  status,
		Checks:  checks,
		Version: "1.0.0",
	}

	if status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response)
}
