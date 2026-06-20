package plugin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testIndexerFactory struct{}

func (testIndexerFactory) Implementation() string { return "test-indexer" }
func (testIndexerFactory) Type() PluginType       { return PluginTypeIndexer }
func (testIndexerFactory) New(_ context.Context, instanceID string, config []byte) (Plugin, error) {
	prefix := "test"
	if len(config) > 0 {
		var cfg struct {
			Prefix string `json:"prefix"`
		}
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, err
		}
		if cfg.Prefix != "" {
			prefix = cfg.Prefix
		}
	}
	return &testIndexer{id: instanceID, prefix: prefix}, nil
}

type testIndexer struct {
	id     string
	prefix string
}

func (t *testIndexer) ID() string              { return t.id }
func (t *testIndexer) Type() PluginType        { return PluginTypeIndexer }
func (t *testIndexer) ContractVersion() string { return ContractVersion }

func (t *testIndexer) Capabilities(_ context.Context) (Capabilities, error) {
	return Capabilities{SupportsSearch: true, Protocols: []Protocol{ProtocolTorrent}}, nil
}

func (t *testIndexer) Search(_ context.Context, q SearchQuery) ([]SearchResult, error) {
	title := q.Title
	if t.prefix != "" {
		title = t.prefix + ":" + title
	}
	return []SearchResult{{
		ReleaseID: t.id + "-release",
		IndexerID: t.id,
		Title:     title,
		Protocol:  ProtocolTorrent,
		SizeBytes: 1,
	}}, nil
}

type testDownloadClientFactory struct{}

func (testDownloadClientFactory) Implementation() string { return "test-client" }
func (testDownloadClientFactory) Type() PluginType       { return PluginTypeDownloadClient }
func (testDownloadClientFactory) New(_ context.Context, instanceID string, _ []byte) (Plugin, error) {
	return &testDownloadClient{id: instanceID, items: make(map[string]DownloadItem)}, nil
}

type testDownloadClient struct {
	id    string
	items map[string]DownloadItem
}

func (t *testDownloadClient) ID() string              { return t.id }
func (t *testDownloadClient) Type() PluginType        { return PluginTypeDownloadClient }
func (t *testDownloadClient) ContractVersion() string { return ContractVersion }

func (t *testDownloadClient) Add(_ context.Context, req AddRequest) (DownloadItem, error) {
	downloadID := t.id + "-dl"
	item := DownloadItem{
		DownloadID: downloadID,
		ClientID:   t.id,
		ReleaseID:  req.ReleaseID,
		Status:     DownloadStatusQueued,
	}
	t.items[downloadID] = item
	return item, nil
}

func (t *testDownloadClient) Status(_ context.Context, downloadID string) (DownloadItem, error) {
	item, ok := t.items[downloadID]
	if !ok {
		return DownloadItem{}, ErrUnsupportedCapability
	}
	return item, nil
}

func TestManager_RegisterFactoryDuplicate(t *testing.T) {
	t.Parallel()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	require.NoError(t, mgr.RegisterFactory(testIndexerFactory{}))
	err := mgr.RegisterFactory(testIndexerFactory{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDuplicateFactory)
}

func TestManager_CreateEnableDisableLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	require.NoError(t, mgr.RegisterFactory(testIndexerFactory{}))

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "test-indexer",
		Name:           "primary-indexer",
		Config:         json.RawMessage(`{"prefix":"test"}`),
	})
	require.NoError(t, err)
	require.Greater(t, id, int64(0))

	_, err = mgr.Indexer(ctx, id)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPluginDisabled)

	require.NoError(t, mgr.Enable(ctx, id))
	idx, err := mgr.Indexer(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, idx)
	assert.Equal(t, "1", idx.ID())

	results, err := idx.Search(ctx, SearchQuery{Title: "Movie", MediaKind: MediaKindMovie})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "test:Movie", results[0].Title)

	require.NoError(t, mgr.Disable(ctx, id))
	_, err = mgr.Indexer(ctx, id)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPluginDisabled)
	assert.Len(t, mgr.EnabledIndexers(), 0)

	rec, err := mgr.Get(ctx, id)
	require.NoError(t, err)
	assert.False(t, rec.Enabled)
}

