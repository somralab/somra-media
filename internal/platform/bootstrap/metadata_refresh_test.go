package bootstrap

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestMetadataRefreshJob(t *testing.T) {
	dbCfg := db.Default()
	dbCfg.DataDir = filepath.Join(t.TempDir(), "data")
	c, err := NewWithStorage(context.Background(), nil, dbCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	t.Setenv("SOMRA_TMDB_API_KEY", "test-key")
	bundle := WireLibrary(c)
	ctx := context.Background()
	dir := t.TempDir()
	_, err = bundle.Library.CreateLibrary(ctx, "Demo", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	job := metadataRefreshJob(c.Logger, bundle.Library, bundle.Metadata)
	require.NoError(t, job.Run(ctx))
}
