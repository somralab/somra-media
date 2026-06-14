-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS playback_session (
    id                     TEXT    PRIMARY KEY,
    user_id                TEXT    NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
    media_item_id          INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    media_file_id          INTEGER NOT NULL REFERENCES media_file (id) ON DELETE CASCADE,
    mode                   TEXT    NOT NULL CHECK (mode IN ('direct_play', 'direct_stream', 'transcode')),
    status                 TEXT    NOT NULL DEFAULT 'active'
        CHECK (status IN ('pending', 'queued', 'active', 'stopped', 'failed', 'expired')),
    cache_path             TEXT,
    start_position_ms      INTEGER NOT NULL DEFAULT 0,
    audio_stream_index     INTEGER,
    subtitle_stream_index  INTEGER,
    expires_at             TEXT    NOT NULL,
    last_access_at         TEXT    NOT NULL DEFAULT (datetime('now')),
    created_at             TEXT    NOT NULL DEFAULT (datetime('now')),
    error_message          TEXT
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_playback_session_user ON playback_session (user_id, created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_playback_session_status ON playback_session (status, last_access_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_playback_session_expires ON playback_session (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_playback_session_expires;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_playback_session_status;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_playback_session_user;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS playback_session;
-- +goose StatementEnd
