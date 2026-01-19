package repository

import (
	"context"
	"testing"

	"github.com/jiin/stale/internal/domain"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func setupSettingsTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	// Create the settings table with default values
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		INSERT OR IGNORE INTO settings (key, value) VALUES
			('schedule_enabled', 'false'),
			('schedule_cron', '0 9 * * *'),
			('email_enabled', 'false'),
			('email_smtp_host', ''),
			('email_smtp_port', '587'),
			('email_smtp_user', ''),
			('email_smtp_pass', ''),
			('email_from', ''),
			('email_to', ''),
			('email_notify_new_outdated', 'true');
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

// Helper functions for creating pointers
func boolPtr(b bool) *bool       { return &b }
func strPtr(s string) *string    { return &s }
func intPtr(i int) *int          { return &i }

func TestSettingsRepository_Get(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	settings, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Verify default values
	if settings.ScheduleEnabled {
		t.Error("Get() ScheduleEnabled should be false by default")
	}
	if settings.ScheduleCron != "0 9 * * *" {
		t.Errorf("Get() ScheduleCron = %q, want %q", settings.ScheduleCron, "0 9 * * *")
	}
	if settings.EmailEnabled {
		t.Error("Get() EmailEnabled should be false by default")
	}
	if settings.EmailSMTPPort != 587 {
		t.Errorf("Get() EmailSMTPPort = %d, want %d", settings.EmailSMTPPort, 587)
	}
	if !settings.EmailNotifyNewOutdated {
		t.Error("Get() EmailNotifyNewOutdated should be true by default")
	}
}

func TestSettingsRepository_Update(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// Update schedule settings
	input := &domain.SettingsInput{
		ScheduleEnabled: boolPtr(true),
		ScheduleCron:    strPtr("0 */6 * * *"),
	}

	err := repo.Update(ctx, input)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify updates
	settings, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !settings.ScheduleEnabled {
		t.Error("Update() ScheduleEnabled should be true")
	}
	if settings.ScheduleCron != "0 */6 * * *" {
		t.Errorf("Update() ScheduleCron = %q, want %q", settings.ScheduleCron, "0 */6 * * *")
	}

	// Verify other settings unchanged
	if settings.EmailEnabled {
		t.Error("Update() should not change EmailEnabled")
	}
}

func TestSettingsRepository_Update_EmailSettings(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	input := &domain.SettingsInput{
		EmailEnabled:  boolPtr(true),
		EmailSMTPHost: strPtr("smtp.gmail.com"),
		EmailSMTPPort: intPtr(465),
		EmailSMTPUser: strPtr("user@gmail.com"),
		EmailFrom:     strPtr("noreply@example.com"),
		EmailTo:       strPtr("team@example.com"),
	}

	err := repo.Update(ctx, input)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	settings, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !settings.EmailEnabled {
		t.Error("Update() EmailEnabled should be true")
	}
	if settings.EmailSMTPHost != "smtp.gmail.com" {
		t.Errorf("Update() EmailSMTPHost = %q, want %q", settings.EmailSMTPHost, "smtp.gmail.com")
	}
	if settings.EmailSMTPPort != 465 {
		t.Errorf("Update() EmailSMTPPort = %d, want %d", settings.EmailSMTPPort, 465)
	}
	if settings.EmailSMTPUser != "user@gmail.com" {
		t.Errorf("Update() EmailSMTPUser = %q, want %q", settings.EmailSMTPUser, "user@gmail.com")
	}
	if settings.EmailFrom != "noreply@example.com" {
		t.Errorf("Update() EmailFrom = %q, want %q", settings.EmailFrom, "noreply@example.com")
	}
	if settings.EmailTo != "team@example.com" {
		t.Errorf("Update() EmailTo = %q, want %q", settings.EmailTo, "team@example.com")
	}
}

func TestSettingsRepository_UpdatePersistence(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// Update
	repo.Update(ctx, &domain.SettingsInput{
		ScheduleEnabled: boolPtr(true),
		ScheduleCron:    strPtr("0 18 * * *"),
	})

	// Create new repo instance to verify persistence
	repo2 := NewSettingsRepository(db)
	settings, err := repo2.Get(ctx)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !settings.ScheduleEnabled {
		t.Error("Settings should persist after update")
	}
	if settings.ScheduleCron != "0 18 * * *" {
		t.Errorf("ScheduleCron = %q, want %q", settings.ScheduleCron, "0 18 * * *")
	}
}

func TestSettingsRepository_PartialUpdate(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// First update - enable schedule
	repo.Update(ctx, &domain.SettingsInput{
		ScheduleEnabled: boolPtr(true),
	})

	// Second update - change cron without touching enabled
	repo.Update(ctx, &domain.SettingsInput{
		ScheduleCron: strPtr("0 12 * * *"),
	})

	settings, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Both should be applied
	if !settings.ScheduleEnabled {
		t.Error("ScheduleEnabled should still be true")
	}
	if settings.ScheduleCron != "0 12 * * *" {
		t.Errorf("ScheduleCron = %q, want %q", settings.ScheduleCron, "0 12 * * *")
	}
}

func TestSettingsRepository_EmptyUpdate(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// Empty update should not error
	err := repo.Update(ctx, &domain.SettingsInput{})
	if err != nil {
		t.Fatalf("Empty Update() should not error, got %v", err)
	}

	// Verify defaults unchanged
	settings, _ := repo.Get(ctx)
	if settings.ScheduleEnabled {
		t.Error("Settings should remain unchanged after empty update")
	}
}

func TestParseIntOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		def      int
		expected int
	}{
		{"valid number", "123", 0, 123},
		{"empty string", "", 587, 587},
		{"invalid string", "abc", 587, 587},
		{"negative number", "-1", 0, -1},
		{"zero", "0", 99, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIntOrDefault(tt.input, tt.def)
			if result != tt.expected {
				t.Errorf("parseIntOrDefault(%q, %d) = %d, want %d", tt.input, tt.def, result, tt.expected)
			}
		})
	}
}

func TestBoolToStr(t *testing.T) {
	if boolToStr(true) != "true" {
		t.Error("boolToStr(true) should return 'true'")
	}
	if boolToStr(false) != "false" {
		t.Error("boolToStr(false) should return 'false'")
	}
}
