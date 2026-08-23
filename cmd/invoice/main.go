package main

import (
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"klapp/db"
	"klapp/internal/models"

	_ "modernc.org/sqlite"
)

type application struct {
	logger        *slog.Logger
	workers       *models.WorkerModel
	customers     *models.CustomerModel
	invoices      *models.InvoiceModel
	catalog       *models.CatalogModel
	templateCache map[string]*template.Template
	session       *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":8083", "invoice site HTTP network address (LAN/WireGuard only)")
	dsn := flag.String("dsn", db.DefaultDSN, "SQLite data source name")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sqlDB, err := db.Open(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer sqlDB.Close()

	if err := db.Migrate(sqlDB); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	session := scs.New()
	session.Lifetime = 2 * time.Hour

	app := &application{
		logger:        logger,
		workers:       &models.WorkerModel{DB: sqlDB},
		customers:     &models.CustomerModel{DB: sqlDB},
		invoices:      &models.InvoiceModel{DB: sqlDB},
		catalog:       &models.CatalogModel{DB: sqlDB},
		templateCache: templateCache,
		session:       session,
	}

	logger.Info("starting invoice site", slog.String("addr", *addr))
	srv := &http.Server{
		Addr:              *addr,
		Handler:           app.session.LoadAndSave(app.routes()),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
