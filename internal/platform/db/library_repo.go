package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// LibraryKind identifies the media type a library holds.
type LibraryKind string

const (
	LibraryKindMovie LibraryKind = "movie"
	LibraryKindTV    LibraryKind = "tv"
	LibraryKindMusic LibraryKind = "music"
)

// Library is a user-defined media collection with one or more root paths.
type Library struct {
	ID           int64
	Name         string
	Kind         LibraryKind
	WatchEnabled bool
	Paths        []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// LibraryRepo persists library definitions.
type LibraryRepo struct {
	q Querier
}

// NewLibraryRepo returns a repository bound to q.
func NewLibraryRepo(q Querier) *LibraryRepo {
	return &LibraryRepo{q: q}
}

var ErrLibraryNotFound = errors.New("db library: not found")

// Create inserts a library and its paths atomically when called inside WithTx.
func (r *LibraryRepo) Create(ctx context.Context, name string, kind LibraryKind, paths []string, watchEnabled bool) (Library, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Library{}, fmt.Errorf("db library create: name is required")
	}
	if len(paths) == 0 {
		return Library{}, fmt.Errorf("db library create: at least one path is required")
	}

	res, err := r.q.ExecContext(ctx, `
		INSERT INTO library (name, kind, watch_enabled, created_at, updated_at)
		VALUES (?, ?, ?, datetime('now'), datetime('now'))
	`, name, kind, boolToInt(watchEnabled))
	if err != nil {
		return Library{}, fmt.Errorf("db library create: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Library{}, fmt.Errorf("db library create id: %w", err)
	}

	for _, p := range paths {
		if _, err := r.q.ExecContext(ctx, `
			INSERT INTO library_path (library_id, path) VALUES (?, ?)
		`, id, p); err != nil {
			return Library{}, fmt.Errorf("db library create path %q: %w", p, err)
		}
	}

	return r.GetByID(ctx, id)
}

// GetByID returns a library with its paths.
func (r *LibraryRepo) GetByID(ctx context.Context, id int64) (Library, error) {
	var lib Library
	var watch int
	var created, updated string
	err := r.q.QueryRowContext(ctx, `
		SELECT id, name, kind, watch_enabled, created_at, updated_at
		FROM library WHERE id = ?
	`, id).Scan(&lib.ID, &lib.Name, &lib.Kind, &watch, &created, &updated)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Library{}, fmt.Errorf("db library get %d: %w", id, ErrLibraryNotFound)
	case err != nil:
		return Library{}, fmt.Errorf("db library get %d: %w", id, err)
	}
	lib.WatchEnabled = watch != 0
	lib.CreatedAt, _ = time.Parse(time.RFC3339, created)
	lib.UpdatedAt, _ = time.Parse(time.RFC3339, updated)

	paths, err := r.listPaths(ctx, id)
	if err != nil {
		return Library{}, err
	}
	lib.Paths = paths
	return lib, nil
}

// List returns all libraries ordered by name.
func (r *LibraryRepo) List(ctx context.Context) ([]Library, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, name, kind, watch_enabled, created_at, updated_at
		FROM library ORDER BY name COLLATE NOCASE
	`)
	if err != nil {
		return nil, fmt.Errorf("db library list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Library
	for rows.Next() {
		var lib Library
		var watch int
		var created, updated string
		if err := rows.Scan(&lib.ID, &lib.Name, &lib.Kind, &watch, &created, &updated); err != nil {
			return nil, fmt.Errorf("db library list scan: %w", err)
		}
		lib.WatchEnabled = watch != 0
		lib.CreatedAt, _ = time.Parse(time.RFC3339, created)
		lib.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		paths, err := r.listPaths(ctx, lib.ID)
		if err != nil {
			return nil, err
		}
		lib.Paths = paths
		out = append(out, lib)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db library list rows: %w", err)
	}
	return out, nil
}

// Update replaces mutable fields and paths.
func (r *LibraryRepo) Update(ctx context.Context, id int64, name string, paths []string, watchEnabled bool) (Library, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Library{}, fmt.Errorf("db library update: name is required")
	}
	if len(paths) == 0 {
		return Library{}, fmt.Errorf("db library update: at least one path is required")
	}

	res, err := r.q.ExecContext(ctx, `
		UPDATE library SET name = ?, watch_enabled = ?, updated_at = datetime('now')
		WHERE id = ?
	`, name, boolToInt(watchEnabled), id)
	if err != nil {
		return Library{}, fmt.Errorf("db library update %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return Library{}, fmt.Errorf("db library update %d: %w", id, ErrLibraryNotFound)
	}

	if _, err := r.q.ExecContext(ctx, `DELETE FROM library_path WHERE library_id = ?`, id); err != nil {
		return Library{}, fmt.Errorf("db library update paths delete: %w", err)
	}
	for _, p := range paths {
		if _, err := r.q.ExecContext(ctx, `
			INSERT INTO library_path (library_id, path) VALUES (?, ?)
		`, id, p); err != nil {
			return Library{}, fmt.Errorf("db library update path %q: %w", p, err)
		}
	}
	return r.GetByID(ctx, id)
}

// Delete removes a library and cascaded children.
func (r *LibraryRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.q.ExecContext(ctx, `DELETE FROM library WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("db library delete %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("db library delete %d: %w", id, ErrLibraryNotFound)
	}
	return nil
}

func (r *LibraryRepo) listPaths(ctx context.Context, libraryID int64) ([]string, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT path FROM library_path WHERE library_id = ? ORDER BY id
	`, libraryID)
	if err != nil {
		return nil, fmt.Errorf("db library paths %d: %w", libraryID, err)
	}
	defer func() { _ = rows.Close() }()
	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("db library paths scan: %w", err)
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
