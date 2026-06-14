-- +goose Up
-- +goose StatementBegin
INSERT INTO settings (key, value, updated_at)
SELECT 'system.installed_at', datetime('now'), datetime('now')
WHERE NOT EXISTS (
    SELECT 1 FROM settings WHERE key = 'system.installed_at'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM settings WHERE key = 'system.installed_at';
-- +goose StatementEnd
