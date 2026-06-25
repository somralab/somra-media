#!/usr/bin/env bash
# CPU/memory profiling helper for Somra hot paths (Sprint 10 perf baseline).
# Usage: bash scripts/profile.sh [package]
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

pkg="${1:-./internal/library/...}"
cpu_out="${TMPDIR:-/tmp}/somra-cpu.prof"
mem_out="${TMPDIR:-/tmp}/somra-mem.prof"

echo ">> CPU profile: $pkg"
go test "$pkg" -count=1 -cpuprofile="$cpu_out" -bench=. -run=^$ 2>/dev/null || \
  go test "$pkg" -count=1 -cpuprofile="$cpu_out"

echo ">> Memory profile: $pkg"
go test "$pkg" -count=1 -memprofile="$mem_out" -bench=. -run=^$ 2>/dev/null || \
  go test "$pkg" -count=1 -memprofile="$mem_out"

echo "Profiles written:"
echo "  CPU: $cpu_out  (go tool pprof -top $cpu_out)"
echo "  Mem: $mem_out  (go tool pprof -top $mem_out)"
echo "See docs/performance-baseline.md for interpretation."
