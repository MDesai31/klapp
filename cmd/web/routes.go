package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.punchForm)
	mux.HandleFunc("GET /privacy", app.privacyPolicy)
	mux.HandleFunc("GET /terms", app.termsAndConditions)
	mux.HandleFunc("GET /punch", app.punchForm)
	mux.HandleFunc("POST /punch", app.punchStatus)
	mux.HandleFunc("POST /punch/in", app.punchIn)
	mux.HandleFunc("POST /punch/out", app.punchOut)
	mux.HandleFunc("GET /punch/late", app.punchLateForm)
	mux.HandleFunc("POST /punch/late", app.punchLate)

	return mux
}

// adminRoutes serves the LAN/WireGuard-only admin site. Everything under
// /admin/ except the login page itself requires a logged-in admin session.
func (app *application) adminRoutes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /admin/login", app.adminLoginForm)
	mux.HandleFunc("POST /admin/login", app.adminLogin)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /admin/{$}", app.adminDashboard)
	protected.HandleFunc("POST /admin/logout", app.adminLogout)
	protected.HandleFunc("GET /admin/timesheet", app.adminTimesheet)
	protected.HandleFunc("GET /admin/summary", app.adminSummary)
	protected.HandleFunc("GET /admin/punches/{id}/edit", app.adminEditPunchForm)
	protected.HandleFunc("POST /admin/punches/{id}/edit", app.adminEditPunch)
	protected.HandleFunc("GET /admin/workers", app.adminWorkers)
	protected.HandleFunc("POST /admin/workers", app.adminCreateWorker)
	protected.HandleFunc("GET /admin/workers/{id}/edit", app.adminEditWorkerForm)
	protected.HandleFunc("POST /admin/workers/{id}/edit", app.adminEditWorker)
	protected.HandleFunc("POST /admin/workers/{id}/toggle-active", app.adminToggleWorkerActive)
	mux.Handle("/admin/", app.requireAdmin(protected))

	return app.sessionManager.LoadAndSave(mux)
}