func TestManager_ConfigureRebuildsEnabledInstance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	require.NoError(t, mgr.RegisterFactory(testIndexerFactory{}))

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "test-indexer",
		Name:           "rebuild-indexer",
		Enabled:        true,
		Config:         json.RawMessage(`{"prefix":"v1"}`),
	})
	require.NoError(t, err)

	idx, err := mgr.Indexer(ctx, id)
	require.NoError(t, err)
	results, err := idx.Search(ctx, SearchQuery{Title: "Show", MediaKind: MediaKindTV})
	require.NoError(t, err)
	assert.Equal(t, "v1:Show", results[0].Title)

	require.NoError(t, mgr.Configure(ctx, id, json.RawMessage(`{"prefix":"v2"}`)))
	idx, err = mgr.Indexer(ctx, id)
	require.NoError(t, err)
	results, err = idx.Search(ctx, SearchQuery{Title: "Show", MediaKind: MediaKindTV})
	require.NoError(t, err)
	assert.Equal(t, "v2:Show", results[0].Title)
}

func TestManager_CreateEnabledOnInsert(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	require.NoError(t, mgr.RegisterFactory(testDownloadClientFactory{}))

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeDownloadClient,
		Implementation: "test-client",
		Name:           "client-a",
		Enabled:        true,
	})
	require.NoError(t, err)

	client, err := mgr.DownloadClient(ctx, id)
	require.NoError(t, err)
	item, err := client.Add(ctx, AddRequest{ReleaseID: "rel-1"})
	require.NoError(t, err)
	assert.Equal(t, DownloadStatusQueued, item.Status)
	assert.Len(t, mgr.EnabledDownloadClients(), 1)
}

func TestManager_LoadEnabled(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newMemoryStore()
	mgr := NewManager(store, ManagerOptions{})
	require.NoError(t, mgr.RegisterFactory(testIndexerFactory{}))

	_, err := store.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "test-indexer",
		Name:           "boot-indexer",
		Enabled:        true,
	})
	require.NoError(t, err)

	require.NoError(t, mgr.LoadEnabled(ctx))
	assert.Len(t, mgr.EnabledIndexers(), 1)
}

func TestManager_CreateUnknownFactory(t *testing.T) {
	t.Parallel()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	_, err := mgr.Create(context.Background(), InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "missing",
		Name:           "x",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrFactoryNotFound)
}

type emptyFactory struct{}

func (emptyFactory) Implementation() string { return "" }
func (emptyFactory) Type() PluginType       { return PluginTypeIndexer }
func (emptyFactory) New(context.Context, string, []byte) (Plugin, error) {
	return nil, nil
}

type badContractPlugin struct {
	id string
}

func (badContractPlugin) ID() string              { return "bad" }
func (badContractPlugin) Type() PluginType        { return PluginTypeIndexer }
func (badContractPlugin) ContractVersion() string { return "0" }

type badFactory struct{}

func (badFactory) Implementation() string { return "bad" }
func (badFactory) Type() PluginType       { return PluginTypeIndexer }
func (badFactory) New(_ context.Context, instanceID string, _ []byte) (Plugin, error) {
	return badContractPlugin{id: instanceID}, nil
}

func TestManager_EnableRejectsIncompatibleContract(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	require.NoError(t, mgr.RegisterFactory(badFactory{}))

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "bad",
		Name:           "bad-indexer",
	})
	require.NoError(t, err)

	err = mgr.Enable(ctx, id)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIncompatibleContract)
}

func TestManager_ListAndGetNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	require.NoError(t, mgr.RegisterFactory(testIndexerFactory{}))

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "test-indexer",
		Name:           "listed",
	})
	require.NoError(t, err)

	rows, err := mgr.List(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, id, rows[0].ID)

	_, err = mgr.Get(ctx, 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPluginNotFound)
}

