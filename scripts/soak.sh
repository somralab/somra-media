#!/usr/bin/env bash
# Long-running soak test: periodic health checks + optional library scan load.
# Usage: SOMRA_SOAK_HOURS=4 bash scripts/soak.sh
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

hours="${SOMRA_SOAK_HOURS:-4}"
interval="${SOMRA_SOAK_INTERVAL_SEC:-60}"
addr="${SOMRA_HTTP_ADDR:-:8080}"
port="${addr##*:}"
base="http://127.0.0.1:${port}"
data_dir="${SOMRA_DATA_DIR:-/tmp/somra-soak-data}"
end=$(( $(date +%s) + hours * 3600 ))

mkdir -p "$data_dir"

if ! curl -fsS "$base/api/v1/health" >/dev/null 2>&1; then
  echo ">> Starting somra (background) with SOMRA_DATA_DIR=$data_dir"
  SOMRA_DATA_DIR="$data_dir" SOMRA_HTTP_ADDR="$addr" go run ./cmd/somra &
  pid=$!
  trap 'kill $pid 2>/dev/null || true' EXIT
  for _ in $(seq 1 30); do
    curl -fsS "$base/api/v1/health" >/dev/null 2>&1 && break
    sleep 1
  done
fi

echo ">> Soak test for ${hours}h (interval ${interval}s) against $base"
while [ "$(date +%s)" -lt "$end" ]; do
  health=$(curl -fsS "$base/api/v1/health" || echo '{"status":"down"}')
  echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) $health"
  sleep "$interval"
done
echo ">> Soak complete — review RSS via process monitor or /api/v1/health checks."
