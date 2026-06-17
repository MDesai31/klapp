// seedadmin creates or resets an admin login. There's no signup page on
// the admin site by design, so this is how the first (or a forgotten)
// admin password gets set.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
	"klapp/db"
	"klapp/internal/models"

	_ "modernc.org/sqlite"
)

func main() {
	dsn := flag.String("dsn", "file:db/klapp.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", "SQLite data source name")
	username := flag.String("username", "", "admin username (required)")
	password := flag.String("password", "", "admin password (required)")
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "usage: seedadmin -username <name> -password <password>")
		os.Exit(1)
	}

	sqlDB, err := sql.Open("sqlite", *dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(db.Migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	admins := &models.AdminModel{DB: sqlDB}
	if err := admins.Upsert(*username, *password); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("admin %q is ready\n", *username)
}
