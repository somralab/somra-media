#!/usr/bin/env bash
# Sprint 10 C1 — enforce frontend JS bundle gzip budget after production build.
#
# Usage:
#   ./scripts/bundle-budget.sh
#   BUNDLE_MAX_KB=900 ./scripts/bundle-budget.sh
#
# Requires: web/dist from `pnpm --dir web run build` (or make build-web)

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

DIST="${DIST:-web/dist/assets}"
MAX_KB="${BUNDLE_MAX_KB:-850}"
WARN_KB="${BUNDLE_WARN_KB:-750}"

if [[ ! -d "$DIST" ]]; then
  echo "bundle-budget: $DIST missing — run make build-web first" >&2
  exit 1
fi

total_bytes=0
while IFS= read -r -d '' file; do
  size=$(gzip -c "$file" | wc -c | tr -d ' ')
  total_bytes=$((total_bytes + size))
done < <(find "$DIST" -type f \( -name '*.js' -o -name '*.css' \) -print0)

total_kb=$(( (total_bytes + 1023) / 1024 ))
max_bytes=$((MAX_KB * 1024))
warn_bytes=$((WARN_KB * 1024))

echo "bundle-budget: gzipped JS+CSS total = ${total_kb} KiB (warn ${WARN_KB} KiB, max ${MAX_KB} KiB)"

if (( total_bytes > max_bytes )); then
  echo "bundle-budget: FAIL — exceeds ${MAX_KB} KiB cap" >&2
  exit 1
fi

if (( total_bytes > warn_bytes )); then
  echo "bundle-budget: WARN — above ${WARN_KB} KiB soft target"
  exit 0
fi

echo "bundle-budget: OK"
