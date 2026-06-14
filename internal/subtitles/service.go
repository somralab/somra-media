package subtitles

import (
	"context"
	"fmt"

	"github.com/somralab/somra-media/internal/platform/db"
)

// MediaLookup resolves title/year for subtitle search.
type MediaLookup interface {
	GetItem(ctx context.Context, id int64) (db.MediaItem, error)
}

// SettingsReader exposes subtitle-related settings.
type SettingsReader interface {
	PreferredLanguages(ctx context.Context) ([]string, error)
	OpenSubtitlesAPIKey(ctx context.Context) (string, error)
}

// Service coordinates search, download, upload, and listing.
type Service struct {
	Repo     *db.SubtitleRepo
	Media    MediaLookup
	Settings SettingsReader
	Storage  *Storage
	Provider Provider
}

// Search finds external subtitles for a media item.
func (s *Service) Search(ctx context.Context, mediaItemID int64, language, query string) ([]SearchResult, error) {
	item, err := s.Media.GetItem(ctx, mediaItemID)
	if err != nil {
		return nil, err
	}
	if query == "" {
		query = item.Title
	}
	if language == "" {
		langs, _ := s.Settings.PreferredLanguages(ctx)
		if len(langs) > 0 {
			language = langs[0]
		}
	}
	provider := s.resolveProvider(ctx)
	if provider == nil {
		return nil, fmt.Errorf("subtitles: no provider configured")
	}
	return provider.Search(ctx, SearchQuery{
		Title:    query,
		Year:     item.Year,
		Language: language,
	})
}

// Download fetches and stores a subtitle from a provider.
func (s *Service) Download(ctx context.Context, mediaItemID int64, providerName, externalID, language string) (db.SubtitleFile, error) {
	provider := s.resolveProvider(ctx)
	if provider == nil || provider.Name() != providerName {
		return db.SubtitleFile{}, fmt.Errorf("subtitles: provider %q not available", providerName)
	}
	data, err := provider.Download(ctx, externalID, language)
	if err != nil {
		return db.SubtitleFile{}, err
	}
	path, err := s.Storage.Save(mediaItemID, language, data)
	if err != nil {
		return db.SubtitleFile{}, err
	}
	id, err := s.Repo.Create(ctx, db.SubtitleFile{
		MediaItemID: mediaItemID,
		Language:    language,
		Source:      db.SubtitleExternal,
		Path:        path,
		Provider:    providerName,
		ExternalID:  externalID,
	})
	if err != nil {
		return db.SubtitleFile{}, err
	}
	files, err := s.Repo.ListByMediaItem(ctx, mediaItemID)
	if err != nil {
		return db.SubtitleFile{}, err
	}
	for _, f := range files {
		if f.ID == id {
			return f, nil
		}
	}
	return db.SubtitleFile{}, fmt.Errorf("subtitles: created file not found")
}

// Upload stores a user-provided subtitle file.
func (s *Service) Upload(ctx context.Context, mediaItemID int64, language string, content []byte) (db.SubtitleFile, error) {
	path, err := s.Storage.Save(mediaItemID, language, content)
	if err != nil {
		return db.SubtitleFile{}, err
	}
	id, err := s.Repo.Create(ctx, db.SubtitleFile{
		MediaItemID: mediaItemID,
		Language:    language,
		Source:      db.SubtitleUploaded,
		Path:        path,
	})
	if err != nil {
		return db.SubtitleFile{}, err
	}
	files, err := s.Repo.ListByMediaItem(ctx, mediaItemID)
	if err != nil {
		return db.SubtitleFile{}, err
	}
	for _, f := range files {
		if f.ID == id {
			return f, nil
		}
	}
	return db.SubtitleFile{}, fmt.Errorf("subtitles: uploaded file not found")
}

// List returns managed subtitles for a media item.
func (s *Service) List(ctx context.Context, mediaItemID int64) ([]db.SubtitleFile, error) {
	return s.Repo.ListByMediaItem(ctx, mediaItemID)
}

// AutoDownloadMissing fetches subtitles for items missing preferred languages.
func (s *Service) AutoDownloadMissing(ctx context.Context, limit int) (int, error) {
	langs, err := s.Settings.PreferredLanguages(ctx)
	if err != nil || len(langs) == 0 {
		return 0, err
	}
	provider := s.resolveProvider(ctx)
	if provider == nil {
		return 0, nil
	}
	ids, err := s.Repo.ListMediaItemsMissingLanguages(ctx, langs, limit)
	if err != nil {
		return 0, err
	}
	downloaded := 0
	for _, id := range ids {
		for _, lang := range langs {
			has, err := s.Repo.HasLanguage(ctx, id, lang)
			if err != nil || has {
				continue
			}
			results, err := s.Search(ctx, id, lang, "")
			if err != nil || len(results) == 0 {
				continue
			}
			best := results[0]
			if _, err := s.Download(ctx, id, best.Provider, best.ExternalID, best.Language); err != nil {
				continue
			}
			downloaded++
			break
		}
	}
	return downloaded, nil
}

func (s *Service) resolveProvider(ctx context.Context) Provider {
	if s.Provider != nil {
		return s.Provider
	}
	key, err := s.Settings.OpenSubtitlesAPIKey(ctx)
	if err != nil || key == "" {
		return nil
	}
	return NewOpenSubtitles(key)
}
