# Action Items (open)

Open, actionable work. Added by `/project-log action`; on completion, `/project-log done`
moves the record to `resolved.md`. Records use the action template from the conventions doc.

### A-20260809-msmtp-server-setup — Wire up msmtp on the server for invoice email
- Workstream: ops
- Status: open
- Created: 2026-08-09

Invoice "Submit & mark reviewed" shells out to `msmtp` targeting `mylawncut@aol.com`; until
msmtp is installed and configured on the server the email silently fails (invoice still marks
reviewed, with a flash). Setup steps live in `README.md` § msmtp setup (AOL app-specific
password required with 2FA). Restart `klapp-invoice.service` after.

### A-20260809-sms-punchout-notifications — SMS punch-out reminders (5 PM auto + manual admin trigger)
- Workstream: time-tracking
- Status: open
- Created: 2026-08-09

Text workers still punched in at 5 PM (or on manual admin trigger) with the late punch-out
link, possibly including their PIN. From `todo.md` Medium + `notification.md`. Related work
in flight: `feature/sms-notifications` branch and the Twilio A2P 10DLC registration pages
(added 412c8859, reverted on main in 49617bb2 pending registration).

Progress 2026-08-22: split in two. **Manual admin trigger — done** on main via the dashboard
Notify link (`DashboardRow.NotifyLink`, 176986de→f2caf3d6, in [[A-20260812-admin-bulk-punch-notify]]):
an `sms:` deep link that opens the admin's own phone messaging app with the punch-out URL +
PIN prefilled in the worker's language — no Twilio, no registration needed. **5 PM automatic
send — still open**, blocked on Twilio A2P 10DLC registration. `feature/sms-notifications`
(Thomas, 2 commits, last touched 2026-06-29, unmerged) holds the building blocks: `internal/sms`
Twilio `Send()`, `cmd/sms` manual CLI, `setup_sms.md`, and the `/privacy` + `/terms` pages
registration requires. It predates the binary split, so reviving it means rebasing onto
`cmd/punch`/`cmd/admin` and wiring the send into the sweep goroutine in `cmd/admin`.

### A-20260809-admin-csrf — Add CSRF protection to admin POST endpoints
- Workstream: security
- Status: open
- Created: 2026-08-09

Medium from `docs/reference/security.md`: admin forms (worker create/edit, punch edit) have no CSRF
tokens. Low real-world risk (WireGuard/LAN-only) but cheap defense in depth (e.g. `gorilla/csrf`).

