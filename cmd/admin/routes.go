package main

import "net/http"

// routes serves the LAN/WireGuard-only admin site. Everything under
// /admin/ except the login page itself requires a logged-in admin session.
func (app *application) routes() http.Handler {
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
	protected.HandleFunc("POST /admin/punch/bulk", app.adminBulkPunch)
	protected.HandleFunc("GET /admin/timesheet", app.adminTimesheet)
	protected.HandleFunc("GET /admin/summary", app.adminSummary)
	protected.HandleFunc("GET /admin/punches/new", app.adminAddPunchForm)
	protected.HandleFunc("POST /admin/punches/new", app.adminAddPunch)
	protected.HandleFunc("GET /admin/punches/{id}/edit", app.adminEditPunchForm)
	protected.HandleFunc("POST /admin/punches/{id}/edit", app.adminEditPunch)
	protected.HandleFunc("POST /admin/punches/{id}/delete", app.adminDeletePunch)
	protected.HandleFunc("GET /admin/workers", app.adminWorkers)
	protected.HandleFunc("POST /admin/workers", app.adminCreateWorker)
	protected.HandleFunc("GET /admin/workers/{id}/edit", app.adminEditWorkerForm)
	protected.HandleFunc("POST /admin/workers/{id}/edit", app.adminEditWorker)
	protected.HandleFunc("POST /admin/workers/{id}/toggle-active", app.adminToggleWorkerActive)

	// invoices
	protected.HandleFunc("GET /admin/invoices", app.adminInvoices)
	protected.HandleFunc("GET /admin/invoices/{id}", app.adminInvoiceView)
	protected.HandleFunc("POST /admin/invoices/{id}/submit", app.adminInvoiceSubmit)

	// catalog (job descriptions + materials)
	protected.HandleFunc("GET /admin/catalog", app.adminJobDescriptions)
	protected.HandleFunc("POST /admin/catalog/jobs", app.adminCreateJobDescription)
	protected.HandleFunc("POST /admin/catalog/jobs/{id}/delete", app.adminDeleteJobDescription)
	protected.HandleFunc("POST /admin/catalog/materials", app.adminCreateMaterial)
	protected.HandleFunc("POST /admin/catalog/materials/{id}/delete", app.adminDeleteMaterial)

	// customers
	protected.HandleFunc("GET /admin/customers", app.adminCustomers)
	protected.HandleFunc("POST /admin/customers", app.adminCreateCustomer)
	protected.HandleFunc("GET /admin/customers/{id}", app.adminCustomerView)
	protected.HandleFunc("GET /admin/customers/{id}/edit", app.adminEditCustomerForm)
	protected.HandleFunc("POST /admin/customers/{id}/edit", app.adminEditCustomer)

	mux.Handle("/admin/", app.requireAdmin(protected))

	return app.sessionManager.LoadAndSave(mux)
}
