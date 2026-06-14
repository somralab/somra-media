package bootstrap_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestNewWiresDefaults(t *testing.T) {
	t.Parallel()
	c, err := bootstrap.New(nil)
	require.NoError(t, err)
	defer func() { _ = c.Close() }()
	require.NotNil(t, c.Scheduler)
	require.NotNil(t, c.I18n)
	require.NotNil(t, c.Diagnostics)
	require.NotNil(t, c.Heartbeat)
	require.NotNil(t, c.Uptime)

	snap := c.Diagnostics.Run(context.Background())
	require.GreaterOrEqual(t, len(snap.Checks), 2)
}

func TestNewWithStorageRegistersDBProvider(t *testing.T) {
	t.Parallel()
	dbCfg := db.Default()
	dbCfg.DataDir = filepath.Join(t.TempDir(), "data")

	c, err := bootstrap.NewWithStorage(context.Background(), nil, dbCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	require.NotNil(t, c.DB)
	snap := c.Diagnostics.Run(context.Background())
	names := make([]string, 0, len(snap.Checks))
	for _, ch := range snap.Checks {
		names = append(names, ch.Name)
	}
	assert.Contains(t, names, "database", "DB provider must be registered")
}
