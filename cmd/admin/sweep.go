package main

import (
	"log/slog"
	"time"
)

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
