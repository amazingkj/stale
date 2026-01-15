-- Add ecosystem column to dependencies
ALTER TABLE dependencies ADD COLUMN ecosystem TEXT NOT NULL DEFAULT 'npm';

-- Update unique constraint to include ecosystem
-- SQLite doesn't support DROP CONSTRAINT, so we need to recreate the table
-- For now, we'll just add the column and update the index

CREATE INDEX IF NOT EXISTS idx_dependencies_ecosystem ON dependencies(ecosystem);

-- Add manifest detection columns to repositories
ALTER TABLE repositories ADD COLUMN has_pom_xml BOOLEAN DEFAULT FALSE;
ALTER TABLE repositories ADD COLUMN has_build_gradle BOOLEAN DEFAULT FALSE;
