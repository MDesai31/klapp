#!/usr/bin/env bash
set -euo pipefail

# First-time setup: directories, systemd units, Caddy. For routine code
# pushes use update.sh instead.
#
#   ./deploy/deploy.sh               admin + invoice sites only (punch site off)
#   ./deploy/deploy.sh --with-punch  also enable the public worker punch site

cd "$(dirname "$0")/.."
source deploy/lib.sh

parse_args "$@"

sudo mkdir -p "$APP_DIR/db"
sudo chown -R "$(whoami):$(whoami)" "$APP_DIR"

build_binaries
install_units
retire_legacy_unit
apply_services

if ! command -v caddy >/dev/null; then
	sudo apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl
	curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
		| sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
	curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
		| sudo tee /etc/apt/sources.list.d/caddy-stable.list
	sudo apt-get update
	sudo apt-get install -y caddy
fi

sudo cp deploy/Caddyfile /etc/caddy/Caddyfile
sudo systemctl enable --now caddy
sudo systemctl reload caddy
