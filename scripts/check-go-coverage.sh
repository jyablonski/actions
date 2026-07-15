#!/usr/bin/env bash

set -euo pipefail

# Keep the coverage policy in one place for Make, CI, and pre-commit; the temporary profile must be cleaned up on every exit path.
minimum_coverage=90
coverage_file="$(mktemp)"
trap 'rm -f "$coverage_file"' EXIT

if [[ -n "${GOTESTSUM_BIN:-}" ]]; then
  "$GOTESTSUM_BIN" --format pkgname -- -coverprofile="$coverage_file" ./...
else
  go run "gotest.tools/gotestsum@${GOTESTSUM_VERSION:-v1.13.0}" --format pkgname -- -coverprofile="$coverage_file" ./...
fi

go tool cover -func="$coverage_file"
coverage="$(go tool cover -func="$coverage_file" | awk '/^total:/{gsub("%", "", $3); print $3}')"

awk -v coverage="$coverage" -v minimum="$minimum_coverage" 'BEGIN {
  if (coverage + 0 < minimum) {
    printf "coverage %s%% is below the required %s%%\n", coverage, minimum
    exit 1
  }
  printf "coverage %s%% meets the required %s%%\n", coverage, minimum
}'
