package main

import (
	"html/template"
	"path/filepath"

	"klapp/internal/models"
)

// templateData holds everything a punch page might need. Fields are
// nil/empty unless that particular page sets them.
type templateData struct {
	Worker    *models.Worker
	OpenPunch *models.TimePunch
	PIN       string // echoed into a hidden field so punch in/out can resend it
	Spanish   bool   // true when the worker's language is spanish (or unknown)

	Flash string
}

// newTemplateCache parses only the worker-facing pages. The admin pages
// live in the same directory but belong to the admin binary, and they pull
// in the admin nav partial and admin-only templateData fields, so parsing
// them here would fail at render time rather than usefully.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/punch*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).ParseFiles("./ui/html/base.tmpl", page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
