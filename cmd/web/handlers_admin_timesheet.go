package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"klapp/internal/models"
)

func (app *application) adminTimesheet(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = models.CurrentPayPeriod(time.Now())
	}

	rows, err := app.timePunches.ForPayPeriod(period)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	periods, err := app.timePunches.PayPeriods()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_timesheet.tmpl", templateData{
		TimesheetRows: rows,
		PayPeriods:    periods,
		CurrentPeriod: period,
	})
}

func (app *application) adminEditPunchForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	punch, err := app.timePunches.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	worker, err := app.workers.Get(punch.WorkerID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_edit_punch.tmpl", templateData{Punch: &punch, Worker: &worker})
}

func (app *application) adminEditPunch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	start, err := time.ParseInLocation("2006-01-02T15:04", r.PostFormValue("start_time"), time.Local)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var end time.Time
	if v := r.PostFormValue("end_time"); v != "" {
		end, err = time.ParseInLocation("2006-01-02T15:04", v, time.Local)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}

	if err := app.timePunches.AdminUpdate(id, start, end); err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	punch, err := app.timePunches.Get(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin/timesheet?period="+punch.PayPeriod, http.StatusSeeOther)
}
