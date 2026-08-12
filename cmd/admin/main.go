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
	"klapp/internal/config"
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
	punchSiteURL   string

	// The summary tab's Print button shells out to the printsched binary,
	// which needs the same database this process is using plus the address
	// of the schedule listener on the home server.
	dsn         string
	printBinary string
	printHost   string
	printPort   int
}

func main() {
	addr := flag.String("addr", ":8082", "admin site HTTP network address (LAN/WireGuard only - do not expose publicly)")
	dsn := flag.String("dsn", db.DefaultDSN, "SQLite data source name")
	configPath := flag.String("config", "config.json", "path to JSON config file (optional; defaults used if missing, see config.example.json)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

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

	sessionManager := scs.New()
	sessionManager.Lifetime = 30 * 24 * time.Hour

	app := &application{
		logger:         logger,
		workers:        &models.WorkerModel{DB: sqlDB},
		timePunches:    &models.TimePunchModel{DB: sqlDB, DailyPunchLimit: cfg.DailyPunchLimit},
		admins:         &models.AdminModel{DB: sqlDB},
		customers:      &models.CustomerModel{DB: sqlDB},
		invoices:       &models.InvoiceModel{DB: sqlDB},
		catalog:        &models.CatalogModel{DB: sqlDB},
		templateCache:  templateCache,
		sessionManager: sessionManager,
		punchSiteURL:   cfg.PunchSiteURL,
		dsn:            *dsn,
		printBinary:    cfg.PrintBinary,
		printHost:      cfg.PrintHost,
		printPort:      cfg.PrintPort,
	}

	// The nightly sweep lives here rather than in the punch binary because
	// the admin site is the one that always runs: the punch site can be
	// left switched off (see deploy/update.sh) and open punches still need
	// closing out.
	go app.runNightlyPunchOut()

	logger.Info("starting admin site", slog.String("addr", *addr))
	logger.Error(newServer(*addr, app.routes()).ListenAndServe().Error())
	os.Exit(1)
}

// newServer wraps a handler in an http.Server with sane timeouts, so slow
// or stalled clients can't hold connections (and their goroutines) open
// indefinitely.
func newServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
}
