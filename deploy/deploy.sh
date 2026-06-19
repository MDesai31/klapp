#!/usr/bin/env bash
set -euo pipefail

# Run from the project root: ~/klapp
APP_DIR=/opt/klapp

sudo mkdir -p "$APP_DIR"
sudo chown "$(whoami):$(whoami)" "$APP_DIR"

go build -o "$APP_DIR/web" ./cmd/web
rsync -a --delete ui/ "$APP_DIR/ui/"

sudo cp deploy/klapp.service /etc/systemd/system/klapp.service
sudo systemctl daemon-reload
sudo systemctl enable --now klapp
