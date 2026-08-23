package main

import (
	"html/template"
	"path/filepath"

	"klapp/internal/models"
)

// templateData holds everything an admin page might need. Fields are
// nil/empty unless that particular page sets them.
type templateData struct {
	Admin *models.Admin

	// timekeeping
	DashboardRows       []models.DashboardRow
	TimesheetRows       []models.TimesheetRow
	SummaryRows         []models.PayPeriodSummaryRow
	SummaryDays         []summaryDay
	PayPeriods          []string
	CurrentPeriod       string
	Worker              *models.Worker
	Workers             []models.Worker
	ShowInactiveWorkers bool
	Punch               *models.TimePunch
	NotifyBaseURL       string

	// invoices
	Invoices    []models.Invoice
	Invoice     *models.Invoice
	CurrentPage int
	TotalPages  int
	PrevPage    int
	NextPage    int

	// customers
	Customers   []models.Customer
	Customer    *models.Customer
	SearchQuery string

	// catalog
	JobDescriptions []models.JobDescription
	Materials       []models.Material

	Flash string
}

// summaryDay is one day column of the pay period summary: Date is the
// "2006-01-02" value the bulk punch form posts, Label the short header the
// table shows.
type summaryDay struct {
	Date  string
	Label string
}

// templateFuncs are helpers exposed to html/template pages.
var templateFuncs = template.FuncMap{
	// safeURL marks a server-built URL (e.g. an "sms:" link) as safe to
	// emit verbatim in an href, bypassing html/template's default
	// http/https/mailto-only scheme allowlist. Only use it on strings
	// built from trusted, already-escaped inputs - never raw user input.
	"safeURL": func(s string) template.URL { return template.URL(s) },
}

// newTemplateCache parses only the admin pages. The worker-facing punch
// pages live in the same directory but belong to the punch binary.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/admin_*.tmpl")
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
