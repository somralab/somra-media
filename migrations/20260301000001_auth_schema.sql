-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_account (
    id             TEXT    PRIMARY KEY,
    username       TEXT    NOT NULL UNIQUE COLLATE NOCASE,
    password_hash  TEXT    NOT NULL,
    disabled       INTEGER NOT NULL DEFAULT 0,
    created_at     TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at     TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS role (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT    NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS permission (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT    NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS role_permission (
    role_id       INTEGER NOT NULL REFERENCES role (id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permission (id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_role (
    user_id TEXT    NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES role (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS session (
    id           TEXT    PRIMARY KEY,
    user_id      TEXT    NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
    device_label TEXT,
    token_hash   TEXT    NOT NULL UNIQUE,
    expires_at   TEXT    NOT NULL,
    revoked_at   TEXT,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    last_used_at TEXT
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_profile (
    user_id            TEXT    PRIMARY KEY REFERENCES user_account (id) ON DELETE CASCADE,
    locale             TEXT    NOT NULL DEFAULT 'en-US'
        CHECK (locale IN ('en-US', 'tr-TR')),
    theme              TEXT    NOT NULL DEFAULT 'cinematic'
        CHECK (theme IN ('cinematic', 'aurora', 'noir', 'minimal')),
    avatar_url         TEXT,
    max_content_rating TEXT
        CHECK (max_content_rating IS NULL OR max_content_rating IN ('G', 'PG', 'PG-13', 'R', 'NC-17')),
    is_child           INTEGER NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS login_attempt (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    identifier   TEXT    NOT NULL,
    kind         TEXT    NOT NULL CHECK (kind IN ('ip', 'username')),
    failed_count INTEGER NOT NULL DEFAULT 0,
    locked_until TEXT,
    updated_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE (identifier, kind)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS watch_state (
    user_id       TEXT    NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
    media_item_id INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    position_ms   INTEGER NOT NULL DEFAULT 0,
    completed     INTEGER NOT NULL DEFAULT 0,
    updated_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (user_id, media_item_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS favorite (
    user_id       TEXT    NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
    media_item_id INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (user_id, media_item_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS watchlist (
    user_id       TEXT    NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
    media_item_id INTEGER NOT NULL REFERENCES media_item (id) ON DELETE CASCADE,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (user_id, media_item_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE media_item ADD COLUMN content_rating TEXT
    CHECK (content_rating IS NULL OR content_rating IN ('G', 'PG', 'PG-13', 'R', 'NC-17'));
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_session_user ON session (user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_session_expires ON session (expires_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_user_role_user ON user_role (user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_watch_state_user ON watch_state (user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_favorite_user ON favorite (user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_watchlist_user ON watchlist (user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_login_attempt_identifier ON login_attempt (identifier, kind);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO role (name) VALUES ('admin'), ('user'), ('child');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO permission (name) VALUES
    ('library:read'),
    ('library:write'),
    ('users:manage'),
    ('profile:edit');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'admin';
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r JOIN permission p ON p.name IN ('library:read', 'library:write', 'profile:edit')
WHERE r.name = 'user';
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r JOIN permission p ON p.name = 'library:read'
WHERE r.name = 'child';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_login_attempt_identifier;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_watchlist_user;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_favorite_user;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_watch_state_user;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_role_user;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_session_expires;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_session_user;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS watchlist;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS favorite;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS watch_state;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS login_attempt;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS user_profile;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS session;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS user_role;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS role_permission;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS permission;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS role;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS user_account;
-- +goose StatementEnd

-- SQLite cannot DROP COLUMN before 3.35; recreate media_item without content_rating is out of scope for down.
