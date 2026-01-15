package domain

import "time"

type ScanStatus string

const (
	ScanStatusPending   ScanStatus = "pending"
	ScanStatusRunning   ScanStatus = "running"
	ScanStatusCompleted ScanStatus = "completed"
	ScanStatusFailed    ScanStatus = "failed"
)

type ScanJob struct {
	ID         int64      `db:"id" json:"id"`
	SourceID   *int64     `db:"source_id" json:"source_id,omitempty"`
	Status     ScanStatus `db:"status" json:"status"`
	ReposFound int        `db:"repos_found" json:"repos_found"`
	DepsFound  int        `db:"deps_found" json:"deps_found"`
	Error      string     `db:"error" json:"error,omitempty"`
	StartedAt  *time.Time `db:"started_at" json:"started_at,omitempty"`
	FinishedAt *time.Time `db:"finished_at" json:"finished_at,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}
