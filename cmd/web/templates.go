package main

import (
	"html/template"
	"path/filepath"

	"klapp/internal/models"
)

// templateData holds everything a page template might need. Fields are
// nil/empty unless that particular page sets them.
type templateData struct {
	// worker site
	Worker    *models.Worker
	OpenPunch *models.TimePunch
	PIN       string // echoed into a hidden field so punch in/out can resend it

	// admin site
	Admin               *models.Admin
	DashboardRows       []models.DashboardRow
	TimesheetRows       []models.TimesheetRow
	SummaryRows         []models.PayPeriodSummaryRow
	SummaryDays         []string
	PayPeriods          []string
	CurrentPeriod       string
	Workers             []models.Worker
	ShowInactiveWorkers bool
	Punch               *models.TimePunch

	Flash string
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).ParseFiles("./ui/html/base.tmpl", "./ui/html/partials/nav.tmpl", page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
