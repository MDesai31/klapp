# Rewrite Plan: Go + MySQL

## Why

The current Next.js 16 + Auth.js v5 beta + Prisma 6 stack has ongoing churn — breaking changes between majors, hundreds of npm dependencies, and framework complexity that outweighs the app's actual needs. This is a small internal CRUD tool and should be treated as one.

Goals:
- Long-running with minimum maintenance
- Simple to edit — no framework magic, no hydration, no RSC lifecycle
- Single binary deployment, no Node runtime
- Stable for years without forced upgrades

## Stack

- **Go** (1.x compatibility guarantee — code written today compiles in 5 years)
- **MySQL** (`database/sql` + `go-sql-driver/mysql`)
- **`html/template`** for server-rendered HTML
- **`alexedwards/scs`** for sessions (admin only)
- **`golang.org/x/crypto/bcrypt`** for PIN hashing

## Project Structure

Follows the same pattern as `~/go-utils/snippetbox`:

```
klapp/
├── cmd/web/
│   ├── main.go         # application struct, DB open, server start
│   ├── routes.go       # all routes, static file server
│   ├── handlers.go     # or split: handlers_jobs.go, handlers_admin.go
│   ├── helpers.go      # serverError, render, etc.
│   ├── middleware.go   # requireAdmin (VPN-only routes)
│   └── templates.go    # templateData struct, template cache
├── internal/models/
│   ├── errors.go
│   ├── users.go
│   ├── customers.go
│   ├── job_types.go
│   ├── materials.go
│   └── job_logs.go
├── go.mod
└── ui/
    ├── html/
    │   ├── base.tmpl
    │   ├── partials/nav.tmpl
    │   └── pages/
    │       ├── login.tmpl       # admin only
    │       ├── punch.tmpl       # worker clock in/out
    │       ├── jobs_new.tmpl    # worker job log submit
    │       ├── hours.tmpl
    │       └── admin_*.tmpl
    └── static/
        └── css/main.css
```

## Authentication

Two separate surfaces with different threat models:

### Worker-facing (public)
- PIN-based, stateless — no sessions, no cookies
- Worker selects their name, enters 4-6 digit PIN
- PIN + worker ID validated on each request
- Actions: clock in, clock out, submit job log
- Rate limiting on PIN attempts to prevent brute force

### Admin-facing (VPN / home wifi only)
- Never exposed to the public internet
- Simple username + password with a session cookie (`alexedwards/scs`)
- Actions: manage users, customers, job types, materials, view reports

## What Goes Away

Compared to the current stack:
- No session management for workers
- No password reset / forgot password flows
- No cookie handling on the worker path
- No ORM — raw SQL, easy to read and change
- No npm, no bundler, no React, no RSC
- No Auth.js beta edge cases
