// Command schedule-listener runs on the home server and turns print jobs
// into PDFs.
//
// klapp's admin site (via the printsched binary) posts a schedule.Payload
// to /print over WireGuard. This server hands it to build_schedule.py,
// which writes one filled-in PDF per worker, and replies with the paths it
// wrote. Actually sending those to a printer is the print_command hook,
// which is off until there is a printer to point it at.
//
// It listens on the WireGuard interface only; it has no authentication,
// because nothing outside that network can reach it.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"klapp/internal/schedule"
)

// maxBodyBytes caps an incoming print job. A pay period for a whole crew
// is a few tens of kilobytes.
const maxBodyBytes = 4 << 20

// buildTimeout bounds one reportlab run over every sheet in a job.
const buildTimeout = 120 * time.Second

type config struct {
	// Addr is the listen address. Default ":5555".
	Addr string `json:"addr"`
	// Python and Script are how build_schedule.py gets run.
	Python string `json:"python"`
	Script string `json:"script"`
	// OutputDir is where the PDFs are written; created if missing.
	OutputDir string `json:"output_dir"`
	// PrintCommand is an argv run once per generated PDF with the file's
	// path appended, e.g. ["lp", "-d", "office"]. Empty means don't
	// print - the job stops at "PDF on disk".
	PrintCommand []string `json:"print_command"`
}

func defaultConfig() config {
	return config{
		Addr:      ":5555",
		Python:    "python3",
		Script:    "./build_schedule.py",
		OutputDir: "./out",
	}
}

func loadConfig(path string) (config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return config{}, fmt.Errorf("reading config file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return config{}, fmt.Errorf("parsing config file %s: %w", path, err)
	}
	return cfg, nil
}

type application struct {
	cfg    config
	logger *slog.Logger
}

func main() {
	configPath := flag.String("config", "config.json", "path to JSON config file (optional; defaults used if missing)")
	addr := flag.String("addr", "", "listen address (default: addr from the config file)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := loadConfig(*configPath)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	if *addr != "" {
		cfg.Addr = *addr
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{cfg: cfg, logger: logger}

	mux := http.NewServeMux()
	mux.HandleFunc("POST "+schedule.Path, app.print)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		// Long enough to cover a full reportlab run plus the reply.
		WriteTimeout: buildTimeout + 30*time.Second,
		IdleTimeout:  2 * time.Minute,
	}

	logger.Info("starting schedule listener", "addr", cfg.Addr, "output_dir", cfg.OutputDir)
	logger.Error(srv.ListenAndServe().Error())
	os.Exit(1)
}

// print builds every sheet in the posted job and replies with the paths.
func (app *application) print(w http.ResponseWriter, r *http.Request) {
	var payload schedule.Payload
	if err := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes)).Decode(&payload); err != nil {
		app.logger.Warn("bad print job", "err", err, "from", r.RemoteAddr)
		http.Error(w, "invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(payload.Sheets) == 0 {
		http.Error(w, "payload has no sheets", http.StatusBadRequest)
		return
	}

	app.logger.Info("print job received", "pay_period", payload.PayPeriod, "sheets", len(payload.Sheets), "from", r.RemoteAddr)

	files, err := app.buildPDFs(r.Context(), payload)
	if err != nil {
		app.logger.Error("build failed", "pay_period", payload.PayPeriod, "err", err)
		http.Error(w, "building the schedule failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Printing is best-effort and reported, not fatal: the PDFs exist and
	// can be printed by hand, so a jammed printer should not read as "the
	// schedule did not build".
	var printErr error
	if len(app.cfg.PrintCommand) > 0 {
		printErr = app.sendToPrinter(r.Context(), files)
		if printErr != nil {
			app.logger.Error("printing failed", "err", printErr)
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "built %d sheet(s) for %s\n", len(files), payload.PayPeriod)
	for _, f := range files {
		fmt.Fprintln(w, f)
	}
	if printErr != nil {
		fmt.Fprintf(w, "warning: printing failed: %v\n", printErr)
	}
}

// buildPDFs pipes the job to build_schedule.py and returns the paths it
// reported on stdout, one per line.
func (app *application) buildPDFs(ctx context.Context, payload schedule.Payload) ([]string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, buildTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, app.cfg.Python, app.cfg.Script, "--json", "-", "--outdir", app.cfg.OutputDir)
	cmd.Stdin = bytes.NewReader(body)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return nil, err
		}
		// reportlab tracebacks are long; the last line is the useful one.
		lines := strings.Split(msg, "\n")
		return nil, fmt.Errorf("%s: %s", err, strings.TrimSpace(lines[len(lines)-1]))
	}

	var files []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	if len(files) == 0 {
		return nil, errors.New("build_schedule.py reported no files")
	}
	return files, nil
}

// sendToPrinter runs PrintCommand once per file, with the file's path as
// the final argument. It stops at the first failure.
func (app *application) sendToPrinter(ctx context.Context, files []string) error {
	ctx, cancel := context.WithTimeout(ctx, buildTimeout)
	defer cancel()

	argv := app.cfg.PrintCommand
	for _, f := range files {
		args := append(append([]string{}, argv[1:]...), f)
		out, err := exec.CommandContext(ctx, argv[0], args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s %s: %w: %s", argv[0], f, err, bytes.TrimSpace(out))
		}
		app.logger.Info("sent to printer", "file", f, "output", strings.TrimSpace(string(out)))
	}
	return nil
}