func TestManager_IndexerWrongType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	require.NoError(t, mgr.RegisterFactory(testDownloadClientFactory{}))

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeDownloadClient,
		Implementation: "test-client",
		Name:           "client-only",
		Enabled:        true,
	})
	require.NoError(t, err)

	_, err = mgr.Indexer(ctx, id)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedCapability)
}

func TestManager_DownloadClientWrongType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	require.NoError(t, mgr.RegisterFactory(testIndexerFactory{}))

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "test-indexer",
		Name:           "indexer-only",
		Enabled:        true,
	})
	require.NoError(t, err)

	_, err = mgr.DownloadClient(ctx, id)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedCapability)
}

func TestManager_RegisterFactoryValidation(t *testing.T) {
	t.Parallel()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	require.Error(t, mgr.RegisterFactory(nil))
	require.Error(t, mgr.RegisterFactory(emptyFactory{}))
}

func TestManager_NilStoreGuards(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(nil, ManagerOptions{})
	_, err := mgr.Create(ctx, InstanceRecord{})
	require.Error(t, err)
	require.Error(t, mgr.Enable(ctx, 1))
	require.Error(t, mgr.Disable(ctx, 1))
	_, err = mgr.Get(ctx, 1)
	require.Error(t, err)
	_, err = mgr.List(ctx)
	require.Error(t, err)
}

func TestManager_ConfigureWithSecrets(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{EncryptionKey: "test-key"})
	require.NoError(t, mgr.RegisterFactory(testIndexerFactory{}))

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "test-indexer",
		Name:           "secret-indexer",
		Enabled:        true,
		Config:         json.RawMessage(`{"prefix":"p","apiKey":"secret-value"}`),
	})
	require.NoError(t, err)

	rec, err := mgr.Get(ctx, id)
	require.NoError(t, err)
	assert.NotContains(t, string(rec.Config), "secret-value")
	assert.NotEmpty(t, rec.SecretsEnc)

	idx, err := mgr.Indexer(ctx, id)
	require.NoError(t, err)
	results, err := idx.Search(ctx, SearchQuery{Title: "X", MediaKind: MediaKindMovie})
	require.NoError(t, err)
	assert.Equal(t, "p:X", results[0].Title)

	pub, err := mgr.PublicConfig(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, true, pub["apiKeySet"])
	assert.NotContains(t, pub, "apiKey")

	require.NoError(t, mgr.PatchConfig(ctx, id, json.RawMessage(`{"apiKey":"rotated-key"}`)))
	pub, err = mgr.PublicConfig(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, true, pub["apiKeySet"])
}

func TestManager_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{EncryptionKey: "k"})
	require.NoError(t, mgr.RegisterFactory(testIndexerFactory{}))

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "test-indexer",
		Name:           "to-delete",
		Enabled:        true,
	})
	require.NoError(t, err)
	require.Len(t, mgr.EnabledIndexers(), 1)

	require.NoError(t, mgr.Delete(ctx, id))
	assert.Len(t, mgr.EnabledIndexers(), 0)
	_, err = mgr.Get(ctx, id)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPluginNotFound)
}

func TestManager_CatalogAndTest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{EncryptionKey: "k"})
	require.NoError(t, mgr.RegisterFactory(testIndexerFactory{}))
	require.NoError(t, mgr.RegisterFactory(testDownloadClientFactory{}))

	catalog := mgr.Catalog()
	require.Len(t, catalog, 2)

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "test-indexer",
		Name:           "testable",
	})
	require.NoError(t, err)

	result, err := mgr.Test(ctx, id)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "plugins.instances.test.success", result.MessageKey)
}

func TestManager_CreateActivateFailureLeavesPersisted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := NewManager(newMemoryStore(), ManagerOptions{})
	require.NoError(t, mgr.RegisterFactory(badFactory{}))

	id, err := mgr.Create(ctx, InstanceRecord{
		PluginType:     PluginTypeIndexer,
		Implementation: "bad",
		Name:           "bad-enabled",
		Enabled:        true,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIncompatibleContract)
	assert.Greater(t, id, int64(0))
}
