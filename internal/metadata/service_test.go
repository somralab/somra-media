package metadata

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestService_AutoMatchAndRematch(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)

	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Movies", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Inception", ptrInt(2010))
	require.NoError(t, err)
	_, err = mediaRepo.UpsertFile(ctx, db.MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID,
		Path: "/tmp/inception.mkv", FileName: "inception.mkv", ParsedTitle: "Inception",
	})
	require.NoError(t, err)

	reg := NewRegistry()
	reg.Register(TestProvider{})
	svc := &Service{
		DB:       &DBStore{DB: d},
		Registry: reg,
		Matcher:  &Matcher{Registry: reg},
	}

	n, err := svc.AutoMatch(ctx, lib.ID, "en-US", 10)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	item, err := mediaRepo.GetItemByID(ctx, itemID, "en-US")
	require.NoError(t, err)
	assert.Equal(t, db.MatchManual, item.MatchStatus)

	require.NoError(t, svc.Rematch(ctx, itemID, "tmdb", "99", "tr-TR"))
}

func TestRateLimiter_Serializes(t *testing.T) {
	l := NewRateLimiter(50)
	ctx := context.Background()
	require.NoError(t, l.Wait(ctx, "tmdb"))
	require.NoError(t, l.Wait(ctx, "tmdb"))
}

func TestResponseCache_GetSet(t *testing.T) {
	c := NewResponseCache(0)
	c.Set("k", []byte("v"))
	got, ok := c.Get("k")
	assert.True(t, ok)
	assert.Equal(t, []byte("v"), got)
}

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	cfg := db.Default()
	cfg.DataDir = t.TempDir()
	d, err := db.Initialize(context.Background(), cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func ptrInt(v int) *int { return &v }
