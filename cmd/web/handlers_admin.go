package main

import (
	"errors"
	"net/http"
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

	app.render(w, r, http.StatusOK, "admin_dashboard.tmpl", templateData{DashboardRows: rows, NotifyBaseURL: app.punchSiteURL})
}
