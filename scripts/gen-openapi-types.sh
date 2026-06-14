#!/usr/bin/env bash
# Generate TypeScript types for the frontend from the OpenAPI 3.1 spec.
#
# The OpenAPI document at api/openapi.yaml is the single source of truth for
# the Somra HTTP API. Frontend code MUST NOT hand-author request/response
# types — they are generated here and consumed via web/src/api/.
#
# Usage:
#   scripts/gen-openapi-types.sh            # default paths
#   make openapi-types                      # equivalent
#
# Requirements:
#   - Node.js (LTS) + npx on PATH. The generator is invoked via `npx -y`
#     so no global install is needed.
#
# This script is idempotent: re-running it overwrites the generated file.
# It creates the output directory if missing so it works even before the
# `web/` SPA has been scaffolded (the file will simply be ready to import
# once the SPA exists).

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
REPO_ROOT="$(cd -- "${SCRIPT_DIR}/.." >/dev/null 2>&1 && pwd)"

SPEC_PATH="${SPEC_PATH:-${REPO_ROOT}/api/openapi.yaml}"
OUTPUT_PATH="${OUTPUT_PATH:-${REPO_ROOT}/web/src/api/generated/openapi.d.ts}"

if [[ ! -f "${SPEC_PATH}" ]]; then
  echo "error: OpenAPI spec not found at: ${SPEC_PATH}" >&2
  exit 1
fi

if ! command -v npx >/dev/null 2>&1; then
  echo "error: npx (Node.js) is required to generate OpenAPI types." >&2
  echo "       Install Node.js LTS and retry." >&2
  exit 1
fi

mkdir -p "$(dirname -- "${OUTPUT_PATH}")"

echo "Generating OpenAPI TypeScript types"
echo "  spec  : ${SPEC_PATH}"
echo "  output: ${OUTPUT_PATH}"

npx -y openapi-typescript@latest "${SPEC_PATH}" -o "${OUTPUT_PATH}"

echo "Done."
