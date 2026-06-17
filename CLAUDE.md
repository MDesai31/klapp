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
