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
| What handlers exist? | `grep -n '^func (app \*application)' cmd/admin/*.go cmd/punch/*.go` | 46 handlers, one line each |
| What's the DB schema? | `sqlite3 <scratch.db> .schema` | See the `smoke-test` skill for a scratch DB |

`go doc` output is generated from the source, so it can never go stale the way a
checked-in cheatsheet would. Prefer it over reading `internal/models/*.go`,
especially `time_punches.go` (755 lines) and `invoices.go` (242).

## Cheap to read whole (under ~110 lines each)

- `cmd/admin/routes.go` (58) and `cmd/punch/routes.go` (22) — every URL on each site
- `cmd/admin/templates.go` (79) / `cmd/punch/templates.go` (45) — that site's
  `templateData` struct: every field its pages can use. `go doc` won't show it
  (package main), so read the file. **The two structs are separate** — a field
  added for an admin page does not exist on the punch site and vice versa.
- `cmd/admin/helpers.go` / `cmd/punch/helpers.go` (36 each, identical) —
  `render`, `serverError`, `clientError`
- `cmd/admin/middleware.go` (14) — `requireAdmin`, the only middleware
- `internal/config/config.go` (70) — tunables and their defaults, shared by both
- `ui/static/css/main.css` (112) — every class available to templates
- `ui/html/base.tmpl` (16) and `ui/html/partials/nav.tmpl` (14)
- Any single page template — the largest is 76 lines

## Layer map

```
cmd/punch/         the public worker punch site (:4000)
  main.go          application struct, flags, startup
  routes.go        routes()
  handlers.go      PIN entry, punch in/out, late notice
  templates.go     punch templateData + cache (globs punch*.tmpl)
  pinlimiter.go    per-IP PIN guess throttle
cmd/admin/         the LAN/WireGuard-only admin site (:8082)
  main.go          application struct, flags, startup
  routes.go        routes()
  handlers.go      login/logout, dashboard, bulk punch
  handlers_*.go    timesheet, summary, workers, invoices, customers
  templates.go     admin templateData + cache (globs admin_*.tmpl) + template funcs
  middleware.go    requireAdmin
  sweep.go         nightly 9 PM auto punch-out
cmd/invoice/       separate binary, separate site (:8083), own templates
cmd/seedadmin/     CLI to create/reset an admin login
internal/config/   Config + Load, shared by punch and admin
internal/models/   one file per area; raw database/sql, no ORM
db/                Open, Migrate, and the embedded migrations
db/migrations/     goose SQL migrations, applied on startup by every binary
ui/html/pages/     one .tmpl per page, admin_*.tmpl and punch*.tmpl
ui/html/partials/  nav.tmpl (admin nav bar)
deploy/            systemd units, Caddyfile, lib.sh + deploy/update scripts
docs/              design docs: refactor.md, time_reporting_plan.md, security.md
```

Only `internal/` and `db/` are shared. Each `cmd/` is its own `package main`
with its own `application` struct, `templateData`, template cache, and
`render`/`serverError` helpers — `helpers.go` is deliberately duplicated.

## Request flow

`routes.go` → handler in the same `cmd/` directory → model method in
`internal/models/` → `app.render(w, r, status, "page.tmpl", templateData{...})`
→ `ui/html/pages/page.tmpl` wrapped by `base.tmpl`.

Admin routes live on a `protected` mux behind `requireAdmin`; the login page and
static files sit outside it. The punch site has no sessions and no middleware at
all — the worker's PIN is posted with every request.

Both sites serve `./ui/static/` and read templates from `./ui/html/`, so both
must run from a directory containing `ui/`.

## Naming gotcha

There is exactly one "dashboard": `/admin` (`admin_dashboard.tmpl`). The
worker-facing page is "Punch" (`punch.tmpl`), not a dashboard.

## Related skills

- `feature-slice` — conventions for actually making a change across these layers
- `smoke-test` — run it locally and verify
- `deploy-ops` — how it runs in production
