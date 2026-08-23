package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"klapp/internal/models"
)

func (app *application) adminLoginForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "admin_login.tmpl", templateData{})
}

func (app *application) adminLogin(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	admin, err := app.admins.Authenticate(username, password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			app.render(w, r, http.StatusOK, "admin_login.tmpl", templateData{Flash: "Incorrect username or password."})
			return
		}
		app.serverError(w, r, err)
		return
	}

	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "adminID", admin.ID)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (app *application) adminLogout(w http.ResponseWriter, r *http.Request) {
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "adminID")
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (app *application) adminDashboard(w http.ResponseWriter, r *http.Request) {
	today := time.Now().Format("2006-01-02")

	rows, err := app.timePunches.DashboardStatus(today)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_dashboard.tmpl", templateData{
		DashboardRows: rows,
		NotifyBaseURL: app.punchSiteURL,
		Flash:         app.sessionManager.PopString(r.Context(), "flash"),
	})
}

// adminBulkPunch punches every worker selected on the dashboard, with the
// two optional time fields deciding what "punch" means: "time_in" alone
// punches them in at that time, "time_out" alone punches them out, both
// together record a whole shift in one go, and neither punches at the
// current time in whichever direction the worker isn't already in. Times
// are on today's date — the dashboard only ever shows today. The summary
// tab has its own bulk punch for the other days of a pay period.
func (app *application) adminBulkPunch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	now := time.Now()

	in, err := parseTimeOnDay(r.PostForm.Get("time_in"), now)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	out, err := parseTimeOnDay(r.PostForm.Get("time_out"), now)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// With both times set every worker gets the same shift, so a backwards
	// pair would skip the whole batch for the same reason. Say so once.
	if !in.IsZero() && !out.IsZero() && !out.After(in) {
		app.sessionManager.Put(r.Context(), "flash", "Time out must be after time in.")
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	res, err := app.applyBulkPunch(r.PostForm["worker_id"], func(id int) (string, error) {
		switch {
		case !in.IsZero():
			// AdminCreate, not PunchIn, so the entry is recorded as
			// admin-made; it still refuses a second open punch. A zero out
			// leaves the punch open.
			_, err := app.timePunches.AdminCreate(id, in, out)
			return "in", err
		case !out.IsZero():
			return "out", app.timePunches.AdminPunchOut(id, out)
		default:
			// No times: punch at the current time, away from wherever the
			// worker is now, so a mixed selection still does the obvious
			// thing for each of them.
			_, openErr := app.timePunches.Open(id)
			switch {
			case openErr == nil:
				return "out", app.timePunches.PunchOut(id, nil, nil, now)
			case errors.Is(openErr, models.ErrNoRecord):
				_, err := app.timePunches.PunchIn(id, nil, nil, now)
				return "in", err
			default:
				return "", openErr
			}
		}
	})
	switch {
	case errors.Is(err, errBadWorkerID):
		app.clientError(w, http.StatusBadRequest)
		return
	case err != nil:
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", bulkPunchFlash(in, out, "", res))
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// bulkPunchResult sorts the workers a bulk punch touched into the lists
// its flash message is built from.
type bulkPunchResult struct {
	punchedIn  []string
	punchedOut []string
	skipped    []string
}

// errBadWorkerID marks a bulk punch that failed on the request itself — a
// worker_id that isn't a number, or isn't a worker — so the handler can
// answer 400 instead of 500.
var errBadWorkerID = errors.New("bad worker id")

// applyBulkPunch calls punch once per distinct worker ID in ids and sorts
// the workers by what happened to them. punch reports which list a success
// belongs in, "in" or "out"; a whole shift counts as a punch in, since
// that's the end of the sentence the flash builds.
//
// A worker's state can change between page load and submit, and a
// selection can legitimately mix workers who are in with workers who are
// out, so a worker the requested punch doesn't fit is skipped and named
// rather than failing the whole batch. Anything else is returned as an
// error and stops the batch.
func (app *application) applyBulkPunch(ids []string, punch func(workerID int) (string, error)) (bulkPunchResult, error) {
	var res bulkPunchResult
	seen := make(map[int]bool)

	for _, v := range ids {
		id, err := strconv.Atoi(v)
		if err != nil {
			return res, errBadWorkerID
		}
		// A worker with two punches today has two dashboard rows, and so
		// two checkboxes carrying the same ID.
		if seen[id] {
			continue
		}
		seen[id] = true

		worker, err := app.workers.Get(id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				return res, errBadWorkerID
			}
			return res, err
		}

		dir, err := punch(id)
		switch {
		case err == nil && dir == "out":
			res.punchedOut = append(res.punchedOut, worker.WorkerName)
		case err == nil:
			res.punchedIn = append(res.punchedIn, worker.WorkerName)
		case errors.Is(err, models.ErrAlreadyOpen):
			res.skipped = append(res.skipped, worker.WorkerName+" (already punched in)")
		case errors.Is(err, models.ErrNoRecord):
			res.skipped = append(res.skipped, worker.WorkerName+" (not punched in)")
		case errors.Is(err, models.ErrDailyLimitExceeded):
			res.skipped = append(res.skipped, worker.WorkerName+" (daily punch-in limit reached)")
		case errors.Is(err, models.ErrEndBeforeStart):
			res.skipped = append(res.skipped, worker.WorkerName+" (that time is before their punch-in)")
		default:
			return res, err
		}
	}

	return res, nil
}

// parseTimeOnDay reads an <input type="time"> value as a time on day's
// date. A blank value gives the zero time, meaning the caller should stamp
// the current time instead.
func parseTimeOnDay(value string, day time.Time) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	// Browsers send "15:04", or "15:04:05" if the input has a seconds step.
	t, err := time.Parse("15:04", value)
	if err != nil {
		t, err = time.Parse("15:04:05", value)
		if err != nil {
			return time.Time{}, err
		}
	}

	return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, time.Local), nil
}

// bulkPunchFlash summarizes what a bulk punch actually did. Non-zero in
// and out times are the ones the admin typed, and are named in the message
// so it's obvious the punch didn't land at "now". With both set, every
// success is a whole shift and arrives in punchedIn. dayLabel names the
// day the punch landed on, for the summary tab where it isn't today; the
// dashboard passes "".
func bulkPunchFlash(in, out time.Time, dayLabel string, res bulkPunchResult) string {
	if len(res.punchedIn) == 0 && len(res.punchedOut) == 0 && len(res.skipped) == 0 {
		return "No workers selected."
	}

	var on string
	if dayLabel != "" {
		on = " on " + dayLabel
	}

	var parts []string
	switch {
	case len(res.punchedIn) > 0 && !in.IsZero() && !out.IsZero():
		parts = append(parts, fmt.Sprintf("Punched %s in at %s and out at %s%s.",
			strings.Join(res.punchedIn, ", "), in.Format("3:04 PM"), out.Format("3:04 PM"), on))
	case len(res.punchedIn) > 0:
		parts = append(parts, fmt.Sprintf("Punched in %s%s%s.", strings.Join(res.punchedIn, ", "), atTime(in), on))
	}
	if len(res.punchedOut) > 0 {
		parts = append(parts, fmt.Sprintf("Punched out %s%s%s.", strings.Join(res.punchedOut, ", "), atTime(out), on))
	}
	if len(res.skipped) > 0 {
		parts = append(parts, fmt.Sprintf("Skipped %s.", strings.Join(res.skipped, ", ")))
	}

	return strings.Join(parts, " ")
}

// atTime renders " at 8:00 AM" for a time the admin chose, and nothing at
// all for a punch that was stamped with the current time.
func atTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return " at " + t.Format("3:04 PM")
}
