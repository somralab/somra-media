-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS quality_profiles (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL UNIQUE,
    spec       TEXT    NOT NULL DEFAULT '{}',
    is_default INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS automation_handoffs (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id INTEGER NOT NULL UNIQUE REFERENCES requests (id) ON DELETE CASCADE,
    status     TEXT    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'grabbed', 'completed', 'failed')),
    error      TEXT,
    created_at TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS automation_downloads (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id         INTEGER REFERENCES requests (id) ON DELETE SET NULL,
    handoff_id         INTEGER REFERENCES automation_handoffs (id) ON DELETE SET NULL,
    client_instance_id INTEGER NOT NULL,
    client_download_id TEXT    NOT NULL,
    release_id         TEXT    NOT NULL DEFAULT '',
    title              TEXT    NOT NULL DEFAULT '',
    protocol           TEXT    NOT NULL CHECK (protocol IN ('torrent', 'usenet')),
    status             TEXT    NOT NULL DEFAULT 'queued'
        CHECK (status IN ('queued', 'downloading', 'paused', 'completed', 'failed')),
    progress           REAL    NOT NULL DEFAULT 0,
    save_path          TEXT    NOT NULL DEFAULT '',
    error              TEXT,
    created_at         TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at         TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_automation_handoffs_status ON automation_handoffs (status, created_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_automation_downloads_status ON automation_downloads (status, updated_at);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO quality_profiles (name, spec, is_default) VALUES (
    'default',
    '{"preferredResolutions":["1080p","720p","any"],"preferredCodecs":["hevc","h264"],"maxSizeBytes":0}',
    1
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_automation_downloads_status;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_automation_handoffs_status;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS automation_downloads;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS automation_handoffs;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS quality_profiles;
-- +goose StatementEnd
