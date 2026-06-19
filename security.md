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
session middleware at all (`cmd/web/routes.go` doesn't wrap with
`sessionManager.LoadAndSave`, unlike `adminRoutes()`).

### Medium: No CSRF protection on admin POST endpoints
`cmd/web/handlers_admin_workers.go`, `handlers_admin_timesheet.go` — admin
forms (create/edit worker, toggle active, edit punch) have no CSRF tokens.
Low real-world risk since admin is WireGuard/LAN-only, but cheap to add for
defense in depth (e.g. `gorilla/csrf`).

### Medium: PIN field is plaintext in admin panel
`ui/html/pages/admin_workers.tmpl:31` — PIN input is `type="text"`, not
`type="password"`, so it's visible on-screen when an admin enters/edits it.

### Low: No lockout on repeated failed PIN/admin password attempts
`internal/models/workers.go:24` (`WorkerModel.Authenticate`), `internal/models/admins.go`
— PINs are bcrypt-hashed and compared via `bcrypt.CompareHashAndPassword`
(constant-time, no timing leak), but nothing stops unlimited guessing.

**In progress:** adding a 4-attempt lockout with the count shown to the
worker as they type. Design wrinkle: `Authenticate` checks the entered PIN
against every active worker's hash (no name/ID picker), so a failed attempt
can't be tied to a specific worker — only to "this browser/device typed a
wrong PIN." Decision: lock by device/session (cookie-based attempt counter),
auto-expiring after a cooldown. No admin-unlock step, no schema tied to
workers — avoids listing employee names on the public kiosk page, which a
name-picker-first flow would require.

### Low: Session cookie security flags not explicit
`cmd/web/main.go:55-56` — `scs.New()` uses defaults (`HttpOnly` true,
`SameSite=Lax`, but `Secure` false). Fine while everything is plain HTTP;
revisit (`sessionManager.Cookie.Secure = true`) if either site moves to HTTPS.

### Low: Both sites run plain HTTP, no TLS
`cmd/web/main.go:71,76` — `http.ListenAndServe` for both the public worker
site and the admin site. Admin is covered by WireGuard/LAN isolation. The
public punch site being plain HTTP is worth revisiting if it's ever exposed
beyond a trusted network path.

### Low: GPS coordinates unvalidated (audit note, not a real vuln)
`cmd/web/handlers_punch.go:135-144` (`parseCoords`) — no bounds checking
(valid range is ±90 lat, ±180 lon), and a client can trivially send fake
coordinates. Acceptable: GPS here is an anti-fraud/audit signal, not an
access control.

## Reviewed and OK

- **PIN hashing/comparison** — bcrypt with constant-time compare, no timing
  attack surface (`internal/models/workers.go`).
- **SQL injection** — all queries use parameterized placeholders (`?`), no
  string concatenation, across all model files.
- **XSS** — templates use `html/template` (auto-escaping), no
  `template.HTML` casts of user input found (`cmd/web/templates.go`).
- **IP-based buddy-punch detection** — considered and rejected. Workers are
  on cellular only (no job-site wifi), so carrier-grade NAT makes IP useless
  as a signal: many unrelated phones share an IP, and a single phone's IP can
  change mid-session. GPS lat/lon (already captured on punch in/out) is the
  meaningful signal instead.
