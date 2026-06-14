-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS library (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL,
    kind          TEXT    NOT NULL CHECK (kind IN ('movie', 'tv', 'music')),
    watch_enabled INTEGER NOT NULL DEFAULT 1,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS library_path (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    library_id INTEGER NOT NULL REFERENCES library (id) ON DELETE CASCADE,
    path       TEXT    NOT NULL,
    UNIQUE (library_id, path)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media_item (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    library_id   INTEGER NOT NULL REFERENCES library (id) ON DELETE CASCADE,
    kind         TEXT    NOT NULL CHECK (kind IN ('movie', 'tv', 'music')),
    sort_title   TEXT,
    year         INTEGER,
    match_status TEXT    NOT NULL DEFAULT 'unmatched'
        CHECK (match_status IN ('unmatched', 'matched', 'manual')),
    match_score  REAL,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS season (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    season_number INTEGER NOT NULL,
    UNIQUE (media_item_id, season_number)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS episode (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    season_id      INTEGER NOT NULL REFERENCES season (id) ON DELETE CASCADE,
    episode_number INTEGER NOT NULL,
    sort_title     TEXT,
    UNIQUE (season_id, episode_number)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media_file (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    library_id    INTEGER NOT NULL REFERENCES library (id) ON DELETE CASCADE,
    media_item_id INTEGER REFERENCES media_item (id) ON DELETE SET NULL,
    episode_id    INTEGER REFERENCES episode (id) ON DELETE SET NULL,
    path          TEXT    NOT NULL UNIQUE,
    file_name     TEXT    NOT NULL,
    size_bytes    INTEGER NOT NULL DEFAULT 0,
    mtime_ns      INTEGER NOT NULL DEFAULT 0,
    content_hash  TEXT,
    parsed_title  TEXT,
    parsed_year   INTEGER,
    parsed_season INTEGER,
    parsed_episode INTEGER,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media_technical (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    media_file_id  INTEGER NOT NULL UNIQUE REFERENCES media_file (id) ON DELETE CASCADE,
    duration_ms    INTEGER,
    container      TEXT,
    video_codec    TEXT,
    video_width    INTEGER,
    video_height   INTEGER,
    audio_codec    TEXT,
    audio_channels INTEGER,
    subtitle_count INTEGER NOT NULL DEFAULT 0,
    raw_json       TEXT,
    probed_at      TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media_stream (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    media_technical_id INTEGER NOT NULL REFERENCES media_technical (id) ON DELETE CASCADE,
    stream_index      INTEGER NOT NULL,
    stream_type       TEXT    NOT NULL CHECK (stream_type IN ('video', 'audio', 'subtitle')),
    codec             TEXT,
    language          TEXT,
    channels          INTEGER,
    width             INTEGER,
    height            INTEGER
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS artwork (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    kind          TEXT    NOT NULL CHECK (kind IN ('poster', 'backdrop', 'logo', 'thumb')),
    source_url    TEXT,
    local_path    TEXT,
    width         INTEGER,
    height        INTEGER,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS person (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media_person (
    media_item_id INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    person_id     INTEGER NOT NULL REFERENCES person (id) ON DELETE CASCADE,
    role          TEXT    NOT NULL CHECK (role IN ('actor', 'director', 'writer', 'artist')),
    sort_order    INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (media_item_id, person_id, role)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS genre (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media_genre (
    media_item_id INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    genre_id      INTEGER NOT NULL REFERENCES genre (id) ON DELETE CASCADE,
    PRIMARY KEY (media_item_id, genre_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tag (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media_tag (
    media_item_id INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    tag_id        INTEGER NOT NULL REFERENCES tag (id) ON DELETE CASCADE,
    PRIMARY KEY (media_item_id, tag_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS provider_id (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    provider      TEXT    NOT NULL CHECK (provider IN ('tmdb', 'tvdb', 'musicbrainz', 'fanart')),
    external_id   TEXT    NOT NULL,
    UNIQUE (media_item_id, provider)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS localized_text (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    locale        TEXT    NOT NULL CHECK (locale IN ('en-US', 'tr-TR')),
    field         TEXT    NOT NULL CHECK (field IN ('title', 'overview', 'tagline')),
    value         TEXT    NOT NULL,
    UNIQUE (media_item_id, locale, field)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS scan_run (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    library_id  INTEGER NOT NULL REFERENCES library (id) ON DELETE CASCADE,
    task_id     TEXT,
    scan_type   TEXT    NOT NULL CHECK (scan_type IN ('full', 'incremental')),
    status      TEXT    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'cancelled')),
    files_total INTEGER NOT NULL DEFAULT 0,
    files_done  INTEGER NOT NULL DEFAULT 0,
    error       TEXT,
    started_at  TEXT,
    finished_at TEXT,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_library_path_library ON library_path (library_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_media_item_library ON media_item (library_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_media_item_match ON media_item (match_status);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_media_file_library ON media_file (library_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_media_file_item ON media_file (media_item_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_scan_run_library ON scan_run (library_id, created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_localized_text_item ON localized_text (media_item_id, locale);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE VIRTUAL TABLE IF NOT EXISTS media_item_fts USING fts5 (
    title,
    tokenize='unicode61'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS media_item_fts;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_localized_text_item;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_scan_run_library;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_file_item;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_file_library;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_item_match;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_item_library;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_library_path_library;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS scan_run;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS localized_text;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS provider_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS media_tag;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS tag;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS media_genre;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS genre;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS media_person;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS person;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS artwork;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS media_stream;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS media_technical;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS media_file;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS episode;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS season;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS media_item;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS library_path;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS library;
-- +goose StatementEnd
