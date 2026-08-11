#!/usr/bin/env bash
# Stop the scratch servers started by scratch-up.sh. Safe to run repeatedly.
set -euo pipefail

ADMIN_PORT="${KLAPP_ADMIN_PORT:-18082}"
WORKER_PORT="${KLAPP_WORKER_PORT:-14000}"
SCRATCH="${KLAPP_SCRATCH_DIR:-/tmp/klapp-scratch}"

for name in admin punch; do
	pidfile="$SCRATCH/$name.pid"
	if [ -f "$pidfile" ]; then
		kill "$(cat "$pidfile")" 2>/dev/null || true
		rm -f "$pidfile"
	fi
done

# Fallback for a server started without the pidfile: kill whatever is actually
# listening on the scratch ports. Deliberately NOT `pkill -f 18082` — that
# matches any command line mentioning the port, including the shell running
# this script, and kills it.
for port in "$ADMIN_PORT" "$WORKER_PORT"; do
	for pid in $(ss -tlnp "sport = :$port" 2>/dev/null | grep -oP 'pid=\K[0-9]+' | sort -u); do
		kill "$pid" 2>/dev/null || true
	done
done

rm -f "$SCRATCH/cookies.txt"
echo "scratch servers stopped"
