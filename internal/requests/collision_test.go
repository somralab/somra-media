package requests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

type stubLibraryLookup struct {
	found  bool
	itemID int64
	err    error
}

func (s stubLibraryLookup) ExistsByProviderID(_ context.Context, _, _ string) (bool, int64, error) {
	return s.found, s.itemID, s.err
}

type stubRequestLookup struct {
	found bool
	reqID int64
	err   error
}

func (s stubRequestLookup) HasPendingByProviderID(_ context.Context, _, _ string) (bool, int64, error) {
	return s.found, s.reqID, s.err
}

func TestCollisionChecker_Check(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("no collision", func(t *testing.T) {
		c := &CollisionChecker{
			Library:  stubLibraryLookup{},
			Requests: stubRequestLookup{},
		}
		out, err := c.Check(ctx, "tmdb", "123")
		require.NoError(t, err)
		assert.False(t, out.Blocked)
	})

	t.Run("in library", func(t *testing.T) {
		c := &CollisionChecker{
			Library: stubLibraryLookup{found: true, itemID: 42},
		}
		out, err := c.Check(ctx, "tmdb", "123")
		require.NoError(t, err)
		assert.True(t, out.Blocked)
		assert.True(t, out.InLibrary)
		assert.Equal(t, int64(42), out.LibraryMediaItemID)
	})

	t.Run("duplicate pending", func(t *testing.T) {
		c := &CollisionChecker{
			Requests: stubRequestLookup{found: true, reqID: 7},
		}
		out, err := c.Check(ctx, "tmdb", "123")
		require.NoError(t, err)
		assert.True(t, out.Blocked)
		assert.True(t, out.DuplicatePending)
		assert.Equal(t, int64(7), out.ExistingRequestID)
	})

	t.Run("library lookup error", func(t *testing.T) {
		c := &CollisionChecker{Library: stubLibraryLookup{err: assert.AnError}}
		_, err := c.Check(ctx, "tmdb", "123")
		require.Error(t, err)
	})

	t.Run("pending lookup error", func(t *testing.T) {
		c := &CollisionChecker{Requests: stubRequestLookup{err: assert.AnError}}
		_, err := c.Check(ctx, "tmdb", "123")
		require.Error(t, err)
	})
}

func TestCollisionChecker_ValidateCreation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c := &CollisionChecker{Library: stubLibraryLookup{found: true}}
	err := c.ValidateCreation(ctx, "tmdb", "99")
	require.ErrorIs(t, err, ErrCollisionInLibrary)

	c = &CollisionChecker{Requests: stubRequestLookup{found: true}}
	err = c.ValidateCreation(ctx, "tmdb", "99")
	require.ErrorIs(t, err, ErrCollisionDuplicatePending)

	c = &CollisionChecker{}
	require.NoError(t, c.ValidateCreation(ctx, "tmdb", "99"))
}

func TestDBLibraryLookup_ExistsByProviderID(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Movies", db.LibraryKindMovie, []string{dir}, true)
	require.NoError(t, err)

	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Inception", intPtr(2010))
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetProviderID(ctx, itemID, "tmdb", "27205"))

	lookup := &DBLibraryLookup{Q: d.Querier()}
	found, gotID, err := lookup.ExistsByProviderID(ctx, "tmdb", "27205")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, itemID, gotID)

	found, _, err = lookup.ExistsByProviderID(ctx, "tmdb", "99999")
	require.NoError(t, err)
	assert.False(t, found)
}

func intPtr(v int) *int { return &v }

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	cfg := db.Default()
	cfg.DataDir = t.TempDir()
	ctx := context.Background()
	d, err := db.Initialize(ctx, cfg, nil)
	require.NoError(t, err)
	return d
}
