# Installation

Somra ships as a single Docker image. No separate database or reverse proxy is required.

## Quick start (one line)

After v1.0.0 is published:

```bash
docker run -d --name somra -p 8080:8080 \
  -e SOMRA_JWT_SECRET="$(openssl rand -base64 32)" \
  -v somra-data:/data -v somra-cache:/cache \
  -v /path/to/media:/media:ro \
  ghcr.io/somralab/somra-media:1.0.0
```

Open **http://localhost:8080** and complete the setup wizard.

## Docker Compose (recommended)

Production example: [`deploy/docker-compose.production.yml`](../../deploy/docker-compose.production.yml)

```bash
mkdir -p deploy/config deploy/data deploy/cache deploy/media
cp deploy/.env.production.example deploy/.env
# Edit deploy/.env — set SOMRA_JWT_SECRET
docker compose -f deploy/docker-compose.production.yml up -d
```

Development build from source:

```bash
docker compose -f deploy/docker-compose.yml up --build
```

## Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SOMRA_JWT_SECRET` | Production yes | 32+ char secret for JWT signing |
| `SOMRA_DATA_DIR` | No | SQLite and state (default `/data`) |
| `SOMRA_CACHE_DIR` | No | Transcode cache (default `/cache`) |
| `SOMRA_HTTP_ADDR` | No | Listen address (default `:8080`) |

## GPU passthrough

See [`docs/gpu-setup.md`](../gpu-setup.md) for VAAPI, NVIDIA, and QSV overlays.

## HTTPS (reverse proxy)

Place Somra behind Caddy or Traefik on your LAN or public domain:

```caddy
somra.example.com {
  reverse_proxy localhost:8080
}
```

Do not expose Somra directly to the internet without TLS.

## System requirements

- **CPU:** 4 cores recommended
- **RAM:** 8 GB recommended
- **Disk:** SSD for `/data` and `/cache`; media on separate storage
- **OS:** Linux with Docker 24+ (amd64 or arm64)
