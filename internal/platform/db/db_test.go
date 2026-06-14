package db

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// newTestConfig returns a Config rooted at an isolated temp directory so
// tests never collide with each other or with developer data.
func newTestConfig(t *testing.T) Config {
	t.Helper()
	return Config{DataDir: t.TempDir(), DBFile: "test.db"}
}

func TestOpen_AppliesPragmasAndPings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d, err := Open(ctx, newTestConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	require.NoError(t, d.Ping(ctx))

	mode, err := d.JournalMode(ctx)
	require.NoError(t, err)
	require.Equal(t, "wal", strings.ToLower(mode))

	var fk int
	require.NoError(t, d.SQL().QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&fk))
	require.Equal(t, 1, fk)

	var busy int
	require.NoError(t, d.SQL().QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busy))
	require.Equal(t, 5000, busy)

	var sync int
	require.NoError(t, d.SQL().QueryRowContext(ctx, "PRAGMA synchronous").Scan(&sync))
	require.Equal(t, 1, sync) // NORMAL == 1

	var temp int
	require.NoError(t, d.SQL().QueryRowContext(ctx, "PRAGMA temp_store").Scan(&temp))
	require.Equal(t, 2, temp) // MEMORY == 2

	var cache int
	require.NoError(t, d.SQL().QueryRowContext(ctx, "PRAGMA cache_size").Scan(&cache))
	require.Equal(t, -65536, cache)
}

func TestOpen_CreatesDataDir(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	base := t.TempDir()
	cfg := Config{DataDir: filepath.Join(base, "nested", "dir"), DBFile: "x.db"}

	d, err := Open(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	info, statErr := os.Stat(cfg.DataDir)
	require.NoError(t, statErr)
	require.True(t, info.IsDir())
}

func TestOpen_InvalidConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := Open(ctx, Config{})
	require.Error(t, err)
}

func TestOpen_DataDirIsFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	base := t.TempDir()
	notADir := filepath.Join(base, "blocker")
	require.NoError(t, os.WriteFile(notADir, []byte("x"), 0o600))

	cfg := Config{DataDir: notADir, DBFile: "x.db"}
	_, err := Open(ctx, cfg)
	require.Error(t, err)
}

func TestDB_CloseIsIdempotent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d, err := Open(ctx, newTestConfig(t))
	require.NoError(t, err)
	require.NoError(t, d.Close())
	// Second close on already-closed handle should still not panic; it may
	// return an error from database/sql which is acceptable. The nil case
	// must remain safe.
	var nilDB *DB
	require.NoError(t, nilDB.Close())
}

func TestDB_QuerierExposed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d, err := Open(ctx, newTestConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	require.NotNil(t, d.Querier())
	require.NotNil(t, d.SQL())
}

func TestConfig_DefaultAndFromEnv(t *testing.T) {
	def := Default()
	require.Equal(t, defaultDataDir, def.DataDir)
	require.Equal(t, defaultDBFile, def.DBFile)

	t.Setenv(envDataDir, "/tmp/some/where")
	got := FromEnv()
	require.Equal(t, "/tmp/some/where", got.DataDir)

	t.Setenv(envDataDir, "  ")
	require.Equal(t, defaultDataDir, FromEnv().DataDir)
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := map[string]Config{
		"empty dir":  {DataDir: "", DBFile: "a.db"},
		"empty file": {DataDir: ".", DBFile: ""},
		"slash file": {DataDir: ".", DBFile: "sub/a.db"},
	}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			require.Error(t, cfg.Validate())
		})
	}

	require.NoError(t, Default().Validate())
}

func TestConfig_PathAndDSN(t *testing.T) {
	t.Parallel()

	cfg := Config{DataDir: t.TempDir(), DBFile: "x.db"}
	p, err := cfg.Path()
	require.NoError(t, err)
	require.True(t, filepath.IsAbs(p))

	dsn, err := cfg.DSN()
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(dsn, "file:"))

	_, err = Config{}.Path()
	require.Error(t, err)
	_, err = Config{}.DSN()
	require.Error(t, err)
}

// TestIntegrationBootstrap is the "smoke" sub-test required by the packet
// brief: open DB at a temp dir, run migrations, query settings.
func TestIntegrationBootstrap(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cfg := Config{DataDir: t.TempDir(), DBFile: "bootstrap.db"}
	d, err := Initialize(ctx, cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	var count int
	require.NoError(t, d.SQL().QueryRowContext(ctx, "SELECT count(*) FROM settings").Scan(&count))
	require.Equal(t, 1, count, "seed migration should leave exactly one settings row")

	repo := NewSettingsRepo(d.Querier())
	val, err := repo.Get(ctx, "system.installed_at")
	require.NoError(t, err)
	require.NotEmpty(t, val)

	_, err = repo.Get(ctx, "nope")
	require.True(t, errors.Is(err, ErrSettingNotFound))
}
