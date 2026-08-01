#!/usr/bin/env bash
# Authenticated curl against the scratch admin site. Logs in on first use and
# reuses the session cookie afterwards.
#
#   admin-curl.sh /admin
#   admin-curl.sh /admin/punch/bulk -d "action=in&worker_id=1"
#   admin-curl.sh /admin/timesheet | grep '<td>'
#
# Always follows redirects (-L), which matters twice: GET /admin 301s to
# /admin/, and every mutation answers 303 See Other, so -L lands you on the
# rendered page (flash message included) rather than an empty body.
#
# Pass a body with -d alone, never "-X POST -d": -X pins the method for every
# hop of the redirect chain, so curl POSTs the 303 target too and the admin mux
# answers 405. -d already implies POST and lets curl switch to GET on the
# redirect, so this script strips a redundant "-X POST" rather than let it
# break the follow.
set -euo pipefail

ADMIN_PORT="${KLAPP_ADMIN_PORT:-18082}"
SCRATCH="${KLAPP_SCRATCH_DIR:-/tmp/klapp-scratch}"
JAR="$SCRATCH/cookies.txt"
BASE="http://localhost:$ADMIN_PORT"

if [ $# -lt 1 ]; then
	echo "usage: admin-curl.sh <path> [curl args...]" >&2
	exit 1
fi
path="$1"
shift

args=()
while [ $# -gt 0 ]; do
	if { [ "$1" = "-X" ] || [ "$1" = "--request" ]; } && [ "${2:-}" = "POST" ]; then
		shift 2
		continue
	fi
	args+=("$1")
	shift
done

if [ ! -s "$JAR" ]; then
	curl -sS -c "$JAR" -b "$JAR" -o /dev/null \
		-d "username=admin&password=scratchpw123" "$BASE/admin/login"
fi

curl -sS -b "$JAR" -c "$JAR" -L ${args+"${args[@]}"} "$BASE$path"
