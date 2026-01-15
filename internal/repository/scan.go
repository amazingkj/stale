package repository

import (
	"context"
	"time"

	"github.com/jiin/stale/internal/domain"
	"github.com/jmoiron/sqlx"
)

type ScanRepository struct {
	db *sqlx.DB
}

func NewScanRepository(db *sqlx.DB) *ScanRepository {
	return &ScanRepository{db: db}
}

func (r *ScanRepository) Create(ctx context.Context, sourceID *int64) (*domain.ScanJob, error) {
	query := `INSERT INTO scan_jobs (source_id, status, created_at)
              VALUES (?, ?, ?)
              RETURNING *`

	var scan domain.ScanJob
	err := r.db.GetContext(ctx, &scan, query, sourceID, domain.ScanStatusPending, time.Now())
	if err != nil {
		return nil, err
	}
	return &scan, nil
}

func (r *ScanRepository) GetByID(ctx context.Context, id int64) (*domain.ScanJob, error) {
	var scan domain.ScanJob
	err := r.db.GetContext(ctx, &scan, "SELECT * FROM scan_jobs WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &scan, nil
}

func (r *ScanRepository) GetAll(ctx context.Context) ([]domain.ScanJob, error) {
	var scans []domain.ScanJob
	err := r.db.SelectContext(ctx, &scans, "SELECT * FROM scan_jobs ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		return nil, err
	}
	return scans, nil
}

func (r *ScanRepository) UpdateStatus(ctx context.Context, id int64, status domain.ScanStatus, err error) error {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	now := time.Now()
	var query string
	var args []any

	if status == domain.ScanStatusRunning {
		query = "UPDATE scan_jobs SET status = ?, started_at = ?, error = ? WHERE id = ?"
		args = []any{status, now, errStr, id}
	} else {
		query = "UPDATE scan_jobs SET status = ?, finished_at = ?, error = ? WHERE id = ?"
		args = []any{status, now, errStr, id}
	}

	_, execErr := r.db.ExecContext(ctx, query, args...)
	return execErr
}

func (r *ScanRepository) UpdateStats(ctx context.Context, id int64, reposFound, depsFound int) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE scan_jobs SET repos_found = ?, deps_found = ? WHERE id = ?",
		reposFound, depsFound, id)
	return err
}

func (r *ScanRepository) GetLatestRunning(ctx context.Context) (*domain.ScanJob, error) {
	var scan domain.ScanJob
	err := r.db.GetContext(ctx, &scan,
		"SELECT * FROM scan_jobs WHERE status = ? ORDER BY created_at DESC LIMIT 1",
		domain.ScanStatusRunning)
	if err != nil {
		return nil, err
	}
	return &scan, nil
}
