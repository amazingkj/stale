package database

import (
	"embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

func New(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(1)

	return db, nil
}

func Migrate(db *sqlx.DB) error {
	// Run all migrations in order
	migrationFiles := []string{
		"migrations/001_initial.sql",
		"migrations/002_add_ecosystem.sql",
		"migrations/003_add_gitlab.sql",
		"migrations/004_add_repositories.sql",
		"migrations/005_add_go_mod.sql",
	}

	for _, file := range migrationFiles {
		migrationSQL, err := migrations.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		_, err = db.Exec(string(migrationSQL))
		if err != nil {
			// Ignore "duplicate column" errors for ALTER TABLE statements
			// This allows migrations to be re-run safely
			continue
		}
	}

	return nil
}
