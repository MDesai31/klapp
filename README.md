# Klaus Field Log

Internal field-operations tool for a landscaping company. Workers punch in/out
from their phones; admins review timesheets, manage invoices, and track hours
over LAN/WireGuard.

Go + SQLite. Three binaries over one database, each serving one site:

| Binary | Default port | Who reaches it |
|---|---|---|
| `cmd/punch` — worker punch site | `:4000` | public via Caddy (HTTPS) |
| `cmd/admin` — admin site | `:8082` | LAN/WireGuard only |
| `cmd/invoice` — invoice submission | `:8083` | LAN/WireGuard only |

The punch site is the only internet-facing surface, and it is deployed
**off by default** — `deploy/update.sh` needs `--with-punch` to turn it on.

---

## Requirements

- **Go 1.22+**
- **msmtp** — for emailing approved invoices. See [msmtp setup](#msmtp-setup) below.
- **Caddy** — reverse proxy for the public worker site (installed during deployment).

---

## Development

### Run locally

```bash
go run ./cmd/punch      # worker punch site :4000
go run ./cmd/admin      # admin site :8082
go run ./cmd/invoice    # invoice site :8083
```

Each is a standalone process; run only the ones you need.

SQLite migrations run automatically on startup. The database file (`db/klapp.db`)
is created on first run and is gitignored.

### First admin login

There is no signup page — bootstrap the first admin from the command line:

```bash
go run ./cmd/seedadmin -username admin -password <something>
```

Then open http://localhost:8082/admin/login.

### Scratch ports

The production services already occupy `:4000` / `:8082` / `:8083`. Use
different ports and a temp DB for ad-hoc testing so you don't collide:

```bash
go run ./cmd/punch -addr=:14000 -dsn="file:/tmp/x.db?_pragma=foreign_keys(1)"
go run ./cmd/admin -addr=:18082 -dsn="file:/tmp/x.db?_pragma=foreign_keys(1)"
```

Kill the child process by PID after — `kill %1` only stops `go run`, not the
binary it spawned.

### Tests

```bash
go test ./...
```

Each test gets its own throwaway SQLite file with fresh migrations applied, so
tests are fully isolated and can run in parallel.

---

## Production deployment

### First-time setup

1. **Clone the repo** on the server (e.g. to `~/klapp`).

2. **Run the setup script** from the project root. It creates `/opt/klapp`,
   builds all three binaries, rsyncs `ui/`, installs and enables the systemd
   units, and installs/configures Caddy:

   ```bash
   ./deploy/deploy.sh                # admin + invoice sites
   ./deploy/deploy.sh --with-punch   # ...and the public worker punch site
   ```

   The Caddyfile proxies `work.klauslandscaping.com` → `localhost:4000`. Caddy
   handles TLS automatically. With the punch site disabled, that hostname
   returns 502 by design.

3. **Seed the first admin** (one-time), straight against the live DB:

   ```bash
   go run ~/klapp/cmd/seedadmin -username admin -password <something> \
     -dsn="file:/opt/klapp/db/klapp.db?_pragma=foreign_keys(1)"
   ```

### Updating an existing deployment

```bash
cd ~/klapp
git pull
./deploy/update.sh                # rebuild + restart; punch site stays OFF
./deploy/update.sh --with-punch   # rebuild + restart; punch site ON
```

`update.sh` rebuilds all three binaries, rsyncs `ui/`, refreshes the systemd
units from the repo, and then enables or disables `klapp-punch` according to
the flag. **Whether the punch site runs is decided by this flag on every
deploy** — a plain `./deploy/update.sh` turns it off.

### Network layout

| Site | Unit | Bind | Reachable from |
|---|---|---|---|
| Worker punch site | `klapp-punch` | `127.0.0.1:4000` | Public via Caddy (HTTPS) — **off unless deployed with `--with-punch`** |
| Admin site | `klapp-admin` | `0.0.0.0:8082` | LAN / WireGuard only — **do not expose publicly** |
| Invoice site | `klapp-invoice` | `0.0.0.0:8083` | LAN / WireGuard only — **do not expose publicly** |

---

## msmtp setup

The invoice **Submit** button sends email via `msmtp`. Set it up once on the
server before the button will work.

1. Install: `sudo apt install msmtp`

2. Create `/etc/msmtprc` (or `~tklaus/.msmtprc`) with mode `600`:

   ```
   defaults
   auth           on
   tls            on
   tls_trust_file /etc/ssl/certs/ca-certificates.crt
   logfile        /var/log/msmtp.log

   account        default
   host           smtp.aol.com
   port           587
   from           mylawncut@aol.com
   user           mylawncut@aol.com
   password       <AOL app password>
   ```

   AOL requires an app-specific password if 2FA is enabled — generate one at
   https://login.aol.com/account/security.

3. Test: `echo -e "To: mylawncut@aol.com\nSubject: test\n\nhello" | msmtp -t`

4. Restart the service: `sudo systemctl restart klapp-invoice`
