package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler is a thin wrapper around robfig/cron/v3 that adds:
//   - per-name overlap protection (a still-running job is skipped),
//   - a structured-logger surface,
//   - a tracked job-record table for diagnostics.
//
// The zero value is not usable; construct with New.
type Scheduler struct {
	logger  *slog.Logger
	cron    *cron.Cron
	tracker *jobTracker

	mu     sync.Mutex
	locks  map[string]*sync.Mutex
	names  map[JobID]string
	idSeq  atomic.Uint64
	state  schedulerState
	stopCh chan struct{}
}

type schedulerState int32

const (
	stateNew schedulerState = iota
	stateRunning
	stateStopped
)

// New builds a Scheduler with the provided logger. A nil logger falls
// back to slog.Default so callers don't need a guard at every call site.
func New(logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{
		logger:  logger,
		cron:    cron.New(cron.WithSeconds()),
		tracker: newJobTracker(),
		locks:   make(map[string]*sync.Mutex),
		names:   make(map[JobID]string),
		stopCh:  make(chan struct{}),
	}
}

// Start begins running scheduled jobs. It is non-blocking: cron runs in
// its own goroutine. When ctx is cancelled the scheduler stops itself
// gracefully (waiting for in-flight jobs up to the cron stop context).
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.state != stateNew {
		s.mu.Unlock()
		return
	}
	s.state = stateRunning
	s.mu.Unlock()

	s.cron.Start()
	s.logger.Info("scheduler.started", slog.String("event", "scheduler.started"))

	go func() {
		select {
		case <-ctx.Done():
			_ = s.Stop(context.Background())
		case <-s.stopCh:
		}
	}()
}

// Stop halts the cron scheduler and waits for in-flight jobs to finish
// (or for ctx to expire). It is safe to call Stop multiple times.
func (s *Scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.state == stateStopped {
		s.mu.Unlock()
		return nil
	}
	s.state = stateStopped
	close(s.stopCh)
	s.mu.Unlock()

	done := s.cron.Stop().Done()
	select {
	case <-done:
		s.logger.Info("scheduler.stopped", slog.String("event", "scheduler.stopped"))
		return nil
	case <-ctx.Done():
		return fmt.Errorf("scheduler stop: %w", ctx.Err())
	}
}

// Running reports whether Start has been called and Stop has not yet
// returned.
func (s *Scheduler) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == stateRunning
}

// Schedule registers job to run on spec (a robfig/cron expression with
// optional seconds). The name is used for overlap protection and as the
// key in the diagnostics tracker.
func (s *Scheduler) Schedule(spec string, name string, job Job) (JobID, error) {
	if name == "" {
		return "", fmt.Errorf("schedule %q: name is required", spec)
	}
	if job == nil {
		return "", fmt.Errorf("schedule %s: job is required", name)
	}
	s.tracker.ensure(name)
	s.getLock(name)

	wrapped := func() {
		s.invoke(context.Background(), name, job)
	}
	entryID, err := s.cron.AddFunc(spec, wrapped)
	if err != nil {
		return "", fmt.Errorf("schedule %s: %w", name, err)
	}

	id := JobID(fmt.Sprintf("job-%d", s.idSeq.Add(1)))
	s.mu.Lock()
	s.names[id] = name
	s.mu.Unlock()
	_ = entryID
	return id, nil
}

// RunOnce executes job immediately on the caller's goroutine, honouring
// overlap protection: if a job with the same name is already running,
// ErrOverlapping is returned and a warning is logged.
func (s *Scheduler) RunOnce(ctx context.Context, name string, job Job) error {
	if name == "" {
		return fmt.Errorf("run once: name is required")
	}
	if job == nil {
		return fmt.Errorf("run once %s: job is required", name)
	}

	lock := s.getLock(name)
	if !lock.TryLock() {
		s.logger.Warn(
			"job.skip_overlapping",
			slog.String("event", "job.skip_overlapping"),
			slog.String("job", name),
		)
		return ErrOverlapping
	}
	defer lock.Unlock()
	return s.runLocked(ctx, name, job)
}

// invoke is the cron-side entrypoint that respects overlap protection.
func (s *Scheduler) invoke(ctx context.Context, name string, job Job) {
	lock := s.getLock(name)
	if !lock.TryLock() {
		s.logger.Warn(
			"job.skip_overlapping",
			slog.String("event", "job.skip_overlapping"),
			slog.String("job", name),
		)
		return
	}
	defer lock.Unlock()
	_ = s.runLocked(ctx, name, job)
}

func (s *Scheduler) runLocked(ctx context.Context, name string, job Job) error {
	start := time.Now()
	s.tracker.markStarted(name, start)
	s.logger.Info(
		"job.started",
		slog.String("event", "job.started"),
		slog.String("job", name),
	)

	err := job.Run(ctx)
	end := time.Now()
	s.tracker.markFinished(name, end, err)

	if err != nil {
		s.logger.Error(
			"job.failed",
			slog.String("event", "job.failed"),
			slog.String("job", name),
			slog.Duration("duration", end.Sub(start)),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("job %s: %w", name, err)
	}

	s.logger.Info(
		"job.success",
		slog.String("event", "job.success"),
		slog.String("job", name),
		slog.Duration("duration", end.Sub(start)),
	)
	return nil
}

func (s *Scheduler) getLock(name string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()
	lock, ok := s.locks[name]
	if !ok {
		lock = &sync.Mutex{}
		s.locks[name] = lock
	}
	return lock
}

// Records returns a snapshot of all tracked job records. The returned
// slice is owned by the caller.
func (s *Scheduler) Records() []JobRecord { return s.tracker.All() }

// Record returns the record for name, or false if no job by that name
// has ever been seen.
func (s *Scheduler) Record(name string) (JobRecord, bool) {
	return s.tracker.Snapshot(name)
}
