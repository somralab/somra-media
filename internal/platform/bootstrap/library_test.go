package bootstrap_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestWireLibrary_RegistersServices(t *testing.T) {
	dbCfg := db.Default()
	dbCfg.DataDir = filepath.Join(t.TempDir(), "data")

	c, err := bootstrap.NewWithStorage(context.Background(), nil, dbCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	t.Setenv("SOMRA_TMDB_API_KEY", "test-key")
	bundle := bootstrap.WireLibrary(c)
	require.NotNil(t, bundle)
	require.NotNil(t, bundle.EventBus)
	require.NotNil(t, bundle.Library)
	require.NotNil(t, bundle.Metadata)

	ctx := context.Background()
	c.Scheduler.Start(ctx)
	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = c.Scheduler.Stop(stopCtx)
	})

	dir := t.TempDir()
	lib, err := bundle.Library.CreateLibrary(ctx, "Demo", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	_, err = bundle.Metadata.AutoMatch(ctx, lib.ID, "en-US", 10)
	require.NoError(t, err)
}

func TestNewWithStorage_InvalidDataDir(t *testing.T) {
	t.Parallel()
	cfg := db.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "blocked")
	require.NoError(t, os.WriteFile(cfg.DataDir, []byte("x"), 0o644))

	_, err := bootstrap.NewWithStorage(context.Background(), nil, cfg)
	require.Error(t, err)
}
