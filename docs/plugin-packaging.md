# Plugin packaging and legal isolation

Somra separates **core media server capabilities** from **content acquisition adapters**
(indexers and download clients). This document is the operational source of truth for how that
boundary is enforced at build time and how contributors add new adapters.

Related: [`.plan/architecture.md`](../.plan/architecture.md) §6 ·
[`.plan/sprint-09-automation-indexers/01-plugin-architecture-tasks.md`](../.plan/sprint-09-automation-indexers/01-plugin-architecture-tasks.md)

## Legal boundary

| Layer | License | Scope |
|---|---|---|
| **Core** | AGPL-3.0 | Library, metadata, streaming, auth, requests, settings — no outbound acquisition |
| **Acquisition adapters** | AGPL-3.0 (same repo today) | Torznab/Newznab indexers, qBittorrent/SABnzbd clients under `internal/plugin/<name>/` |

The core product must remain fully usable with **zero enabled acquisition plugins** (library
scanning, playback, request workflow). Integration tests verify this plugin-less mode.

Acquisition code is isolated in clearly named packages and can be omitted from builds using the
`acquisition` Go build tag. A future separate distribution (e.g. `somra-acquisition-plugins`) is
supported without changing the core contract.

## Build variants

| Target | Build tag | Adapters registered | Use case |
|---|---|---|---|
| **Core (default)** | _(none)_ | `stub` only | CI default, neutral distribution |
| **Full** | `acquisition` | stub + torznab + newznab + qbittorrent + sabnzbd | Home-lab images with acquisition |

```bash
# Core binary (default)
make build-go

# Full binary with acquisition adapters
make build-acquisition
```

Docker: pass `BUILD_TAGS=acquisition` as a build arg for full images (see `deploy/Dockerfile`).

## Package layout

```
internal/plugin/
  contract.go, indexer.go, download.go, manager.go   # core-facing interfaces
  stub/                                              # always built; tests + neutral mode
  torznab/       //go:build acquisition
  newznab/       //go:build acquisition
  qbittorrent/   //go:build acquisition
  sabnzbd/       //go:build acquisition
  nzbindexer/    //go:build acquisition   # shared Torznab/Newznab HTTP + XML

internal/automation/                               # orchestration (search, grab, import)
internal/platform/outbound/                        # pinned-host SSRF-safe HTTP client
```

Factory registration:

- [`internal/platform/bootstrap/plugins.go`](../internal/platform/bootstrap/plugins.go) — always registers `stub`.
- [`internal/platform/bootstrap/plugins_acquisition.go`](../internal/platform/bootstrap/plugins_acquisition.go) — registers acquisition adapters when `-tags acquisition`.

## Adapter configuration schemas

### Torznab (torrent indexer)

```json
{
  "baseUrl": "https://indexer.example/torznab/all",
  "categories": [2000, 5000]
}
```

Secrets: `apiKey` (optional depending on indexer).

### Newznab (Usenet indexer)

```json
{
  "baseUrl": "https://indexer.example/api",
  "categories": [2000, 5070]
}
```

Secrets: `apiKey`.

### qBittorrent (torrent download client)

```json
{
  "baseUrl": "http://qbittorrent:8080",
  "category": "somra",
  "savePath": "/downloads/incomplete"
}
```

Secrets: `username`, `password`.

### SABnzbd (Usenet download client)

```json
{
  "baseUrl": "http://sabnzbd:8080",
  "category": "somra"
}
```

Secrets: `apiKey`.

## Adding a new adapter

1. Implement `plugin.Factory` and the relevant interface (`Indexer` or `DownloadClient`).
2. Place code under `internal/plugin/<implementation>/` with `//go:build acquisition` if acquisition-related.
3. Register the factory in `plugins_acquisition.go`.
4. Document config/secrets in this file.
5. Add unit tests with `httptest` mock servers (no real network in CI).
6. Bump `plugin.ContractVersion` only when breaking interface/DTO changes (requires adapter recompile).

## Non-goals

- Runtime `.so` / CGO plugin loading
- Third-party plugin marketplace
- Bundling copyrighted indexer definitions or credentials
