-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS settings (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    key        TEXT    NOT NULL UNIQUE,
    value      TEXT,
    updated_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_settings_key;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS settings;
-- +goose StatementEnd
