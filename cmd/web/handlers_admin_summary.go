package main

import (
	"net/http"
	"time"

	"klapp/internal/models"
)

func (app *application) adminSummary(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = models.CurrentPayPeriod(time.Now())
	}

	rows, err := app.timePunches.PayPeriodSummary(period)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	days, err := models.PayPeriodDays(period)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	dayLabels := make([]string, len(days))
	for i, d := range days {
		dayLabels[i] = models.DayLabel(d)
	}

	periods, err := app.timePunches.PayPeriods()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_summary.tmpl", templateData{
		SummaryRows:   rows,
		SummaryDays:   dayLabels,
		PayPeriods:    periods,
		CurrentPeriod: period,
	})
}
