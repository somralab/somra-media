package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// TaskID identifies a queued asynchronous job instance.
type TaskID string

// TaskStatus represents the lifecycle of a queued task.
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskSucceeded TaskStatus = "succeeded"
	TaskFailed    TaskStatus = "failed"
	TaskCancelled TaskStatus = "cancelled"
)

// ErrTaskNotFound is returned by JobQueue implementations when no task
// matches the given TaskID.
var ErrTaskNotFound = errors.New("task not found")

// Options carry per-enqueue metadata. Today only Name is used; the
// struct exists so additional fields (priority, retry policy, ...) can
// be introduced without breaking the public API.
type Options struct {
	Name string
}

// Option mutates Options when passed to JobQueue.Enqueue.
type Option func(*Options)

// WithName labels the queued task; surfaced by Status and in logs.
func WithName(name string) Option {
	return func(o *Options) { o.Name = name }
}

// JobQueue is the public contract for asynchronous background work.
//
// Implementations must be safe for concurrent use and must guarantee
// that Status reflects every state transition observable to callers.
// The contract is intentionally minimal — features such as retries,
// priorities and persistence are deferred to later sprints.
type JobQueue interface {
	Enqueue(ctx context.Context, job Job, opts ...Option) (TaskID, error)
	Status(ctx context.Context, id TaskID) (TaskStatus, error)
	Cancel(ctx context.Context, id TaskID) error
}

// MemoryQueue is an in-memory FIFO JobQueue served by a fixed-size
// worker pool. It is intended for development, tests, and Sprint 02's
// library scan; persistent queues will live behind the same interface.
type MemoryQueue struct {
	logger  *slog.Logger
	workers int
	taskCh  chan *queuedTask

	mu     sync.RWMutex
	tasks  map[TaskID]*queuedTask
	idSeq  atomic.Uint64
	wg     sync.WaitGroup
	closed atomic.Bool
}

type queuedTask struct {
	id              TaskID
	name            string
	job             Job
	status          TaskStatus
	err             error
	cancel          context.CancelFunc
	cancelRequested bool
	mu              sync.Mutex
}

// MemoryQueueConfig configures a MemoryQueue.
type MemoryQueueConfig struct {
	Workers int
	Buffer  int
	Logger  *slog.Logger
}

// NewMemoryQueue starts a MemoryQueue with cfg.Workers worker goroutines
// and a channel buffer of cfg.Buffer. Both default to 1 when zero or
// negative.
func NewMemoryQueue(cfg MemoryQueueConfig) *MemoryQueue {
	workers := cfg.Workers
	if workers <= 0 {
		workers = 1
	}
	buffer := cfg.Buffer
	if buffer <= 0 {
		buffer = 1
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	q := &MemoryQueue{
		logger:  logger,
		workers: workers,
		taskCh:  make(chan *queuedTask, buffer),
		tasks:   make(map[TaskID]*queuedTask),
	}
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
	return q
}

// Enqueue adds job to the queue with the supplied options. ctx governs
// the enqueue operation (typically used to honour caller cancellation
// when the channel is full), not the lifetime of job execution.
func (q *MemoryQueue) Enqueue(ctx context.Context, job Job, opts ...Option) (TaskID, error) {
	if q.closed.Load() {
		return "", errors.New("queue closed")
	}
	if job == nil {
		return "", errors.New("enqueue: job is required")
	}
	o := Options{}
	for _, apply := range opts {
		apply(&o)
	}
	id := TaskID(fmt.Sprintf("task-%d", q.idSeq.Add(1)))
	task := &queuedTask{
		id:     id,
		name:   o.Name,
		job:    job,
		status: TaskPending,
	}

	q.mu.Lock()
	q.tasks[id] = task
	q.mu.Unlock()

	select {
	case q.taskCh <- task:
		return id, nil
	case <-ctx.Done():
		q.mu.Lock()
		delete(q.tasks, id)
		q.mu.Unlock()
		return "", fmt.Errorf("enqueue: %w", ctx.Err())
	}
}

// Status returns the current status of the task identified by id.
func (q *MemoryQueue) Status(_ context.Context, id TaskID) (TaskStatus, error) {
	q.mu.RLock()
	task, ok := q.tasks[id]
	q.mu.RUnlock()
	if !ok {
		return "", ErrTaskNotFound
	}
	task.mu.Lock()
	defer task.mu.Unlock()
	return task.status, nil
}

// Cancel transitions a pending task to TaskCancelled or signals a
// running task's context. Already-completed tasks return nil so the
// operation is idempotent.
func (q *MemoryQueue) Cancel(_ context.Context, id TaskID) error {
	q.mu.RLock()
	task, ok := q.tasks[id]
	q.mu.RUnlock()
	if !ok {
		return ErrTaskNotFound
	}

	task.mu.Lock()
	switch task.status {
	case TaskPending:
		task.status = TaskCancelled
		task.mu.Unlock()
		return nil
	case TaskRunning:
		task.cancelRequested = true
		cancel := task.cancel
		task.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		return nil
	default:
		task.mu.Unlock()
		return nil
	}
}

// Close stops the worker pool. Pending tasks are not drained; callers
// should ensure quiescence before calling Close. Close blocks until all
// workers have exited.
func (q *MemoryQueue) Close() {
	if !q.closed.CompareAndSwap(false, true) {
		return
	}
	close(q.taskCh)
	q.wg.Wait()
}

func (q *MemoryQueue) worker(id int) {
	defer q.wg.Done()
	for task := range q.taskCh {
		q.execute(id, task)
	}
}

func (q *MemoryQueue) execute(workerID int, task *queuedTask) {
	task.mu.Lock()
	if task.status == TaskCancelled {
		task.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	task.cancel = cancel
	task.status = TaskRunning
	task.mu.Unlock()

	start := time.Now()
	q.logger.Info(
		"queue.task.started",
		slog.String("event", "queue.task.started"),
		slog.String("task", string(task.id)),
		slog.String("name", task.name),
		slog.Int("worker", workerID),
	)

	err := task.job.Run(ctx)
	cancel()

	task.mu.Lock()
	defer task.mu.Unlock()
	task.cancel = nil
	if task.cancelRequested {
		task.status = TaskCancelled
		task.err = err
		q.logger.Warn(
			"queue.task.cancelled",
			slog.String("event", "queue.task.cancelled"),
			slog.String("task", string(task.id)),
			slog.Duration("duration", time.Since(start)),
		)
		return
	}
	if err != nil {
		task.status = TaskFailed
		task.err = err
		q.logger.Error(
			"queue.task.failed",
			slog.String("event", "queue.task.failed"),
			slog.String("task", string(task.id)),
			slog.Duration("duration", time.Since(start)),
			slog.String("error", err.Error()),
		)
		return
	}
	task.status = TaskSucceeded
	q.logger.Info(
		"queue.task.succeeded",
		slog.String("event", "queue.task.succeeded"),
		slog.String("task", string(task.id)),
		slog.Duration("duration", time.Since(start)),
	)
}
