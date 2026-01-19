-- Ignored dependencies table
CREATE TABLE IF NOT EXISTS ignored_dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    ecosystem TEXT,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, ecosystem)
);
