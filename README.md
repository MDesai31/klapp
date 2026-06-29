# Klaus Field Log

Internal field-operations tool for a landscaping company. Workers punch in/out
from their phones; admins review timesheets, manage invoices, and track hours
over LAN/WireGuard.

Go + SQLite. Two binaries:

| Binary | Default port | Who reaches it |
|---|---|---|
| `cmd/web` — worker punch + admin site | `:4000` (worker), `:8082` (admin) | worker site is public via Caddy; admin is LAN/WireGuard only |
| `cmd/invoice` — invoice submission | `:8083` | LAN/WireGuard only |

---

## Requirements

- **Go 1.22+**
- **msmtp** — for emailing approved invoices. See [msmtp setup](#msmtp-setup) below.
- **Caddy** — reverse proxy for the public worker site (installed during deployment).

---

## Development

### Run locally

```bash
go run ./cmd/web        # worker site :4000  +  admin site :8082
go run ./cmd/invoice    # invoice site :8083
```

SQLite migrations run automatically on startup. The database file (`db/klapp.db`)
is created on first run and is gitignored.

### First admin login

There is no signup page — bootstrap the first admin from the command line:

```bash
go run ./cmd/seedadmin -username admin -password <something>
```

Then open http://localhost:8082/admin/login.

### Scratch ports

The production service already occupies `:4000` / `:8082` / `:8083`. Use
different ports and a temp DB for ad-hoc testing so you don't collide:

```bash
go run ./cmd/web \
  -addr=:14000 \
  -admin-addr=:18082 \
  -dsn="file:/tmp/x.db?_pragma=foreign_keys(1)"
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

2. **Create the app directory** and build the binaries:

   ```bash
   sudo mkdir -p /opt/klapp/db
   sudo chown -R "$(whoami):$(whoami)" /opt/klapp
   go build -o /opt/klapp/web ./cmd/web
   go build -o /opt/klapp/invoice ./cmd/invoice
   rsync -a --delete ui/ /opt/klapp/ui/
   ```

3. **Symlink the systemd units** — linking instead of copying means edits to
   the repo files take effect after a `daemon-reload`, with no re-copy step:

   ```bash
   sudo ln -s /opt/klapp/deploy/klapp.service         /etc/systemd/system/klapp.service
   sudo ln -s /opt/klapp/deploy/klapp-invoice.service /etc/systemd/system/klapp-invoice.service
   sudo systemctl daemon-reload
   sudo systemctl enable --now klapp
   sudo systemctl enable --now klapp-invoice
   ```

4. **Install Caddy** (if not already present):

   ```bash
   sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
   curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
     | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
   curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
     | sudo tee /etc/apt/sources.list.d/caddy-stable.list
   sudo apt update && sudo apt install -y caddy
   ```

5. **Symlink the Caddyfile**:

   ```bash
   sudo ln -sf /opt/klapp/deploy/Caddyfile /etc/caddy/Caddyfile
   sudo systemctl enable --now caddy
   sudo systemctl reload caddy
   ```

   The Caddyfile proxies `work.klauslandscaping.com` → `localhost:4000`. Caddy
   handles TLS automatically.

6. **Seed the first admin** (one-time). Stop the service, run seedadmin against
   the live DB, then restart:

   ```bash
   sudo systemctl stop klapp
   /opt/klapp/web -dsn="file:/opt/klapp/db/klapp.db?_pragma=foreign_keys(1)" &
   go run ~/klapp/cmd/seedadmin -username admin -password <something>
   kill %1
   sudo systemctl start klapp
   ```

### Updating an existing deployment

```bash
cd ~/klapp
git pull
go build -o /opt/klapp/web ./cmd/web
go build -o /opt/klapp/invoice ./cmd/invoice
rsync -a --delete ui/ /opt/klapp/ui/
sudo systemctl restart klapp
sudo systemctl restart klapp-invoice
```

Or use the convenience script (run from the project root):

```bash
bash deploy/update.sh
```

### Network layout

| Site | Bind | Reachable from |
|---|---|---|
| Worker punch site | `127.0.0.1:4000` | Public via Caddy (HTTPS) |
| Admin site | `0.0.0.0:8082` | LAN / WireGuard only — **do not expose publicly** |
| Invoice site | `0.0.0.0:8083` | LAN / WireGuard only — **do not expose publicly** |

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
