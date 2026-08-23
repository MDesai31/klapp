---
name: smoke-test
description: Run klapp locally against a throwaway SQLite database on scratch ports and drive the admin or worker punch site with curl. Use to verify a change in the real running app, reproduce a bug end to end, inspect rendered pages, or exercise the print path without reportlab — anything beyond what `go test` covers. Handles seeding an admin and workers, session-authenticated requests, and cleanup.
---

# Smoke-testing klapp locally

`klapp-admin.service` and `klapp-punch.service` run the real apps on `:8082` /
`:4000` against `/opt/klapp/db/klapp.db`. **Never reuse those ports or that
database.** The scripts here use `:18082` / `:14000` and a throwaway DB in
`/tmp/klapp-scratch`.

## Start

```bash
.claude/skills/smoke-test/scripts/scratch-up.sh
```

Builds `cmd/admin` and `cmd/punch`, creates a fresh DB with migrations, seeds
admin `admin` / `scratchpw123` and three active workers (Ana PIN 1111 / Spanish,
Bob 2222 / English, Cid 3333 / English, all punched out), starts **both**
servers against the same DB, waits for them to answer, and prints the URLs and
PIDs. Add `--keep-db` to restart them without wiping data.

They are separate processes with separate logs (`/tmp/klapp-scratch/admin.log`,
`punch.log`). If only one site matters for what you're testing, you can still
start just it by hand — but the script's shared DB is what makes a punch-then-
check-the-dashboard test work.

## Drive the admin site

```bash
S=.claude/skills/smoke-test/scripts

$S/admin-curl.sh /admin                                        # dashboard
$S/admin-curl.sh /admin/punch/bulk -d "action=in&worker_id=1"   # a mutation
```

`admin-curl.sh` logs in on first use, reuses the cookie, and follows redirects —
so a mutation prints the page you land on, flash message included. Grep rather
than dumping whole pages:

```bash
$S/admin-curl.sh /admin/punch/bulk -d "action=in&worker_id=1" | grep '<p class="flash"'
#   <p class="flash">Punched in Ana.</p>
```

