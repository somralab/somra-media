package streaming

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
)

// ProcessManager limits concurrent ffmpeg processes and ensures cleanup.
type ProcessManager struct {
	maxConcurrent int
	sem           chan struct{}
	mu            sync.Mutex
	procs         map[string]*exec.Cmd
	ffmpegBin     string
}

// ProcessManagerConfig configures ffmpeg process limits.
type ProcessManagerConfig struct {
	MaxConcurrent int
	FFmpegBin     string
}

// NewProcessManager returns a manager with concurrency cap.
func NewProcessManager(cfg ProcessManagerConfig) *ProcessManager {
	limit := cfg.MaxConcurrent
	if limit <= 0 {
		limit = 2
	}
	bin := cfg.FFmpegBin
	if bin == "" {
		bin = "ffmpeg"
	}
	return &ProcessManager{
		maxConcurrent: limit,
		sem:           make(chan struct{}, limit),
		procs:         make(map[string]*exec.Cmd),
		ffmpegBin:     bin,
	}
}

// Acquire blocks until a transcode slot is available or ctx is cancelled.
func (m *ProcessManager) Acquire(ctx context.Context) error {
	select {
	case m.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees a transcode slot.
func (m *ProcessManager) Release() {
	select {
	case <-m.sem:
	default:
	}
}

// Start launches ffmpeg with args tied to sessionID.
func (m *ProcessManager) Start(ctx context.Context, sessionID string, args []string) (*exec.Cmd, error) {
	allArgs := append([]string{"-hide_banner", "-loglevel", "error", "-y"}, args...)
	cmd := exec.CommandContext(ctx, m.ffmpegBin, allArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	m.mu.Lock()
	if old, ok := m.procs[sessionID]; ok {
		m.kill(old)
	}
	m.procs[sessionID] = cmd
	m.mu.Unlock()

	if err := cmd.Start(); err != nil {
		m.Release()
		return nil, fmt.Errorf("ffmpeg start session %q: %w", sessionID, err)
	}

	go func() {
		_ = cmd.Wait()
		m.mu.Lock()
		delete(m.procs, sessionID)
		m.mu.Unlock()
		m.Release()
	}()

	return cmd, nil
}

// Stop terminates ffmpeg for sessionID.
func (m *ProcessManager) Stop(sessionID string) {
	m.mu.Lock()
	cmd, ok := m.procs[sessionID]
	m.mu.Unlock()
	if ok {
		m.kill(cmd)
	}
}

// RunningCount returns active ffmpeg processes.
func (m *ProcessManager) RunningCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.procs)
}

func (m *ProcessManager) kill(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	if runtime.GOOS != "windows" {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		return
	}
	_ = cmd.Process.Kill()
}

// RemoveDir deletes a session cache directory.
func RemoveDir(path string) error {
	if path == "" {
		return nil
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("streaming remove cache %q: %w", path, err)
	}
	return nil
}
