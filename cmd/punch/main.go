package main

import (
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"klapp/db"
	"klapp/internal/config"
	"klapp/internal/models"

	"github.com/alexedwards/scs/v2"
	_ "modernc.org/sqlite"
)

type application struct {
	logger        *slog.Logger
	workers       *models.WorkerModel
	timePunches   *models.TimePunchModel
	templateCache map[string]*template.Template
	punchSessions *scs.SessionManager
	pinLimiter    *pinLimiter
	pinCheckDelay time.Duration
}

func main() {
	addr := flag.String("addr", ":4000", "worker punch site HTTP network address")
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

	// Worker punch sessions hold only the worker ID between PIN entry and
	// the punch button, so the PIN never round-trips through the client
	// (docs/reference/security.md). Short-lived by design: a worker
	// re-enters their PIN on each visit, same as before.
	punchSessions := newPunchSessionManager()

	app := &application{
		logger:  logger,
		workers: &models.WorkerModel{DB: sqlDB},
		timePunches: &models.TimePunchModel{
			DB:              sqlDB,
			DailyPunchLimit: cfg.DailyPunchLimit,
		},
		templateCache: templateCache,
		punchSessions: punchSessions,
		pinLimiter: newPinLimiter(
			cfg.PinLockoutThreshold,
			time.Duration(cfg.PinLockoutWindowMinutes)*time.Minute,
			time.Duration(cfg.PinLockoutCooldownMinutes)*time.Minute,
		),
		pinCheckDelay: time.Duration(cfg.PinCheckDelayMs) * time.Millisecond,
	}
	go app.pinLimiter.cleanupLoop()

	logger.Info("starting worker punch site", slog.String("addr", *addr))
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
