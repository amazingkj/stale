package domain

import "time"

type Repository struct {
	ID             int64      `db:"id" json:"id"`
	SourceID       int64      `db:"source_id" json:"source_id"`
	Name           string     `db:"name" json:"name"`
	FullName       string     `db:"full_name" json:"full_name"`
	DefaultBranch  string     `db:"default_branch" json:"default_branch"`
	HTMLURL        string     `db:"html_url" json:"html_url"`
	HasPackageJSON bool       `db:"has_package_json" json:"has_package_json"`
	HasPomXML      bool       `db:"has_pom_xml" json:"has_pom_xml"`
	HasBuildGradle bool       `db:"has_build_gradle" json:"has_build_gradle"`
	HasGoMod       bool       `db:"has_go_mod" json:"has_go_mod"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
	LastScanAt     *time.Time `db:"last_scan_at" json:"last_scan_at,omitempty"`
}
