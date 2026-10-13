#!/usr/bin/env bash
# Enforce the coverage thresholds documented in .plan/definition-of-done.md §4.1.
#
# Inputs (auto-detected, overridable via env):
#   GO_COVER          Go coverprofile (default: coverage.go.out)
#   WEB_COVER_SUMMARY Vitest c8 json-summary file
#                     (default: web/coverage/coverage-summary.json)
#
# Thresholds (line coverage, percent):
#   - Core Go business logic              >= 80
#   - Critical Go modules (auth, jobs,    >= 90
#     platform/db, platform/errors,
#     platform/i18n, platform/diagnostics)
#   - Frontend statements                 >= 70
#
# Emits a human-readable summary. Exits non-zero when any threshold is unmet.

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

GO_COVER="${GO_COVER:-coverage.go.out}"
WEB_COVER_SUMMARY="${WEB_COVER_SUMMARY:-web/coverage/coverage-summary.json}"

CORE_THRESHOLD="${CORE_THRESHOLD:-80}"
CRITICAL_THRESHOLD="${CRITICAL_THRESHOLD:-90}"
WEB_THRESHOLD="${WEB_THRESHOLD:-70}"

# Critical modules: each path matches a Go import path prefix.
CRITICAL_PACKAGES=(
  "github.com/somralab/somra-media/internal/jobs"
  "github.com/somralab/somra-media/internal/auth"
  "github.com/somralab/somra-media/internal/platform/db"
  "github.com/somralab/somra-media/internal/platform/errors"
  "github.com/somralab/somra-media/internal/platform/i18n"
  "github.com/somralab/somra-media/internal/platform/diagnostics"
  "github.com/somralab/somra-media/internal/streaming"
  "github.com/somralab/somra-media/internal/requests"
  "github.com/somralab/somra-media/internal/notifications"
)

# Entry-point binaries we intentionally exclude from the coverage gate. Their
# main() wiring is exercised by integration smoke tests, not by unit tests.
EXCLUDED_PACKAGES=(
  "github.com/somralab/somra-media/cmd/somra"
  "github.com/somralab/somra-media/internal/automation/automationtest"
)

emit() {
  printf '%s\n' "$1"
}

failed=0

check_go() {
  if [[ ! -f "$GO_COVER" ]]; then
    emit "❌ Go coverage profile not found at $GO_COVER"
    failed=1
    return
  fi
  if ! command -v go >/dev/null 2>&1; then
    emit "❌ go toolchain missing — cannot evaluate Go coverage"
    failed=1
    return
  fi

  emit "### Go coverage gate"
  emit ""
  emit "| Package | Coverage | Threshold | Status |"
  emit "|---|---|---|---|"

  # Aggregate coverprofile blocks per package. Entries have the form
  #   <file>:<startLine>.<startCol>,<endLine>.<endCol> <numStmts> <count>
  # Done in awk rather than bash associative arrays so the gate also runs on
  # macOS's stock bash 3.2 (there is no CI; this runs on developer machines).
  local per_pkg
  per_pkg=$(awk '
    /^mode:/ || NF == 0 { next }
    {
      file = $1; sub(/:.*/, "", file)
      pkg = file; sub(/\/[^\/]*$/, "", pkg)
      total[pkg] += $2
      if ($3 != "0") covered[pkg] += $2
    }
    END { for (p in total) printf "%s %d %d\n", p, covered[p], total[p] }
  ' "$GO_COVER" | sort)

  local total_stmts=0 total_cov=0

  is_critical() {
    local pkg="$1"
    for c in "${CRITICAL_PACKAGES[@]}"; do
      if [[ "$pkg" == "$c" || "$pkg" == "$c/"* ]]; then return 0; fi
    done
    return 1
  }

  is_excluded() {
    local pkg="$1"
    for c in "${EXCLUDED_PACKAGES[@]}"; do
      if [[ "$pkg" == "$c" || "$pkg" == "$c/"* ]]; then return 0; fi
    done
    return 1
  }

  local pkg cov stmts threshold status pct
  while read -r pkg cov stmts; do
    if [[ -z "$pkg" ]]; then continue; fi
    total_stmts=$(( total_stmts + stmts ))
    total_cov=$(( total_cov + cov ))
    if (( stmts == 0 )); then continue; fi
    if is_excluded "$pkg"; then
      pct=$(awk -v c="$cov" -v s="$stmts" 'BEGIN{printf "%.1f", (s==0?100:c*100/s)}')
      emit "| ${pkg#github.com/somralab/somra-media/} | ${pct}% | (excluded — entrypoint) | ⏭️ |"
      continue
    fi
    if is_critical "$pkg"; then threshold=$CRITICAL_THRESHOLD; else threshold=$CORE_THRESHOLD; fi
    pct=$(awk -v c="$cov" -v s="$stmts" 'BEGIN{printf "%.1f", (s==0?100:c*100/s)}')
    local ok
    ok=$(awk -v p="$pct" -v t="$threshold" 'BEGIN{print (p+0 >= t+0) ? 1 : 0}')
    if [[ "$ok" == "1" ]]; then status="✅"; else status="❌"; failed=1; fi
    emit "| ${pkg#github.com/somralab/somra-media/} | ${pct}% | ${threshold}% | ${status} |"
  done <<< "$per_pkg"

  local total_pct
  total_pct=$(awk -v c="$total_cov" -v s="$total_stmts" 'BEGIN{printf "%.1f", (s==0?100:c*100/s)}')
  emit ""
  emit "Go aggregate: ${total_pct}% (${total_cov}/${total_stmts} statements)"
  emit ""
}

check_web() {
  emit "### Frontend coverage gate"
  emit ""
  if [[ ! -f "$WEB_COVER_SUMMARY" ]]; then
    emit "❌ Frontend coverage summary not found at $WEB_COVER_SUMMARY"
    failed=1
    return
  fi
  if ! command -v node >/dev/null 2>&1; then
    emit "❌ node runtime missing — cannot evaluate frontend coverage"
    failed=1
    return
  fi

  local pct
  pct=$(node -e '
    const fs = require("node:fs");
    const path = process.argv[1];
    const s = JSON.parse(fs.readFileSync(path, "utf8"));
    const t = s.total || {};
    const stmts = t.statements || t.lines || {};
    process.stdout.write(String(stmts.pct ?? 0));
  ' "$WEB_COVER_SUMMARY")

  local ok status
  ok=$(awk -v p="$pct" -v t="$WEB_THRESHOLD" 'BEGIN{print (p+0 >= t+0) ? 1 : 0}')
  if [[ "$ok" == "1" ]]; then status="✅"; else status="❌"; failed=1; fi
  emit "| Metric | Coverage | Threshold | Status |"
  emit "|---|---|---|---|"
  emit "| statements | ${pct}% | ${WEB_THRESHOLD}% | ${status} |"
  emit ""
}

check_go
check_web

if [[ "$failed" -ne 0 ]]; then
  emit "❌ coverage-gate: thresholds not met"
  exit 1
fi
emit "✅ coverage-gate: all thresholds met"
