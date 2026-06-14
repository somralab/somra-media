package streaming

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.incActive()
	m.incActive()
	m.decActive()
	m.incErrors()
	m.incStarts()
	m.setQueue(3)
	assert.Equal(t, int64(1), m.ActiveSessions())
	assert.Equal(t, int64(1), m.TranscodeErrors())
	assert.Equal(t, int64(1), m.TranscodeStarts())
	assert.Equal(t, int64(3), m.QueueDepth())
}

func TestWriteMediaPlaylist(t *testing.T) {
	t.Parallel()
	out := WriteMediaPlaylist(4, "init.mp4", []SegmentRef{{URI: "seg_00001.m4s", DurationSec: 4.0}})
	assert.Contains(t, out, "#EXT-X-ENDLIST")
	assert.Contains(t, out, "seg_00001.m4s")
}

func TestEnsureOutputDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir() + "/nested/session"
	assert.NoError(t, EnsureOutputDir(dir))
}

func TestSegmentName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "seg_00001.m4s", SegmentName(1))
}

func TestDefaultBrowserCapabilities(t *testing.T) {
	t.Parallel()
	caps := DefaultBrowserCapabilities()
	assert.Contains(t, caps.VideoCodecs, "h264")
}

func TestMediaProbe_EstimatedBitrate(t *testing.T) {
	t.Parallel()
	p := MediaProbe{VideoWidth: 1920, VideoHeight: 1080, DurationMs: 3600000, Bitrate: 5_000_000}
	assert.Equal(t, int64(5_000_000), p.EstimatedBitrate())
}

func TestProcessManager_AcquireRelease(t *testing.T) {
	t.Parallel()
	pm := NewProcessManager(ProcessManagerConfig{MaxConcurrent: 1})
	assert.NoError(t, pm.Acquire(context.Background()))
	pm.Release()
}

func TestRemoveDir_Empty(t *testing.T) {
	t.Parallel()
	assert.NoError(t, RemoveDir(""))
}

func TestWaitForManifest_Timeout(t *testing.T) {
	t.Parallel()
	err := WaitForManifest(t.TempDir(), 50*time.Millisecond)
	assert.Error(t, err)
}
