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
	pinLimiter     *pinLimiter
	pinCheckDelay  time.Duration
	punchSiteURL   string
}

func main() {
	addr := flag.String("addr", ":4000", "worker site HTTP network address")
	adminAddr := flag.String("admin-addr", ":8082", "admin site HTTP network address (LAN/WireGuard only - do not expose publicly)")
	dsn := flag.String("dsn", "file:db/klapp.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", "SQLite data source name")
	configPath := flag.String("config", "config.json", "path to JSON config file (optional; defaults used if missing, see config.example.json)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := loadConfig(*configPath)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

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
		logger:  logger,
		workers: &models.WorkerModel{DB: sqlDB},
		timePunches: &models.TimePunchModel{
			DB:              sqlDB,
			DailyPunchLimit: cfg.DailyPunchLimit,
		},
		admins:         &models.AdminModel{DB: sqlDB},
		customers:      &models.CustomerModel{DB: sqlDB},
		invoices:       &models.InvoiceModel{DB: sqlDB},
		catalog:        &models.CatalogModel{DB: sqlDB},
		templateCache:  templateCache,
		sessionManager: sessionManager,
		pinLimiter: newPinLimiter(
			cfg.PinLockoutThreshold,
			time.Duration(cfg.PinLockoutWindowMinutes)*time.Minute,
			time.Duration(cfg.PinLockoutCooldownMinutes)*time.Minute,
		),
		pinCheckDelay: time.Duration(cfg.PinCheckDelayMs) * time.Millisecond,
		punchSiteURL:  cfg.PunchSiteURL,
	}
	go app.pinLimiter.cleanupLoop()

	errc := make(chan error, 2)

	go func() {
		logger.Info("starting worker site", slog.String("addr", *addr))
		errc <- newServer(*addr, app.routes()).ListenAndServe()
	}()

	go func() {
		logger.Info("starting admin site", slog.String("addr", *adminAddr))
		errc <- newServer(*adminAddr, app.adminRoutes()).ListenAndServe()
	}()

	go app.runNightlyPunchOut()

	logger.Error((<-errc).Error())
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

// nightlySweepAfter returns the next 9 PM local time strictly after now.
func nightlySweepAfter(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, now.Location())
	if !now.Before(next) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// runNightlyPunchOut sleeps until 9 PM each day, then auto-closes any
// punch still open and marks it non_compliant. On startup it first runs a
// catch-up sweep for the most recent 9 PM: a restart between 9 PM and
// midnight would otherwise skip that night's sweep entirely, leaving
// punches open until the following evening.
func (app *application) runNightlyPunchOut() {
	app.sweepOpenPunches(nightlySweepAfter(time.Now().Local()).AddDate(0, 0, -1))

	for {
		next := nightlySweepAfter(time.Now().Local())
		time.Sleep(time.Until(next))
		app.sweepOpenPunches(next)
	}
}

func (app *application) sweepOpenPunches(cutoff time.Time) {
	n, err := app.timePunches.AutoPunchOutNonCompliant(cutoff)
	if err != nil {
		app.logger.Error("auto punch-out failed", slog.Any("err", err))
	} else if n > 0 {
		app.logger.Info("auto punch-out complete", slog.Int("workers", n))
	}
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
