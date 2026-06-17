# Day 1 Summary — Go Rewrite, Phase 1 (Punch Card)

Scope: worker-facing punch in/out flow only, per refactor.md and time_reporting_plan.md.
Nothing in the existing Next.js app (`src/`, `prisma/`, etc.) was touched — this is a
new, separate Go module living alongside it.

## What's new

```
go.mod / go.sum            module "klapp"
cmd/web/
  main.go                  app struct, opens SQLite, runs migrations, starts server
  routes.go                 GET/POST routes for the punch flow
  handlers_punch.go         PIN entry, punch in/out, late punch-out
  helpers.go                serverError/clientError/render
  templates.go               templateData + template cache
internal/models/
  errors.go                  ErrNoRecord, ErrInvalidPIN, ErrAlreadyOpen
  workers.go                  WorkerModel: Authenticate(pin), Get(id)
  time_punches.go             TimePunchModel: Open, PunchIn, PunchOut, PunchOutLate
  *_test.go                   unit tests for all of the above (go test ./... passes)
db/
  migrations.go                embeds db/migrations/*.sql for goose
  migrations/0001_init.sql     workers + time_punches tables
ui/
  html/base.tmpl                shared page shell
  html/pages/punch.tmpl          PIN entry -> status -> punch in/out button
  html/pages/punch_late.tmpl     late punch-out (PIN + finish time, no GPS)
  static/css/main.css            big green/red buttons, mobile-friendly
```

## How it works

- Worker hits `/punch`, enters PIN. There's no username/session — the PIN alone
  identifies them, checked via bcrypt against every active worker (fine at ~15 workers).
- If not punched in: green "Punch In" button. Browser geolocation fills hidden
  lat/lon fields before the button enables; submitting posts to `/punch/in`.
- If punched in: shows the punch-in time and a red "Punch Out" button posting to
  `/punch/out`.
