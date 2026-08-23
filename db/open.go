package db

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

// DefaultDSN is the SQLite data source every binary defaults to, relative to
// its working directory.
const DefaultDSN = "file:db/klapp.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"

// Open opens dsn and verifies the connection before handing it back, so a
// bad path or locked file fails at startup rather than on the first query.
// The caller is responsible for registering a "sqlite" driver (importing
// modernc.org/sqlite for its side effects).
func Open(dsn string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return sqlDB, nil
}

// Migrate applies the embedded goose migrations. It is idempotent and safe
// to run from every binary at startup, including concurrently.
func Migrate(sqlDB *sql.DB) error {
	goose.SetBaseFS(Migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	return goose.Up(sqlDB, "migrations")
}
