package repository

import (
	"context"
	"time"

	"github.com/jiin/stale/internal/domain"
	"github.com/jmoiron/sqlx"
)

type SourceRepository struct {
	db *sqlx.DB
}

func NewSourceRepository(db *sqlx.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

func (r *SourceRepository) Create(ctx context.Context, input domain.SourceInput) (*domain.Source, error) {
	query := `INSERT INTO sources (name, type, token, organization, url, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?)
              RETURNING id, name, type, token, organization, url, created_at, updated_at, last_scan_at`

	now := time.Now()
	var source domain.Source
	err := r.db.GetContext(ctx, &source, query, input.Name, input.Type, input.Token, input.Organization, input.URL, now, now)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *SourceRepository) GetAll(ctx context.Context) ([]domain.Source, error) {
	var sources []domain.Source
	err := r.db.SelectContext(ctx, &sources, "SELECT * FROM sources ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	return sources, nil
}

func (r *SourceRepository) GetByID(ctx context.Context, id int64) (*domain.Source, error) {
	var source domain.Source
	err := r.db.GetContext(ctx, &source, "SELECT * FROM sources WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *SourceRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sources WHERE id = ?", id)
	return err
}

func (r *SourceRepository) UpdateLastScan(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE sources SET last_scan_at = ?, updated_at = ? WHERE id = ?",
		time.Now(), time.Now(), id)
	return err
}
