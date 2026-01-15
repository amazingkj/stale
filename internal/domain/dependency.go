package domain

import "time"

type Dependency struct {
	ID             int64     `db:"id" json:"id"`
	RepositoryID   int64     `db:"repository_id" json:"repository_id"`
	Name           string    `db:"name" json:"name"`
	CurrentVersion string    `db:"current_version" json:"current_version"`
	LatestVersion  string    `db:"latest_version" json:"latest_version"`
	Type           string    `db:"type" json:"type"`
	Ecosystem      string    `db:"ecosystem" json:"ecosystem"` // npm, maven, gradle
	IsOutdated     bool      `db:"is_outdated" json:"is_outdated"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type DependencyWithRepo struct {
	Dependency
	RepoName     string `db:"repo_name" json:"repo_name"`
	RepoFullName string `db:"repo_full_name" json:"repo_full_name"`
	SourceName   string `db:"source_name" json:"source_name"`
}

type DependencyStats struct {
	TotalDependencies int            `json:"total_dependencies"`
	OutdatedCount     int            `json:"outdated_count"`
	UpToDateCount     int            `json:"up_to_date_count"`
	ByType            map[string]int `json:"by_type"`
}
