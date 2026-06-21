#!/usr/bin/env bash
# Sprint 10 D1 — lightweight soak test against a running Somra instance.
#
# Usage:
#   make build-go && ./scripts/soak-test.sh
#   SOAK_DURATION=30m SOAK_INTERVAL=30s ./scripts/soak-test.sh
#
# Env:
#   SOMRA_BASE_URL   default http://127.0.0.1:8080
#   SOAK_DURATION    default 15m (use 4-8h pre-release)
#   SOAK_INTERVAL    default 60s between health probes
#   SOAK_MAX_RSS_MB  default 512 (warn threshold from baseline doc)

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

BASE_URL="${SOMRA_BASE_URL:-http://127.0.0.1:8080}"
DURATION="${SOAK_DURATION:-15m}"
INTERVAL="${SOAK_INTERVAL:-60s}"
MAX_RSS_MB="${SOAK_MAX_RSS_MB:-512}"
BIN="${BIN:-bin/somra}"

if [[ ! -x "$BIN" ]]; then
  echo "soak-test: build binary first (make build-go)" >&2
  exit 1
fi

data_dir="$(mktemp -d "${TMPDIR:-/tmp}/somra-soak.XXXXXX")"
cleanup() {
  if [[ -n "${server_pid:-}" ]] && kill -0 "$server_pid" 2>/dev/null; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  rm -rf "$data_dir"
}
trap cleanup EXIT

export SOMRA_DATA_DIR="$data_dir"
export SOMRA_HTTP_ADDR="127.0.0.1:18080"
export SOMRA_LOG_LEVEL=warn

echo ">> starting server (data: $data_dir)"
"$BIN" &
server_pid=$!

for _ in $(seq 1 60); do
  if curl -sf "http://127.0.0.1:18080/api/v1/health" >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
done

BASE_URL="http://127.0.0.1:18080"
end_epoch=$(( $(date +%s) + $(python3 - <<PY
import os
d=os.environ.get("SOAK_DURATION","15m").strip()
if d.endswith("h"):
    print(int(float(d[:-1])*3600))
elif d.endswith("m"):
    print(int(float(d[:-1])*60))
else:
    print(int(d))
PY
) ))

echo ">> soaking $BASE_URL for $DURATION (interval $INTERVAL)"
failures=0
while (( $(date +%s) < end_epoch )); do
  if ! curl -sf "$BASE_URL/api/v1/health" -o /tmp/somra-soak-health.json; then
    echo "soak-test: health probe failed" >&2
    failures=$((failures + 1))
  else
    rss_kb=$(ps -o rss= -p "$server_pid" 2>/dev/null | tr -d ' ' || echo 0)
    rss_mb=$(( (rss_kb + 1023) / 1024 ))
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) health=ok rss=${rss_mb}MiB"
    if (( rss_mb > MAX_RSS_MB )); then
      echo "soak-test: WARN RSS ${rss_mb}MiB exceeds soft cap ${MAX_RSS_MB}MiB" >&2
    fi
  fi
  sleep "$INTERVAL"
done

if (( failures > 0 )); then
  echo "soak-test: FAIL ($failures health errors)" >&2
  exit 1
fi

echo "soak-test: OK"
