package repository

import (
	"context"
	"time"

	"github.com/jiin/stale/internal/domain"
	"github.com/jmoiron/sqlx"
)

type DependencyRepository struct {
	db *sqlx.DB
}

func NewDependencyRepository(db *sqlx.DB) *DependencyRepository {
	return &DependencyRepository{db: db}
}

func (r *DependencyRepository) Upsert(ctx context.Context, dep domain.Dependency) error {
	query := `INSERT INTO dependencies (repository_id, name, current_version, latest_version, type, ecosystem, is_outdated, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)
              ON CONFLICT(repository_id, name, type) DO UPDATE SET
                  current_version = excluded.current_version,
                  latest_version = excluded.latest_version,
                  ecosystem = excluded.ecosystem,
                  is_outdated = excluded.is_outdated,
                  updated_at = excluded.updated_at`

	ecosystem := dep.Ecosystem
	if ecosystem == "" {
		ecosystem = "npm"
	}

	_, err := r.db.ExecContext(ctx, query,
		dep.RepositoryID, dep.Name, dep.CurrentVersion, dep.LatestVersion,
		dep.Type, ecosystem, dep.IsOutdated, time.Now())
	return err
}

func (r *DependencyRepository) GetByRepoID(ctx context.Context, repoID int64) ([]domain.Dependency, error) {
	var deps []domain.Dependency
	err := r.db.SelectContext(ctx, &deps,
		"SELECT * FROM dependencies WHERE repository_id = ? ORDER BY name", repoID)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func (r *DependencyRepository) GetAll(ctx context.Context) ([]domain.DependencyWithRepo, error) {
	query := `SELECT d.*, r.name as repo_name, r.full_name as repo_full_name, s.name as source_name
              FROM dependencies d
              JOIN repositories r ON d.repository_id = r.id
              JOIN sources s ON r.source_id = s.id
              ORDER BY d.name`

	var deps []domain.DependencyWithRepo
	err := r.db.SelectContext(ctx, &deps, query)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func (r *DependencyRepository) GetOutdated(ctx context.Context) ([]domain.DependencyWithRepo, error) {
	query := `SELECT d.*, r.name as repo_name, r.full_name as repo_full_name, s.name as source_name
              FROM dependencies d
              JOIN repositories r ON d.repository_id = r.id
              JOIN sources s ON r.source_id = s.id
              WHERE d.is_outdated = TRUE
              ORDER BY d.name`

	var deps []domain.DependencyWithRepo
	err := r.db.SelectContext(ctx, &deps, query)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func (r *DependencyRepository) GetStats(ctx context.Context) (*domain.DependencyStats, error) {
	var total, outdated int

	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM dependencies")
	if err != nil {
		return nil, err
	}

	err = r.db.GetContext(ctx, &outdated, "SELECT COUNT(*) FROM dependencies WHERE is_outdated = TRUE")
	if err != nil {
		return nil, err
	}

	type typeCount struct {
		Type  string `db:"type"`
		Count int    `db:"count"`
	}
	var typeCounts []typeCount
	err = r.db.SelectContext(ctx, &typeCounts,
		"SELECT type, COUNT(*) as count FROM dependencies GROUP BY type")
	if err != nil {
		return nil, err
	}

	byType := make(map[string]int)
	for _, tc := range typeCounts {
		byType[tc.Type] = tc.Count
	}

	return &domain.DependencyStats{
		TotalDependencies: total,
		OutdatedCount:     outdated,
		UpToDateCount:     total - outdated,
		ByType:            byType,
	}, nil
}

func (r *DependencyRepository) DeleteByRepoID(ctx context.Context, repoID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM dependencies WHERE repository_id = ?", repoID)
	return err
}

func (r *DependencyRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM dependencies")
	return count, err
}
