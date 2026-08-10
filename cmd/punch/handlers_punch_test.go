package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"klapp/db"
	"klapp/internal/models"
)

// testPIN is distinctive enough that it cannot appear in rendered times,
// IDs, or other page content by coincidence.
const testPIN = "246813"

// newTestApp builds an application backed by a throwaway SQLite file with
// migrations applied, mirroring the wiring in main().
func newTestApp(t *testing.T) *application {
	t.Helper()
	t.Chdir(projectRoot())

	sqlDB, err := db.Open("file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	if err := db.Migrate(sqlDB); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatalf("newTemplateCache: %v", err)
	}

	app := &application{
		logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		workers:       &models.WorkerModel{DB: sqlDB},
		timePunches:   &models.TimePunchModel{DB: sqlDB, DailyPunchLimit: 3},
		templateCache: templateCache,
		punchSessions: newPunchSessionManager(),
		pinLimiter:    newPinLimiter(10, 15*time.Minute, 15*time.Minute),
		pinCheckDelay: 0,
	}
	return app
}

// newPunchClient returns a test server for the worker site plus a client
// that carries cookies between requests, like a phone browser would.
func newPunchClient(t *testing.T, app *application) (*httptest.Server, *http.Client) {
	t.Helper()
	srv := httptest.NewServer(app.routes())
	t.Cleanup(srv.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar: %v", err)
	}
	return srv, &http.Client{Jar: jar}
}

func postForm(t *testing.T, c *http.Client, url string, form url.Values) string {
	t.Helper()
	resp, err := c.PostForm(url, form)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST %s: status %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(body)
}

// The worker's PIN must never be written into page HTML: it would be
// visible in page source, browser history, and proxy logs.
func TestPunchStatusDoesNotEchoPIN(t *testing.T) {
	app := newTestApp(t)
	if _, err := app.workers.Create("Ana", testPIN, "", 20, "spanish", false); err != nil {
		t.Fatalf("create worker: %v", err)
	}
	srv, client := newPunchClient(t, app)

	body := postForm(t, client, srv.URL+"/punch", url.Values{"pin": {testPIN}})

	if !strings.Contains(body, "Ana") {
		t.Fatalf("expected authenticated page greeting the worker, got:\n%s", body)
	}
	if strings.Contains(body, testPIN) {
		t.Errorf("PIN %q is echoed into the page HTML", testPIN)
	}
}

// After entering the PIN once, punch in must work off the server-side
// session — the browser sends no PIN with the punch form.
func TestPunchInUsesSessionNotPIN(t *testing.T) {
	app := newTestApp(t)
	workerID, err := app.workers.Create("Ana", testPIN, "", 20, "spanish", false)
	if err != nil {
		t.Fatalf("create worker: %v", err)
	}
	srv, client := newPunchClient(t, app)

	postForm(t, client, srv.URL+"/punch", url.Values{"pin": {testPIN}})
	body := postForm(t, client, srv.URL+"/punch/in", url.Values{})

	if strings.Contains(body, "PIN no reconocido") {
		t.Fatal("punch in without a pin field was treated as a failed PIN attempt")
	}
	if _, err := app.timePunches.Open(workerID); err != nil {
		t.Errorf("expected an open punch after punch in via session, got: %v", err)
	}
}

// Same for punch out: the session, not a resent PIN, identifies the worker.
func TestPunchOutUsesSessionNotPIN(t *testing.T) {
	app := newTestApp(t)
	workerID, err := app.workers.Create("Ana", testPIN, "", 20, "spanish", false)
	if err != nil {
		t.Fatalf("create worker: %v", err)
	}
	if _, err := app.timePunches.PunchIn(workerID, nil, nil, time.Now()); err != nil {
		t.Fatalf("punch in fixture: %v", err)
	}
	srv, client := newPunchClient(t, app)

	postForm(t, client, srv.URL+"/punch", url.Values{"pin": {testPIN}})
	body := postForm(t, client, srv.URL+"/punch/out", url.Values{})

	if strings.Contains(body, "PIN no reconocido") {
		t.Fatal("punch out without a pin field was treated as a failed PIN attempt")
	}
	if _, err := app.timePunches.Open(workerID); err == nil {
		t.Error("expected no open punch after punch out via session")
	}
}

// Hitting punch in with no session (expired or never established) must not
// error out — the worker lands back on the PIN form.
func TestPunchInWithoutSessionShowsPINForm(t *testing.T) {
	app := newTestApp(t)
	if _, err := app.workers.Create("Ana", testPIN, "", 20, "spanish", false); err != nil {
		t.Fatalf("create worker: %v", err)
	}
	srv, client := newPunchClient(t, app)

	body := postForm(t, client, srv.URL+"/punch/in", url.Values{})

	if !strings.Contains(body, `name="pin"`) {
		t.Errorf("expected the PIN entry form, got:\n%s", body)
	}
}
