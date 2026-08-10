package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"klapp/internal/models"
)

// newPunchSessionManager configures the worker-site session store: it holds
// only the authenticated worker's ID so punch in/out never resend the PIN.
func newPunchSessionManager() *scs.SessionManager {
	sm := scs.New()
	sm.Lifetime = 30 * time.Minute
	// Distinct cookie name so it can never collide with the admin
	// session when both sites are reached via the same host.
	sm.Cookie.Name = "punch_session"
	return sm
}

func (app *application) punchForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "punch.tmpl", templateData{Spanish: true})
}

// punchStatus identifies the worker by PIN, stores their ID in the punch
// session, and shows whether they're currently punched in. The PIN itself
// never goes back to the client.
func (app *application) punchStatus(w http.ResponseWriter, r *http.Request) {
	pin := r.PostFormValue("pin")

	worker, err := app.authenticateWorker(r, pin)
	if err != nil {
		app.showInvalidPIN(w, r, "punch.tmpl", err)
		return
	}

	if err := app.punchSessions.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.punchSessions.Put(r.Context(), "workerID", worker.ID)

	spanish := worker.Language != "english"
	data := templateData{Worker: &worker, Spanish: spanish}

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
	worker, ok := app.sessionWorker(w, r)
	if !ok {
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
	app.render(w, r, http.StatusOK, "punch.tmpl", templateData{Worker: &worker, OpenPunch: &open, Spanish: spanish})
}

func (app *application) punchOut(w http.ResponseWriter, r *http.Request) {
	worker, ok := app.sessionWorker(w, r)
	if !ok {
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
	// A clock time that hasn't happened yet today means the worker is
	// reporting yesterday's finish after midnight (e.g. entering "22:30"
	// at 7 AM the next morning) - attach it to yesterday instead.
	if endTime.After(now) {
		endTime = endTime.AddDate(0, 0, -1)
	}

	err = app.timePunches.PunchOutLate(worker.ID, endTime)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			flash := pickMsg(spanish, "No hay entrada abierta para hoy.", "No open punch found for you today.")
			app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{Worker: &worker, Spanish: spanish, Flash: flash})
			return
		}
		if errors.Is(err, models.ErrEndBeforeStart) {
			flash := pickMsg(spanish, "Esa hora es antes de tu hora de entrada. Revisa e inténtalo de nuevo.", "That time is before your punch-in time. Check it and try again.")
			app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{Worker: &worker, Spanish: spanish, Flash: flash})
			return
		}
		app.serverError(w, r, err)
		return
	}

	flash := pickMsg(spanish, "¡Listo, gracias por avisarnos!", "Got it, thanks for letting us know.")
	app.render(w, r, http.StatusOK, "punch_late.tmpl", templateData{Worker: &worker, Spanish: spanish, Flash: flash})
}

// sessionWorker resolves the punch session to an active worker. When the
// session is missing/expired (or the worker was deactivated mid-session) it
// renders the PIN entry form and reports !ok — the caller just returns.
func (app *application) sessionWorker(w http.ResponseWriter, r *http.Request) (models.Worker, bool) {
	id := app.punchSessions.GetInt(r.Context(), "workerID")
	if id != 0 {
		worker, err := app.workers.Get(id)
		if err == nil && worker.Active {
			return worker, true
		}
		if err != nil && !errors.Is(err, models.ErrNoRecord) {
			app.serverError(w, r, err)
			return models.Worker{}, false
		}
		app.punchSessions.Remove(r.Context(), "workerID")
	}

	// Worker is unknown here; default to Spanish like showInvalidPIN.
	app.render(w, r, http.StatusOK, "punch.tmpl", templateData{Spanish: true, Flash: "Tu sesión expiró. Ingresa tu PIN de nuevo."})
	return models.Worker{}, false
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
