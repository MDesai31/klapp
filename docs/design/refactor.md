# Rewrite Plan: Go + SQLite

---

## Why

The current Next.js 16 + Auth.js v5 beta + Prisma 6 stack has ongoing churn — breaking changes between majors, hundreds of npm dependencies, and framework complexity that outweighs the app's actual needs. This is a small internal CRUD tool and should be treated as one.

Goals:
- Long-running with minimum maintenance
- Simple to edit — no framework magic, no hydration, no RSC lifecycle
- Single binary deployment, no Node runtime
- Stable for years without forced upgrades

---

## Stack

- **Go** (1.x compatibility guarantee — code written today compiles in 5 years)
- **SQLite** (`database/sql` + `modernc.org/sqlite` — pure Go, no cgo, keeps cross-compilation and the single-binary build simple)
- **`html/template`** for server-rendered HTML
- **`golang.org/x/crypto/bcrypt`** for PIN and password hashing
- **`alexedwards/scs`** for admin session cookies (encrypted, no DB table)
- **`pressly/goose`** for database migrations (numbered SQL files, `goose up/down`)
- **Caddy** as reverse proxy for the public-facing worker site (automatic HTTPS)

---

## Project Structure

Follows the same pattern as `~/go-utils/snippetbox`:

```
klapp/
├── cmd/web/
│   ├── main.go           # application struct, DB open, server start
│   ├── routes.go         # all routes, static file server
│   ├── handlers.go       # or split per phase: handlers_punch.go, handlers_jobs.go, handlers_admin.go
│   ├── helpers.go        # serverError, render, etc.
│   ├── middleware.go     # requireAdmin
│   └── templates.go      # templateData struct, template cache
├── internal/models/
│   ├── errors.go
│   ├── workers.go        # Phase 1
│   ├── time_punches.go   # Phase 1
│   ├── config.go         # Phase 1 (storage coords, radius, etc.)
│   ├── customers.go      # Phase 2
│   ├── job_types.go      # Phase 2
│   ├── materials.go      # Phase 2
│   └── job_logs.go       # Phase 2
├── db/
│   ├── klapp.db           # SQLite database file (gitignored)
│   └── migrations/       # goose SQL files (0001_init.sql, 0002_add_jobs.sql, etc.)
├── go.mod
└── ui/
    ├── html/
    │   ├── base.tmpl
    │   ├── partials/nav.tmpl
    │   └── pages/
    │       ├── punch.tmpl          # Phase 1 — worker clock in/out
    │       ├── punch_late.tmpl     # Phase 1 — late punch-out via SMS link
    │       ├── admin_dashboard.tmpl
    │       ├── admin_timesheet.tmpl
    │       ├── admin_workers.tmpl
    │       ├── admin_settings.tmpl
    │       ├── admin_login.tmpl
    │       ├── jobs_new.tmpl       # Phase 2
    │       └── admin_jobs.tmpl     # Phase 2
    └── static/
        └── css/main.css
```

---

## Authentication

### Worker-facing (public)
- PIN-based, stateless — no sessions, no cookies
- Worker enters PIN → backend identifies worker, returns name + punch status
- PIN + worker ID validated on each request via bcrypt
- Rate limiting on PIN attempts

### Admin-facing (WireGuard VPN only)
- Username + password login form
- `alexedwards/scs` with encrypted cookie-based sessions — no extra DB table
- Long session lifetime (days/weeks) — log in once, stay logged in
- `requireAdmin` middleware redirects to login if session is missing
- HTTP is fine — WireGuard encrypts the tunnel

---

## Database Schema

SQLite has no enforced `VARCHAR(n)` length and no native `DECIMAL`/`DATETIME` types — columns below use SQLite's actual storage classes (`TEXT`, `INTEGER`, `REAL`). Dates/times are stored as ISO-8601 `TEXT` (sorts and compares correctly as a string). Two pragmas are set on every connection: `foreign_keys = ON` (off by default in SQLite) and `journal_mode = WAL` (lets reads proceed without blocking on the single writer).

### `workers`
| Column | Type | Notes |
|---|---|---|
| id | INTEGER PK AUTOINCREMENT | |
| worker_name | TEXT | |
| pin | TEXT | bcrypt hash |
| phone | TEXT | for SMS |
| active | BOOLEAN DEFAULT TRUE | |

