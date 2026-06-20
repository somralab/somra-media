package bootstrap

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/plugin/stub"
)

func testComponents(t *testing.T) *Components {
	t.Helper()
	dbCfg := db.Default()
	dbCfg.DataDir = filepath.Join(t.TempDir(), "data")
	c, err := NewWithStorage(context.Background(), nil, dbCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestWirePlugins_EmptyDatabase(t *testing.T) {
	c := testComponents(t)

	bundle, err := WirePlugins(c)
	require.NoError(t, err)
	require.NotNil(t, bundle)
	require.NotNil(t, bundle.Manager)
	assert.Len(t, bundle.Manager.EnabledIndexers(), 0)
	assert.Len(t, bundle.Manager.EnabledDownloadClients(), 0)
}

func TestWirePlugins_HydratesEnabledInstance(t *testing.T) {
	ctx := context.Background()
	c := testComponents(t)

	repo := db.NewPluginInstanceRepo(c.DB.Querier())
	id, err := repo.Create(ctx, db.PluginInstance{
		PluginType:     db.PluginInstanceTypeIndexer,
		Implementation: stub.Implementation,
		Name:           "boot-indexer",
		Config:         `{"prefix":"boot"}`,
		Enabled:        true,
	})
	require.NoError(t, err)

	bundle, err := WirePlugins(c)
	require.NoError(t, err)
	require.Len(t, bundle.Manager.EnabledIndexers(), 1)

	idx, err := bundle.Manager.Indexer(ctx, id)
	require.NoError(t, err)
	results, err := idx.Search(ctx, plugin.SearchQuery{Title: "Film", MediaKind: plugin.MediaKindMovie})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "boot:Film", results[0].Title)
}

func TestWirePlugins_SkipsBrokenEnabledInstance(t *testing.T) {
	ctx := context.Background()
	c := testComponents(t)

	repo := db.NewPluginInstanceRepo(c.DB.Querier())
	_, err := repo.Create(ctx, db.PluginInstance{
		PluginType:     db.PluginInstanceTypeIndexer,
		Implementation: "unknown-impl",
		Name:           "broken-indexer",
		Enabled:        true,
	})
	require.NoError(t, err)

	bundle, err := WirePlugins(c)
	require.NoError(t, err)
	assert.Len(t, bundle.Manager.EnabledIndexers(), 0)
}

func TestWirePlugins_RequiresDB(t *testing.T) {
	c, err := New(nil)
	require.NoError(t, err)
	_, err = WirePlugins(c)
	require.Error(t, err)
}

func TestPluginStore_CRUD(t *testing.T) {
	ctx := context.Background()
	c := testComponents(t)

	store := newPluginStore(db.NewPluginInstanceRepo(c.DB.Querier()))
	id, err := store.Create(ctx, plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeIndexer,
		Implementation: stub.Implementation,
		Name:           "store-indexer",
		Config:         json.RawMessage(`{"prefix":"db"}`),
		Enabled:        true,
	})
	require.NoError(t, err)

	rec, err := store.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "store-indexer", rec.Name)
	assert.True(t, rec.Enabled)

	require.NoError(t, store.UpdateConfig(ctx, id, json.RawMessage(`{"prefix":"new"}`)))
	rec, err = store.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Contains(t, string(rec.Config), "new")

	require.NoError(t, store.SetEnabled(ctx, id, false))
	rec, err = store.GetByID(ctx, id)
	require.NoError(t, err)
	assert.False(t, rec.Enabled)

	_, err = store.GetByID(ctx, 9999)
	require.Error(t, err)
	assert.ErrorIs(t, err, plugin.ErrPluginNotFound)
}

func TestPluginStore_CreateDuplicateMapsError(t *testing.T) {
	ctx := context.Background()
	c := testComponents(t)

	store := newPluginStore(db.NewPluginInstanceRepo(c.DB.Querier()))
	_, err := store.Create(ctx, plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeIndexer,
		Implementation: stub.Implementation,
		Name:           "dup",
		Config:         json.RawMessage(`{}`),
	})
	require.NoError(t, err)

	_, err = store.Create(ctx, plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeIndexer,
		Implementation: stub.Implementation,
		Name:           "dup",
		Config:         json.RawMessage(`{}`),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, plugin.ErrDuplicateInstance)
}
