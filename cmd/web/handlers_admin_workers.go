package main

import (
	"net/http"
	"strconv"
)

func (app *application) adminWorkers(w http.ResponseWriter, r *http.Request) {
	workers, err := app.workers.List()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_workers.tmpl", templateData{Workers: workers})
}

func (app *application) adminCreateWorker(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("worker_name")
	pin := r.PostFormValue("pin")
	phone := r.PostFormValue("phone")

	if name == "" || pin == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if _, err := app.workers.Create(name, pin, phone); err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin/workers", http.StatusSeeOther)
}

func (app *application) adminToggleWorkerActive(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	worker, err := app.workers.Get(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if err := app.workers.SetActive(id, !worker.Active); err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin/workers", http.StatusSeeOther)
}
