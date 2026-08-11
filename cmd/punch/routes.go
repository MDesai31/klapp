package main

import "net/http"

// routes serves the public, internet-facing worker punch site. It has no
// sessions and no admin surface of any kind - see docs/security.md.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.punchForm)
	mux.HandleFunc("GET /punch", app.punchForm)
	mux.HandleFunc("POST /punch", app.punchStatus)
	mux.HandleFunc("POST /punch/in", app.punchIn)
	mux.HandleFunc("POST /punch/out", app.punchOut)
	mux.HandleFunc("GET /punch/late", app.punchLateForm)
	mux.HandleFunc("POST /punch/late", app.punchLate)

	return mux
}
