package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepos_ErrorsAfterDBClose(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	libRepo := NewLibraryRepo(d.Querier())
	scanRepo := NewScanRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	require.NoError(t, d.Close())

	_, err := libRepo.List(ctx)
	require.Error(t, err)

	_, err = libRepo.GetByID(ctx, 1)
	require.Error(t, err)

	_, err = libRepo.Create(ctx, "x", LibraryKindMovie, []string{"/tmp"}, false)
	require.Error(t, err)

	_, err = scanRepo.CreateRun(ctx, 1, ScanFull, "x")
	require.Error(t, err)

	_, err = scanRepo.GetByID(ctx, 1)
	require.Error(t, err)

	_, err = mediaRepo.CreateItem(ctx, 1, LibraryKindMovie, "x", nil)
	require.Error(t, err)

	_, err = mediaRepo.UpsertFile(ctx, MediaFile{LibraryID: 1, Path: "/x", FileName: "x"})
	require.Error(t, err)

	require.Error(t, mediaRepo.SetMatch(ctx, 1, MatchUnmatched, nil))
	require.Error(t, mediaRepo.SetLocalizedText(ctx, 1, "en-US", "title", "x"))
	require.Error(t, mediaRepo.SetProviderID(ctx, 1, "tmdb", "1"))
	require.Error(t, mediaRepo.UpsertArtwork(ctx, 1, "poster", "http://x", ""))
	require.Error(t, mediaRepo.UpsertTechnical(ctx, 1, 0, "", "", 0, 0, "", 0, 0, ""))
	require.Error(t, scanRepo.Finish(ctx, 1, ScanFailed, "x"))
	require.Error(t, scanRepo.MarkRunning(ctx, 1))
	_, err = scanRepo.ListByLibrary(ctx, 1, 5)
	require.Error(t, err)
}
