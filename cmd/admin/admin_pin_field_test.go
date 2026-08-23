package main

import (
	"bytes"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"klapp/internal/models"
)

// projectRoot returns the repo root by walking up from this source file.
func projectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "../..")
}

// Admin PIN inputs must be password fields so a worker's PIN isn't readable
// on-screen while an admin types or edits it (docs/reference/security.md).
func TestAdminPINInputsArePasswordFields(t *testing.T) {
	t.Chdir(projectRoot())

	cache, err := newTemplateCache()
	if err != nil {
		t.Fatalf("newTemplateCache: %v", err)
	}

	pinInput := regexp.MustCompile(`<input[^>]*name="pin"[^>]*>`)

	cases := []struct {
		page string
		data templateData
	}{
		{"admin_workers.tmpl", templateData{}},
		{"admin_edit_worker.tmpl", templateData{Worker: &models.Worker{WorkerName: "Ana", PIN: "246813"}}},
	}

	for _, tc := range cases {
		t.Run(tc.page, func(t *testing.T) {
			ts := cache[tc.page]
			if ts == nil {
				t.Fatalf("%s missing from template cache", tc.page)
			}
			var buf bytes.Buffer
			if err := ts.ExecuteTemplate(&buf, "base", tc.data); err != nil {
				t.Fatalf("render error: %v", err)
			}
			input := pinInput.FindString(buf.String())
			if input == "" {
				t.Fatal("no pin input found in rendered page")
			}
			if !strings.Contains(input, `type="password"`) {
				t.Errorf("pin input is not a password field: %s", input)
			}
		})
	}
}
