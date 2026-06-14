package bootstrap

import (
	"context"
	"log/slog"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/settings"
	"github.com/somralab/somra-media/internal/subtitles"
)

// SettingsBundle groups Sprint 06 settings/onboarding services.
type SettingsBundle struct {
	Settings   *settings.Service
	Onboarding *settings.Onboarding
}

// SubtitlesBundle groups subtitle automation services.
type SubtitlesBundle struct {
	Service *subtitles.Service
}

type mediaItemLookup struct {
	repo *db.MediaRepo
}

func (m mediaItemLookup) GetItem(ctx context.Context, id int64) (db.MediaItem, error) {
	return m.repo.GetItemByID(ctx, id, "en-US")
}

// WireSettings constructs settings and onboarding services.
func WireSettings(c *Components, authSvc *auth.Service) *SettingsBundle {
	repo := db.NewSettingsRepo(c.DB.Querier())
	svc := settings.NewService(repo)
	onb := settings.NewOnboarding(repo, svc, authSvc)
	return &SettingsBundle{Settings: svc, Onboarding: onb}
}

// WireSubtitles constructs subtitle search/download automation.
func WireSubtitles(c *Components, cfg config.Config, settingsSvc *settings.Service, logger *slog.Logger) *SubtitlesBundle {
	if c == nil || c.DB == nil || settingsSvc == nil {
		return nil
	}
	cacheDir := cfg.Data.CacheDir
	if cacheDir == "" {
		cacheDir = cfg.Data.Dir + "/cache"
	}
	svc := &subtitles.Service{
		Repo:     db.NewSubtitleRepo(c.DB.Querier()),
		Media:    mediaItemLookup{repo: db.NewMediaRepo(c.DB.Querier())},
		Settings: settingsSvc,
		Storage:  &subtitles.Storage{Root: cacheDir},
	}
	if c.Scheduler != nil {
		subSvc := svc
		_, _ = c.Scheduler.Schedule("0 30 4 * * *", "subtitle-auto-download", jobs.JobFunc(func(ctx context.Context) error {
			_, err := subSvc.AutoDownloadMissing(ctx, 50)
			return err
		}))
	}
	return &SubtitlesBundle{Service: svc}
}
