package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"klapp/internal/models"
)

func (app *application) punchForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "punch.tmpl", templateData{})
}

func (app *application) privacyPolicy(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "privacy.tmpl", templateData{})
}

func (app *application) termsAndConditions(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "terms.tmpl", templateData{})
}

// punchStatus identifies the worker by PIN and shows whether they're
// currently punched in.
func (app *application) punchStatus(w http.ResponseWriter, r *http.Request) {
	pin := r.PostFormValue("pin")

	worker, err := app.workers.Authenticate(pin)
	if err != nil {
		app.showInvalidPIN(w, r, "punch.tmpl", err)
		return
	}

	data := templateData{Worker: &worker, PIN: pin}

	open, err := app.timePunches.Open(worker.ID)
	if err == nil {
		data.OpenPunch = &open
	} else if !errors.Is(err, models.ErrNoRecord) {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "punch.tmpl", data)
}

func (app *application) punchIn(w http.ResponseWriter, r *http.Request) {
	pin := r.PostFormValue("pin")
	worker, err := app.workers.Authenticate(pin)
	if err != nil {
		app.showInvalidPIN(w, r, "punch.tmpl", err)
		return
	}

	lat, lon, err := parseCoords(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	_, err = app.timePunches.PunchIn(worker.ID, lat, lon, time.Now())
	if err != nil && !errors.Is(err, models.ErrAlreadyOpen) {
		app.serverError(w, r, err)
		return
	}

	open, err := app.timePunches.Open(worker.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "punch.tmpl", templateData{Worker: &worker, OpenPunch: &open, PIN: pin})
}

func (app *application) punchOut(w http.ResponseWriter, r *http.Request) {
	pin := r.PostFormValue("pin")
	worker, err := app.workers.Authenticate(pin)
	if err != nil {
		app.showInvalidPIN(w, r, "punch.tmpl", err)
		return
	}

	lat, lon, err := parseCoords(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.timePunches.PunchOut(worker.ID, lat, lon, time.Now())
	if err != nil && !errors.Is(err, models.ErrNoRecord) {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "punch.tmpl", templateData{Worker: &worker, Flash: "Punched out. See you next time!"})
}

func (app *application) punchLateForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{})
}

// punchLate handles the 9pm late-notice link: worker enters their PIN and
// the time they finished, no location capture.
func (app *application) punchLate(w http.ResponseWriter, r *http.Request) {
	pin := r.PostFormValue("pin")
	worker, err := app.workers.Authenticate(pin)
	if err != nil {
		app.showInvalidPIN(w, r, "punch_late.tmpl", err)
		return
	}

	clockTime, err := time.Parse("15:04", r.PostFormValue("end_time"))
	if err != nil {
		app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{Worker: &worker, Flash: "Enter a valid time."})
		return
	}
	now := time.Now()
	endTime := time.Date(now.Year(), now.Month(), now.Day(), clockTime.Hour(), clockTime.Minute(), 0, 0, time.Local)

	err = app.timePunches.PunchOutLate(worker.ID, endTime)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{Worker: &worker, Flash: "No open punch found for you today."})
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{Worker: &worker, Flash: "Got it, thanks for letting us know."})
}

func (app *application) showInvalidPIN(w http.ResponseWriter, r *http.Request, page string, err error) {
	if errors.Is(err, models.ErrInvalidPIN) {
		app.render(w, r, http.StatusOK, page, templateData{Flash: "PIN not recognized. Try again."})
		return
	}
	app.serverError(w, r, err)
}

func parseCoords(r *http.Request) (lat, lon float64, err error) {
	lat, err = strconv.ParseFloat(r.PostFormValue("lat"), 64)
	if err != nil {
		return 0, 0, err
	}
	lon, err = strconv.ParseFloat(r.PostFormValue("lon"), 64)
	if err != nil {
		return 0, 0, err
	}
	return lat, lon, nil
}
