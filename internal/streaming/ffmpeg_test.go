package streaming

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProcessManager_StartAndStop(t *testing.T) {
	pm := NewProcessManager(ProcessManagerConfig{MaxConcurrent: 1, FFmpegBin: "/bin/sleep"})
	require.NoError(t, pm.Acquire(context.Background()))
	_, err := pm.Start(context.Background(), "sleep-session", []string{"2"})
	require.NoError(t, err)
	require.Equal(t, 1, pm.RunningCount())
	pm.Stop("sleep-session")
	time.Sleep(50 * time.Millisecond)
	pm.Release()
}
