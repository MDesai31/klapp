package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"klapp/internal/models"
)

func (app *application) punchForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "punch.tmpl", templateData{Spanish: true})
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

	worker, err := app.authenticateWorker(r, pin)
	if err != nil {
		app.showInvalidPIN(w, r, "punch.tmpl", err)
		return
	}

	spanish := worker.Language != "english"
	data := templateData{Worker: &worker, PIN: pin, Spanish: spanish}

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
	worker, err := app.authenticateWorker(r, pin)
	if err != nil {
		app.showInvalidPIN(w, r, "punch.tmpl", err)
		return
	}

	lat, lon, err := parseCoords(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if worker.RequireLocation && (lat == nil || lon == nil) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	_, err = app.timePunches.PunchIn(worker.ID, lat, lon, time.Now())
	if errors.Is(err, models.ErrDailyLimitExceeded) {
		spanish := worker.Language != "english"
		flash := pickMsg(spanish, "Has alcanzado el límite de entradas por hoy.", "You've reached today's punch-in limit.")
		app.render(w, r, http.StatusOK, "punch.tmpl", templateData{Worker: &worker, Spanish: spanish, Flash: flash})
		return
	}
	if err != nil && !errors.Is(err, models.ErrAlreadyOpen) {
		app.serverError(w, r, err)
		return
	}

	open, err := app.timePunches.Open(worker.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	spanish := worker.Language != "english"
	app.render(w, r, http.StatusOK, "punch.tmpl", templateData{Worker: &worker, OpenPunch: &open, PIN: pin, Spanish: spanish})
}

func (app *application) punchOut(w http.ResponseWriter, r *http.Request) {
	pin := r.PostFormValue("pin")
	worker, err := app.authenticateWorker(r, pin)
	if err != nil {
		app.showInvalidPIN(w, r, "punch.tmpl", err)
		return
	}

	lat, lon, err := parseCoords(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if worker.RequireLocation && (lat == nil || lon == nil) {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.timePunches.PunchOut(worker.ID, lat, lon, time.Now())
	if err != nil && !errors.Is(err, models.ErrNoRecord) {
		app.serverError(w, r, err)
		return
	}

	spanish := worker.Language != "english"
	flash := pickMsg(spanish, "¡Salida registrada! Hasta la próxima.", "Punched out. See you next time!")
	app.render(w, r, http.StatusOK, "punch.tmpl", templateData{Worker: &worker, Spanish: spanish, Flash: flash})
}

func (app *application) punchLateForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{Spanish: true})
}

// punchLate handles the 9pm late-notice link: worker enters their PIN and
// the time they finished, no location capture.
func (app *application) punchLate(w http.ResponseWriter, r *http.Request) {
	pin := r.PostFormValue("pin")
	worker, err := app.authenticateWorker(r, pin)
	if err != nil {
		app.showInvalidPIN(w, r, "punch_late.tmpl", err)
		return
	}

	spanish := worker.Language != "english"

	clockTime, err := time.Parse("15:04", r.PostFormValue("end_time"))
	if err != nil {
		flash := pickMsg(spanish, "Ingresa una hora válida.", "Enter a valid time.")
		app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{Worker: &worker, Spanish: spanish, Flash: flash})
		return
	}
	now := time.Now()
	endTime := time.Date(now.Year(), now.Month(), now.Day(), clockTime.Hour(), clockTime.Minute(), 0, 0, time.Local)

	err = app.timePunches.PunchOutLate(worker.ID, endTime)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			flash := pickMsg(spanish, "No hay entrada abierta para hoy.", "No open punch found for you today.")
			app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{Worker: &worker, Spanish: spanish, Flash: flash})
			return
		}
		app.serverError(w, r, err)
		return
	}

	flash := pickMsg(spanish, "¡Listo, gracias por avisarnos!", "Got it, thanks for letting us know.")
	app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{Worker: &worker, Spanish: spanish, Flash: flash})
}

// authenticateWorker looks up the worker for pin, applying a fixed delay to
// every check (to slow scripted guessing regardless of outcome) and an IP
// lockout after repeated failures. See pinlimiter.go.
func (app *application) authenticateWorker(r *http.Request, pin string) (models.Worker, error) {
	time.Sleep(app.pinCheckDelay)

	ip := clientIP(r)
	if !app.pinLimiter.allow(ip) {
		return models.Worker{}, errPinLockedOut
	}

	worker, err := app.workers.Authenticate(pin)
	if err != nil {
		if errors.Is(err, models.ErrInvalidPIN) {
			app.pinLimiter.recordFailure(ip)
		}
		return models.Worker{}, err
	}

	app.pinLimiter.recordSuccess(ip)
	return worker, nil
}

func (app *application) showInvalidPIN(w http.ResponseWriter, r *http.Request, page string, err error) {
	if errors.Is(err, errPinLockedOut) {
		app.render(w, r, http.StatusOK, page, templateData{Spanish: true, Flash: "Demasiados intentos. Espera unos minutos e inténtalo de nuevo."})
		return
	}
	if errors.Is(err, models.ErrInvalidPIN) {
		// Worker is unknown at this point; default to Spanish since most workers are Spanish-speaking.
		app.render(w, r, http.StatusOK, page, templateData{Spanish: true, Flash: "PIN no reconocido. Inténtalo de nuevo."})
		return
	}
	app.serverError(w, r, err)
}

func pickMsg(spanish bool, es, en string) string {
	if spanish {
		return es
	}
	return en
}

// parseCoords returns nil pointers when lat/lon fields are empty (location not
// captured). Returns an error only if the fields are non-empty but unparseable.
func parseCoords(r *http.Request) (lat, lon *float64, err error) {
	latStr := r.PostFormValue("lat")
	lonStr := r.PostFormValue("lon")
	if latStr == "" || lonStr == "" {
		return nil, nil, nil
	}
	latVal, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return nil, nil, err
	}
	lonVal, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return nil, nil, err
	}
	return &latVal, &lonVal, nil
}
