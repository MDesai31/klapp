# Security review notes

Findings from a review of the punch-card flow and admin auth. Threat model:
small business internal tool, punch site is public-facing (workers use their
own phones, cellular only, no job-site wifi), admin site is LAN/WireGuard-only.

## Open issues

### High: PIN echoed into page HTML
`ui/html/pages/punch.tmpl:20,27` — after a worker authenticates, their PIN is
written into a hidden form field (`<input type="hidden" name="pin" value="{{.PIN}}">`)
and resubmitted on punch in/out. Visible in page source, browser history, and
any proxy logs. Fix: hold `workerID` server-side (session) instead of
round-tripping the PIN through the client. The punch site currently has no
session middleware at all (`cmd/punch/routes.go` doesn't wrap with
`sessionManager.LoadAndSave`, unlike `cmd/admin/routes.go`) — and since the
split, the punch binary has no session manager at all.

### Medium: No CSRF protection on admin POST endpoints
`cmd/admin/handlers_workers.go`, `cmd/admin/handlers_timesheet.go` — admin
forms (create/edit worker, toggle active, edit punch) have no CSRF tokens.
Low real-world risk since admin is WireGuard/LAN-only, but cheap to add for
defense in depth (e.g. `gorilla/csrf`).

### Medium: PIN field is plaintext in admin panel
`ui/html/pages/admin_workers.tmpl:31` — PIN input is `type="text"`, not
`type="password"`, so it's visible on-screen when an admin enters/edits it.

### Fixed: No lockout on repeated failed PIN attempts
`internal/models/workers.go` (`WorkerModel.Authenticate`) — worker PINs are
compared with a plain `==` (not bcrypt/hashed at all — the earlier note here
claiming bcrypt comparison was wrong; only admin passwords go through
`bcrypt.CompareHashAndPassword` in `internal/models/admins.go`). Nothing
stopped unlimited guessing.

Design wrinkle: `Authenticate` checks the entered PIN against every active
worker's hash (no name/ID picker), so a failed attempt can't be tied to a
specific worker — only to "this client sent a wrong PIN." A cookie/device
counter was considered first, but rejected: a scripted attacker's HTTP
client can simply never persist/resend the cookie, so it provides no real
protection against automated guessing — only against a human retrying in
the same browser tab, which is already self-limiting.

Implemented instead (`cmd/punch/pinlimiter.go`, `cmd/punch/handlers.go`):
an IP-keyed lockout (`pinLimiter`) plus a fixed per-attempt delay applied
regardless of outcome (`app.pinCheckDelay`). Workers are cellular-only (no
job-site wifi — see "Reviewed and OK" below), so carrier-grade NAT means an
IP can be shared by unrelated workers; the lockout threshold is set high
(default 10 failures/15 min, tunable in `config.json`) specifically to make
collateral lockout of a shared connection unlikely while still stopping a
scripted attacker hammering one PIN space. The flat delay (default 250ms)
isn't bypassable by clearing cookies/IP and directly slows any brute-force
attempt. Also added: a per-worker daily punch-in cap (default 3,
`TimePunchModel.DailyPunchLimit`) so a guessed PIN can't be used to spam
punch records even before a lockout trips.

### Low: Session cookie security flags not explicit
`cmd/admin/main.go` — `scs.New()` uses defaults (`HttpOnly` true,
`SameSite=Lax`, but `Secure` false). Fine while everything is plain HTTP;
revisit (`sessionManager.Cookie.Secure = true`) if either site moves to HTTPS.

### Low: Both sites run plain HTTP, no TLS
`cmd/punch/main.go` and `cmd/admin/main.go` — `ListenAndServe` for both the
public worker site and the admin site. Admin is covered by WireGuard/LAN
isolation, and the punch site is deployed off by default (`deploy/update.sh`
needs `--with-punch`), which shrinks the exposed window further. The
public punch site being plain HTTP is worth revisiting if it's ever exposed
beyond a trusted network path.

### Low: GPS coordinates unvalidated (audit note, not a real vuln)
`cmd/punch/handlers.go:135-144` (`parseCoords`) — no bounds checking
(valid range is ±90 lat, ±180 lon), and a client can trivially send fake
coordinates. Acceptable: GPS here is an anti-fraud/audit signal, not an
access control.

### Low: worker PINs stored and compared in plaintext
`internal/models/workers.go` (`WorkerModel.Authenticate`) — PINs are stored
as plain `TEXT` and compared with `==`, not hashed. Unlike admin passwords
(bcrypt, `internal/models/admins.go`), a DB dump reveals every worker's PIN
directly. Lower severity than it sounds: PINs are short by design (a phone
kiosk numeric entry, not a real password) and the DB isn't otherwise
exposed, but worth hashing for defense in depth if this is ever revisited.
Not addressed by the IP-lockout/delay work above, which only slows guessing
attempts — it doesn't protect the PIN at rest.

## Reviewed and OK

- **SQL injection** — all queries use parameterized placeholders (`?`), no
  string concatenation, across all model files.
- **XSS** — templates use `html/template` (auto-escaping), no
  `template.HTML` casts of user input found (`cmd/admin/templates.go`).
- **IP-based buddy-punch detection** — considered and rejected. Workers are
  on cellular only (no job-site wifi), so carrier-grade NAT makes IP useless
  as a signal: many unrelated phones share an IP, and a single phone's IP can
  change mid-session. GPS lat/lon (already captured on punch in/out) is the
  meaningful signal instead.
