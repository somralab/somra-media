#!/usr/bin/env bash
# Generate static Redoc HTML from the OpenAPI spec.
#
# Output: docs/api/index.html
# Requires: npx (Node.js) — uses @redocly/cli, no global install.

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

spec="${OPENAPI_SPEC:-api/openapi.yaml}"
out_dir="docs/api"
out_file="${out_dir}/index.html"

if [[ ! -f "$spec" ]]; then
  echo "gen-api-docs: spec not found: $spec" >&2
  exit 1
fi

mkdir -p "$out_dir"

echo ">> npx @redocly/cli build-docs $spec -o $out_file"
npx -y @redocly/cli@1 build-docs "$spec" -o "$out_file"

echo "gen-api-docs: wrote $out_file"
