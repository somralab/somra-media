package bootstrap

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/somralab/somra-media/internal/api"
	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/library"
	"github.com/somralab/somra-media/internal/metadata"
)

// LibraryBundle groups Sprint 02 library/metadata services.
type LibraryBundle struct {
	EventBus *api.EventBus
	Library  *library.Service
	Metadata *metadata.Service
}

// WireLibrary constructs library scan and metadata services on top of
// existing platform components.
func WireLibrary(c *Components) *LibraryBundle {
	bus := api.NewEventBus()
	scanner := library.NewScanner(library.ScannerConfig{
		Logger:   c.Logger,
		DB:       c.DB,
		Prober:   library.NewProber(""),
		Progress: api.ScanProgressPublisher{Bus: bus},
	})
	svc := library.NewService(library.ServiceConfig{
		Logger:  c.Logger,
		DB:      c.DB,
		Queue:   c.Queue,
		Scanner: scanner,
	})

	reg := metadata.NewRegistry()
	reg.Register(&metadata.TVDBProvider{})
	reg.Register(&metadata.MusicBrainzProvider{})
	reg.Register(&metadata.FanartProvider{})
	if key := os.Getenv("SOMRA_TMDB_API_KEY"); key != "" {
		reg.Register(metadata.NewTMDBProvider(key, nil))
	}

	metaSvc := &metadata.Service{
		DB:       &metadata.DBStore{DB: c.DB},
		Registry: reg,
		Matcher:  &metadata.Matcher{Registry: reg, Limiter: metadata.NewRateLimiter(250 * time.Millisecond)},
	}

	_, _ = c.Scheduler.Schedule("0 0 3 * * *", "metadata-refresh", metadataRefreshJob(c.Logger, svc, metaSvc))

	return &LibraryBundle{EventBus: bus, Library: svc, Metadata: metaSvc}
}

func metadataRefreshJob(logger *slog.Logger, libSvc *library.Service, metaSvc *metadata.Service) jobs.Job {
	return jobs.JobFunc(func(ctx context.Context) error {
		libs, err := libSvc.ListLibraries(ctx)
		if err != nil {
			return err
		}
		for _, lib := range libs {
			_, err := metaSvc.AutoMatch(ctx, lib.ID, "en-US", 50)
			if err != nil {
				logger.Warn("metadata refresh", slog.Int64("libraryId", lib.ID), slog.Any("error", err))
			}
		}
		return nil
	})
}
