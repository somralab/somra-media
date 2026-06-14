package jobs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
)

func TestMemoryQueueEnqueueRuns(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 2, Buffer: 4, Logger: quietLogger()})
	defer q.Close()

	done := make(chan struct{}, 1)
	id, err := q.Enqueue(context.Background(), jobs.JobFunc(func(context.Context) error {
		done <- struct{}{}
		return nil
	}), jobs.WithName("greeter"))
	require.NoError(t, err)
	require.NotEmpty(t, id)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("queued job did not run")
	}

	require.Eventually(t, func() bool {
		st, err := q.Status(context.Background(), id)
		return err == nil && st == jobs.TaskSucceeded
	}, time.Second, 5*time.Millisecond)
}

func TestMemoryQueueFailedTask(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2, Logger: quietLogger()})
	defer q.Close()

	id, err := q.Enqueue(context.Background(), jobs.JobFunc(func(context.Context) error {
		return errors.New("nope")
	}))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		st, _ := q.Status(context.Background(), id)
		return st == jobs.TaskFailed
	}, time.Second, 5*time.Millisecond)
}

func TestMemoryQueueCancelPending(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 4, Logger: quietLogger()})
	defer q.Close()

	block := make(chan struct{})
	first, err := q.Enqueue(context.Background(), jobs.JobFunc(func(ctx context.Context) error {
		<-block
		return nil
	}))
	require.NoError(t, err)

	pending, err := q.Enqueue(context.Background(), jobs.JobFunc(func(context.Context) error {
		t.Fatalf("cancelled task must not run")
		return nil
	}))
	require.NoError(t, err)

	require.NoError(t, q.Cancel(context.Background(), pending))
	st, err := q.Status(context.Background(), pending)
	require.NoError(t, err)
	require.Equal(t, jobs.TaskCancelled, st)

	close(block)
	require.Eventually(t, func() bool {
		s, _ := q.Status(context.Background(), first)
		return s == jobs.TaskSucceeded
	}, time.Second, 5*time.Millisecond)
}

func TestMemoryQueueCancelRunning(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 1, Logger: quietLogger()})
	defer q.Close()

	started := make(chan struct{})
	id, err := q.Enqueue(context.Background(), jobs.JobFunc(func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}))
	require.NoError(t, err)

	<-started
	require.NoError(t, q.Cancel(context.Background(), id))

	require.Eventually(t, func() bool {
		st, _ := q.Status(context.Background(), id)
		return st == jobs.TaskCancelled
	}, time.Second, 5*time.Millisecond)
}

func TestMemoryQueueStatusUnknown(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 1, Logger: quietLogger()})
	defer q.Close()

	_, err := q.Status(context.Background(), "missing")
	require.ErrorIs(t, err, jobs.ErrTaskNotFound)

	require.ErrorIs(t, q.Cancel(context.Background(), "missing"), jobs.ErrTaskNotFound)
}

func TestMemoryQueueRejectsNilJob(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 1, Logger: quietLogger()})
	defer q.Close()

	_, err := q.Enqueue(context.Background(), nil)
	require.Error(t, err)
}

func TestMemoryQueueCloseRejectsEnqueue(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 1, Logger: quietLogger()})
	q.Close()
	_, err := q.Enqueue(context.Background(), jobs.JobFunc(func(context.Context) error { return nil }))
	require.Error(t, err)
	q.Close()
}

func TestMemoryQueueDefaultConfig(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{})
	defer q.Close()

	done := make(chan struct{})
	_, err := q.Enqueue(context.Background(), jobs.JobFunc(func(context.Context) error {
		close(done)
		return nil
	}))
	require.NoError(t, err)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("default-config queue must run a task")
	}
}

func TestMemoryQueueEnqueueRespectsContext(t *testing.T) {
	t.Parallel()
	q := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 1, Logger: quietLogger()})
	defer q.Close()

	block := make(chan struct{})
	_, err := q.Enqueue(context.Background(), jobs.JobFunc(func(context.Context) error {
		<-block
		return nil
	}))
	require.NoError(t, err)

	_, err = q.Enqueue(context.Background(), jobs.JobFunc(func(context.Context) error { return nil }))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err = q.Enqueue(ctx, jobs.JobFunc(func(context.Context) error { return nil }))
	require.Error(t, err)

	close(block)
}
