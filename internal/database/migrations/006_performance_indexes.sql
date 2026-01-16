-- Performance optimization indexes

-- Compound index for filtering outdated dependencies by ecosystem
CREATE INDEX IF NOT EXISTS idx_dependencies_outdated_ecosystem ON dependencies(is_outdated, ecosystem);

-- Compound index for repository manifest filtering
CREATE INDEX IF NOT EXISTS idx_repositories_source_manifests ON repositories(source_id, has_package_json, has_pom_xml, has_build_gradle, has_go_mod);

-- Compound index for scan jobs ordering by status and date
CREATE INDEX IF NOT EXISTS idx_scan_jobs_status_created ON scan_jobs(status, created_at DESC);

-- Index for dependency type filtering
CREATE INDEX IF NOT EXISTS idx_dependencies_type ON dependencies(type);
