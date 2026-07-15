#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${DEADCODE_BIN:-}" ]]; then
  exec "$DEADCODE_BIN" -test ./...
fi

exec go run "golang.org/x/tools/cmd/deadcode@${DEADCODE_VERSION:-v0.48.0}" -test ./...
