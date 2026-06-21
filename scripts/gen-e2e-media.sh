#!/usr/bin/env bash
# Generate a minimal H.264/AAC MP4 for playback e2e tests.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out_dir="$repo_root/testdata/media"
out_file="$out_dir/Sample Movie (2010).mp4"

mkdir -p "$out_dir"

if [[ -f "$out_file" ]] && [[ "${FORCE:-0}" != "1" ]]; then
  echo "e2e media fixture exists: $out_file"
  exit 0
fi

if ! command -v ffmpeg >/dev/null 2>&1; then
  echo "ffmpeg required to generate e2e media fixture" >&2
  exit 1
fi

ffmpeg -y -f lavfi -i testsrc=size=320x180:rate=24 -f lavfi -i sine=frequency=440:duration=5 \
  -c:v libx264 -preset ultrafast -t 5 -c:a aac -shortest "$out_file" \
  -loglevel error

echo "Generated $out_file"
