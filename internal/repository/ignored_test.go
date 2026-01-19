package repository

import (
	"context"
	"testing"

	"github.com/jiin/stale/internal/domain"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	// Create the ignored_dependencies table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ignored_dependencies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			ecosystem TEXT,
			reason TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, ecosystem)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestIgnoredRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewIgnoredRepository(db)
	ctx := context.Background()

	input := &domain.IgnoredDependencyInput{
		Name:      "lodash",
		Ecosystem: "npm",
		Reason:    "known vulnerability in old version",
	}

	result, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if result.ID == 0 {
		t.Error("Create() should return non-zero ID")
	}
	if result.Name != input.Name {
		t.Errorf("Create() name = %q, want %q", result.Name, input.Name)
	}
	if result.Ecosystem != input.Ecosystem {
		t.Errorf("Create() ecosystem = %q, want %q", result.Ecosystem, input.Ecosystem)
	}
	if result.Reason != input.Reason {
		t.Errorf("Create() reason = %q, want %q", result.Reason, input.Reason)
	}
}

func TestIgnoredRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewIgnoredRepository(db)
	ctx := context.Background()

	// Initially empty
	list, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(list) != 0 {
		t.Errorf("GetAll() should return empty list initially, got %d", len(list))
	}

	// Create some entries
	repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "lodash", Ecosystem: "npm"})
	repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "axios", Ecosystem: "npm"})
	repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "junit", Ecosystem: "maven"})

	list, err = repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(list) != 3 {
		t.Errorf("GetAll() should return 3 items, got %d", len(list))
	}

	// Verify sorted by name
	if list[0].Name != "axios" {
		t.Errorf("GetAll() first item should be axios, got %q", list[0].Name)
	}
}

func TestIgnoredRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewIgnoredRepository(db)
	ctx := context.Background()

	// Create an entry
	created, _ := repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "lodash", Ecosystem: "npm"})

	// Delete it
	err := repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	list, _ := repo.GetAll(ctx)
	if len(list) != 0 {
		t.Errorf("Delete() should remove the item, got %d items", len(list))
	}
}

func TestIgnoredRepository_IsIgnored(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewIgnoredRepository(db)
	ctx := context.Background()

	// Add some ignored dependencies
	repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "lodash", Ecosystem: "npm"})
	repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "axios", Ecosystem: ""}) // all ecosystems

	tests := []struct {
		name      string
		depName   string
		ecosystem string
		expected  bool
	}{
		{"exact match", "lodash", "npm", true},
		{"wrong ecosystem", "lodash", "maven", false},
		{"global match any ecosystem", "axios", "npm", true},
		{"global match another ecosystem", "axios", "maven", true},
		{"not ignored", "react", "npm", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.IsIgnored(ctx, tt.depName, tt.ecosystem)
			if err != nil {
				t.Fatalf("IsIgnored() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("IsIgnored(%q, %q) = %v, want %v", tt.depName, tt.ecosystem, result, tt.expected)
			}
		})
	}
}

func TestIgnoredRepository_GetIgnoredNames(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewIgnoredRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "lodash", Ecosystem: "npm"})
	repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "axios", Ecosystem: ""})

	names, err := repo.GetIgnoredNames(ctx)
	if err != nil {
		t.Fatalf("GetIgnoredNames() error = %v", err)
	}

	// Should contain plain names and ecosystem-qualified names
	if !names["lodash"] {
		t.Error("GetIgnoredNames() should contain 'lodash'")
	}
	if !names["lodash:npm"] {
		t.Error("GetIgnoredNames() should contain 'lodash:npm'")
	}
	if !names["axios"] {
		t.Error("GetIgnoredNames() should contain 'axios'")
	}
}

func TestIgnoredRepository_UniqueConstraint(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewIgnoredRepository(db)
	ctx := context.Background()

	// Create first entry
	_, err := repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "lodash", Ecosystem: "npm"})
	if err != nil {
		t.Fatalf("First Create() error = %v", err)
	}

	// Try to create duplicate
	_, err = repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "lodash", Ecosystem: "npm"})
	if err == nil {
		t.Error("Duplicate Create() should fail due to unique constraint")
	}

	// Same name different ecosystem should succeed
	_, err = repo.Create(ctx, &domain.IgnoredDependencyInput{Name: "lodash", Ecosystem: "maven"})
	if err != nil {
		t.Errorf("Different ecosystem Create() should succeed, got error = %v", err)
	}
}
