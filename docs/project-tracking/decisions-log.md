# Decisions Log

Append-only record of *why* we chose X. Added by `/project-log decision` with a `D-YYYYMMDD-slug`
ID. Never rewrite history — only add. Records use the decision template from the conventions doc.

### D-20260809-go-rewrite — Rewrite klapp in Go + SQLite, replacing the Next.js stack
- Workstream: ops
- Created: 2026-08-09
- Status: accepted
- Rationale: Next.js 16 + Auth.js v5 beta + Prisma 6 churn (breaking majors, hundreds of npm deps, framework complexity) outweighed the needs of a small internal CRUD tool. Go's 1.x compatibility guarantee, SQLite as a single file, html/template, single-binary deploy — stable for years without forced upgrades. Full rationale in `docs/refactor.md`.
- Consequences: no ORM (raw SQL), no worker sessions/cookies, no CSRF on worker path, no separate DB server; tests run parallel on throwaway SQLite files.
- Spawns: none

Adopted from `docs/refactor.md` § Why (decision originally made ~2026-06-16 by Thomas; recorded at adoption time).

### D-20260809-promote-go-rewrite-to-main — Fast-forward main to refactor/go-rewrite
- Workstream: ops
- Created: 2026-08-09
- Status: accepted
- Rationale: the Go rewrite is the real app going forward; the branch was based exactly on main's tip, so main fast-forwarded cleanly (no merge commit, linear history preserved). The Next.js app is preserved on the `legacy/nextjs` branch.
- Consequences: main = Go app as of 49617bb2; old workspace-os dogfood artifacts from the Next.js era were archived locally and re-adopted fresh (see [[nextjs-era-history]]).
- Spawns: none

### D-20260809-pin-lockout-design — IP-keyed lockout + flat delay + daily cap for PIN brute-force defense
- Workstream: security
- Created: 2026-08-09
- Status: accepted
- Rationale: a cookie/device failure counter was rejected — a scripted client simply never resends the cookie, so it only limits humans who are already self-limiting. Workers are cellular-only (carrier-grade NAT), so the IP lockout threshold is set high (10/15min) to avoid collateral-locking a shared connection, backed by a flat 250ms per-attempt delay that no client behavior can bypass, plus a per-worker daily punch-in cap (3) so a guessed PIN can't spam punch records.
- Consequences: defense is connection-shaped, not account-shaped (PINs can't be attributed to a worker on failure — see [[worker-auth-model]]).
- Spawns: none

Adopted from `docs/security.md` (implemented in e88fc117).

### D-20260809-missed-punch-admin-corrects — Missed punch-ins are admin-corrected, not worker-self-entered
- Workstream: time-tracking
- Created: 2026-08-09
- Status: accepted
- Rationale: letting workers back-enter their own start time is the path of least resistance and would devolve into nobody punching in live. Having to ask the admin (and be lightly scolded) disincentivizes forgetting. Option 1 vs Option 2 analysis in `docs/time_reporting_plan.md` § Worker Missed Punch In.
- Consequences: admin Timesheet has "+ Add entry" and edit flows as the correction mechanism.
- Spawns: none
