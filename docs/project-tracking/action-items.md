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

### A-20260809-pin-echoed-in-html — Stop round-tripping the worker PIN through page HTML
- Workstream: security
- Status: open
- Created: 2026-08-09

High-severity from `docs/reference/security.md`: after auth, the PIN is written into a hidden form field
(`punch.tmpl`) and resubmitted on punch — visible in page source, history, proxy logs. Fix:
hold `workerID` server-side (session) instead; the punch site currently has no session
middleware at all.

### A-20260809-admin-csrf — Add CSRF protection to admin POST endpoints
- Workstream: security
- Status: open
- Created: 2026-08-09

Medium from `docs/reference/security.md`: admin forms (worker create/edit, punch edit) have no CSRF
tokens. Low real-world risk (WireGuard/LAN-only) but cheap defense in depth (e.g. `gorilla/csrf`).

### A-20260809-pin-field-plaintext — Use a password input for the PIN field in the admin panel
- Workstream: security
- Status: open
- Created: 2026-08-09

Medium from `docs/reference/security.md`: `admin_workers.tmpl` PIN input is `type="text"`, visible
on-screen while an admin enters/edits it.
