-- +goose Up
-- Plugin secrets are encrypted at rest in secrets_enc; config holds public fields only.
-- Re-encrypting after JWT secret rotation is not automatic (single-tenant home server trade-off).
-- +goose StatementBegin
ALTER TABLE plugin_instances ADD COLUMN secrets_enc TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO permission (name) VALUES ('plugins:manage');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r
JOIN permission p ON p.name = 'plugins:manage'
WHERE r.name = 'admin';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM role_permission
WHERE permission_id IN (SELECT id FROM permission WHERE name = 'plugins:manage');
-- +goose StatementEnd

-- +goose StatementBegin
DELETE FROM permission WHERE name = 'plugins:manage';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plugin_instances DROP COLUMN secrets_enc;
-- +goose StatementEnd
