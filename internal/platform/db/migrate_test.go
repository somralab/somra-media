package db

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/migrations"
)

func TestMigrateUp_Idempotent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d, err := Open(ctx, newTestConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	require.NoError(t, MigrateUp(ctx, d, logger))

	// Second call must be a no-op and must not error.
	require.NoError(t, MigrateUp(ctx, d, nil))

	// Settings table must exist after migration.
	var name string
	err = d.SQL().QueryRowContext(ctx,
		`SELECT name FROM sqlite_master WHERE type='table' AND name='settings'`,
	).Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "settings", name)
}

func TestMigrateStatus_ReportsTargetVersion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d, err := Open(ctx, newTestConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	require.NoError(t, MigrateUp(ctx, d, nil))

	cur, tgt, err := MigrateStatus(ctx, d)
	require.NoError(t, err)
	require.Greater(t, tgt, int64(0))
	require.Equal(t, tgt, cur)
}

func TestMigrateStatus_TargetMatchesEmbeddedFS(t *testing.T) {
	t.Parallel()

	// Walk the embedded FS independently of goose to make sure the helper
	// reports the highest version present.
	var max int64
	entries, err := fs.ReadDir(migrations.FS, ".")
	require.NoError(t, err)
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		// Format is <version>_<description>.sql
		parts := strings.SplitN(name, "_", 2)
		require.Len(t, parts, 2)
		var v int64
		for _, c := range parts[0] {
			v = v*10 + int64(c-'0')
		}
		if v > max {
			max = v
		}
	}

	d, err := Open(context.Background(), newTestConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	require.NoError(t, MigrateUp(context.Background(), d, nil))
	_, tgt, err := MigrateStatus(context.Background(), d)
	require.NoError(t, err)
	require.Equal(t, max, tgt)
}

func TestIntegrityCheck_Ok(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d, err := Open(ctx, newTestConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	require.NoError(t, MigrateUp(ctx, d, nil))

	result, err := IntegrityCheck(ctx, d)
	require.NoError(t, err)
	require.Equal(t, "ok", strings.TrimSpace(result))
}

func TestMigrate_NilHandles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	require.Error(t, MigrateUp(ctx, nil, nil))
	_, _, err := MigrateStatus(ctx, nil)
	require.Error(t, err)
	_, err = IntegrityCheck(ctx, nil)
	require.Error(t, err)
}

func TestMigrateStatus_BeforeMigrate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d, err := Open(ctx, newTestConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	// Calling Status on an un-migrated DB must not panic; goose returns 0 or
	// creates its bookkeeping table, both of which are acceptable.
	cur, tgt, err := MigrateStatus(ctx, d)
	require.NoError(t, err)
	require.GreaterOrEqual(t, tgt, int64(1))
	require.GreaterOrEqual(t, cur, int64(0))
}
