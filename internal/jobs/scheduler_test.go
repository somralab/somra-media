package jobs_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
)

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSchedulerStartStop(t *testing.T) {
	t.Parallel()
	s := jobs.New(quietLogger())
	require.False(t, s.Running())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)
	require.True(t, s.Running())

	require.NoError(t, s.Stop(context.Background()))
	require.NoError(t, s.Stop(context.Background()), "Stop must be idempotent")
}

func TestSchedulerScheduleFires(t *testing.T) {
	t.Parallel()
	s := jobs.New(quietLogger())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{}, 1)
	var fired atomic.Int32

	_, err := s.Schedule("@every 50ms", "tick", jobs.JobFunc(func(_ context.Context) error {
		if fired.Add(1) == 1 {
			done <- struct{}{}
		}
		return nil
	}))
	require.NoError(t, err)

	s.Start(ctx)
	defer func() { _ = s.Stop(context.Background()) }()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("scheduled job did not fire in time")
	}

	rec, ok := s.Record("tick")
	require.True(t, ok)
	require.GreaterOrEqual(t, rec.Runs, uint64(1))
}

func TestSchedulerScheduleValidatesInputs(t *testing.T) {
	t.Parallel()
	s := jobs.New(quietLogger())
	_, err := s.Schedule("@every 1s", "", jobs.JobFunc(func(context.Context) error { return nil }))
	require.Error(t, err)

	_, err = s.Schedule("@every 1s", "x", nil)
	require.Error(t, err)

	_, err = s.Schedule("not a cron spec", "x", jobs.JobFunc(func(context.Context) error { return nil }))
	require.Error(t, err)
}

func TestSchedulerRunOnceTracksSuccessAndFailure(t *testing.T) {
	t.Parallel()
	s := jobs.New(quietLogger())

	require.NoError(t, s.RunOnce(context.Background(), "ok", jobs.JobFunc(func(context.Context) error { return nil })))
	rec, ok := s.Record("ok")
	require.True(t, ok)
	require.Equal(t, jobs.StatusSuccess, rec.Status)
	require.Equal(t, uint64(1), rec.Runs)
	require.Equal(t, uint64(0), rec.Failures)
	require.False(t, rec.LastStarted.IsZero())
	require.False(t, rec.LastEnded.IsZero())

	boom := errors.New("boom")
	err := s.RunOnce(context.Background(), "bad", jobs.JobFunc(func(context.Context) error { return boom }))
	require.Error(t, err)
	require.ErrorIs(t, err, boom)
	rec, ok = s.Record("bad")
	require.True(t, ok)
	require.Equal(t, jobs.StatusFailed, rec.Status)
	require.Equal(t, "boom", rec.LastError)
	require.Equal(t, uint64(1), rec.Failures)
}

func TestSchedulerRunOnceRejectsBadInput(t *testing.T) {
	t.Parallel()
	s := jobs.New(nil)
	require.Error(t, s.RunOnce(context.Background(), "", jobs.JobFunc(func(context.Context) error { return nil })))
	require.Error(t, s.RunOnce(context.Background(), "x", nil))
}

func TestSchedulerOverlapProtection(t *testing.T) {
	t.Parallel()
	s := jobs.New(quietLogger())

	started := make(chan struct{})
	release := make(chan struct{})
	go func() {
		_ = s.RunOnce(context.Background(), "slow", jobs.JobFunc(func(ctx context.Context) error {
			close(started)
			<-release
			return nil
		}))
	}()
	<-started

	err := s.RunOnce(context.Background(), "slow", jobs.JobFunc(func(context.Context) error { return nil }))
	require.ErrorIs(t, err, jobs.ErrOverlapping)

	close(release)

	require.Eventually(t, func() bool {
		rec, _ := s.Record("slow")
		return rec.Status == jobs.StatusSuccess
	}, time.Second, 10*time.Millisecond)
}

func TestSchedulerCronOverlapSkipped(t *testing.T) {
	t.Parallel()
	s := jobs.New(quietLogger())

	var hits atomic.Int32
	release := make(chan struct{})
	closeOnce := sync.OnceFunc(func() { close(release) })
	defer closeOnce()

	_, err := s.Schedule("@every 20ms", "slow-cron", jobs.JobFunc(func(ctx context.Context) error {
		hits.Add(1)
		<-release
		return nil
	}))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)
	defer func() { _ = s.Stop(context.Background()) }()

	time.Sleep(200 * time.Millisecond)
	closeOnce()
	require.LessOrEqual(t, hits.Load(), int32(2), "overlap must skip subsequent fires")
}

func TestSchedulerStopRespectsContext(t *testing.T) {
	t.Parallel()
	s := jobs.New(quietLogger())

	release := make(chan struct{})
	defer close(release)

	_, err := s.Schedule("@every 10ms", "slow", jobs.JobFunc(func(ctx context.Context) error {
		<-release
		return nil
	}))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	require.Eventually(t, func() bool {
		rec, ok := s.Record("slow")
		return ok && rec.Status == jobs.StatusRunning
	}, time.Second, 5*time.Millisecond, "expected the scheduled job to start running")

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stopCancel()
	err = s.Stop(stopCtx)
	require.Error(t, err)
}

func TestSchedulerRecords(t *testing.T) {
	t.Parallel()
	s := jobs.New(quietLogger())
	require.NoError(t, s.RunOnce(context.Background(), "a", jobs.JobFunc(func(context.Context) error { return nil })))
	require.NoError(t, s.RunOnce(context.Background(), "b", jobs.JobFunc(func(context.Context) error { return nil })))
	require.Len(t, s.Records(), 2)

	_, ok := s.Record("missing")
	require.False(t, ok)
}