### `time_punches`
| Column | Type | Notes |
|---|---|---|
| id | INTEGER PK AUTOINCREMENT | |
| worker_id | INTEGER FK → workers.id | |
| pay_period | TEXT | ISO-8601 date, start of the bi-weekly pay period |
| day | TEXT | ISO-8601 date |
| start_time | TEXT | ISO-8601 datetime |
| end_time | TEXT NULL | null until punched out |
| start_lat | REAL | |
| start_lon | REAL | |
| end_lat | REAL NULL | |
| end_lon | REAL NULL | |
| late | BOOLEAN DEFAULT FALSE | true if submitted via 9pm SMS link |
| modified_by_admin | BOOLEAN DEFAULT FALSE | |

### `config`
| Column | Type | Notes |
|---|---|---|
| key | TEXT PK | |
| value | TEXT | |

Config keys: `storage_lat`, `storage_lon`, `storage_radius_miles`

Job-related tables (Phase 2): `customers`, `job_types`, `materials`, `job_logs`, junction tables.

---

## Deployment

- **Worker site**: Public HTTPS via Caddy, region-locked to USA (IP geolocation, best-effort)
- **Admin site**: HTTP only, accessible over WireGuard VPN — no public exposure
- **SMS**: Telit modem on the server, AT commands for 6pm reminders and 9pm notices
- **Cron**: 6pm reminder and 9pm late-notice run as system cron jobs
- **Database**: SQLite file on local disk, no separate server process — back up by copying the file (or stream continuously with `litestream` if off-box backup is wanted)

---

## Build Phases

### Phase 1 — Punch In / Out
Full design at time_reporting_plan.md
Worker site:
- Enter PIN → identify worker, show punch status
- Large green Punch In / red Punch Out button
- Browser sends GPS coords + datetime on punch
- Late punch-out flow: SMS link → worker enters PIN → submits finish time, `late = true`

Admin site:
- Dashboard: who is in / not in today
- Timesheet: full punch table for current pay period, late entries highlighted, out-of-radius flags
- Edit entry (sets `modified_by_admin = true`)
- Worker management: list, add, remove
- Pay period history
- Settings: storage GPS coords + radius
- Manual SMS triggers: 6pm reminder, 9pm notice
- Pay period report: PDF per worker, email to admin, print via `lp`

### Phase 2 — Job Log
- Worker submits job log form (customer, job types, materials, workers on site, notes, times)
- Admin views and manages job logs
- Admin manages lookup tables (customers, job types, materials)

### Phase 3 — QuickBooks Integration
- Export pay period data in a format QuickBooks can import
- Possibly a standalone Python script using the `python-quickbooks` SDK
- OAuth flow to Intuit — likely easier in Python than Go given available libraries
- Runs separately from the main Go app; not coupled to it

---

## Testing

Go's built-in `testing` package — no external framework needed. Test files live alongside the code they test (`_test.go`). Run with `go test ./...`.

### What to test

**Model methods** (`internal/models/`) — highest priority. Silent bugs here lose punch records or accept wrong PINs, which directly affects payroll.
- `workers.go`: PIN lookup and bcrypt verification
- `time_punches.go`: insert punch, fetch current pay period, late flag logic

**PIN validation handler** — wrong behavior here locks workers out entirely.

**Skip for now:**
- Admin CRUD handlers — low risk, easy to spot manually with 15 users
- End-to-end tests — overkill for this scope

### Test database

Each test run uses its own throwaway SQLite database — a temp file (`t.TempDir()`) or an in-memory DB (`file::memory:?cache=shared`). Goose migrations apply at the start of the run. No separate test server, no `DATABASE_URL_TEST`, no shared-state races — tests can run in parallel instead of needing `maxWorkers: 1` like the current Next.js setup.

Reference: Let's Go chapter 13 (`~/go-utils/lets-go-professional-package.html/13.00-testing.html`) covers the full pattern including mocking, integration tests, and assertion helpers.

---

## What Goes Away (vs current Next.js stack)

- No session management for workers
- No password reset / forgot password flows
- No cookies on the worker path
- No ORM — raw SQL, easy to read and change
- No separate database server to install, patch, or back up — SQLite is one file next to the binary
- No npm, no bundler, no React, no RSC
- No Auth.js beta edge cases
- No CSRF (worker path has no cookies; admin is VPN-only)
