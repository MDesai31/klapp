#!/usr/bin/env bash
set -euo pipefail

# Run from the project root: ~/klapp
APP_DIR=/opt/klapp

sudo mkdir -p "$APP_DIR/db"
sudo chown -R "$(whoami):$(whoami)" "$APP_DIR"

go build -o "$APP_DIR/web" ./cmd/web
rsync -a --delete ui/ "$APP_DIR/ui/"

sudo cp deploy/klapp.service /etc/systemd/system/klapp.service
sudo systemctl daemon-reload
sudo systemctl enable --now klapp

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
