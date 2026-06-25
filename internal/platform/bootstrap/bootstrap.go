// Package bootstrap wires Somra's platform-level components together
// for use by cmd/somra and tests.
//
// Paket 5 introduced the scheduler / i18n bundle / diagnostics
// registry. Paket 8 (QA) extends Components with the DB Pinger from
// Paket 4 so the integration tests and the production binary share the
// same wiring path.
//
// Keeping wiring in one place avoids accidental global state and lets
// the integration packet add new dependencies without touching every
// caller.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/platform/diagnostics"
	"github.com/somralab/somra-media/internal/platform/i18n"
)

// Components groups the platform singletons that other packages depend
// on. Fields are exported because callers (cmd/somra, tests) must reach
// into them.
type Components struct {
	Logger      *slog.Logger
	Scheduler   *jobs.Scheduler
	Queue       *jobs.MemoryQueue
	I18n        *i18n.Bundle
	Diagnostics *diagnostics.Registry
	Heartbeat   *jobs.Heartbeat
	// DB is the migrated SQLite handle. Nil when the bootstrap was built
	// without storage (e.g. lightweight unit tests via New).
	DB *db.DB

	// Uptime is exposed so cmd/somra can surface it on /api/v1/version.
	Uptime *diagnostics.UptimeProvider
}

// New builds the default Somra platform components. The caller owns
// lifecycle: invoke Components.Scheduler.Start(ctx) when ready and
// Components.Scheduler.Stop on shutdown.
func New(logger *slog.Logger) (*Components, error) {
	if logger == nil {
		logger = slog.Default()
	}

	bundle, err := i18n.NewBundle()
	if err != nil {
		return nil, fmt.Errorf("bootstrap i18n: %w", err)
	}

	sched := jobs.New(logger)
	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfigFromEnv(logger))
	heartbeat := jobs.NewHeartbeat(logger)

	registry := diagnostics.NewRegistry()
	uptime := diagnostics.NewUptimeProvider()
	registry.Register(uptime)
	registry.Register(diagnostics.NewSchedulerProvider(sched))

	return &Components{
		Logger:      logger,
		Scheduler:   sched,
		Queue:       queue,
		I18n:        bundle,
		Diagnostics: registry,
		Heartbeat:   heartbeat,
		Uptime:      uptime,
	}, nil
}

// NewWithStorage builds the default Components and additionally opens
// and migrates the SQLite database described by dbCfg, registering it
// as a critical diagnostics provider. The caller still owns lifecycle:
// invoke Components.Scheduler.Start(ctx) when ready and
// Components.Close on shutdown.
func NewWithStorage(ctx context.Context, logger *slog.Logger, dbCfg db.Config) (*Components, error) {
	c, err := New(logger)
	if err != nil {
		return nil, err
	}
	d, err := db.Initialize(ctx, dbCfg, c.Logger)
	if err != nil {
		c.Queue.Close()
		return nil, fmt.Errorf("bootstrap storage: %w", err)
	}
	c.DB = d
	c.Diagnostics.Register(diagnostics.NewDBProvider(d))
	return c, nil
}

// Close releases the resources owned by Components (queue, DB). It is
// safe to call on a partially constructed value: nil fields are
// skipped.
func (c *Components) Close() error {
	if c == nil {
		return nil
	}
	var errs []error
	if c.Queue != nil {
		c.Queue.Close()
	}
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close db: %w", err))
		}
	}
	return errors.Join(errs...)
}
