# API client (placeholder)

This directory is reserved for the typed HTTP client and OpenAPI-generated types.

- Generated TypeScript types will be written to `web/src/api/generated/` by
  `scripts/gen-openapi-types.sh` (owned by **Paket 2 — OpenAPI + Makefile**).
- The runtime API client, query hooks and SSE/WebSocket wiring land in
  **Paket 6 — Frontend integration**. Until then, pages render static i18n
  placeholders and the SPA does not call the backend.

Do not commit generated files; they are produced by the build pipeline.
