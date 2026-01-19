package repository

import (
	"context"

	"github.com/jiin/stale/internal/domain"
	"github.com/jmoiron/sqlx"
)

type IgnoredRepository struct {
	db *sqlx.DB
}

func NewIgnoredRepository(db *sqlx.DB) *IgnoredRepository {
	return &IgnoredRepository{db: db}
}

func (r *IgnoredRepository) GetAll(ctx context.Context) ([]domain.IgnoredDependency, error) {
	var ignored []domain.IgnoredDependency
	err := r.db.SelectContext(ctx, &ignored,
		"SELECT * FROM ignored_dependencies ORDER BY name")
	if err != nil {
		return nil, err
	}
	return ignored, nil
}

func (r *IgnoredRepository) Create(ctx context.Context, input *domain.IgnoredDependencyInput) (*domain.IgnoredDependency, error) {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO ignored_dependencies (name, ecosystem, reason) VALUES (?, ?, ?)",
		input.Name, input.Ecosystem, input.Reason)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &domain.IgnoredDependency{
		ID:        id,
		Name:      input.Name,
		Ecosystem: input.Ecosystem,
		Reason:    input.Reason,
	}, nil
}

func (r *IgnoredRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM ignored_dependencies WHERE id = ?", id)
	return err
}

func (r *IgnoredRepository) IsIgnored(ctx context.Context, name, ecosystem string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM ignored_dependencies WHERE name = ? AND (ecosystem = ? OR ecosystem = '' OR ecosystem IS NULL)",
		name, ecosystem)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetIgnoredNames returns a set of ignored dependency names for quick lookup
func (r *IgnoredRepository) GetIgnoredNames(ctx context.Context) (map[string]bool, error) {
	var ignored []domain.IgnoredDependency
	err := r.db.SelectContext(ctx, &ignored, "SELECT name, ecosystem FROM ignored_dependencies")
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for _, ig := range ignored {
		// Key format: "name" or "name:ecosystem"
		result[ig.Name] = true
		if ig.Ecosystem != "" {
			result[ig.Name+":"+ig.Ecosystem] = true
		}
	}
	return result, nil
}