- The PIN is re-sent as a hidden field on every action (stateless, no cookies —
  matches refactor.md's auth design for the worker site).
- `/punch/late` is the page a 9pm SMS link would point to: PIN + a finish time,
  no location capture, sets `late = true`. (Sending the actual SMS is not built —
  out of scope for this slice.)
- SQLite migrations run automatically on startup via embedded goose migrations —
  no separate CLI step needed.

## Verified

- `go build ./...`, `go vet ./...`, `gofmt -l .` all clean.
- `go test ./...` passes (PIN auth incl. inactive-worker rejection, punch in/out,
  double-punch-in guard, double-punch-out guard, late flow, pay-period math).
- Ran the server for real and curled through: wrong PIN → correct PIN → punch in →
  re-check status → idempotent punch in → punch out → punch in again → late
  punch-out (wrong PIN, correct PIN, then no-open-punch case). Checked the SQLite
  rows directly to confirm timestamps, lat/lon, and the `late` flag landed correctly.

## Judgment calls worth confirming with you

- **Pay period anchor**: `internal/models/time_punches.go` assumes 14-day pay
  periods counted from 2024-01-01 (a Monday). I made this up to have *something*
  consistent for the `pay_period` column — confirm the real anchor date before
  this is used for actual payroll grouping.
- **No `config` table yet** (storage GPS coords/radius). It's listed under Phase 1
  in refactor.md but is really an admin/settings concern with nothing to attach to
  yet, so I left it out of this slice. Punch in/out already captures lat/lon, so
  wiring up radius flagging later is just adding the config table + a distance
  check, no schema changes needed.

## Not built yet (intentionally out of scope for this slice)

- Admin site entirely (dashboard, timesheet, worker management, settings, SMS
  triggers, pay period reports) — that's the rest of Phase 1.
- Actual SMS sending (Telit modem / AT commands) — only the web page the link
  would point to exists.
- Phase 2 (job log) and Phase 3 (QuickBooks) — untouched.

## Try it yourself

```
go run ./cmd/web
# then, in another shell, insert a worker (no admin UI yet):
# see the seed snippet pattern used during testing - bcrypt-hash a PIN and
# INSERT INTO workers (worker_name, pin, phone, active) VALUES (...)
```
Then visit `http://localhost:4000/punch`. The dev DB (`db/klapp.db`) is gitignored.

---

# Update — Cleanup + Admin Site

## Repo cleanup (this is now a Go-only repo)

`git rm`'d the entire previous Next.js/Prisma/Postgres app and its tests/config:
`src/`, `prisma/`, `public/`, `tests/`, `package.json` + lockfile, `tsconfig.json`,
`next.config.ts`, `eslint.config.mjs`, `playwright.config.ts`, `vitest.config.ts`,
`prisma.config.ts`, `.env.example`, `RUNNING.md`, the old `README.md`, `AGENTS.md`
(Next.js-training-data warning, no longer relevant), and `docs/superpowers/`
(the old Next.js MVP plan/design docs — superseded by `refactor.md` and
`time_reporting_plan.md`). Kept `docs/ideas.md` (stack-agnostic feature
brainstorm, still useful). Wrote a new `README.md` for the Go app and rewrote
`CLAUDE.md` to describe the Go/SQLite stack instead of Prisma/Auth.js/Next.js.
Trimmed `.gitignore` down to the Go-relevant entries.

**Nothing was committed** — all of this is staged (`git rm` stages by nature)
but waiting on you, per your standing "don't commit" instruction.

One thing worth flagging: partway through, `CLAUDE.md` turned up deleted from
disk on its own — not from any command I ran. I restored it from git history
(`git checkout HEAD -- CLAUDE.md`) before editing it, so no content was lost,
but I don't have an explanation for how it happened. Worth keeping an eye out
for if it recurs.

Also fixed per your request: `time_reporting_plan.md` said Nginx for the
reverse proxy; changed to Caddy to match `refactor.md`. And the pay-period
anchor in `internal/models/time_punches.go` is now the real value you gave me
— Monday, 2026-06-08 — replacing the placeholder from last time.

## Admin site (new)

A second HTTP server, started alongside the worker site from the same
`cmd/web` binary:

```
-addr        ":4000"  worker punch site (unchanged)
-admin-addr  ":8082"  admin site - binds all interfaces, reachable over LAN and WireGuard
```

```
db/migrations/0002_admins.sql   admins table (username, password_hash)
internal/models/admins.go        AdminModel: Authenticate(user,pass), Upsert (for seedadmin)
internal/models/workers.go        + List(), Create(), SetActive() (soft delete)
internal/models/time_punches.go    + DashboardStatus(day), ForPayPeriod(period),
                                    PayPeriods(), AdminUpdate(id,start,end), Get(id)
                                    + small display-helper methods used directly
                                    from templates (EndTimeDisplay, StatusLabel, etc.)
cmd/web/middleware.go              requireAdmin (session check, redirects to login)
cmd/web/handlers_admin*.go         login/logout/dashboard, timesheet/edit, worker mgmt
cmd/seedadmin/                     CLI to bootstrap the first admin (no signup page exists)
ui/html/partials/nav.tmpl          admin nav bar
ui/html/pages/admin_*.tmpl         login, dashboard, timesheet, edit-punch, workers
```

Sessions are `alexedwards/scs` with the default in-memory store (matches
refactor.md's "no extra DB table"), 30-day lifetime, so an admin logs in once
and stays in. Worker site is untouched - still no cookies, no sessions.

**Dashboard** (`/admin`): every active worker, today's status - not in / in
since HH:MM / out at HH:MM.

**Timesheet** (`/admin/timesheet?period=YYYY-MM-DD`): every punch in a pay
period, late entries highlighted, an "edited by admin" column, and a map
link per lat/lon pair instead of automatic radius flagging (see below). Links
to switch between every pay period that has data.

**Edit entry** (`/admin/punches/{id}/edit`): corrects start/end time,
sets `modified_by_admin = true`. Leaving end time blank reopens the punch.

**Worker management** (`/admin/workers`): list, add (name + PIN + phone),
deactivate/reactivate.

Bootstrap the first admin (no UI can do this, chicken-and-egg):
```
go run ./cmd/seedadmin -username owner -password <something>
```

## Judgment calls worth knowing about

- **GPS radius**: per your steer, I didn't build the `config` table or any
  distance math. Lat/lon is still captured on every punch and the timesheet
  shows it as a clickable map link (`google.com/maps?q=lat,lon`) for both in
  and out, so a far-away punch is something you'd notice by eye, not
  something the system flags automatically.
- **"Remove worker" is a soft delete.** `SetActive(id, false)` deactivates -
  their PIN stops working and they drop off the dashboard - rather than
  deleting the row, since punch history has a foreign key to `worker_id` and
  payroll records shouldn't disappear. The workers page calls this
  Deactivate/Reactivate rather than Remove.
- **Admin auth is plaintext HTTP, now reachable from the LAN, not just the
  WireGuard tunnel.** refactor.md's original reasoning for skipping HTTPS on
  the admin site ("WireGuard encrypts the tunnel") only holds for traffic
  that actually goes through WireGuard. A device on the plain LAN now sees
  the admin username/password in cleartext on login. I built exactly what
  was asked (bind all interfaces, no TLS), but flagging this in case it
  changes your risk calculus — e.g. restricting port 8082 at the firewall to
  the WireGuard interface only, even though the app itself will happily
  serve LAN requests too.
- No CSRF protection on admin forms, unchanged from refactor.md's original
  call ("admin is VPN-only") - now slightly less true given the above.

## Verified

- `go build ./...`, `go vet ./...`, `gofmt -l .` clean across the whole module.
- `go test ./...` passes - added tests for `AdminModel`, the new
  `TimePunchModel`/`WorkerModel` methods (dashboard status, pay-period
  listing/filtering, admin update, create/list/deactivate).
- Ran both servers for real: logged in, hit `/admin` unauthenticated (redirects
  to login), added a worker through the admin UI, punched them in from the
  worker site, confirmed the dashboard showed "In since" live, checked the
  timesheet (pay period, map links, late/edited columns), edited the entry's
  times and confirmed the change and the `modified_by_admin` flag persisted,
  deactivated the worker and confirmed their PIN stopped working and they
  dropped off the dashboard, and confirmed logout invalidates the session.

## Not built yet

- Settings page (would only matter once/if GPS radius flagging comes back)
- Manual SMS triggers, pay period PDF/email/print reports
- Phase 2 (job log) and Phase 3 (QuickBooks) - still untouched

---

# Update — systemd unit + a real bug found while testing on a phone

## Added `deploy/klapp.service`

A template systemd unit (not installed anywhere — just checked into the repo)
covering the whole app as one process (worker site + admin site both start
from `cmd/web`'s `main()`). Placeholders to fill in before use: `User`,
`WorkingDirectory`, `ExecStart` path. `cmd/seedadmin` isn't a service — still
a one-off command you run by hand. To actually deploy: `go build -o web
./cmd/web`, copy the binary + unit file to the server, `systemctl daemon-reload
&& systemctl enable --now klapp`.

## Bug found: geolocation fails on the worker site over WireGuard (Safari)

You tested `/punch` on Safari over WireGuard and got "Location unavailable,"
unable to clock in. Root cause: the worker site is plain HTTP, and Safari (like
other modern browsers) refuses the Geolocation API on any origin that isn't
`https://` or `localhost` — it fails before even prompting for permission.
WireGuard encrypting the tunnel doesn't make the browser consider the page a
secure context; that's a separate, protocol-level check.

Worth knowing regardless of TLS: the punch handlers currently **hard-reject**
a punch if lat/lon are missing — a worker literally cannot clock in without a
GPS reading. We hadn't decided whether a permission/signal hiccup should ever
be allowed to block clock-in entirely (vs. letting the punch through with a
null location). Flagged, not yet resolved either way.

## Next session: set up Caddy

Decided: Caddy goes in front of the worker site for automatic HTTPS, which
directly fixes the Safari/geolocation issue above (admin site stays off
Caddy/public HTTPS by design — it's not meant to be reachable from the public
internet at all). Still need to figure out, before writing the Caddyfile:
whether there's a real public domain pointed at this box (lets Caddy get a
normal Let's Encrypt cert), or whether this stays WireGuard/LAN-only, in which
case Safari will only trust the cert if Caddy's internal CA root is manually
installed on test phones.
