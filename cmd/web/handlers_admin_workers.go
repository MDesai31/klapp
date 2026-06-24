package main

import (
	"errors"
	"net/http"
	"strconv"

	"klapp/internal/models"
)

func (app *application) adminWorkers(w http.ResponseWriter, r *http.Request) {
	showInactive := r.URL.Query().Has("show_inactive")

	workers, err := app.workers.List()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !showInactive {
		workers = activeWorkers(workers)
	}

	app.render(w, r, http.StatusOK, "admin_workers.tmpl", templateData{Workers: workers, ShowInactiveWorkers: showInactive})
}

// activeWorkers filters a worker list down to active workers only, for the
// default (deactivated workers hidden) view of the admin workers page.
func activeWorkers(workers []models.Worker) []models.Worker {
	out := make([]models.Worker, 0, len(workers))
	for _, w := range workers {
		if w.Active {
			out = append(out, w)
		}
	}
	return out
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
		if errors.Is(err, models.ErrDuplicatePIN) {
			workers, err := app.workers.List()
			if err != nil {
				app.serverError(w, r, err)
				return
			}
			app.render(w, r, http.StatusOK, "admin_workers.tmpl", templateData{
				Workers: activeWorkers(workers),
				Flash:   "That PIN is already in use by another worker.",
			})
			return
		}
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin/workers", http.StatusSeeOther)
}

func (app *application) adminEditWorkerForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	worker, err := app.workers.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_edit_worker.tmpl", templateData{Worker: &worker})
}

func (app *application) adminEditWorker(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	name := r.PostFormValue("worker_name")
	pin := r.PostFormValue("pin")
	phone := r.PostFormValue("phone")

	if name == "" || pin == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := app.workers.Update(id, name, pin, phone); err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		if errors.Is(err, models.ErrDuplicatePIN) {
			worker := models.Worker{ID: id, WorkerName: name, PIN: pin, Phone: phone}
			app.render(w, r, http.StatusOK, "admin_edit_worker.tmpl", templateData{
				Worker: &worker,
				Flash:  "That PIN is already in use by another worker.",
			})
			return
		}
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
