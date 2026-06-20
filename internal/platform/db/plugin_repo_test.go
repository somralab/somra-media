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

func TestPluginInstanceRepo_CreateValidation(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewPluginInstanceRepo(d.Querier())
	_, err := repo.Create(ctx, PluginInstance{})
	require.Error(t, err)

	_, err = repo.Create(ctx, PluginInstance{PluginType: PluginInstanceTypeIndexer})
	require.Error(t, err)
}

func TestPluginInstanceRepo_CreateDefaultsEmptyConfig(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewPluginInstanceRepo(d.Querier())
	id, err := repo.Create(ctx, PluginInstance{
		PluginType:     PluginInstanceTypeDownloadClient,
		Implementation: "stub",
		Name:           "client-default-config",
		Config:         "",
	})
	require.NoError(t, err)

	inst, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "{}", inst.Config)
}

func TestPluginInstanceRepo_UpdateConfigEmpty(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewPluginInstanceRepo(d.Querier())
	id, err := repo.Create(ctx, PluginInstance{
		PluginType:     PluginInstanceTypeIndexer,
		Implementation: "stub",
		Name:           "config-empty",
		Config:         `{"prefix":"keep"}`,
	})
	require.NoError(t, err)

	require.NoError(t, repo.UpdateConfig(ctx, id, ""))
	inst, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "{}", inst.Config)
}

func TestPluginInstanceRepo_SameNameDifferentType(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewPluginInstanceRepo(d.Querier())
	_, err := repo.Create(ctx, PluginInstance{
		PluginType:     PluginInstanceTypeIndexer,
		Implementation: "stub",
		Name:           "shared-name",
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, PluginInstance{
		PluginType:     PluginInstanceTypeDownloadClient,
		Implementation: "stub",
		Name:           "shared-name",
	})
	require.NoError(t, err)

	rows, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, rows, 2)
}

func TestPluginInstanceRepo_SetEnabledToggle(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewPluginInstanceRepo(d.Querier())
	id, err := repo.Create(ctx, PluginInstance{
		PluginType:     PluginInstanceTypeIndexer,
		Implementation: "stub",
		Name:           "toggle-indexer",
		Enabled:        true,
	})
	require.NoError(t, err)

	require.NoError(t, repo.SetEnabled(ctx, id, false))
	inst, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.False(t, inst.Enabled)

	require.NoError(t, repo.SetEnabled(ctx, id, true))
	inst, err = repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.True(t, inst.Enabled)
}

func TestPluginInstanceRepo_ListEmpty(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewPluginInstanceRepo(d.Querier())
	rows, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestPluginInstanceRepo_CreateOnClosedDB(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	require.NoError(t, d.Close())

	repo := NewPluginInstanceRepo(d.Querier())
	_, err := repo.Create(ctx, PluginInstance{
		PluginType:     PluginInstanceTypeIndexer,
		Implementation: "stub",
		Name:           "closed-db",
	})
	require.Error(t, err)

	_, err = repo.GetByID(ctx, 1)
	require.Error(t, err)

	_, err = repo.List(ctx)
	require.Error(t, err)

	require.Error(t, repo.UpdateConfig(ctx, 1, `{}`))
	require.Error(t, repo.SetEnabled(ctx, 1, false))
}
