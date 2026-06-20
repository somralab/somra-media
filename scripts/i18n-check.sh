#!/usr/bin/env bash
# Locale parity check for Somra.
#
# Delegates the actual key-set comparison to the Go tool at
# cmd/i18n-check (uses the same toml/json dependencies already in the
# module). Run from the repository root.
#
# See .plan/i18n-localization.md §6 for the binding rules.

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

if ! command -v go >/dev/null 2>&1; then
  echo "i18n-check: go toolchain is required" >&2
  exit 2
fi

exec go run ./cmd/i18n-check "$@"
