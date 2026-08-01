#!/usr/bin/env bash
# Stop the scratch server started by scratch-up.sh. Safe to run repeatedly.
set -euo pipefail

ADMIN_PORT="${KLAPP_ADMIN_PORT:-18082}"
SCRATCH="${KLAPP_SCRATCH_DIR:-/tmp/klapp-scratch}"
PIDFILE="$SCRATCH/server.pid"

if [ -f "$PIDFILE" ]; then
	kill "$(cat "$PIDFILE")" 2>/dev/null || true
	rm -f "$PIDFILE"
fi

# Fallback for a server started without the pidfile: kill whatever is actually
# listening on the scratch admin port. Deliberately NOT `pkill -f 18082` — that
# matches any command line mentioning the port, including the shell running
# this script, and kills it.
for pid in $(ss -tlnp "sport = :$ADMIN_PORT" 2>/dev/null | grep -oP 'pid=\K[0-9]+' | sort -u); do
	kill "$pid" 2>/dev/null || true
done

rm -f "$SCRATCH/cookies.txt"
echo "scratch server stopped"
