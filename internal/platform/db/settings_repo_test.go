package db

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func newMigratedDB(t *testing.T) *DB {
	t.Helper()
	ctx := context.Background()
	d, err := Open(ctx, newTestConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })
	require.NoError(t, MigrateUp(ctx, d, nil))
	return d
}

func TestSettingsRepo_CRUD(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d := newMigratedDB(t)
	repo := NewSettingsRepo(d.Querier())

	t.Run("get missing returns ErrSettingNotFound", func(t *testing.T) {
		_, err := repo.Get(ctx, "missing.key")
		require.True(t, errors.Is(err, ErrSettingNotFound))
	})

	t.Run("set then get round-trips", func(t *testing.T) {
		require.NoError(t, repo.Set(ctx, "foo.bar", "baz"))
		v, err := repo.Get(ctx, "foo.bar")
		require.NoError(t, err)
		require.Equal(t, "baz", v)
	})

	t.Run("set updates existing value", func(t *testing.T) {
		require.NoError(t, repo.Set(ctx, "foo.bar", "qux"))
		v, err := repo.Get(ctx, "foo.bar")
		require.NoError(t, err)
		require.Equal(t, "qux", v)
	})

	t.Run("set tolerates empty value", func(t *testing.T) {
		require.NoError(t, repo.Set(ctx, "empty.key", ""))
		v, err := repo.Get(ctx, "empty.key")
		require.NoError(t, err)
		require.Equal(t, "", v)
	})

	t.Run("delete removes row", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, "foo.bar"))
		_, err := repo.Get(ctx, "foo.bar")
		require.True(t, errors.Is(err, ErrSettingNotFound))
	})

	t.Run("delete missing key is idempotent", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, "never.existed"))
	})
}

func TestSettingsRepo_UniqueViolation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := NewSettingsRepo(newMigratedDB(t).Querier())
	require.NoError(t, repo.Insert(ctx, "dup.key", "first"))

	err := repo.Insert(ctx, "dup.key", "second")
	require.Error(t, err)
	require.Contains(t, strings.ToUpper(err.Error()), "UNIQUE")
}

func TestSettingsRepo_EmptyKeyRejected(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := NewSettingsRepo(newMigratedDB(t).Querier())

	_, err := repo.Get(ctx, "")
	require.Error(t, err)
	_, err = repo.Get(ctx, "   ")
	require.Error(t, err)
	require.Error(t, repo.Set(ctx, "", "x"))
	require.Error(t, repo.Insert(ctx, "", "x"))
	require.Error(t, repo.Delete(ctx, ""))
}

func TestSettingsRepo_NullValueReadsAsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d := newMigratedDB(t)
	_, err := d.SQL().ExecContext(ctx,
		`INSERT INTO settings (key, value, updated_at) VALUES (?, NULL, datetime('now'))`,
		"null.key",
	)
	require.NoError(t, err)

	repo := NewSettingsRepo(d.Querier())
	v, err := repo.Get(ctx, "null.key")
	require.NoError(t, err)
	require.Equal(t, "", v)
}
