## Project
Klaus Field Log — Go + SQLite. Two binaries: the public worker punch site and the
LAN/WireGuard-only admin site. See `refactor.md` and `time_reporting_plan.md` for
the design.

## Running
- Worker site: `go run ./cmd/web` (default `:4000`)
- Admin site: started by the same binary on a second port (default `:8082`,
  flag `-admin-addr`), bound to all interfaces so it's reachable over LAN and WireGuard
- Bootstrap the first admin login: `go run ./cmd/seedadmin -username <name> -password <pw>`
- DB migrations (goose, embedded) run automatically on startup against `db/klapp.db`
  (gitignored, created on first run)
- Tunable settings (PIN lockout threshold/window/cooldown, per-attempt delay, daily
  punch-in cap) live in a JSON config file, flag `-config` (default `config.json`,
  gitignored; see `config.example.json`). Missing file falls back to built-in
  defaults — see `cmd/web/config.go`.

## Conventions
- Project layout follows `~/go-utils/snippetbox` (app struct in `cmd/web/main.go`,
  models in `internal/models/`, html/template pages in `ui/html/pages/`)
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
- `klapp.service` (systemd) already runs the real binary on default ports `:4000`/`:8082`
  against `/opt/klapp`'s db — never reuse those ports or that db for ad-hoc testing.
  `.claude/skills/smoke-test/scripts/scratch-up.sh` starts a seeded throwaway
  instance on `:14000`/`:18082` instead; `scratch-down.sh` stops it.
- `workers.phone` is nullable in the schema; the Go model scans it through
  `COALESCE(phone, '')`, so hand-inserted rows with `NULL` phone are safe, but prefer
  `phone = ''` for consistency with what the app writes.

## Key files
- Worker site handlers: `cmd/web/handlers_punch.go`. Admin site handlers:
  `cmd/web/handlers_admin.go`. Routes: `cmd/web/routes.go`. Template data struct:
  `cmd/web/templates.go`.
- There is exactly one "dashboard": `/admin` (`admin_dashboard.tmpl`), backed by
  `models.DashboardStatus`/`DashboardRow` in `internal/models/time_punches.go`.
  The worker-facing page is `punch.tmpl` ("Punch"), not a dashboard.
