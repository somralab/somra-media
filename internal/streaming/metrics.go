package streaming

import (
	"sync/atomic"
)

// Metrics tracks playback/transcode counters for diagnostics.
type Metrics struct {
	activeSessions  atomic.Int64
	transcodeErrors atomic.Int64
	transcodeStarts atomic.Int64
	queueDepth      atomic.Int64
	hwActive        atomic.Int64
	hwStarts        atomic.Int64
	hwErrors        atomic.Int64
	hwFallbacks     atomic.Int64
}

// NewMetrics returns zeroed streaming metrics.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// ActiveSessions returns the current active session count.
func (m *Metrics) ActiveSessions() int64 {
	return m.activeSessions.Load()
}

// TranscodeErrors returns cumulative transcode failures.
func (m *Metrics) TranscodeErrors() int64 {
	return m.transcodeErrors.Load()
}

// TranscodeStarts returns cumulative transcode starts.
func (m *Metrics) TranscodeStarts() int64 {
	return m.transcodeStarts.Load()
}

// QueueDepth returns pending transcode queue depth.
func (m *Metrics) QueueDepth() int64 {
	return m.queueDepth.Load()
}

// HWActiveSessions returns active hardware transcode sessions.
func (m *Metrics) HWActiveSessions() int64 {
	return m.hwActive.Load()
}

// HWStarts returns cumulative HW transcode starts.
func (m *Metrics) HWStarts() int64 {
	return m.hwStarts.Load()
}

// HWErrors returns cumulative HW transcode errors.
func (m *Metrics) HWErrors() int64 {
	return m.hwErrors.Load()
}

// HWFallbacks returns HW→SW fallback count.
func (m *Metrics) HWFallbacks() int64 {
	return m.hwFallbacks.Load()
}

func (m *Metrics) incActive()       { m.activeSessions.Add(1) }
func (m *Metrics) decActive()       { m.activeSessions.Add(-1) }
func (m *Metrics) incErrors()       { m.transcodeErrors.Add(1) }
func (m *Metrics) incStarts()       { m.transcodeStarts.Add(1) }
func (m *Metrics) setQueue(n int64) { m.queueDepth.Store(n) }
func (m *Metrics) incHWActive()     { m.hwActive.Add(1) }
func (m *Metrics) decHWActive()     { m.hwActive.Add(-1) }
func (m *Metrics) incHWStarts()     { m.hwStarts.Add(1) }
func (m *Metrics) incHWErrors()     { m.hwErrors.Add(1) }
func (m *Metrics) incHWFallbacks()  { m.hwFallbacks.Add(1) }
