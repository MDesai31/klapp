---
name: deploy-ops
description: How klapp runs in production on this box — the systemd units, the /opt/klapp layout, pushing code with deploy/update.sh, reading logs, and the live database. Use when deploying, restarting, checking service status or logs, or reasoning about anything that touches /opt/klapp or the live data.
---

# Running klapp in production

Production is this same machine. Dev work happens in `~/klapp`; the live app
runs from `/opt/klapp`. **Confirm with the user before restarting or deploying**
— real workers punch in against this.

## What's running

| Unit | Binary | Listens | Exposure |
|---|---|---|---|
| `klapp-admin` | `/opt/klapp/admin` | `:8082` | all interfaces, LAN/WireGuard only |
| `klapp-punch` | `/opt/klapp/punch` | `127.0.0.1:4000` | public via Caddy TLS — **disabled by default** |
| `klapp-invoice` | `/opt/klapp/invoice` | `:8083` | invoice site, LAN/WireGuard only |
| `caddy` | — | `:443` | `work.klauslandscaping.com` → `localhost:4000` |

The admin port is deliberately bound to all interfaces and deliberately not
behind Caddy — it is reachable only over LAN/WireGuard, and that trust boundary
is what justifies the admin site having no CSRF protection.

`/opt/klapp` holds `admin`, `punch`, `invoice`, `printsched`, `ui/` (rsynced
from the repo) and `db/klapp.db` — the live database. Never point a dev or test
command at it.

`printsched` has no unit of its own: it is a short-lived command the admin
site execs when someone presses Print on the summary tab (`print_binary` in
`config.json`, `./printsched` by default, resolved against the unit's
`WorkingDirectory=/opt/klapp`). It can also be run by hand:

```bash
cd /opt/klapp && ./printsched -period 2026-06-08
```

The pre-split `klapp` unit and its `/opt/klapp/web` binary are gone; the deploy
scripts remove them on the first run after the split.

## Push a code change

From the repo root, after committing:

```bash
./deploy/update.sh                # admin + invoice up, punch site OFF
./deploy/update.sh --with-punch   # ...and punch site ON
```

Builds the four binaries into `/opt/klapp`, rsyncs `ui/`, refreshes the systemd
units from the repo, restarts `klapp-admin` and `klapp-invoice`, and then
`enable --now` / `disable --now`s `klapp-punch` per the flag. Needs sudo.

**The flag is re-applied on every deploy.** Running plain `./deploy/update.sh`
while the punch site is up will take it down — workers get a Caddy 502. Ask the
user which they want before deploying.

`deploy/deploy.sh` is **first-time setup only** — it also creates `/opt/klapp`
and installs/configures Caddy. Don't re-run it for a routine push. Both scripts
share `deploy/lib.sh`.

## Things that surprise people

- **Templates are read from disk at startup.** A template-only change still
  needs `update.sh` (the rsync *and* the restart) to take effect.
- **Migrations run automatically at startup**, forward-only (`goose.Up`). Every
  migration file has a `-- +goose Down` section but nothing ever runs it, so
  rolling code back does **not** roll the schema back. A rollback across a
  migration needs a plan for the data.
- **`config.json` lives in `/opt/klapp`, not the repo** (it's gitignored). Both
  the admin and punch binaries read it. If it's missing they run on built-in
  defaults — see `internal/config/config.go`.
- **The 9 PM sweep runs in the admin binary**, not the punch one, precisely
  because the punch site can be switched off. It auto-closes open punches and
  flags them non-compliant; a restart between 9 PM and midnight re-runs a
  catch-up sweep on startup, by design (`runNightlyPunchOut` in
  `cmd/admin/sweep.go`).

## The schedule listener is on a different machine

Pressing Print does not finish on this box. `printsched` posts the pay period
to `schedule-listener` on the home server (`10.9.0.7:5555`, over WireGuard),
which runs `build_schedule.py` and writes the PDFs there. **`update.sh` does
not touch it** — deploying that half is a separate, manual, cross-host job
documented in `schedule_listener/README.md`. Its logs are on that machine:

```bash
curl http://10.9.0.7:5555/healthz              # -> ok
ssh 10.9.0.7 'journalctl -u schedule-listener -f'
```

If Print fails, the admin site shows the reason as a flash and logs the whole
`printsched` output at ERROR (`grep print` in `journalctl -u klapp-admin`).
Connection refused there almost always means the listener is down or WireGuard
is, not that anything in klapp broke.

## Status and logs

```bash
systemctl status klapp-admin
journalctl -u klapp-admin -f        # also: -u klapp-punch, -u klapp-invoice, -u caddy
journalctl -u klapp-admin --since "1 hour ago"
```

The app logs structured slog lines to stdout, so journald has everything.

## Admin logins

There's no signup page. Create or reset one against the live DB:

```bash
go run ./cmd/seedadmin -username <name> -password <pw> \
  -dsn "file:/opt/klapp/db/klapp.db?_pragma=foreign_keys(1)"
```

## Related skills

- `smoke-test` — verify on scratch ports before shipping
- `codebase-map` — find what you're deploying
