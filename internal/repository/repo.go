package repository

import (
	"context"
	"time"

	"github.com/jiin/stale/internal/domain"
	"github.com/jmoiron/sqlx"
)

type RepoRepository struct {
	db *sqlx.DB
}

func NewRepoRepository(db *sqlx.DB) *RepoRepository {
	return &RepoRepository{db: db}
}

func (r *RepoRepository) Upsert(ctx context.Context, repo domain.Repository) (int64, error) {
	query := `INSERT INTO repositories (source_id, name, full_name, default_branch, html_url, has_package_json, has_pom_xml, has_build_gradle, has_go_mod, created_at, updated_at, last_scan_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
              ON CONFLICT(full_name) DO UPDATE SET
                  name = excluded.name,
                  default_branch = excluded.default_branch,
                  html_url = excluded.html_url,
                  has_package_json = excluded.has_package_json,
                  has_pom_xml = excluded.has_pom_xml,
                  has_build_gradle = excluded.has_build_gradle,
                  has_go_mod = excluded.has_go_mod,
                  updated_at = excluded.updated_at,
                  last_scan_at = excluded.last_scan_at
              RETURNING id`

	now := time.Now()
	var id int64
	err := r.db.GetContext(ctx, &id, query,
		repo.SourceID, repo.Name, repo.FullName, repo.DefaultBranch,
		repo.HTMLURL, repo.HasPackageJSON, repo.HasPomXML, repo.HasBuildGradle, repo.HasGoMod, now, now, now)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *RepoRepository) GetAll(ctx context.Context) ([]domain.Repository, error) {
	query := `SELECT r.*,
		COALESCE((SELECT COUNT(*) FROM dependencies d WHERE d.repository_id = r.id), 0) as dependency_count,
		COALESCE((SELECT COUNT(*) FROM dependencies d WHERE d.repository_id = r.id AND d.is_outdated = TRUE), 0) as outdated_count
		FROM repositories r
		ORDER BY r.full_name`
	var repos []domain.Repository
	err := r.db.SelectContext(ctx, &repos, query)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func (r *RepoRepository) GetBySourceID(ctx context.Context, sourceID int64) ([]domain.Repository, error) {
	query := `SELECT r.*,
		COALESCE((SELECT COUNT(*) FROM dependencies d WHERE d.repository_id = r.id), 0) as dependency_count,
		COALESCE((SELECT COUNT(*) FROM dependencies d WHERE d.repository_id = r.id AND d.is_outdated = TRUE), 0) as outdated_count
		FROM repositories r
		WHERE r.source_id = ?
		ORDER BY r.full_name`
	var repos []domain.Repository
	err := r.db.SelectContext(ctx, &repos, query, sourceID)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func (r *RepoRepository) GetByID(ctx context.Context, id int64) (*domain.Repository, error) {
	var repo domain.Repository
	err := r.db.GetContext(ctx, &repo, "SELECT * FROM repositories WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *RepoRepository) DeleteBySourceID(ctx context.Context, sourceID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM repositories WHERE source_id = ?", sourceID)
	return err
}

func (r *RepoRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM repositories")
	return count, err
}

func (r *RepoRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM repositories WHERE id = ?", id)
	return err
}
