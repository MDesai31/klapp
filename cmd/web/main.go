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
}

func main() {
	addr := flag.String("addr", ":4000", "worker site HTTP network address")
	adminAddr := flag.String("admin-addr", ":8082", "admin site HTTP network address (LAN/WireGuard only - do not expose publicly)")
	dsn := flag.String("dsn", "file:db/klapp.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", "SQLite data source name")
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
	}
	go app.pinLimiter.cleanupLoop()

	errc := make(chan error, 2)

	go func() {
		logger.Info("starting worker site", slog.String("addr", *addr))
		errc <- http.ListenAndServe(*addr, app.routes())
	}()

	go func() {
		logger.Info("starting admin site", slog.String("addr", *adminAddr))
		errc <- http.ListenAndServe(*adminAddr, app.adminRoutes())
	}()

	go app.runNightlyPunchOut()

	logger.Error((<-errc).Error())
	os.Exit(1)
}

// runNightlyPunchOut sleeps until 9 PM each day, then auto-closes any
// punch still open and marks it non_compliant.
func (app *application) runNightlyPunchOut() {
	for {
		now := time.Now().Local()
		next := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, time.Local)
		if !now.Before(next) {
			next = next.AddDate(0, 0, 1)
		}
		time.Sleep(time.Until(next))
		n, err := app.timePunches.AutoPunchOutNonCompliant(next)
		if err != nil {
			app.logger.Error("auto punch-out failed", slog.Any("err", err))
		} else if n > 0 {
			app.logger.Info("auto punch-out complete", slog.Int("workers", n))
		}
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
