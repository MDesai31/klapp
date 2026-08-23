# Ideas (unscoped / future)

Future intents — captured by `/project-plan`, not started. Scope before acting; promote to
`action-items.md` when an idea goes active.

### separate-time-reporting-site — Split Timesheet/Summary out of the admin dashboard
- Workstream: admin
- Priority: low
- Intended start: someday
- Why/context: divide invoicing and time-keeping concerns more clearly — own site or section. From `todo.md` Low.
- To start, future-us needs: decide site vs section; routes live in `cmd/admin/routes.go`.

### punch-map-visualization — Map tab of punch-out locations
- Workstream: admin
- Priority: low
- Intended start: someday
- Why/context: markers per punch-out; clicking one shows the worker's name and where they punched out. From `todo.md` Low.
- To start, future-us needs: a map library choice that works offline-ish on LAN; coords already stored on `time_punches`.

### daily-hours-email — Email the admin a daily summary of hours per worker
- Workstream: time-tracking
- Priority: low
- Intended start: someday
- Why/context: from `todo.md` Low; msmtp delivery path will already exist once A-20260809-msmtp-server-setup is done.
- To start, future-us needs: msmtp working; a daily trigger (the 9 PM sweep goroutine is a natural hook).

### geofence-punch-flag — Flag punches outside a radius of locations of interest
- Workstream: time-tracking
- Priority: low
- Intended start: someday
- Why/context: from `todo.md` Nice-to-have (verbatim, sentence trails off: "if a worker punched out, outside of Mary"). GPS is an audit signal, not access control — see `docs/reference/security.md`.
- To start, future-us needs: the locations-of-interest list and radius config; per-worker `require_location` flag already exists.

### optimize-claude-md — Slim CLAUDE.md so sessions need less context
- Workstream: ops
- Priority: someday
- Intended start: someday
- Why/context: from `todo.md` Nice-to-have. Partially addressed by the workspace-os memory layer (consult-when-relevant facts now live in `docs/memory/`).
- To start, future-us needs: apply the boundary test line-by-line to `CLAUDE.md`.

### photo-attachments — Before/after photos per invoice
- Workstream: invoicing
- Priority: someday
- Intended start: someday
- Why/context: carried over from the Next.js-era `docs/ideas.md`; still relevant to the Go app's invoice form.
- To start, future-us needs: local-disk storage decision (S3 was the old assumption; single-server SQLite app suggests local files).

### audit-log — Track who created/edited/deleted what and when
- Workstream: admin
- Priority: someday
- Intended start: someday
- Why/context: carried over from Next.js-era `docs/ideas.md`; `modified_by_admin` flag is the only audit signal today.
- To start, future-us needs: decide scope (punches only vs all admin CRUD).

### csv-pdf-export — Export hours/invoices for payroll or records
- Workstream: admin
- Priority: someday
- Intended start: someday
- Why/context: carried over from Next.js-era `docs/ideas.md`; the original design (`docs/design/time_reporting_plan.md`) imagined per-worker PDF pay-period reports emailed and printed via `lp`.
- To start, future-us needs: pick format(s); Summary tab already computes the pay-period grid.

### offline-pwa — Punch without cell signal, sync later
- Workstream: time-tracking
- Priority: someday
- Intended start: someday
- Why/context: carried over from Next.js-era `docs/ideas.md`; workers are cellular-only so dead zones are plausible. The original plan's fallback is "text the admin".
- To start, future-us needs: decide if the complexity is warranted for a rare case; conflicts with the stateless-PIN design.
