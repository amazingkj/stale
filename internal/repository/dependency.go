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

func (r *DependencyRepository) GetPaginated(ctx context.Context, page, limit int, upgradableOnly bool, repoFilter string) (*domain.PaginatedDependencies, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Build WHERE clause
	where := "1=1"
	args := []interface{}{}
	if upgradableOnly {
		where += " AND d.is_outdated = TRUE"
	}
	if repoFilter != "" {
		where += " AND r.full_name = ?"
		args = append(args, repoFilter)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM dependencies d
                   JOIN repositories r ON d.repository_id = r.id
                   WHERE ` + where
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Get paginated data
	dataQuery := `SELECT d.*, r.name as repo_name, r.full_name as repo_full_name, s.name as source_name
                  FROM dependencies d
                  JOIN repositories r ON d.repository_id = r.id
                  JOIN sources s ON r.source_id = s.id
                  WHERE ` + where + `
                  ORDER BY d.name
                  LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	var deps []domain.DependencyWithRepo
	err = r.db.SelectContext(ctx, &deps, dataQuery, args...)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.PaginatedDependencies{
		Data:       deps,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (r *DependencyRepository) GetUpgradable(ctx context.Context) ([]domain.DependencyWithRepo, error) {
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

func (r *DependencyRepository) DeleteBySourceID(ctx context.Context, sourceID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM dependencies WHERE repository_id IN (SELECT id FROM repositories WHERE source_id = ?)`,
		sourceID)
	return err
}

// GetFiltered returns dependencies with database-level filtering for better performance
func (r *DependencyRepository) GetFiltered(ctx context.Context, filter, repoFilter string) ([]domain.DependencyWithRepo, error) {
	query := `SELECT d.*, r.name as repo_name, r.full_name as repo_full_name, s.name as source_name
              FROM dependencies d
              JOIN repositories r ON d.repository_id = r.id
              JOIN sources s ON r.source_id = s.id
              WHERE 1=1`
	args := []interface{}{}

	// Apply repository filter
	if repoFilter != "" {
		query += " AND r.full_name = ?"
		args = append(args, repoFilter)
	}

	// Apply status filter
	switch filter {
	case "upgradable":
		query += " AND d.is_outdated = TRUE"
	case "uptodate":
		query += " AND d.is_outdated = FALSE"
	case "prod":
		query += " AND d.type = 'dependency'"
	case "dev":
		query += " AND d.type = 'devDependency'"
	}

	query += " ORDER BY d.name"

	var deps []domain.DependencyWithRepo
	err := r.db.SelectContext(ctx, &deps, query, args...)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

// GetRepositoryNames returns unique repository full names for dropdowns
func (r *DependencyRepository) GetRepositoryNames(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT r.full_name
              FROM repositories r
              JOIN dependencies d ON d.repository_id = r.id
              ORDER BY r.full_name`

	var names []string
	err := r.db.SelectContext(ctx, &names, query)
	if err != nil {
		return nil, err
	}
	return names, nil
}

// MarkPreviouslyOutdated marks currently outdated dependencies before a new scan
func (r *DependencyRepository) MarkPreviouslyOutdated(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "UPDATE dependencies SET previously_outdated = is_outdated")
	return err
}

// GetNewlyOutdated returns dependencies that became outdated in the latest scan
func (r *DependencyRepository) GetNewlyOutdated(ctx context.Context) ([]domain.DependencyWithRepo, error) {
	query := `SELECT d.*, r.name as repo_name, r.full_name as repo_full_name, s.name as source_name
              FROM dependencies d
              JOIN repositories r ON d.repository_id = r.id
              JOIN sources s ON r.source_id = s.id
              WHERE d.is_outdated = TRUE AND (d.previously_outdated = FALSE OR d.previously_outdated IS NULL)
              ORDER BY r.full_name, d.name`

	var deps []domain.DependencyWithRepo
	err := r.db.SelectContext(ctx, &deps, query)
	if err != nil {
		return nil, err
	}
	return deps, nil
}
