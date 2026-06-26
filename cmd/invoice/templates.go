package main

import (
	"html/template"
	"path/filepath"

	"klapp/internal/models"
)

type templateData struct {
	Worker  *models.Worker
	Spanish bool
	Flash   string

	// Re-populate form on validation error
	FormDate         string
	FormHouseNumber  string
	FormCustomerName string
	FormCustomerID   string
	FormNoOfWorkers  string
	FormTimeArrived  string
	FormTimeLeft     string
	FormJobs         []string
	FormMaterials    []string
	FormComments     string
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/invoice/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles("./ui/html/invoice/base.tmpl", page)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}

	return cache, nil
}
