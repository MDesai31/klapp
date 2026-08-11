@docs/memory/MEMORY.md

## Project
Klaus Field Log — Go + SQLite. Three binaries, one database: the public worker
punch site (`cmd/punch`), the LAN/WireGuard-only admin site (`cmd/admin`), and the
invoice site (`cmd/invoice`). See `refactor.md` and `time_reporting_plan.md` for
the design.

## Running
- Worker punch site: `go run ./cmd/punch` (default `:4000`)
- Admin site: `go run ./cmd/admin` (default `:8082`), bound to all interfaces so
  it's reachable over LAN and WireGuard
- The two are independent processes — the admin site runs fine with the punch site
  switched off, which is the default deploy (see `deploy/update.sh`). The nightly
  9 PM punch-out sweep therefore lives in `cmd/admin`, the always-on binary.
- Bootstrap the first admin login: `go run ./cmd/seedadmin -username <name> -password <pw>`
- DB migrations (goose, embedded) run automatically on startup against `db/klapp.db`
  (gitignored, created on first run) — every binary applies them, goose is idempotent
- Tunable settings (PIN lockout threshold/window/cooldown, per-attempt delay, daily
  punch-in cap, punch site URL) live in a JSON config file read by both sites, flag
  `-config` (default `config.json`, gitignored; see `config.example.json`). Missing
  file falls back to built-in defaults — see `internal/config/config.go`.

## Conventions
- Project layout follows `~/go-utils/snippetbox` (app struct in each binary's
  `main.go`, models in `internal/models/`, html/template pages in `ui/html/pages/`)
- Each binary is its own `package main` with its own `templateData`, template cache,
  and `render`/`serverError` helpers. Only `internal/` and `db/` are shared. The
  template cache globs just that site's pages (`punch*.tmpl` vs `admin_*.tmpl`).
- Dates/times are stored as ISO-8601 `TEXT` in SQLite (UTC instants, local-date
  day/pay-period buckets) — see `internal/models/time_punches.go`
- No ORM — raw SQL via `database/sql`

## Testing
- `go test ./...` — model tests use a throwaway SQLite file per test (`t.TempDir()`),
  migrations applied fresh each run. No shared test DB, no race conditions, tests
  can run in parallel.

## Skills
`.claude/skills/` holds task guides Claude loads on demand: `codebase-map`
(finding things without reading whole files), `feature-slice` (conventions for a
change spanning model/handler/route/template), `smoke-test` (scripted local run
+ curl helpers), `deploy-ops` (systemd, /opt/klapp, update.sh).

## Manual/local testing
- `klapp-admin.service` and `klapp-punch.service` (systemd) run the real binaries on
  default ports `:8082`/`:4000` against `/opt/klapp`'s db — never reuse those ports
  or that db for ad-hoc testing. `.claude/skills/smoke-test/scripts/scratch-up.sh`
  starts both, seeded, on `:18082`/`:14000` instead; `scratch-down.sh` stops them.
- `workers.phone` is nullable in the schema; the Go model scans it through
  `COALESCE(phone, '')`, so hand-inserted rows with `NULL` phone are safe, but prefer
  `phone = ''` for consistency with what the app writes.

## Key files
- Worker site: `cmd/punch/handlers.go`, `cmd/punch/routes.go`, `cmd/punch/templates.go`.
- Admin site: `cmd/admin/handlers*.go`, `cmd/admin/routes.go`, `cmd/admin/templates.go`.
- Deploy: `deploy/lib.sh` holds the build/unit/service logic shared by
  `deploy/deploy.sh` (first-time setup) and `deploy/update.sh` (routine push).
- There is exactly one "dashboard": `/admin` (`admin_dashboard.tmpl`), backed by
  `models.DashboardStatus`/`DashboardRow` in `internal/models/time_punches.go`.
  The worker-facing page is `punch.tmpl` ("Punch"), not a dashboard.
