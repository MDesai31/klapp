#!/usr/bin/env bash
set -euo pipefail

# Install (or update) the schedule listener on the home server, from a repo
# checkout on that same machine. Run from anywhere:
#
#   schedule_listener/install.sh
#
# It builds the binary, drops it and build_schedule.py in /opt/schedule-listener,
# installs the systemd unit, and restarts the service. config.json is written
# only if it isn't there yet, so a hand-edited one survives an update.

cd "$(dirname "$0")/.."

APP_DIR=/opt/schedule-listener
UNIT=schedule-listener.service

echo "building..."
go build -o /tmp/schedule-listener ./schedule_listener

sudo mkdir -p "$APP_DIR"
sudo chown "$USER" "$APP_DIR"

install -m 755 /tmp/schedule-listener "$APP_DIR/schedule-listener"
install -m 755 schedule_listener/build_schedule.py "$APP_DIR/build_schedule.py"
mkdir -p "$APP_DIR/out"

if [ -f "$APP_DIR/config.json" ]; then
	echo "keeping existing $APP_DIR/config.json"
else
	install -m 644 schedule_listener/config.example.json "$APP_DIR/config.json"
	echo "wrote $APP_DIR/config.json from the example — check addr, python and print_command"
fi

sudo install -m 644 "schedule_listener/$UNIT" "/etc/systemd/system/$UNIT"
sudo systemctl daemon-reload
sudo systemctl enable --now "$UNIT"
sudo systemctl restart "$UNIT"

systemctl --no-pager --lines=0 status "$UNIT" || true

# A bind failure is the usual first-run surprise (WireGuard not up yet), so
# say plainly whether the thing is actually answering.
addr=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["addr"])' "$APP_DIR/config.json")
echo
echo -n "health check http://$addr/healthz: "
curl -sS --max-time 5 "http://$addr/healthz" || echo "unreachable — journalctl -u $UNIT"
