package subtitles_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/settings"
	"github.com/somralab/somra-media/internal/subtitles"
)

type stubSettings struct {
	langs  []string
	apiKey string
}

func (s stubSettings) PreferredLanguages(context.Context) ([]string, error) {
	return s.langs, nil
}

func (s stubSettings) OpenSubtitlesAPIKey(context.Context) (string, error) {
	return s.apiKey, nil
}

type stubMedia struct {
	item db.MediaItem
}

func (m stubMedia) GetItem(_ context.Context, _ int64) (db.MediaItem, error) {
	return m.item, nil
}

type langLister struct {
	langs []string
}

func (l langLister) ListLanguages(_ context.Context, _ int64) ([]string, error) {
	return l.langs, nil
}

func TestStorageSave(t *testing.T) {
	root := t.TempDir()
	store := &subtitles.Storage{Root: root}
	path, err := store.Save(42, "en", []byte("subtitle content"))
	require.NoError(t, err)
	assert.Contains(t, path, "42")
	assert.FileExists(t, path)

	_, err = (&subtitles.Storage{}).Save(1, "en", []byte("x"))
	require.Error(t, err)
}

func TestServiceSearchDownloadUploadList(t *testing.T) {
	ctx := context.Background()
	d := openSubtitleTestDB(t)
	repo := db.NewSubtitleRepo(d.Querier())
	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Films", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	year := 2010
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Inception", &year)
	require.NoError(t, err)
	_, err = mediaRepo.UpsertFile(ctx, db.MediaFile{
		MediaItemID: &itemID,
		LibraryID:   lib.ID,
		Path:        filepath.Join(dir, "inception.mkv"),
	})
	require.NoError(t, err)

	mock := &mockProvider{
		results: []subtitles.SearchResult{{
			Provider: "mock", ExternalID: "99", Language: "en", Score: 95,
		}},
		data: []byte("1\n00:00:01,000 --> 00:00:02,000\nHello"),
	}
	svc := &subtitles.Service{
		Repo:     repo,
		Media:    stubMedia{item: db.MediaItem{ID: itemID, Title: "Inception", Year: &year}},
		Settings: stubSettings{langs: []string{"en"}, apiKey: "key"},
		Storage:  &subtitles.Storage{Root: t.TempDir()},
		Provider: mock,
	}

	results, err := svc.Search(ctx, itemID, "en", "")
	require.NoError(t, err)
	require.Len(t, results, 1)

	file, err := svc.Download(ctx, itemID, "mock", "99", "en")
	require.NoError(t, err)
	assert.Equal(t, "en", file.Language)
	assert.Equal(t, db.SubtitleExternal, file.Source)

	uploaded, err := svc.Upload(ctx, itemID, "tr", []byte("uploaded"))
	require.NoError(t, err)
	assert.Equal(t, "tr", uploaded.Language)

	list, err := svc.List(ctx, itemID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 2)

	count, err := svc.AutoDownloadMissing(ctx, 5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestAutoDownloadMissingDownloadsSubtitle(t *testing.T) {
	ctx := context.Background()
	d := openSubtitleTestDB(t)
	repo := db.NewSubtitleRepo(d.Querier())
	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	settingsSvc := settings.NewService(db.NewSettingsRepo(d.Querier()))
	_, err := settingsSvc.PatchCategory(ctx, settings.CategorySubtitles, map[string]any{
		"preferredLanguages": []any{"en"},
	})
	require.NoError(t, err)

	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Films", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	year := 2010
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Inception", &year)
	require.NoError(t, err)
	_, err = mediaRepo.UpsertFile(ctx, db.MediaFile{
		MediaItemID: &itemID,
		LibraryID:   lib.ID,
		Path:        filepath.Join(dir, "inception.mkv"),
	})
	require.NoError(t, err)

	mock := &mockProvider{
		results: []subtitles.SearchResult{{Provider: "mock", ExternalID: "9", Language: "en", Score: 90}},
		data:    []byte("subtitle bytes"),
	}
	svc := &subtitles.Service{
		Repo:     repo,
		Media:    stubMedia{item: db.MediaItem{ID: itemID, Title: "Inception", Year: &year}},
		Settings: settingsSvc,
		Storage:  &subtitles.Storage{Root: t.TempDir()},
		Provider: mock,
	}

	count, err := svc.AutoDownloadMissing(ctx, 5)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDetectorMissingForItem(t *testing.T) {
	ctx := context.Background()
	det := &subtitles.Detector{
		Preferred: func(context.Context) ([]string, error) { return []string{"en", "tr"}, nil },
		List:      langLister{langs: []string{"en"}},
	}
	missing, err := det.MissingForItem(ctx, 1, []string{"fr"})
	require.NoError(t, err)
	assert.Equal(t, []string{"tr"}, missing)
}

func TestServiceNoProvider(t *testing.T) {
	ctx := context.Background()
	svc := &subtitles.Service{
		Settings: stubSettings{},
		Media:    stubMedia{item: db.MediaItem{ID: 1, Title: "X"}},
	}
	_, err := svc.Search(ctx, 1, "en", "X")
	require.Error(t, err)

	n, err := svc.AutoDownloadMissing(ctx, 5)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestServiceResolveProviderFromSettings(t *testing.T) {
	ctx := context.Background()
	d := openSubtitleTestDB(t)
	settingsSvc := settings.NewService(db.NewSettingsRepo(d.Querier()))
	_, err := settingsSvc.PatchCategory(ctx, settings.CategorySubtitles, map[string]any{
		"apiKey": "configured-key",
	})
	require.NoError(t, err)

	svc := &subtitles.Service{
		Repo:     db.NewSubtitleRepo(d.Querier()),
		Settings: settingsSvc,
		Media:    stubMedia{item: db.MediaItem{ID: 1, Title: "X"}},
	}
	_, err = svc.Search(ctx, 1, "en", "X")
	require.Error(t, err)
}

func openSubtitleTestDB(t *testing.T) *db.DB {
	t.Helper()
	cfg := db.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")
	d, err := db.Initialize(context.Background(), cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })
	return d
}
