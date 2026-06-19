package models

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"

	gdb "klapp/db"

	_ "modernc.org/sqlite"
)

// newTestDB gives each test its own throwaway SQLite file with migrations
// applied, so tests can run independently and in parallel.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	goose.SetBaseFS(gdb.Migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		t.Fatal(err)
	}

	return db
}

func mustInsertWorker(t *testing.T, db *sql.DB, name, pin string, active bool) int {
	t.Helper()

	result, err := db.Exec(`INSERT INTO workers (worker_name, pin, phone, active) VALUES (?, ?, ?, ?)`,
		name, pin, "555-0100", active)
	if err != nil {
		t.Fatal(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	return int(id)
}
