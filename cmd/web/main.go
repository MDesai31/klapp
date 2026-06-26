package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/pressly/goose/v3"
	"klapp/db"
	"klapp/internal/models"

	_ "modernc.org/sqlite"
)

type application struct {
	logger         *slog.Logger
	workers        *models.WorkerModel
	timePunches    *models.TimePunchModel
	admins         *models.AdminModel
	customers      *models.CustomerModel
	invoices       *models.InvoiceModel
	catalog        *models.CatalogModel
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "worker site HTTP network address")
	adminAddr := flag.String("admin-addr", ":8082", "admin site HTTP network address (LAN/WireGuard only - do not expose publicly)")
	dsn := flag.String("dsn", "file:db/klapp.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", "SQLite data source name")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sqlDB, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer sqlDB.Close()

	if err := migrate(sqlDB); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 30 * 24 * time.Hour

	app := &application{
		logger:         logger,
		workers:        &models.WorkerModel{DB: sqlDB},
		timePunches:    &models.TimePunchModel{DB: sqlDB},
		admins:         &models.AdminModel{DB: sqlDB},
		customers:      &models.CustomerModel{DB: sqlDB},
		invoices:       &models.InvoiceModel{DB: sqlDB},
		catalog:        &models.CatalogModel{DB: sqlDB},
		templateCache:  templateCache,
		sessionManager: sessionManager,
	}

	errc := make(chan error, 2)

	go func() {
		logger.Info("starting worker site", slog.String("addr", *addr))
		errc <- http.ListenAndServe(*addr, app.routes())
	}()

	go func() {
		logger.Info("starting admin site", slog.String("addr", *adminAddr))
		errc <- http.ListenAndServe(*adminAddr, app.adminRoutes())
	}()

	logger.Error((<-errc).Error())
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
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

func migrate(sqlDB *sql.DB) error {
	goose.SetBaseFS(db.Migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	return goose.Up(sqlDB, "migrations")
}
