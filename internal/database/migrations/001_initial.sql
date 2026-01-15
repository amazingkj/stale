-- Sources table (GitHub tokens/orgs)
CREATE TABLE IF NOT EXISTS sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'github',
    token TEXT NOT NULL,
    organization TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_scan_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_sources_type ON sources(type);

-- Repositories table
CREATE TABLE IF NOT EXISTS repositories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id INTEGER NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    full_name TEXT NOT NULL UNIQUE,
    default_branch TEXT NOT NULL DEFAULT 'main',
    html_url TEXT NOT NULL,
    has_package_json BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_scan_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_repositories_source_id ON repositories(source_id);
CREATE INDEX IF NOT EXISTS idx_repositories_has_package_json ON repositories(has_package_json);

-- Dependencies table
CREATE TABLE IF NOT EXISTS dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repository_id INTEGER NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    current_version TEXT NOT NULL,
    latest_version TEXT,
    type TEXT NOT NULL DEFAULT 'dependency',
    is_outdated BOOLEAN DEFAULT FALSE,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(repository_id, name, type)
);

CREATE INDEX IF NOT EXISTS idx_dependencies_repository_id ON dependencies(repository_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_is_outdated ON dependencies(is_outdated);
CREATE INDEX IF NOT EXISTS idx_dependencies_name ON dependencies(name);

-- Scan jobs table
CREATE TABLE IF NOT EXISTS scan_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id INTEGER REFERENCES sources(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    repos_found INTEGER DEFAULT 0,
    deps_found INTEGER DEFAULT 0,
    error TEXT,
    started_at DATETIME,
    finished_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_scan_jobs_status ON scan_jobs(status);
CREATE INDEX IF NOT EXISTS idx_scan_jobs_source_id ON scan_jobs(source_id);
