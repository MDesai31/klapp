---
name: codebase-map
description: Orient in the klapp codebase without reading whole files. Where each layer lives (routes, handlers, models, templates, migrations) plus the cheap lookup commands that answer "what methods, fields, or routes exist?" in a few hundred tokens instead of thousands. Use at the start of any klapp task, before opening source files.
---

# Finding things in klapp

Go + SQLite, ~5.3k lines. Small enough that most files are cheap to read whole —
but a few are not, and there are faster ways to answer the common questions.

## Don't read these whole — query them instead

| Question | Command | Why |
|---|---|---|
| What models/types exist? | `go doc klapp/internal/models` | ~110 words for all 17 types |
| What can I call on a model? | `go doc klapp/internal/models.TimePunchModel` | ~150 words vs 755 lines of `time_punches.go` |
| What fields/methods on a row type? | `go doc klapp/internal/models.DashboardRow` | Includes the doc comments |
| What sentinel errors exist? | `go doc klapp/internal/models` (var block at top) | Or read `internal/models/errors.go`, 13 lines |
| What handlers exist? | `grep -n '^func (app \*application)' cmd/web/*.go` | 46 handlers, one line each |
| What's the DB schema? | `sqlite3 <scratch.db> .schema` | See the `smoke-test` skill for a scratch DB |

`go doc` output is generated from the source, so it can never go stale the way a
checked-in cheatsheet would. Prefer it over reading `internal/models/*.go`,
especially `time_punches.go` (755 lines) and `invoices.go` (242).

## Cheap to read whole (under ~110 lines each)

- `cmd/web/routes.go` (75) — every URL on both sites, in one place
- `cmd/web/templates.go` (81) — the `templateData` struct: every field any page can use.
  `go doc` won't show it (package main), so read the file.
- `cmd/web/helpers.go` (36) — `render`, `serverError`, `clientError`
- `cmd/web/middleware.go` (14) — `requireAdmin`, the only middleware
- `cmd/web/config.go` (66) — tunables and their defaults
- `ui/static/css/main.css` (112) — every class available to templates
- `ui/html/base.tmpl` (16) and `ui/html/partials/nav.tmpl` (14)
- Any single page template — the largest is 76 lines

## Layer map

```
cmd/web/           the main binary: worker punch site (:4000) + admin site (:8082)
  main.go          application struct, flags, startup, nightly 9PM sweep
  routes.go        routes() = worker site; adminRoutes() = admin site
  handlers_punch.go        worker-facing: PIN entry, punch in/out, late notice
  handlers_admin.go        admin login/logout, dashboard, bulk punch
  handlers_admin_*.go      timesheet, summary, workers, invoices, customers
  templates.go     templateData + newTemplateCache + template funcs
  pinlimiter.go    per-IP PIN guess throttle
cmd/invoice/       separate binary, separate site (:8083), own templates
cmd/seedadmin/     CLI to create/reset an admin login
internal/models/   one file per area; raw database/sql, no ORM
db/migrations/     goose SQL migrations, embedded, applied on startup
ui/html/pages/     one .tmpl per page, admin_*.tmpl and punch*.tmpl
ui/html/partials/  nav.tmpl (admin nav bar)
deploy/            systemd units, Caddyfile, deploy/update scripts
docs/              design docs: refactor.md, time_reporting_plan.md, security.md
```

## Request flow

`routes.go` → handler in `cmd/web/handlers_*.go` → model method in
`internal/models/` → `app.render(w, r, status, "page.tmpl", templateData{...})`
→ `ui/html/pages/page.tmpl` wrapped by `base.tmpl`.

Admin routes live on a `protected` mux behind `requireAdmin`; the login page and
static files sit outside it. The worker site and admin site are two `http.Server`s
over two different `http.Handler`s from the same binary and the same DB.

## Naming gotcha

There is exactly one "dashboard": `/admin` (`admin_dashboard.tmpl`). The
worker-facing page is "Punch" (`punch.tmpl`), not a dashboard.

## Related skills

- `feature-slice` — conventions for actually making a change across these layers
- `smoke-test` — run it locally and verify
- `deploy-ops` — how it runs in production
