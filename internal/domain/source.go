package domain

import "time"

type Source struct {
	ID           int64      `db:"id" json:"id"`
	Name         string     `db:"name" json:"name"`
	Type         string     `db:"type" json:"type"`       // github or gitlab
	Token        string     `db:"token" json:"-"`
	Organization string     `db:"organization" json:"organization,omitempty"` // GitHub org or GitLab group
	URL          string     `db:"url" json:"url,omitempty"`                   // For self-hosted GitLab
	Repositories string     `db:"repositories" json:"repositories,omitempty"` // Comma-separated list of repos to scan (empty = all)
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	LastScanAt   *time.Time `db:"last_scan_at" json:"last_scan_at,omitempty"`
}

type SourceInput struct {
	Name         string `json:"name"`
	Type         string `json:"type"`                     // github or gitlab
	Token        string `json:"token"`
	Organization string `json:"organization,omitempty"`   // GitHub org or GitLab group
	URL          string `json:"url,omitempty"`            // For self-hosted GitLab
	Repositories string `json:"repositories,omitempty"`   // Comma-separated list of repos to scan (empty = all)
}
