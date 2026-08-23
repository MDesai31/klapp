package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"klapp/internal/models"
)

func (app *application) adminSummary(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = models.CurrentPayPeriod(time.Now())
	}

	days, err := models.PayPeriodDays(period)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	rows, err := app.timePunches.PayPeriodSummary(period)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	summaryDays := make([]summaryDay, len(days))
	for i, d := range days {
		summaryDays[i] = summaryDay{Date: d, Label: models.DayLabel(d)}
	}

	periods, err := app.timePunches.PayPeriods()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_summary.tmpl", templateData{
		SummaryRows:   rows,
		SummaryDays:   summaryDays,
		PayPeriods:    periods,
		CurrentPeriod: period,
		Flash:         app.sessionManager.PopString(r.Context(), "flash"),
	})
}

// adminSummaryBulkPunch records the same punch for every worker selected
// on the summary tab, on any day of the pay period rather than only today.
// It is the dashboard's bulk punch with a day picker: "time_in" alone
// punches them in at that time on that day, "time_out" alone closes the
// punch they have open on that day, both together record a whole shift.
//
// There is no "punch at the current time" fallback here — a day that has
// already passed has no meaningful "now" — so one of the two times is
// required.
func (app *application) adminSummaryBulkPunch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	period := r.PostForm.Get("period")
	days, err := models.PayPeriodDays(period)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	back := "/admin/summary?period=" + url.QueryEscape(period)

	// The day decides which column the punch lands in, and the form only
	// ever offers this period's own days. It starts unchosen so a bulk
	// punch can't quietly land on a default day nobody meant.
	day := r.PostForm.Get("day")
	if day == "" {
		app.sessionManager.Put(r.Context(), "flash", "Choose a day for the punch.")
		http.Redirect(w, r, back, http.StatusSeeOther)
		return
	}
	if !slices.Contains(days, day) {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	date, err := time.ParseInLocation("2006-01-02", day, time.Local)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	in, err := parseTimeOnDay(r.PostForm.Get("time_in"), date)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	out, err := parseTimeOnDay(r.PostForm.Get("time_out"), date)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	switch {
	case in.IsZero() && out.IsZero():
		app.sessionManager.Put(r.Context(), "flash", "Enter a time in, a time out, or both.")
		http.Redirect(w, r, back, http.StatusSeeOther)
		return
	case !in.IsZero() && !out.IsZero() && !out.After(in):
		app.sessionManager.Put(r.Context(), "flash", "Time out must be after time in.")
		http.Redirect(w, r, back, http.StatusSeeOther)
		return
	}

	res, err := app.applyBulkPunch(r.PostForm["worker_id"], func(id int) (string, error) {
		if !in.IsZero() {
			// A zero out leaves the punch open, same as on the dashboard.
			_, err := app.timePunches.AdminCreate(id, in, out)
			return "in", err
		}
		return "out", app.timePunches.AdminPunchOutDay(id, day, out)
	})
	switch {
	case errors.Is(err, errBadWorkerID):
		app.clientError(w, http.StatusBadRequest)
		return
	case err != nil:
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", bulkPunchFlash(in, out, models.DayLabel(day), res))
	http.Redirect(w, r, back, http.StatusSeeOther)
}

// printExecTimeout bounds the printsched run. It sits under the server's
// 60s WriteTimeout so a wedged listener produces a flash message rather
// than a dropped connection.
const printExecTimeout = 45 * time.Second

// adminSummaryPrint sends this pay period's hours to the schedule listener
// on the home server, which builds the printable PDFs.
//
// The work is done by the printsched binary rather than in this process:
// it is the same command an admin can run from a shell, so there is one
// definition of what a print job contains and one place to debug it.
func (app *application) adminSummaryPrint(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	period := r.PostForm.Get("period")
	if _, err := models.PayPeriodDays(period); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	back := "/admin/summary?period=" + url.QueryEscape(period)

	ctx, cancel := context.WithTimeout(r.Context(), printExecTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, app.printBinary,
		"-period", period,
		"-host", app.printHost,
		"-port", strconv.Itoa(app.printPort),
		"-dsn", app.dsn,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// printsched's own failures - nobody to print, listener down -
		// are the admin's to see and act on, so they become a flash
		// rather than a 500. Log the full output for the operator.
		app.logger.Error("print job failed", "period", period, "err", err, "output", string(out))
		app.sessionManager.Put(r.Context(), "flash", "Print failed: "+printFailureReason(out, err))
		http.Redirect(w, r, back, http.StatusSeeOther)
		return
	}

	app.logger.Info("print job sent", "period", period, "output", string(out))
	app.sessionManager.Put(r.Context(), "flash", "Sent this pay period to the printer at "+app.printHost+".")
	http.Redirect(w, r, back, http.StatusSeeOther)
}

// printFailureReason picks the one line worth showing an admin out of a
// failed printsched run: its last line of output, which is where its own
// error message lands. Falls back to the exec error when it produced none
// (a missing binary, say).
func printFailureReason(out []byte, err error) string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	last := strings.TrimSpace(lines[len(lines)-1])
	last = strings.TrimPrefix(last, "printsched: ")
	if last == "" {
		return err.Error()
	}
	if len(last) > 200 {
		last = last[:200] + "..."
	}
	return last
}
