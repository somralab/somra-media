package bootstrap_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestWireStreaming(t *testing.T) {
	dbCfg := db.Default()
	dbCfg.DataDir = filepath.Join(t.TempDir(), "data")

	c, err := bootstrap.NewWithStorage(context.Background(), nil, dbCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	cfg := config.Default()
	cfg.Data.CacheDir = filepath.Join(t.TempDir(), "cache")
	cfg.Streaming.SessionTTL = time.Hour
	cfg.Streaming.IdleTimeout = 5 * time.Minute

	bundle := bootstrap.WireStreaming(c, cfg, nil)
	require.NotNil(t, bundle)
	require.NotNil(t, bundle.Service)
}

func TestWireStreaming_NilComponents(t *testing.T) {
	require.Nil(t, bootstrap.WireStreaming(nil, config.Default(), nil))
}

func TestWireStreaming_NilDB(t *testing.T) {
	c, err := bootstrap.New(nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })
	require.Nil(t, bootstrap.WireStreaming(c, config.Default(), nil))
}
