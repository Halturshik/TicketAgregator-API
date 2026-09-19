package postgres

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pressly/goose"
)

func TestMigrationsAreOrderedAndCollectable(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	dir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
	migrations, err := goose.CollectMigrations(dir, 0, goose.MaxVersion)
	if err != nil {
		t.Fatalf("collect migrations: %v", err)
	}
	if len(migrations) != 9 {
		t.Fatalf("migration count = %d, want 9", len(migrations))
	}
	for i, migration := range migrations {
		wantVersion := int64(i + 1)
		if migration.Version != wantVersion {
			t.Fatalf("migration[%d].Version = %d, want %d", i, migration.Version, wantVersion)
		}
	}
}

func TestSupplierMigrationsAreOrderedAndCollectable(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	dir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "supplier-migrations")
	migrations, err := goose.CollectMigrations(dir, 0, goose.MaxVersion)
	if err != nil {
		t.Fatalf("collect supplier migrations: %v", err)
	}
	if len(migrations) != 1 || migrations[0].Version != 1 {
		t.Fatalf("supplier migrations = %+v, want one migration with version 1", migrations)
	}
}
