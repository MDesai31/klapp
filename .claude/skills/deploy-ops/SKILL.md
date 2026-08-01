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
| `klapp` | `/opt/klapp/web` | `127.0.0.1:4000` worker site, `:8082` admin site | worker site via Caddy TLS; admin on all interfaces for LAN/WireGuard only |
| `klapp-invoice` | `/opt/klapp/invoice` | `:8083` | invoice site |
| `caddy` | — | `:443` | `work.klauslandscaping.com` → `localhost:4000` |

The admin port is deliberately bound to all interfaces and deliberately not
behind Caddy — it is reachable only over LAN/WireGuard, and that trust boundary
is what justifies the admin site having no CSRF protection.

`/opt/klapp` holds `web`, `invoice`, `ui/` (rsynced from the repo) and
`db/klapp.db` — the live database. Never point a dev or test command at it.

## Push a code change

From the repo root, after committing:

```bash
./deploy/update.sh
```

Builds both binaries into `/opt/klapp`, rsyncs `ui/`, restarts `klapp` and
`klapp-invoice`. Needs sudo for the restarts. That's the whole deploy.

`deploy/deploy.sh` is **first-time setup only** — it installs the systemd units,
enables them, and installs/configures Caddy. Don't re-run it for a routine push.

## Things that surprise people

- **Templates are read from disk at startup.** A template-only change still
  needs `update.sh` (the rsync *and* the restart) to take effect.
- **Migrations run automatically at startup**, forward-only (`goose.Up`). Every
  migration file has a `-- +goose Down` section but nothing ever runs it, so
  rolling code back does **not** roll the schema back. A rollback across a
  migration needs a plan for the data.
- **`config.json` lives in `/opt/klapp`, not the repo** (it's gitignored). If
  it's missing the app runs on built-in defaults — see `cmd/web/config.go`.
- **The 9 PM sweep** auto-closes open punches and flags them non-compliant. A
  restart between 9 PM and midnight re-runs a catch-up sweep on startup, by
  design (`runNightlyPunchOut` in `cmd/web/main.go`).

## Status and logs

```bash
systemctl status klapp
journalctl -u klapp -f              # also: -u klapp-invoice, -u caddy
journalctl -u klapp --since "1 hour ago"
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
