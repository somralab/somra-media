package db

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/migrations"
)

func TestListEmbeddedVersions_Sorted(t *testing.T) {
	t.Parallel()
	versions, err := ListEmbeddedVersions()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(versions), 2)
	for i := 1; i < len(versions); i++ {
		assert.Less(t, versions[i-1], versions[i])
	}
}

func TestMigrateUpTo_StopsAtVersion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	versions, err := ListEmbeddedVersions()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(versions), 2)
	penultimate := versions[len(versions)-2]

	d, err := Open(ctx, newTestConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	require.NoError(t, MigrateUpTo(ctx, d, penultimate, nil))
	cur, tgt, err := MigrateStatus(ctx, d)
	require.NoError(t, err)
	require.Equal(t, penultimate, cur)
	require.Equal(t, tgt, versions[len(versions)-1])

	require.NoError(t, MigrateUp(ctx, d, nil))
	cur, tgt, err = MigrateStatus(ctx, d)
	require.NoError(t, err)
	require.Equal(t, tgt, cur)
}

func TestMigrateUpTo_PreservesRowsAcrossFinalMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	versions, err := ListEmbeddedVersions()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(versions), 2)
	penultimate := versions[len(versions)-2]

	dataDir := t.TempDir()
	cfg := Config{DataDir: dataDir, DBFile: "upgrade.db"}
	d, err := Open(ctx, cfg)
	require.NoError(t, err)
	require.NoError(t, MigrateUpTo(ctx, d, penultimate, nil))

	const markerKey = "upgrade_test_marker"
	_, err = d.SQL().ExecContext(ctx,
		`INSERT INTO settings (key, value, updated_at) VALUES (?, 'keep', datetime('now'))`,
		markerKey,
	)
	require.NoError(t, err)
	dbPath := filepath.Join(dataDir, cfg.DBFile)
	require.NoError(t, d.Close())

	copyDir := t.TempDir()
	copyPath := filepath.Join(copyDir, "snapshot.db")
	raw, err := os.ReadFile(dbPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(copyPath, raw, 0o600))

	cfg2 := Config{DataDir: copyDir, DBFile: "snapshot.db"}
	d2, err := Open(ctx, cfg2)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d2.Close() })

	require.NoError(t, MigrateUp(ctx, d2, nil))
	var value string
	err = d2.SQL().QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, markerKey).Scan(&value)
	require.NoError(t, err)
	require.Equal(t, "keep", value)

	result, err := IntegrityCheck(ctx, d2)
	require.NoError(t, err)
	require.Equal(t, "ok", result)

	cur, tgt, err := MigrateStatus(ctx, d2)
	require.NoError(t, err)
	require.Equal(t, tgt, cur)
}

func TestListEmbeddedVersions_MatchesFS(t *testing.T) {
	t.Parallel()
	versions, err := ListEmbeddedVersions()
	require.NoError(t, err)
	var count int
	_ = fs.WalkDir(migrations.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".sql" {
			count++
		}
		return nil
	})
	require.Equal(t, count, len(versions))
}
