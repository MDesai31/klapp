---
name: feature-slice
description: Conventions and checklist for changing klapp across its layers — migration, model method, handler, route, template, test. Use when implementing any klapp feature or fix that touches more than one file, so the change matches existing patterns (sentinel errors, flash messages, POST-redirect-GET, templateData, bilingual worker strings).
---

# Making a change in klapp

Work outside in: model first (it's the testable part), then handler, route,
template. Use `codebase-map` to locate things without reading whole files.

## Checklist

1. **Migration** (only if the schema changes) — `db/migrations/NNNN_name.sql`,
   next number in sequence, with `-- +goose Up` and `-- +goose Down` sections.
   Embedded via `db/migrations.go`; applied automatically on startup *and* in
   every model test. Comment the *why* in SQL comments — existing migrations do
   (see `0012_one_open_punch.sql`).

2. **Model method** — `internal/models/<area>.go`, a method on the area's
   `*XModel` (which holds `DB *sql.DB`). Raw `database/sql`, no ORM. Return a
   sentinel from `internal/models/errors.go` for expected failures
   (`ErrNoRecord`, `ErrAlreadyOpen`, `ErrDuplicatePIN`, `ErrEndBeforeStart`,
   `ErrDailyLimitExceeded`, `ErrInvalidPIN`, `ErrInvalidCredentials`) — add a new
   one there rather than returning bare `errors.New` from a model.

3. **Handler** — `cmd/admin/handlers*.go` or `cmd/punch/handlers.go`. Pick the
   binary first: the two are separate `package main`s and share nothing but
   `internal/` and `db/`.
   ```go
   func (app *application) adminThing(w http.ResponseWriter, r *http.Request) {
       // r.PathValue("id") for path params, r.PostFormValue / r.PostForm for body
       // app.clientError(w, http.StatusBadRequest) for bad input
       // app.serverError(w, r, err) for anything unexpected
       // app.render(w, r, http.StatusOK, "page.tmpl", templateData{...})
   }
   ```
   Map expected model errors to a user-facing message; only unexpected ones go to
   `serverError`. `punchTimeErrorFlash` in `cmd/admin/handlers_timesheet.go` is
   the pattern for reusing that mapping across handlers.

4. **Route** — `cmd/admin/routes.go` or `cmd/punch/routes.go`. Admin routes go on
   the `protected` mux (behind `requireAdmin`); only the login page and static
   files sit outside it. Method + path patterns, e.g.
   `protected.HandleFunc("POST /admin/punch/bulk", app.adminBulkPunch)`.

5. **templateData** — add a field in that binary's `templates.go`, in the
   commented section it belongs to. **Each binary has its own struct** — a field
   added to `cmd/admin/templates.go` does not exist on the punch site. One struct
   serves every page of that site; unused fields stay zero.

6. **Template** — `ui/html/pages/*.tmpl`, defining `"title"` and `"main"`. The
   filename decides which binary parses it: `admin_*.tmpl` → admin, `punch*.tmpl`
   → punch. A new page named outside those patterns will never load. Admin
   pages start with `{{template "nav" .}}`. Style with the classes already in
   `ui/static/css/main.css`. Any JS goes in an inline `<script>` at the bottom of
   the page — there is no build step and no external JS/CSS.

7. **Test** — model methods get tests in `internal/models/*_test.go` using
   `newTestDB(t)` (throwaway SQLite file per test, migrations applied fresh, safe
   in parallel) and helpers like `mustInsertWorker`. Handlers have no HTTP test
   harness; verify them with the `smoke-test` skill. `cmd/punch/templates_test.go`
   guards the punch templates against nil-field panics — extend its case table
   when a page gains a field that can be nil.

Then: `go build ./... && go vet ./... && go test ./...`

## Conventions that bite if you miss them

**Flash messages, two flavors.** After a mutation that redirects, stash it in the
session and pop it in the GET handler:
```go
app.sessionManager.Put(r.Context(), "flash", "Punched in Ana, Bob.")
http.Redirect(w, r, "/admin", http.StatusSeeOther)
// ...and in the GET handler:
Flash: app.sessionManager.PopString(r.Context(), "flash"),
```
When re-rendering a form immediately after a validation error, skip the session
and pass `Flash:` straight into `templateData`. Templates render it as
`{{if .Flash}}<p class="flash">{{.Flash}}</p>{{end}}`.

**POST → redirect → GET** for every mutation, so a refresh doesn't resubmit.

**No CSRF middleware, by design.** The admin site is LAN/WireGuard-only and
fully trusted; the public punch site is the hostile surface. Don't add CSRF
tokens to admin forms without discussing it — and do treat worker-facing input
as untrusted (see `pinlimiter.go`, `docs/security.md`).

**Times.** ISO-8601 `TEXT` in SQLite: UTC RFC3339 for instants
(`t.UTC().Format(time.RFC3339)`), local `2006-01-02` for the `day` and
`pay_period` bucket columns. Parse back with `time.Parse(time.RFC3339, s)` and
display with `.Local()`. Pay periods are fixed 14-day blocks from a Monday
anchor — use `payPeriodStart` / `CurrentPayPeriod`, don't recompute.

**Worker-facing text is bilingual**, admin text is English-only. Workers default
to Spanish: `spanish := worker.Language != "english"`, then
`pickMsg(spanish, "es text", "en text")`. Templates branch on `.Spanish`.

**Nullable columns** scan through `COALESCE(col, '')` or `sql.NullString`. A
model method that returns nothing found returns `ErrNoRecord`, not a zero value.

**`safeURL`** (in `templateFuncs`) is needed for hrefs with non-http schemes like
`sms:` — only for server-built strings, never raw user input.

## Related skills

- `codebase-map` — where things live, cheap lookups
- `smoke-test` — run the change and verify it end to end
- `deploy-ops` — ship it
