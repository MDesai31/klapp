// printsched turns one pay period into a print job and sends it to the
// schedule listener on the home server, which builds the PDFs.
//
// It is a separate binary rather than code inside the admin site so the same
// print can be run by hand:
//
//	printsched -period 2026-06-08
//	printsched -period 2026-06-08 -host 10.9.0.7 -port 5555
//
// The admin site's Print button runs exactly this, passing its own -dsn and
// the host and port from its config file.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"klapp/db"
	"klapp/internal/config"
	"klapp/internal/models"
	"klapp/internal/schedule"

	_ "modernc.org/sqlite"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "printsched:", err)
		os.Exit(1)
	}
}

func run() error {
	period := flag.String("period", "", `pay period start date, "2006-01-02" (default: the current pay period)`)
	host := flag.String("host", "", "schedule listener host (default: print_host from the config file)")
	port := flag.Int("port", 0, "schedule listener port (default: print_port from the config file)")
	dsn := flag.String("dsn", db.DefaultDSN, "SQLite data source name")
	configPath := flag.String("config", "config.json", "path to JSON config file (optional; defaults used if missing, see config.example.json)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if *host == "" {
		*host = cfg.PrintHost
	}
	if *port == 0 {
		*port = cfg.PrintPort
	}
	if *period == "" {
		*period = models.CurrentPayPeriod(time.Now())
	}
	// Fail on a bad period here rather than after opening the database, so
	// a typo costs nothing and says so plainly.
	if _, err := models.PayPeriodDays(*period); err != nil {
		return fmt.Errorf("bad -period %q: %w", *period, err)
	}

	sqlDB, err := db.Open(*dsn)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	// No db.Migrate here: printsched is a reader, and it runs alongside a
	// live admin site that has already migrated the same file.
	workerModel := &models.WorkerModel{DB: sqlDB}
	punchModel := &models.TimePunchModel{DB: sqlDB}

	workers, err := workerModel.List()
	if err != nil {
		return err
	}
	punches, err := punchModel.ForPayPeriod(*period)
	if err != nil {
		return err
	}

	payload, err := schedule.Build(*period, workers, punches)
	if err != nil {
		return err
	}
	if len(payload.Sheets) == 0 {
		return fmt.Errorf("no active workers have hours in pay period %s", *period)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	reply, err := schedule.Send(ctx, *host, *port, payload)
	if err != nil {
		return err
	}

	fmt.Printf("sent %d sheet(s) for %s to %s:%d\n", len(payload.Sheets), *period, *host, *port)
	if reply != "" {
		fmt.Println(reply)
	}
	return nil
}
