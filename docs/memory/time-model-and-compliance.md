---
name: time-model-and-compliance
description: pay-period/day bucketing rules and the late vs non-compliant punch distinction
type: domain
---

Times are ISO-8601 `TEXT` in SQLite: UTC instants for start/end, **local-date** buckets for
`day` and `pay_period` (14-day blocks) — see `internal/models/time_punches.go`; bucketing is
DST-aware. When an admin edits a punch's start time, `day`/`pay_period` are **recalculated**,
so moving an entry to another day refiles it automatically.

Two distinct "didn't punch out properly" states on the Timesheet:
- **late** (yellow): worker used the 9 PM SMS late punch-out link themselves.
- **non-compliant** (red, sorted to top): worker was still clocked in at 9 PM and a nightly
  sweep auto-punched them out (migration 0009). A catch-up sweep also runs on startup so an
  evening restart can't skip the 9 PM pass. An admin-edited non-compliant punch turns light
  blue and returns to normal sort position.

Missed punch-*ins* are deliberately admin-corrected, not worker-self-entered — see
[[D-20260809-missed-punch-admin-corrects]].
