package main

import (
	"bytes"
	"path/filepath"
	"runtime"
	"testing"

	"klapp/internal/models"
)

// projectRoot returns the repo root by walking up from this source file.
func projectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "../..")
}

// TestPunchTmplNilWorker guards against the punch.tmpl template accessing
// fields on a nil Worker when rendering the initial PIN-entry page.
func TestPunchTmplNilWorker(t *testing.T) {
	t.Chdir(projectRoot())

	cache, err := newTemplateCache()
	if err != nil {
		t.Fatalf("newTemplateCache: %v", err)
	}

	cases := []struct {
		name string
		data templateData
	}{
		{"nil worker (initial page load)", templateData{Spanish: true}},
		{"nil worker english", templateData{}},
		{"worker no open punch", templateData{Worker: &models.Worker{WorkerName: "Ana"}}},
		{"worker with open punch", templateData{Worker: &models.Worker{WorkerName: "Ana"}, OpenPunch: &models.TimePunch{}}},
		{"worker require location", templateData{Worker: &models.Worker{WorkerName: "Ana", RequireLocation: true}}},
	}

	ts := cache["punch.tmpl"]
	if ts == nil {
		t.Fatal("punch.tmpl missing from template cache")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := ts.ExecuteTemplate(&buf, "base", tc.data); err != nil {
				t.Errorf("render error: %v", err)
			}
		})
	}
}
