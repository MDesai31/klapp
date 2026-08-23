#!/usr/bin/env bash
# Start both klapp sites on scratch ports against a throwaway SQLite DB, seeded
# with an admin login and three workers. Never touches the live ports
# (:4000/:8082) or the live DB (/opt/klapp/db/klapp.db) that klapp-admin.service
# and klapp-punch.service use.
#
#   scratch-up.sh              start (rebuilds, reseeds from empty)
#   scratch-up.sh --keep-db    restart the servers, keep existing scratch data
#
# Override with env vars: KLAPP_ADMIN_PORT, KLAPP_WORKER_PORT, KLAPP_SCRATCH_DIR
set -euo pipefail

ADMIN_PORT="${KLAPP_ADMIN_PORT:-18082}"
WORKER_PORT="${KLAPP_WORKER_PORT:-14000}"
SCRATCH="${KLAPP_SCRATCH_DIR:-/tmp/klapp-scratch}"
ADMIN_USER=admin
ADMIN_PASS=scratchpw123

DB="$SCRATCH/klapp.db"
SEED_LOG="$SCRATCH/seed.log"

# The binaries read ./ui/html and ./ui/static relative to their working
# directory, so they must run from the repo root.
cd "$(git rev-parse --show-toplevel)"

mkdir -p "$SCRATCH"
"$(dirname "$0")/scratch-down.sh" >/dev/null 2>&1 || true

if [ "${1:-}" != "--keep-db" ]; then
	rm -f "$DB" "$DB-shm" "$DB-wal"
fi

echo "building..."
go build -o "$SCRATCH/admin" ./cmd/admin
go build -o "$SCRATCH/punch" ./cmd/punch

if [ ! -f "$DB" ]; then
	# seedadmin runs the migrations, so this also creates the schema.
	go run ./cmd/seedadmin -username "$ADMIN_USER" -password "$ADMIN_PASS" \
		-dsn "file:$DB?_pragma=foreign_keys(1)" >"$SEED_LOG" 2>&1
	sqlite3 "$DB" "INSERT INTO workers (worker_name, pin, phone, hourly_rate, language, active) VALUES
		('Ana', '1111', '', 20, 'spanish', 1),
		('Bob', '2222', '', 22, 'english', 1),
		('Cid', '3333', '', 18, 'english', 1);"
fi

"$SCRATCH/admin" -addr=":$ADMIN_PORT" \
	-dsn "file:$DB?_pragma=foreign_keys(1)" >"$SCRATCH/admin.log" 2>&1 &
echo $! >"$SCRATCH/admin.pid"

"$SCRATCH/punch" -addr=":$WORKER_PORT" \
	-dsn "file:$DB?_pragma=foreign_keys(1)" >"$SCRATCH/punch.log" 2>&1 &
echo $! >"$SCRATCH/punch.pid"

for _ in $(seq 1 40); do
	if curl -sf -o /dev/null "http://localhost:$ADMIN_PORT/admin/login" &&
		curl -sf -o /dev/null "http://localhost:$WORKER_PORT/"; then
		cat <<EOF
ready (admin pid $(cat "$SCRATCH/admin.pid"), punch pid $(cat "$SCRATCH/punch.pid"))
  admin site   http://localhost:$ADMIN_PORT/admin   $ADMIN_USER / $ADMIN_PASS
  punch site   http://localhost:$WORKER_PORT/       PINs: Ana 1111, Bob 2222, Cid 3333
  db           $DB
  logs         $SCRATCH/admin.log  $SCRATCH/punch.log
  stop         .claude/skills/smoke-test/scripts/scratch-down.sh
EOF
		exit 0
	fi
	sleep 0.25
done

echo "servers did not become ready; last log lines:" >&2
tail -20 "$SCRATCH/admin.log" "$SCRATCH/punch.log" >&2
exit 1
