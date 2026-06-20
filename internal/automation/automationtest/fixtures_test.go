package automationtest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/plugin/stub"
)

func TestOpenDBAndManager(t *testing.T) {
	d := OpenDB(t)
	mgr := NewManager(t, d)
	idxID := CreateStubIndexer(t, mgr, "fixture-idx")
	dlID := CreateStubDownloadClient(t, mgr, "fixture-dl")
	reqID := CreateApprovedRequest(t, d, "Fixture Movie")
	require.Positive(t, idxID)
	require.Positive(t, dlID)
	require.Positive(t, reqID)
}

func TestPluginStoreLifecycle(t *testing.T) {
	ctx := context.Background()
	d := OpenDB(t)
	mgr := NewManager(t, d)
	id := CreateStubIndexer(t, mgr, "lifecycle")

	require.NoError(t, mgr.UpdateName(ctx, id, "renamed"))
	require.NoError(t, mgr.Disable(ctx, id))
	require.NoError(t, mgr.Enable(ctx, id))
	require.NoError(t, mgr.Configure(ctx, id, []byte(`{"prefix":"x"}`)))

	rec, err := mgr.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "renamed", rec.Name)

	all, err := mgr.List(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, all)

	require.NoError(t, mgr.Delete(ctx, id))
	_, err = mgr.Get(ctx, id)
	require.Error(t, err)

	_, err = mgr.Create(ctx, plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeIndexer,
		Implementation: stub.Implementation,
		Name:           "dup-name",
		Config:         []byte("{}"),
		Enabled:        true,
	})
	require.NoError(t, err)
	_, err = mgr.Create(ctx, plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeIndexer,
		Implementation: stub.Implementation,
		Name:           "dup-name",
		Config:         []byte("{}"),
		Enabled:        true,
	})
	require.Error(t, err)
}
