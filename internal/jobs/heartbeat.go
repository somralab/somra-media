package jobs

import (
	"context"
	"log/slog"
	"sync/atomic"
)

// HeartbeatJobName is the canonical name used when registering the
// heartbeat job with the scheduler. The constant is exported so other
// packages (e.g. diagnostics) can reference the same string.
const HeartbeatJobName = "scheduler.heartbeat"

// Heartbeat is the canonical example periodic job. It emits a single
// structured log entry on every tick and exposes Ticks for tests.
//
// Heartbeat is safe for concurrent use; the scheduler's overlap
// protection ensures only one invocation runs at a time per name.
type Heartbeat struct {
	logger *slog.Logger
	ticks  atomic.Uint64
}

// NewHeartbeat builds a Heartbeat that logs to logger (defaulting to
// slog.Default when nil).
func NewHeartbeat(logger *slog.Logger) *Heartbeat {
	if logger == nil {
		logger = slog.Default()
	}
	return &Heartbeat{logger: logger}
}

// Run implements Job. It logs the canonical heartbeat event so
// operators can confirm the scheduler is alive end-to-end.
func (h *Heartbeat) Run(ctx context.Context) error {
	n := h.ticks.Add(1)
	h.logger.Info(
		"scheduler.heartbeat",
		slog.String("event", "scheduler.heartbeat"),
		slog.Uint64("tick", n),
	)
	return nil
}

// Ticks returns the number of times Run has been invoked. Primarily
// useful for tests and diagnostics.
func (h *Heartbeat) Ticks() uint64 { return h.ticks.Load() }
