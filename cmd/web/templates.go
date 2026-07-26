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
	Spanish   bool   // true when the worker's language is spanish (or unknown)

	// admin site — timekeeping
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
	NotifyBaseURL       string

	// admin site — invoices
	Invoices    []models.Invoice
	Invoice     *models.Invoice
	CurrentPage int
	TotalPages  int
	PrevPage    int
	NextPage    int

	// admin site — customers
	Customers   []models.Customer
	Customer    *models.Customer
	SearchQuery string

	// admin site — catalog
	JobDescriptions []models.JobDescription
	Materials       []models.Material

	Flash string
}

// templateFuncs are helpers exposed to html/template pages.
var templateFuncs = template.FuncMap{
	// safeURL marks a server-built URL (e.g. an "sms:" link) as safe to
	// emit verbatim in an href, bypassing html/template's default
	// http/https/mailto-only scheme allowlist. Only use it on strings
	// built from trusted, already-escaped inputs - never raw user input.
	"safeURL": func(s string) template.URL { return template.URL(s) },
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(templateFuncs).ParseFiles("./ui/html/base.tmpl", "./ui/html/partials/nav.tmpl", page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
