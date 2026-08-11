#!/usr/bin/env bash
# Shared helpers for deploy.sh and update.sh. Source, don't execute.

APP_DIR=/opt/klapp

# parse_args sets WITH_PUNCH from the command line. The public worker punch
# site is off unless explicitly asked for, so the default deploy leaves only
# the LAN-only admin site (and the invoice site) exposed.
WITH_PUNCH=0
parse_args() {
	for arg in "$@"; do
		case "$arg" in
		--with-punch)
			WITH_PUNCH=1
			;;
		-h | --help)
			echo "usage: $0 [--with-punch]"
			echo
			echo "  --with-punch  also enable and start the public worker punch site"
			echo "                (default: punch site is stopped and disabled)"
			exit 0
			;;
		*)
			echo "unknown option: $arg" >&2
			echo "usage: $0 [--with-punch]" >&2
			exit 2
			;;
		esac
	done
}

build_binaries() {
	go build -o "$APP_DIR/admin" ./cmd/admin
	go build -o "$APP_DIR/punch" ./cmd/punch
	go build -o "$APP_DIR/invoice" ./cmd/invoice
	rsync -a --delete ui/ "$APP_DIR/ui/"
}

# install_units keeps the systemd units in sync with the repo on every push,
# so a unit file edited here never silently fails to take effect.
#
# Each destination is removed first. An older setup here symlinked the unit in
# /etc/systemd/system back to the copy in this repo; plain `cp` onto such a
# symlink resolves to source and destination being the same file, which fails
# ("are the same file") and, under `set -e`, aborts the whole deploy before any
# service is touched. Removing the destination path unlinks the symlink itself,
# leaving the repo file alone, and replaces it with a real file.
install_units() {
	local unit
	for unit in klapp-admin klapp-punch klapp-invoice; do
		sudo rm -f "/etc/systemd/system/$unit.service"
		sudo cp "deploy/$unit.service" "/etc/systemd/system/$unit.service"
	done
	sudo systemctl daemon-reload
}

# retire_legacy_unit removes the pre-split `klapp` unit, which ran a single
# `web` binary serving both sites. Harmless to run once the unit is gone.
#
# Test -L as well as -e: the legacy unit is a symlink to deploy/klapp.service,
# which the split renamed to klapp-admin.service, so it is now dangling and
# -e/-f (which follow symlinks) both report false while the unit is still
# loaded and running.
retire_legacy_unit() {
	local unit=/etc/systemd/system/klapp.service

	if [ -e "$unit" ] || [ -L "$unit" ]; then
		echo "retiring legacy klapp.service (replaced by klapp-admin + klapp-punch)"
		# Stop before disable, not `disable --now`: with the unit file missing
		# systemd calls the unit "bad" and `disable` fails, which would skip
		# the stop and leave the old binary holding :4000 and :8082.
		sudo systemctl stop klapp.service || true
		sudo systemctl disable klapp.service || true
		sudo rm -f "$unit"
		sudo rm -f "$APP_DIR/web"
		sudo systemctl daemon-reload
		sudo systemctl reset-failed klapp.service 2>/dev/null || true
	fi
}

# apply_services brings the units to the state WITH_PUNCH asks for. The
# admin and invoice sites always run; the punch site is enabled or disabled
# outright rather than merely stopped, so a reboot doesn't resurrect it.
apply_services() {
	sudo systemctl enable --now klapp-admin
	sudo systemctl enable --now klapp-invoice
	sudo systemctl restart klapp-admin
	sudo systemctl restart klapp-invoice

	if [ "$WITH_PUNCH" -eq 1 ]; then
		sudo systemctl enable --now klapp-punch
		sudo systemctl restart klapp-punch
		echo "worker punch site: ENABLED (public via Caddy on :4000)"
	else
		sudo systemctl disable --now klapp-punch || true
		echo "worker punch site: DISABLED (Caddy will return 502 for work.klauslandscaping.com)"
		echo "  re-run with --with-punch to bring it back up"
	fi
}
