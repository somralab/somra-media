package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginInstanceRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewPluginInstanceRepo(d.Querier())

	id, err := repo.Create(ctx, PluginInstance{
		PluginType:     PluginInstanceTypeIndexer,
		Implementation: "stub",
		Name:           "indexer-a",
		Config:         `{"prefix":"demo"}`,
		Enabled:        true,
	})
	require.NoError(t, err)
	require.Greater(t, id, int64(0))

	inst, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, PluginInstanceTypeIndexer, inst.PluginType)
	assert.Equal(t, "stub", inst.Implementation)
	assert.True(t, inst.Enabled)
	assert.False(t, inst.CreatedAt.IsZero())

	rows, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	require.NoError(t, repo.UpdateConfig(ctx, id, `{"prefix":"updated"}`))
	inst, err = repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Contains(t, inst.Config, "updated")

	require.NoError(t, repo.SetEnabled(ctx, id, false))
	inst, err = repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.False(t, inst.Enabled)
}

func TestPluginInstanceRepo_DuplicateName(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewPluginInstanceRepo(d.Querier())
	_, err := repo.Create(ctx, PluginInstance{
		PluginType:     PluginInstanceTypeIndexer,
		Implementation: "stub",
		Name:           "same-name",
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, PluginInstance{
		PluginType:     PluginInstanceTypeIndexer,
		Implementation: "stub",
		Name:           "same-name",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPluginInstanceDuplicate)
}

func TestPluginInstanceRepo_NotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewPluginInstanceRepo(d.Querier())
	_, err := repo.GetByID(ctx, 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPluginInstanceNotFound)

	require.Error(t, repo.SetEnabled(ctx, 999, false))
	require.Error(t, repo.UpdateConfig(ctx, 999, `{}`))
}
