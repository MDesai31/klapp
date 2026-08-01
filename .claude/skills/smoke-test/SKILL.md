---
name: smoke-test
description: Run klapp locally against a throwaway SQLite database on scratch ports and drive the admin or worker punch site with curl. Use to verify a change in the real running app, reproduce a bug end to end, or inspect rendered pages — anything beyond what `go test` covers. Handles seeding an admin and workers, session-authenticated requests, and cleanup.
---

# Smoke-testing klapp locally

`klapp.service` runs the real app on `:4000` / `:8082` against
`/opt/klapp/db/klapp.db`. **Never reuse those ports or that database.** The
scripts here use `:14000` / `:18082` and a throwaway DB in `/tmp/klapp-scratch`.

## Start

```bash
.claude/skills/smoke-test/scripts/scratch-up.sh
```

Builds `cmd/web`, creates a fresh DB with migrations, seeds admin `admin` /
`scratchpw123` and three active workers (Ana PIN 1111 / Spanish, Bob 2222 /
English, Cid 3333 / English, all punched out), starts the server, waits for it
to answer, and prints the URLs and PID. Add `--keep-db` to restart the server
without wiping data.

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

No session needed — the worker's PIN is posted with each request:

```bash
curl -sS -X POST -d "pin=1111" http://localhost:14000/punch | grep -E 'flash|button'
curl -sS -X POST -d "pin=1111" http://localhost:14000/punch/in
```

Each PIN check sleeps `pin_check_delay_ms` (250ms default), so worker-site
requests are deliberately slow. Ten bad PINs from one IP triggers a lockout.

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

- **Working directory** — the binary globs `./ui/html/pages/*.tmpl` and serves
  `./ui/static/` relative to its CWD, so it must run from the repo root.
- **`GET /admin` 301s to `/admin/`** — without `-L` you get an empty body and
  may misread it as a lost session.
- **Mutations answer 303** — `-L` follows through to the rendered page so you
  can see the flash message.
- **`go run`'s child outlives `kill %1`** — the scripts build a real binary and
  track its PID instead.
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