Pass a body with `-d` alone. **Don't write `-X POST -d`** — `-X` pins the method
for every hop, so curl re-POSTs the 303 target and the admin mux answers 405.
`-d` already implies POST. (The script strips a redundant `-X POST` for you, but
plain `curl` won't.)

To check a status code instead of a body:

```bash
$S/admin-curl.sh /admin/punch/bulk -d "action=bogus" -o /dev/null -w "%{http_code}\n"
```

## Drive the worker punch site

The PIN is posted once, to `/punch`; that sets a short-lived `punch_session`
cookie holding the worker ID, and `/punch/in` and `/punch/out` read the worker
from that cookie (the PIN is never resent — see `docs/reference/security.md`).
So carry a cookie jar across the calls:

```bash
J=$(mktemp)
curl -sS -c "$J" -d "pin=1111" http://localhost:14000/punch | grep -E 'flash|button'
curl -sS -b "$J" -d "" http://localhost:14000/punch/in
```

Hitting `/punch/in` without the cookie just re-renders the PIN form.

Each PIN check sleeps `pin_check_delay_ms` (250ms default), so worker-site
requests are deliberately slow. Ten bad PINs from one IP triggers a lockout.

## Test the print path

The summary tab's Print button execs `printsched`, which POSTs to the schedule
listener on the home server, which runs `build_schedule.py`. All three links can
be exercised here.

**This box is the home server** (`10.9.0.7` on `wg1`), so the real thing is
available: reportlab lives in `~/venv` and `lp` talks to the printer. Point the
listener's `python` at `/home/tklaus/venv/bin/python3` and you get actual PDFs,
readable with `pdftotext -layout` or `pdftoppm -png`. Keep `print_command` empty
while testing, or hold the jobs (`["lp", "-d", "HP_OfficeJet_9120e", "-H",
"hold"]`) and `cancel` them afterwards, so a smoke test never spends paper.

The `reportlab-stub` below is the fallback for a box without reportlab — it
dumps the table `build_schedule.py` would have drawn instead of rendering it.

```bash
STUB=.claude/skills/smoke-test/scripts/reportlab-stub
SCRATCH=/tmp/klapp-scratch

# 1. the listener, with the stub reportlab on its PYTHONPATH
go build -o $SCRATCH/schedule-listener ./schedule_listener
cat > $SCRATCH/listener.json <<EOF
{"addr": "127.0.0.1:15555", "python": "python3",
 "script": "$PWD/schedule_listener/build_schedule.py",
 "output_dir": "$SCRATCH/out", "print_command": []}
EOF
( cd $SCRATCH && PYTHONPATH=$PWD/$STUB ./schedule-listener -config listener.json > listener.log 2>&1 & )

# 2. point the scratch admin site at it (scratch-up.sh does not do this for you)
cat > $SCRATCH/config.json <<'EOF'
{"print_host": "127.0.0.1", "print_port": 15555,
 "print_binary": "/tmp/klapp-scratch/printsched"}
EOF
go build -o $SCRATCH/printsched ./cmd/printsched
#   ...then restart the scratch admin with -config /tmp/klapp-scratch/config.json

# 3. press the button
.claude/skills/smoke-test/scripts/admin-curl.sh /admin/summary/print \
  -d "period=2026-06-08" | grep '<p class="flash"'

rm -f $SCRATCH/schedule-rows.txt   # before each run; the stub appends
cat $SCRATCH/schedule-rows.txt     # the rows each sheet would print
```

`printsched` alone is enough to test the wire without a browser or a session:

```bash
$SCRATCH/printsched -period 2026-06-08 -host 127.0.0.1 -port 15555 \
  -dsn "file:$SCRATCH/klapp.db?_pragma=busy_timeout(5000)" -config /nonexistent.json
```

Notes:

- **`-config /nonexistent.json` is deliberate** — it forces built-in defaults
  rather than picking up a real `config.json` pointing at `10.9.0.7`. Getting
  this wrong sends a live print job to the home server.
- **Seed multi-punch and open-punch days by hand** (`sqlite3` INSERTs into
  `time_punches`), since that is what exercises the spill rows under `TOTAL`.
  `scratch-up.sh`'s workers are all punched out with no history.
- **The stub reports `rows=`, `heights=` and `min_row_h=`** per sheet. The two
  counts must match, and `min_row_h` shrinking is how you confirm a sheet with
  many spill rows still fits its half page.
- **A blank sheet is 21 rows.** Anything else means the spill rows grew.

## Inspect the database directly

```bash
sqlite3 -header /tmp/klapp-scratch/klapp.db "SELECT * FROM time_punches;"
sqlite3 /tmp/klapp-scratch/klapp.db .schema
```

Times are stored UTC RFC3339; `day` and `pay_period` are local dates.

## Stop when done

```bash
.claude/skills/smoke-test/scripts/scratch-down.sh
```

## Gotchas these scripts already handle

- **Working directory** — each binary globs its own pages out of
  `./ui/html/pages/` (`admin_*.tmpl` for admin, `punch*.tmpl` for punch) and
  serves `./ui/static/` relative to its CWD, so both must run from the repo root.
- **`GET /admin` 301s to `/admin/`** — without `-L` you get an empty body and
  may misread it as a lost session.
- **Mutations answer 303** — `-L` follows through to the rendered page so you
  can see the flash message.
- **`go run`'s child outlives `kill %1`** — the scripts build real binaries and
  track their PIDs instead (`admin.pid`, `punch.pid`).
- **`pkill -f 18082` kills your own shell** — the pattern matches any command
  line mentioning the port, including the one running `pkill`. `scratch-down.sh`
  uses the pidfile, falling back to whatever `ss` shows actually listening.
- **`daily_punch_limit` defaults to 3** even with no `config.json`, so the
  fourth punch-in for a worker on the same day is refused by design.
- **`punch_site_url` defaults to the production URL**, so `Notify` links on a
  scratch dashboard point at the real site unless you pass `-config`.

## Related skills

- `feature-slice` — conventions for the change you're testing
- `codebase-map` — finding the handler or route involved
