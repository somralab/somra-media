-- +goose Up
-- Browse/discover hot paths at 10k+ media items (Sprint 10 A2).

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_media_item_library_created
    ON media_item (library_id, created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_media_item_library_sort
    ON media_item (library_id, sort_title COLLATE NOCASE);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_media_item_library_year
    ON media_item (library_id, year DESC, sort_title COLLATE NOCASE);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_watch_state_item
    ON watch_state (media_item_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_watch_state_user_updated
    ON watch_state (user_id, updated_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_media_genre_genre
    ON media_genre (genre_id, media_item_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_artwork_item_kind
    ON artwork (media_item_id, kind);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_artwork_item_kind;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_genre_genre;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_watch_state_user_updated;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_watch_state_item;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_item_library_year;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_item_library_sort;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_item_library_created;
-- +goose StatementEnd
