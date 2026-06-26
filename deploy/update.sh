#!/usr/bin/env bash
set -euo pipefail

# Run from the project root: ~/klapp
# For pushing code changes to an already-deployed instance (see deploy.sh
# for first-time setup: systemd unit, Caddy, etc).
APP_DIR=/opt/klapp

go build -o "$APP_DIR/web" ./cmd/web
go build -o "$APP_DIR/invoice" ./cmd/invoice
rsync -a --delete ui/ "$APP_DIR/ui/"

sudo systemctl restart klapp
sudo systemctl restart klapp-invoice
