package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.pinForm)
	mux.HandleFunc("POST /{$}", app.pinSubmit)
	mux.HandleFunc("GET /form", app.invoiceForm)
	mux.HandleFunc("POST /form", app.invoiceSubmit)
	mux.HandleFunc("GET /success", app.successPage)

	// JSON autocomplete endpoints
	mux.HandleFunc("GET /api/customers", app.apiCustomers)
	mux.HandleFunc("GET /api/jobs", app.apiJobs)
	mux.HandleFunc("GET /api/materials", app.apiMaterials)

	return mux
}
