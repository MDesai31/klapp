#!/usr/bin/env bash
set -euo pipefail

# Run from the project root: ~/klapp
# For pushing code changes to an already-deployed instance (see deploy.sh
# for first-time setup: directories, Caddy, etc).
#
#   ./deploy/update.sh               admin + invoice sites only (punch site off)
#   ./deploy/update.sh --with-punch  also enable the public worker punch site

cd "$(dirname "$0")/.."
source deploy/lib.sh

parse_args "$@"
build_binaries
install_units
retire_legacy_unit
apply_services
