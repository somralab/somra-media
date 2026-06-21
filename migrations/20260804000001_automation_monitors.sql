-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS automation_monitors (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         TEXT    NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
    title           TEXT    NOT NULL,
    provider        TEXT    NOT NULL DEFAULT 'tmdb',
    external_id     TEXT    NOT NULL,
    quality_profile TEXT    NOT NULL DEFAULT '',
    enabled         INTEGER NOT NULL DEFAULT 1,
    last_season     INTEGER NOT NULL DEFAULT 0,
    last_episode    INTEGER NOT NULL DEFAULT 0,
    last_checked_at TEXT,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE (provider, external_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS automation_monitor_episodes (
    monitor_id INTEGER NOT NULL REFERENCES automation_monitors (id) ON DELETE CASCADE,
    season     INTEGER NOT NULL,
    episode    INTEGER NOT NULL,
    request_id INTEGER REFERENCES requests (id) ON DELETE SET NULL,
    PRIMARY KEY (monitor_id, season, episode)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_automation_monitors_enabled ON automation_monitors (enabled, updated_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_automation_monitors_enabled;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS automation_monitor_episodes;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS automation_monitors;
-- +goose StatementEnd
