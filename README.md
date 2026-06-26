# Klaus Field Log

Internal field-operations tool for a landscaping company: workers punch in/out
from a phone, submit job invoices, and admins review time entries and invoices
over LAN/WireGuard.

Go + SQLite. See `refactor.md` for the stack rationale and
`time_reporting_plan.md` for the feature design.

## Requirements

- **Go 1.22+**
- **msmtp** — used to email approved invoices to `mylawncut@aol.com`. Must be
  installed and configured at `/etc/msmtprc` on the server before the invoice
  submit button will deliver email. Install: `sudo apt install msmtp`.

## Quick Start

Automated tests:
```
go test ./...
```

To try it for real:

1. Start the app (runs both servers + migrations):
   ```
   go run ./cmd/web
   ```
2. In another terminal, create your admin login:
   ```
   go run ./cmd/seedadmin -username owner -password <pick-something>
   ```
3. Go to http://localhost:8082/admin/login and log in.
4. On the **Workers** tab, add a worker with a PIN (e.g. name "Test", PIN "1234").
5. Go to http://localhost:4000/punch, enter that PIN, hit **Punch In** (your
   browser will ask for location permission — allow it).
6. Back on the admin site, check `/admin` (dashboard) — should show "In since ...".
7. Punch out from the worker page, then check `/admin/timesheet` — should show
   the full entry with a map link for the location.
8. Try `/admin/punches/{id}/edit` (linked from the timesheet) to correct a time
   and confirm the "Edited" column flips to yes.
9. Try http://localhost:4000/punch/late for the late-punch-out flow.

The dev DB is `db/klapp.db` — delete it (plus the `-shm`/`-wal` files) anytime
to start fresh.

## Run it

```
go run ./cmd/web      # worker punch site (:4000) + admin site (:8082)
go run ./cmd/invoice  # invoice submission site (:8083)
```

- Worker punch site: http://localhost:4000/punch
- Admin site: http://localhost:8082/admin (all interfaces — LAN/WireGuard)
- Invoice site: http://localhost:8083 (all interfaces — LAN/WireGuard only)

SQLite migrations run automatically on startup. The database file
(`db/klapp.db`) is created on first run and is gitignored.

### First admin login

There's no signup page — bootstrap the first admin from the command line:

```
go run ./cmd/seedadmin -username admin -password <something>
```

## Test

```
go test ./...
```

Each test gets its own throwaway SQLite file with migrations applied fresh,
so tests run independently and in parallel — no shared test database to set up.
