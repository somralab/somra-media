# FAQ & Troubleshooting

## Installation

**Container exits immediately** — check logs: `docker logs somra`. Verify `SOMRA_DATA_DIR` is writable.

**Cannot reach UI** — confirm port mapping and firewall. Health: `curl localhost:8080/api/v1/health`.

## Library

**Scan finds no files** — verify the path is mounted read-only inside the container and file extensions are supported.

**Metadata missing** — set `SOMRA_TMDB_API_KEY` or use test metadata in dev (`SOMRA_USE_TEST_METADATA=1`).

## Playback

**ffmpeg not found** — production image includes ffmpeg; host `go run` needs ffmpeg installed locally.

**Transcode fails** — inspect logs; ensure cache volume has free space (`health` → `disk` check).

## Database

**SQLite locked** — only one Somra instance must write to the same `data/` directory. Stop duplicate containers.

## GPU

See [gpu-setup.md](../gpu-setup.md). Verify `/api/v1/system/detect` reports accelerators.

## Plugins

**Connection test fails** — check plugin URL, credentials, and network from the container.

## Getting help

- [GitHub Issues](https://github.com/somralab/somra-media/issues)
- [SECURITY.md](../../SECURITY.md) for vulnerabilities
