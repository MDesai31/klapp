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

// adminBulkPunch punches every worker selected on the dashboard in or out.
// The dashboard only offers one action for the whole selection, but a
// worker's state can change between page load and submit, so a worker who
// is already in the requested state is skipped and named in the flash
// rather than failing the whole batch.
func (app *application) adminBulkPunch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	action := r.PostForm.Get("action")
	if action != "in" && action != "out" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	now := time.Now()
	var punched, skipped []string
	seen := make(map[int]bool)

	for _, v := range r.PostForm["worker_id"] {
		id, err := strconv.Atoi(v)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
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
				app.clientError(w, http.StatusBadRequest)
				return
			}
			app.serverError(w, r, err)
			return
		}

		if action == "in" {
			_, err = app.timePunches.PunchIn(id, nil, nil, now)
		} else {
			err = app.timePunches.PunchOut(id, nil, nil, now)
		}

		switch {
		case err == nil:
			punched = append(punched, worker.WorkerName)
		case errors.Is(err, models.ErrAlreadyOpen):
			skipped = append(skipped, worker.WorkerName+" (already punched in)")
		case errors.Is(err, models.ErrNoRecord):
			skipped = append(skipped, worker.WorkerName+" (not punched in)")
		case errors.Is(err, models.ErrDailyLimitExceeded):
			skipped = append(skipped, worker.WorkerName+" (daily punch-in limit reached)")
		default:
			app.serverError(w, r, err)
			return
		}
	}

	app.sessionManager.Put(r.Context(), "flash", bulkPunchFlash(action, punched, skipped))
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// bulkPunchFlash summarizes what a bulk punch in/out actually did.
func bulkPunchFlash(action string, punched, skipped []string) string {
	if len(punched) == 0 && len(skipped) == 0 {
		return "No workers selected."
	}

	var parts []string
	if len(punched) > 0 {
		verb := "Punched in"
		if action == "out" {
			verb = "Punched out"
		}
		parts = append(parts, fmt.Sprintf("%s %s.", verb, strings.Join(punched, ", ")))
	}
	if len(skipped) > 0 {
		parts = append(parts, fmt.Sprintf("Skipped %s.", strings.Join(skipped, ", ")))
	}

	return strings.Join(parts, " ")
}
