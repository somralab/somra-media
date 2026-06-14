// Package jobs provides a lightweight scheduler and queue abstraction
// used across Somra to run recurring and background work.
//
// The scheduler wraps robfig/cron/v3 with overlap protection and an
// in-memory job-record table that other packages can introspect via the
// diagnostics provider. The package also defines an asynchronous job
// queue contract (JobQueue) for future use by features such as the
// library scan in Sprint 02.
package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// JobStatus represents the lifecycle of a tracked job.
type JobStatus string

const (
	StatusIdle    JobStatus = "idle"
	StatusRunning JobStatus = "running"
	StatusSuccess JobStatus = "success"
	StatusFailed  JobStatus = "failed"
)

// JobID is an opaque identifier returned by Scheduler.Schedule.
type JobID string

// Job is the minimal contract a scheduled or queued unit of work must satisfy.
type Job interface {
	Run(ctx context.Context) error
}

// JobFunc adapts a plain function to the Job interface.
type JobFunc func(ctx context.Context) error

// Run implements Job for JobFunc.
func (f JobFunc) Run(ctx context.Context) error { return f(ctx) }

// JobRecord captures the most recent execution state of a named job.
//
// The struct is value-copy friendly: callers receive snapshots, not
// references to the scheduler's internal state.
type JobRecord struct {
	Name        string
	Status      JobStatus
	LastStarted time.Time
	LastEnded   time.Time
	LastError   string
	Runs        uint64
	Failures    uint64
}

// jobTracker maintains per-name execution records protected by a mutex.
//
// It is intentionally small and avoids per-job goroutines; the scheduler
// drives state transitions inline.
type jobTracker struct {
	mu      sync.RWMutex
	records map[string]*JobRecord
}

func newJobTracker() *jobTracker {
	return &jobTracker{records: make(map[string]*JobRecord)}
}

func (t *jobTracker) ensure(name string) *JobRecord {
	t.mu.Lock()
	defer t.mu.Unlock()
	rec, ok := t.records[name]
	if !ok {
		rec = &JobRecord{Name: name, Status: StatusIdle}
		t.records[name] = rec
	}
	return rec
}

func (t *jobTracker) markStarted(name string, at time.Time) {
	rec := t.ensure(name)
	t.mu.Lock()
	defer t.mu.Unlock()
	rec.Status = StatusRunning
	rec.LastStarted = at
}

func (t *jobTracker) markFinished(name string, at time.Time, err error) {
	rec := t.ensure(name)
	t.mu.Lock()
	defer t.mu.Unlock()
	rec.LastEnded = at
	rec.Runs++
	if err != nil {
		rec.Status = StatusFailed
		rec.LastError = err.Error()
		rec.Failures++
		return
	}
	rec.Status = StatusSuccess
	rec.LastError = ""
}

// Snapshot returns a copy of the record for name, or false if unknown.
func (t *jobTracker) Snapshot(name string) (JobRecord, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	rec, ok := t.records[name]
	if !ok {
		return JobRecord{}, false
	}
	return *rec, true
}

// All returns a snapshot of every tracked job, ordered by name insertion
// is not preserved; callers should sort if needed.
func (t *jobTracker) All() []JobRecord {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]JobRecord, 0, len(t.records))
	for _, rec := range t.records {
		out = append(out, *rec)
	}
	return out
}

// ErrOverlapping is returned by RunOnce when a previous invocation of the
// same named job is still running. Callers can use errors.Is to detect this
// condition and back off.
var ErrOverlapping = fmt.Errorf("job is already running")
