# Resolved (the record)

Completed action records, archived here by `/project-log done` (with completion date + commit
ref). Append-only.

### A-20260617-go-rewrite-core — Replace the Next.js/Prisma app with the Go + SQLite rewrite
- Workstream: ops
- Status: done
- Created: 2026-06-16
- Completed: 2026-06-17
- Commit: 191c358a (plus plan docs cc0b1499, 9b13880e, 128020cd)

Phase-1 rewrite per `docs/design/refactor.md` and `docs/design/time_reporting_plan.md`: cmd/web with worker
punch site + admin site, internal/models, goose migrations, html/template UI.
Imported from git history by /tracking-adopt git.

### A-20260619-deploy-pipeline — systemd /opt/klapp deploy, Caddy on the production domain
- Workstream: ops
- Status: done
- Created: 2026-06-19
- Completed: 2026-06-19
- Commit: 442b0fae, 62c01f4a

Generalized systemd deploy to /opt/klapp, pointed the Caddyfile at work.klauslandscaping.com,
added Caddy bootstrap and the binary-only update script. 442b0fae also switched worker PINs to
plaintext storage (noted in `docs/reference/security.md` as a Low open issue).
Imported from git history by /tracking-adopt git.

### A-20260624-admin-worker-management — Worker editing, unique PINs, deactivated-worker hiding
- Workstream: admin
- Status: done
- Created: 2026-06-19
- Completed: 2026-06-24
- Commit: 2ca4188b, f4dfc9b5, f87ebcd7

Admin Workers page: edit flow, duplicate-PIN rejection on create/edit, deactivated workers
hidden by default with a show toggle.
Imported from git history by /tracking-adopt git.

### A-20260624-dashboard-summary — Dashboard elapsed-time labels and the pay-period Summary tab
- Workstream: time-tracking
- Status: done
- Created: 2026-06-24
- Completed: 2026-06-24
- Commit: 849dab18, cb679692

Dashboard shows elapsed/worked time per worker; new Summary tab with per-day hours grid and
salary estimate per pay period.
Imported from git history by /tracking-adopt git.

### A-20260625-worker-rate-language — Hourly rate, language field, and Spanish worker UI
- Workstream: time-tracking
- Status: done
- Created: 2026-06-25
- Completed: 2026-06-25
- Commit: cd660ad9

Migrations 0003/0004; per-worker hourly rate feeds the Summary salary column, per-worker
language renders the punch site in Spanish or English (see [[bilingual-ui]]).
Imported from git history by /tracking-adopt git.

### A-20260626-invoice-system — Invoice system: cmd/invoice binary, admin tabs, schema, msmtp, tests
- Workstream: invoicing
- Status: done
- Created: 2026-06-25
- Completed: 2026-06-26
- Commit: 3bc54642, 146ebc94, 0e7e66f9, 69a5ce90

Per `docs/reference/invoices.md`: new cmd/invoice binary (:8083), migrations 0005–0008 (customers,
job_descriptions, materials, invoices + junctions), admin Invoices/Customers/Catalog tabs,
msmtp email delivery, model tests. Deploy scripts and service units updated.
Imported from git history by /tracking-adopt git.

### A-20260629-timesheet-compliance — Day/pay-period edit fix, add/delete entries, 9 PM auto-punch-out
- Workstream: time-tracking
- Status: done
- Created: 2026-06-29
- Completed: 2026-06-29
- Commit: 3da0b3bf, ee7bbbda, 8d4c5465

Covers all three struck-through High items in `todo.md` (cross-matched): editing a punch
recalculates day/pay_period (the Friday/Saturday bug), per-row delete + "+ Add entry", and the
non-compliant flow (migration 0009, nightly 9 PM sweep, red rows pinned to top,
newest-first sort). Also replaces overlapping punches on admin edit.
Imported from git history by /tracking-adopt git.

### A-20260629-docs-guides — Full deployment README and the admin guide
- Workstream: ops
- Status: done
- Created: 2026-06-29
- Completed: 2026-06-29
- Commit: b785f88d, 3b7df0fe

README deployment guide (symlinked systemd units, Caddy, msmtp) and `docs/guides/admin-guide.md` covering
every tab and common tasks.
Imported from git history by /tracking-adopt git.

### A-20260701-timesheet-ux — require_location escape hatch, pay-period dropdown, edited-punch styling
- Workstream: admin
- Status: done
- Created: 2026-06-29
- Completed: 2026-07-01
- Commit: 7dbe6b59, 755298d0, 0740a2c7

Per-worker require_location flag as an admin escape hatch, pay-period links replaced with a
dropdown on Timesheet/Summary, edited non-compliant punches shown light blue at normal sort
position.
Imported from git history by /tracking-adopt git.

### A-20260702-pin-hardening — IP-based PIN lockout, per-attempt delay, daily punch-in cap
- Workstream: security
- Status: done
- Created: 2026-07-02
- Completed: 2026-07-02
- Commit: e88fc117 (plus 05187763 X-Forwarded-For loopback trust, 803e314e handler/query hardening)

Implements [[D-20260809-pin-lockout-design]]: `cmd/web/pinlimiter.go` IP lockout, flat
per-attempt delay, per-worker daily punch-in cap, all tunable via config.json.
Imported from git history by /tracking-adopt git.

### A-20260702-punch-correctness — Punch-edit integrity, DST bucketing, sweep catch-up, robustness
- Workstream: time-tracking
- Status: done
- Created: 2026-06-30
- Completed: 2026-07-02
- Commit: 3e46266f, 52d8ce68, aebcf060, 079faac3, 7f132cb4, 0c9ba2d4, b873b41d, e0a101ca

Correctness batch: admin punch edits no longer delete later punches (one open punch enforced,
migration 0012), flash on invalid admin times, late punch-out link works after the 9 PM sweep,
DST-safe pay-period bucketing, startup catch-up sweep, nil-panic fix on punch.tmpl, HTTP server
timeouts + SQLite busy_timeout, geolocation timeout so the punch button can't hang.
Imported from git history by /tracking-adopt git.

### A-20260702-invoice-email-semantics — Settle invoice review-vs-email-failure behavior
- Workstream: invoicing
- Status: done
- Created: 2026-07-02
- Completed: 2026-07-02
- Commit: e4c8ebe6 (supersedes f9188594)

First made email failure block the reviewed flag, then reversed: mark reviewed even when the
email fails, but tell the admin so (see [[invoice-review-flow]]).
Imported from git history by /tracking-adopt git.
