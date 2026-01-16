package repository

import (
	"context"
	"strconv"

	"github.com/jiin/stale/internal/domain"
	"github.com/jmoiron/sqlx"
)

type SettingsRepository struct {
	db *sqlx.DB
}

func NewSettingsRepository(db *sqlx.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get(ctx context.Context) (*domain.Settings, error) {
	rows, err := r.db.QueryxContext(ctx, "SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		values[key] = value
	}

	settings := &domain.Settings{
		ScheduleEnabled:        values["schedule_enabled"] == "true",
		ScheduleCron:           values["schedule_cron"],
		EmailEnabled:           values["email_enabled"] == "true",
		EmailSMTPHost:          values["email_smtp_host"],
		EmailSMTPPort:          parseIntOrDefault(values["email_smtp_port"], 587),
		EmailSMTPUser:          values["email_smtp_user"],
		EmailSMTPPass:          values["email_smtp_pass"],
		EmailFrom:              values["email_from"],
		EmailTo:                values["email_to"],
		EmailNotifyNewOutdated: values["email_notify_new_outdated"] != "false",
	}

	return settings, nil
}

func (r *SettingsRepository) Update(ctx context.Context, input *domain.SettingsInput) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateSetting := func(key string, value string) error {
		_, err := tx.ExecContext(ctx,
			"INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
			key, value)
		return err
	}

	if input.ScheduleEnabled != nil {
		if err := updateSetting("schedule_enabled", boolToStr(*input.ScheduleEnabled)); err != nil {
			return err
		}
	}
	if input.ScheduleCron != nil {
		if err := updateSetting("schedule_cron", *input.ScheduleCron); err != nil {
			return err
		}
	}
	if input.EmailEnabled != nil {
		if err := updateSetting("email_enabled", boolToStr(*input.EmailEnabled)); err != nil {
			return err
		}
	}
	if input.EmailSMTPHost != nil {
		if err := updateSetting("email_smtp_host", *input.EmailSMTPHost); err != nil {
			return err
		}
	}
	if input.EmailSMTPPort != nil {
		if err := updateSetting("email_smtp_port", strconv.Itoa(*input.EmailSMTPPort)); err != nil {
			return err
		}
	}
	if input.EmailSMTPUser != nil {
		if err := updateSetting("email_smtp_user", *input.EmailSMTPUser); err != nil {
			return err
		}
	}
	if input.EmailSMTPPass != nil {
		if err := updateSetting("email_smtp_pass", *input.EmailSMTPPass); err != nil {
			return err
		}
	}
	if input.EmailFrom != nil {
		if err := updateSetting("email_from", *input.EmailFrom); err != nil {
			return err
		}
	}
	if input.EmailTo != nil {
		if err := updateSetting("email_to", *input.EmailTo); err != nil {
			return err
		}
	}
	if input.EmailNotifyNewOutdated != nil {
		if err := updateSetting("email_notify_new_outdated", boolToStr(*input.EmailNotifyNewOutdated)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func parseIntOrDefault(s string, def int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return def
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
