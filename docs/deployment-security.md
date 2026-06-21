# Deployment security defaults

Somra ships with conservative defaults for local development. Production
deployments must override secrets and network settings explicitly.

## JWT and cookies

| Setting | Default | Production |
| ------- | ------- | ---------- |
| `SOMRA_JWT_SECRET` | empty (ephemeral random at startup) | **Required** — ≥32 random bytes |
| `SOMRA_AUTH_SECURE_COOKIE` | `false` | `true` behind HTTPS |
| `SOMRA_REFRESH_PEPPER` | falls back to JWT secret | unique random value |

When `SOMRA_JWT_SECRET` is missing or short, Somra logs a warning and generates
an ephemeral secret — tokens become invalid after restart.

## CORS

Default allowed origin: `http://localhost:5173` (Vite dev server). Wildcard `*`
is never used.

Set explicit origins for your deployment:

```bash
export SOMRA_CORS_ORIGINS="https://somra.example"
```

## Rate limiting

- **HTTP gateway:** no global rate limiter by default (`NoopRateLimiter`); put
  reverse-proxy rate limits in front for public exposure.
- **Auth brute-force:** login lockout is enabled (5 failures → 15 minute lockout
  per IP/username). See Sprint 03 security tasks.
- **Metadata providers:** per-provider token buckets in the metadata module.

## HTTPS reverse proxy

Somra listens on plain HTTP inside the container (`:8080`). Terminate TLS at a
reverse proxy and forward `X-Forwarded-*` headers.

### Caddy

```caddy
somra.example {
    reverse_proxy 127.0.0.1:8080
}
```

### Traefik (Docker labels excerpt)

```yaml
labels:
  - traefik.http.routers.somra.rule=Host(`somra.example`)
  - traefik.http.routers.somra.entrypoints=websecure
  - traefik.http.routers.somra.tls.certresolver=letsencrypt
  - traefik.http.services.somra.loadbalancer.server.port=8080
```

After enabling HTTPS, set `SOMRA_CORS_ORIGINS` to your public origin and
`SOMRA_AUTH_SECURE_COOKIE=true`.

## Health diagnostics

`/api/v1/health` includes non-critical checks for disk space, ffmpeg/ffprobe,
and active transcode count. Use these for support triage — they do not replace
host-level monitoring.

## Related

- [backup-restore.md](./backup-restore.md)
- [plugin-packaging.md](./plugin-packaging.md) — acquisition isolation
- [SECURITY.md](../SECURITY.md) — disclosure process
