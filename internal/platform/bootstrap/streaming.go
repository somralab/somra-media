package bootstrap

import (
	"context"
	"log/slog"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/streaming"
)

// StreamingBundle groups streaming dependencies for API wiring.
type StreamingBundle struct {
	Service *streaming.Service
}

// WireStreaming constructs the streaming service and registers idle reaper job.
func WireStreaming(c *Components, cfg config.Config, logger *slog.Logger) *StreamingBundle {
	if c == nil || c.DB == nil {
		return nil
	}
	if logger == nil {
		logger = c.Logger
	}
	cacheDir := cfg.Data.CacheDir
	if cacheDir == "" {
		cacheDir = cfg.Data.Dir + "/cache"
	}
	svc := streaming.NewService(streaming.ServiceConfig{
		CacheDir:          cacheDir,
		SessionTTL:        cfg.Streaming.SessionTTL,
		IdleTimeout:       cfg.Streaming.IdleTimeout,
		MaxConcurrent:     cfg.Streaming.MaxConcurrentTranscodes,
		MaxTranscodeQueue: cfg.Streaming.MaxTranscodeQueue,
		FFmpegBin:         cfg.Streaming.FFmpegBin,
		FFprobeBin:        cfg.Streaming.FFprobeBin,
	}, db.NewPlaybackRepo(c.DB.Querier()), db.NewMediaRepo(c.DB.Querier()), logger)

	if c.Scheduler != nil {
		streamSvc := svc
		_, _ = c.Scheduler.Schedule("0 */1 * * * *", "streaming-idle-reaper", jobs.JobFunc(func(ctx context.Context) error {
			return streamSvc.ReapIdle(ctx)
		}))
	}
	return &StreamingBundle{Service: svc}
}
