package library

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/platform/db"
)

// Watcher debounces filesystem events and triggers incremental scans.
type Watcher struct {
	logger   *slog.Logger
	lib      db.Library
	queue    jobs.JobQueue
	scanner  *Scanner
	debounce time.Duration

	mu      sync.Mutex
	watcher *fsnotify.Watcher
	timer   *time.Timer
	stopCh  chan struct{}
}

// NewWatcher builds a filesystem watcher for lib.
func NewWatcher(logger *slog.Logger, lib db.Library, queue jobs.JobQueue, scanner *Scanner, debounce time.Duration) *Watcher {
	if logger == nil {
		logger = slog.Default()
	}
	if debounce <= 0 {
		debounce = 2 * time.Second
	}
	return &Watcher{
		logger:   logger,
		lib:      lib,
		queue:    queue,
		scanner:  scanner,
		debounce: debounce,
		stopCh:   make(chan struct{}),
	}
}

// Start begins watching library paths.
func (w *Watcher) Start(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.logger.Error("watch init failed", slog.Int64("libraryId", w.lib.ID), slog.Any("error", err))
		return
	}
	w.mu.Lock()
	w.watcher = watcher
	w.mu.Unlock()

	for _, p := range w.lib.Paths {
		if err := watcher.Add(p); err != nil {
			w.logger.Warn("watch add path failed", slog.String("path", p), slog.Any("error", err))
		}
	}

	go w.loop(ctx)
}

// Stop stops the watcher.
func (w *Watcher) Stop() {
	close(w.stopCh)
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timer != nil {
		w.timer.Stop()
	}
	if w.watcher != nil {
		_ = w.watcher.Close()
	}
}

func (w *Watcher) loop(ctx context.Context) {
	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				w.scheduleScan(ctx)
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.logger.Warn("watch error", slog.Int64("libraryId", w.lib.ID), slog.Any("error", err))
		}
	}
}

func (w *Watcher) scheduleScan(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(w.debounce, func() {
		runID, err := db.NewScanRepo(w.scanner.db.Querier()).CreateRun(ctx, w.lib.ID, db.ScanIncremental, "")
		if err != nil {
			w.logger.Warn("watch scan create run", slog.Any("error", err))
			return
		}
		_, err = w.queue.Enqueue(ctx, &ScanJob{
			Scanner:   w.scanner,
			LibraryID: w.lib.ID,
			ScanType:  db.ScanIncremental,
			RunID:     runID,
		}, jobs.WithName("library-watch-scan"))
		if err != nil {
			w.logger.Warn("watch scan enqueue", slog.Any("error", err))
		}
	})
}
