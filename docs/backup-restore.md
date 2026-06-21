# Backup and Restore

Somra stores all state under configurable data directories. Back up before upgrades.

## What to Back Up

| Path (container) | Host mount | Contents |
|------------------|------------|----------|
| `/data` | `./data` | SQLite DB, user data, watch state |
| `/config` | `./config` | Optional runtime overrides |
| `/cache` | `./cache` | Transcode scratch (optional) |

Media files live on your library mounts (`./media` in compose) — back up separately.

## Backup Procedure

1. **Stop Somra** to flush SQLite WAL:
   ```bash
   docker compose -f deploy/docker-compose.production.yml stop somra
   ```
2. **Copy directories:**
   ```bash
   tar czf somra-backup-$(date +%Y%m%d).tar.gz deploy/data deploy/config
   ```
3. **Verify integrity** (optional):
   ```bash
   sqlite3 deploy/data/somra.db "PRAGMA integrity_check;"
   ```
4. **Start Somra:**
   ```bash
   docker compose -f deploy/docker-compose.production.yml start somra
   ```

## Restore Procedure

1. Stop the container.
2. Replace `data/` and `config/` from your backup archive.
3. Ensure file ownership matches the container user (usually root in default image).
4. Start Somra — migrations run automatically on boot if the image is newer.

## Upgrade Path

Automated upgrade coverage: `internal/platform/bootstrap/upgrade_integration_test.go`
(migrates from penultimate schema revision with sample row intact).

After upgrading the image tag, watch logs for migration messages and verify:

```bash
curl -s http://localhost:8080/api/v1/health | jq
curl -s http://localhost:8080/api/v1/version | jq
```

## Disaster Recovery

- Keep at least one off-site backup of `data/`.
- Document your `SOMRA_JWT_SECRET` securely — tokens invalidate if changed without migration plan.
- For corruption, restore from backup; do not delete WAL files manually while the server is running.
