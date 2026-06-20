package stub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/plugin"
)

func TestStubIndexerFactory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := NewIndexerFactory()
	assert.Equal(t, Implementation, f.Implementation())
	assert.Equal(t, plugin.PluginTypeIndexer, f.Type())

	p, err := f.New(ctx, "7", []byte(`{"prefix":"demo"}`))
	require.NoError(t, err)
	idx, ok := p.(plugin.Indexer)
	require.True(t, ok)

	caps, err := idx.Capabilities(ctx)
	require.NoError(t, err)
	assert.True(t, caps.SupportsSearch)

	results, err := idx.Search(ctx, plugin.SearchQuery{Title: "Movie", MediaKind: plugin.MediaKindMovie})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "demo:Movie", results[0].Title)
}

func TestStubDownloadClientFactory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := NewDownloadClientFactory()
	p, err := f.New(ctx, "3", []byte(`{"prefix":"/downloads"}`))
	require.NoError(t, err)
	client, ok := p.(plugin.DownloadClient)
	require.True(t, ok)

	item, err := client.Add(ctx, plugin.AddRequest{ReleaseID: "rel-1"})
	require.NoError(t, err)
	assert.Equal(t, "/downloads/rel-1", item.SavePath)
}

func TestStubDownloadClientFactory_StatusNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := NewDownloadClientFactory()
	p, err := f.New(ctx, "3", nil)
	require.NoError(t, err)
	client, ok := p.(plugin.DownloadClient)
	require.True(t, ok)

	_, err = client.Status(ctx, "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, plugin.ErrUnsupportedCapability)
}

func TestStubIndexerFactory_InvalidConfig(t *testing.T) {
	t.Parallel()
	f := NewIndexerFactory()
	_, err := f.New(context.Background(), "1", []byte(`not-json`))
	require.Error(t, err)
}

func TestStubDownloadClientFactory_InvalidConfig(t *testing.T) {
	t.Parallel()
	f := NewDownloadClientFactory()
	_, err := f.New(context.Background(), "1", []byte(`not-json`))
	require.Error(t, err)
}

var _ plugin.Indexer = (*stubIndexer)(nil)
var _ plugin.DownloadClient = (*stubDownloadClient)(nil)
