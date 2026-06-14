-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS subtitle_file (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL REFERENCES media_item(id) ON DELETE CASCADE,
    language      TEXT    NOT NULL,
    source        TEXT    NOT NULL CHECK (source IN ('embedded', 'external', 'uploaded')),
    path          TEXT    NOT NULL,
    provider      TEXT,
    external_id   TEXT,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_subtitle_file_media ON subtitle_file(media_item_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_subtitle_file_media;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS subtitle_file;
-- +goose StatementEnd
