---
name: worker-auth-model
description: PIN auth has no worker picker — and why cookie counters and IP buddy-punch detection were rejected
type: domain
---

Worker auth is a bare PIN: `WorkerModel.Authenticate` checks the entered PIN against **every
active worker** (no name/ID picker), so a failed attempt can't be attributed to a specific
worker. Brute-force defense is therefore connection-shaped, not account-shaped
(`cmd/punch/pinlimiter.go`): an IP-keyed lockout (default 10 failures/15 min, tunable in
`config.json`), a flat per-attempt delay (default 250ms) applied regardless of outcome, and a
per-worker daily punch-in cap (default 3).

Two alternatives were considered and rejected — workers are **cellular-only** (no job-site
wifi), so carrier-grade NAT shapes both calls:
- A cookie/device failure counter: a scripted client just never resends the cookie — no real
  protection. See [[D-20260809-pin-lockout-design]].
- IP-based buddy-punch detection: NAT makes many unrelated phones share an IP and one phone's
  IP can change mid-session. GPS lat/lon captured on punch is the anti-fraud signal instead.

The lockout threshold is deliberately high because NAT also means one locked IP can collateral-
lock unrelated workers. Full review: `docs/reference/security.md`.
