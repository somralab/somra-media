#!/usr/bin/env bash
# Enforce the coverage thresholds documented in plan/definition-of-done.md §4.1.
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
# Emits a human-readable summary and, when running under GitHub Actions,
# appends the same summary to $GITHUB_STEP_SUMMARY. Exits non-zero when any
# threshold is unmet.

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
)

# Entry-point binaries we intentionally exclude from the coverage gate. Their
# main() wiring is exercised by integration smoke tests, not by unit tests.
EXCLUDED_PACKAGES=(
  "github.com/somralab/somra-media/cmd/somra"
)

summary_file=""
if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  summary_file="$GITHUB_STEP_SUMMARY"
fi

emit() {
  printf '%s\n' "$1"
  if [[ -n "$summary_file" ]]; then
    printf '%s\n' "$1" >> "$summary_file"
  fi
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

  # Per-package coverage from go tool cover -func. The "total:" line at the
  # end is the aggregate; per-file lines come before it and we aggregate
  # them back to package level by stripping the file suffix.
  local func_out
  func_out=$(go tool cover -func="$GO_COVER")

  declare -A pkg_covered=()
  declare -A pkg_total=()

  # go tool cover -func prints lines like:
  #   github.com/x/y/file.go:12:  Func    50.0%
  # We need to aggregate per package. Easier path: re-run go tool cover -func
  # and re-parse coverage profile blocks per package. For accuracy we parse
  # the raw profile directly.

  # Parse coverprofile blocks: mode line + entries of the form
  #   <file>:<startLine>.<startCol>,<endLine>.<endCol> <numStmts> <count>
  while IFS= read -r line; do
    case "$line" in
      mode:*) continue ;;
      "") continue ;;
    esac
    local left rest stmts count file pkg
    left="${line%% *}"            # file:range
    rest="${line#* }"             # numStmts count
    stmts="${rest%% *}"
    count="${rest##* }"
    file="${left%%:*}"
    pkg="$(dirname "$file")"
    pkg_total["$pkg"]=$(( ${pkg_total[$pkg]:-0} + stmts ))
    if [[ "$count" != "0" ]]; then
      pkg_covered["$pkg"]=$(( ${pkg_covered[$pkg]:-0} + stmts ))
    fi
  done < "$GO_COVER"

  local total_stmts=0 total_cov=0
  for pkg in "${!pkg_total[@]}"; do
    total_stmts=$(( total_stmts + pkg_total[$pkg] ))
    total_cov=$(( total_cov + ${pkg_covered[$pkg]:-0} ))
  done

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

  local sorted
  sorted=$(printf '%s\n' "${!pkg_total[@]}" | sort)

  while IFS= read -r pkg; do
    local stmts=${pkg_total[$pkg]} cov=${pkg_covered[$pkg]:-0} threshold status pct
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
  done <<< "$sorted"

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
