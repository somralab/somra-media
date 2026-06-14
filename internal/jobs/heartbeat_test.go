package jobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
)

func TestHeartbeatRunIncrementsTicks(t *testing.T) {
	t.Parallel()
	hb := jobs.NewHeartbeat(quietLogger())
	require.Zero(t, hb.Ticks())
	require.NoError(t, hb.Run(context.Background()))
	require.Equal(t, uint64(1), hb.Ticks())
	require.NoError(t, hb.Run(context.Background()))
	require.Equal(t, uint64(2), hb.Ticks())
}

func TestHeartbeatNilLoggerDefaults(t *testing.T) {
	t.Parallel()
	hb := jobs.NewHeartbeat(nil)
	require.NoError(t, hb.Run(context.Background()))
}

func TestHeartbeatWithScheduler(t *testing.T) {
	t.Parallel()
	s := jobs.New(quietLogger())
	hb := jobs.NewHeartbeat(quietLogger())

	_, err := s.Schedule("@every 50ms", jobs.HeartbeatJobName, hb)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)
	defer func() { _ = s.Stop(context.Background()) }()

	require.Eventually(t, func() bool {
		return hb.Ticks() >= 1
	}, 2*time.Second, 10*time.Millisecond, "heartbeat should fire under sped-up spec")

	rec, ok := s.Record(jobs.HeartbeatJobName)
	require.True(t, ok)
	require.GreaterOrEqual(t, rec.Runs, uint64(1))
}
