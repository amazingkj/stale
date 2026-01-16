package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/repository"
	"github.com/jiin/stale/internal/service/email"
	"github.com/jiin/stale/internal/service/scheduler"
	"github.com/robfig/cron/v3"
)

type SettingsHandler struct {
	repo         *repository.SettingsRepository
	scheduler    *scheduler.Scheduler
	emailService *email.Service
}

func NewSettingsHandler(
	repo *repository.SettingsRepository,
	scheduler *scheduler.Scheduler,
	emailService *email.Service,
) *SettingsHandler {
	return &SettingsHandler{
		repo:         repo,
		scheduler:    scheduler,
		emailService: emailService,
	}
}

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := h.repo.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Mask SMTP password in response
	if settings.EmailSMTPPass != "" {
		settings.EmailSMTPPass = "********"
	}

	json.NewEncoder(w).Encode(settings)
}

func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input domain.SettingsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate cron expression if provided
	if input.ScheduleCron != nil {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(*input.ScheduleCron); err != nil {
			http.Error(w, "invalid cron expression: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Don't update password if it's the masked value
	if input.EmailSMTPPass != nil && *input.EmailSMTPPass == "********" {
		input.EmailSMTPPass = nil
	}

	if err := h.repo.Update(r.Context(), &input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reload scheduler if schedule settings changed
	if input.ScheduleEnabled != nil || input.ScheduleCron != nil {
		h.scheduler.ReloadSchedule()
	}

	// Return updated settings
	settings, err := h.repo.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Mask SMTP password in response
	if settings.EmailSMTPPass != "" {
		settings.EmailSMTPPass = "********"
	}

	json.NewEncoder(w).Encode(settings)
}

func (h *SettingsHandler) TestEmail(w http.ResponseWriter, r *http.Request) {
	settings, err := h.repo.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if settings.EmailSMTPHost == "" {
		http.Error(w, "SMTP host not configured", http.StatusBadRequest)
		return
	}

	if err := h.emailService.TestConnection(settings); err != nil {
		http.Error(w, "Failed to send test email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Test email sent successfully"})
}

type NextScanResponse struct {
	Enabled  bool    `json:"enabled"`
	NextRun  *string `json:"next_run,omitempty"`
	CronExpr string  `json:"cron_expr"`
}

func (h *SettingsHandler) GetNextScan(w http.ResponseWriter, r *http.Request) {
	settings, err := h.repo.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := NextScanResponse{
		Enabled:  settings.ScheduleEnabled,
		CronExpr: settings.ScheduleCron,
	}

	if settings.ScheduleEnabled {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		schedule, err := parser.Parse(settings.ScheduleCron)
		if err == nil {
			nextTime := schedule.Next(time.Now())
			nextStr := nextTime.Format("2006-01-02 15:04:05")
			response.NextRun = &nextStr
		}
	}

	json.NewEncoder(w).Encode(response)
}
