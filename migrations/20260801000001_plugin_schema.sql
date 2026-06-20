-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS plugin_instances (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_type    TEXT    NOT NULL CHECK (plugin_type IN ('indexer', 'download_client')),
    implementation TEXT    NOT NULL,
    name           TEXT    NOT NULL DEFAULT '',
    config         TEXT    NOT NULL DEFAULT '{}',
    enabled        INTEGER NOT NULL DEFAULT 0,
    created_at     TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at     TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_instances_type_name ON plugin_instances (plugin_type, name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_plugin_instances_type_name;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS plugin_instances;
-- +goose StatementEnd
