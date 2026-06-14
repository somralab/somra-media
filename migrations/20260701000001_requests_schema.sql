-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS requests (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id             TEXT    NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
    media_kind          TEXT    NOT NULL CHECK (media_kind IN ('movie', 'tv')),
    provider            TEXT    NOT NULL,
    external_id         TEXT    NOT NULL,
    title               TEXT    NOT NULL,
    poster_url          TEXT,
    quality_resolution  TEXT    NOT NULL DEFAULT 'any'
        CHECK (quality_resolution IN ('1080p', '720p', 'any')),
    quality_profile     TEXT,
    status              TEXT    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected', 'completed', 'cancelled')),
    collision_flag      INTEGER NOT NULL DEFAULT 0,
    admin_note          TEXT,
    created_at          TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at          TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS request_policies (
    id                    INTEGER PRIMARY KEY CHECK (id = 1),
    auto_approve_roles      TEXT    NOT NULL DEFAULT '[]',
    user_quota_per_month  INTEGER NOT NULL DEFAULT 10 CHECK (user_quota_per_month >= 0),
    admin_settings        TEXT    NOT NULL DEFAULT '{}',
    updated_at            TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notification_channels (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_type TEXT    NOT NULL CHECK (channel_type IN ('webhook', 'discord', 'email')),
    name         TEXT    NOT NULL DEFAULT '',
    config       TEXT    NOT NULL DEFAULT '{}',
    enabled      INTEGER NOT NULL DEFAULT 1,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notification_preferences (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id          TEXT    NOT NULL REFERENCES user_account (id) ON DELETE CASCADE,
    event_type       TEXT    NOT NULL,
    channel_id       INTEGER NOT NULL REFERENCES notification_channels (id) ON DELETE CASCADE,
    enabled          INTEGER NOT NULL DEFAULT 1,
    debounce_seconds INTEGER NOT NULL DEFAULT 0 CHECK (debounce_seconds >= 0),
    UNIQUE (user_id, event_type, channel_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_requests_user ON requests (user_id, created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_requests_status ON requests (status, created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_requests_provider ON requests (provider, external_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_notification_pref_user ON notification_preferences (user_id);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO request_policies (id, auto_approve_roles, user_quota_per_month, admin_settings)
VALUES (1, '["admin"]', 10, '{}');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO permission (name) VALUES
    ('requests:create'),
    ('requests:read'),
    ('requests:manage'),
    ('notifications:manage');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r
JOIN permission p ON p.name IN (
    'requests:create',
    'requests:read',
    'requests:manage',
    'notifications:manage'
)
WHERE r.name = 'admin';
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r
JOIN permission p ON p.name IN ('requests:create', 'requests:read')
WHERE r.name = 'user';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM role_permission
WHERE permission_id IN (
    SELECT id FROM permission WHERE name IN (
        'requests:create',
        'requests:read',
        'requests:manage',
        'notifications:manage'
    )
);
-- +goose StatementEnd

-- +goose StatementBegin
DELETE FROM permission WHERE name IN (
    'requests:create',
    'requests:read',
    'requests:manage',
    'notifications:manage'
);
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_notification_pref_user;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_requests_provider;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_requests_status;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_requests_user;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS notification_preferences;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS notification_channels;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS request_policies;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS requests;
-- +goose StatementEnd
