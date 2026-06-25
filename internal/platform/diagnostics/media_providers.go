package diagnostics

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// SessionCounter exposes active transcode session counts for health checks.
type SessionCounter interface {
	ActiveSessions() int64
	HWActiveSessions() int64
}

// DiskSpaceProvider reports free space for a data directory.
type DiskSpaceProvider struct {
	name    string
	dataDir string
}

// NewDiskSpaceProvider builds a non-critical disk space check for dir.
func NewDiskSpaceProvider(dataDir string) *DiskSpaceProvider {
	return &DiskSpaceProvider{name: "disk", dataDir: dataDir}
}

func (p *DiskSpaceProvider) Name() string   { return p.name }
func (p *DiskSpaceProvider) Critical() bool { return false }

func (p *DiskSpaceProvider) Check(_ context.Context) Check {
	if p.dataDir == "" {
		return Check{Name: p.name, Status: StatusDegraded, Detail: "data dir not configured"}
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(p.dataDir, &stat); err != nil {
		return Check{Name: p.name, Status: StatusDegraded, Detail: fmt.Sprintf("stat failed: %v", err)}
	}
	free := stat.Bavail * uint64(stat.Bsize)
	total := stat.Blocks * uint64(stat.Bsize)
	freeGB := float64(free) / (1 << 30)
	totalGB := float64(total) / (1 << 30)
	status := StatusOK
	if freeGB < 1 {
		status = StatusDegraded
	}
	return Check{
		Name:   p.name,
		Status: status,
		Detail: fmt.Sprintf("%.1f GiB free / %.1f GiB total", freeGB, totalGB),
	}
}

// FFmpegProvider verifies ffmpeg and ffprobe binaries are present.
type FFmpegProvider struct {
	ffmpegBin  string
	ffprobeBin string
}

// NewFFmpegProvider builds a non-critical media toolchain check.
func NewFFmpegProvider(ffmpegBin, ffprobeBin string) *FFmpegProvider {
	return &FFmpegProvider{ffmpegBin: ffmpegBin, ffprobeBin: ffprobeBin}
}

func (p *FFmpegProvider) Name() string   { return "ffmpeg" }
func (p *FFmpegProvider) Critical() bool { return false }

func (p *FFmpegProvider) Check(_ context.Context) Check {
	if err := probeBinary(p.ffmpegBin); err != nil {
		return Check{Name: p.Name(), Status: StatusDegraded, Detail: fmt.Sprintf("ffmpeg: %v", err)}
	}
	if err := probeBinary(p.ffprobeBin); err != nil {
		return Check{Name: p.Name(), Status: StatusDegraded, Detail: fmt.Sprintf("ffprobe: %v", err)}
	}
	return Check{Name: p.Name(), Status: StatusOK, Detail: "ffmpeg and ffprobe available"}
}

func probeBinary(bin string) error {
	if bin == "" {
		return fmt.Errorf("path not configured")
	}
	if _, err := exec.LookPath(bin); err != nil {
		if _, statErr := os.Stat(bin); statErr != nil {
			return fmt.Errorf("%q not found", bin)
		}
	}
	return nil
}

// TranscodeProvider reports active software and hardware transcode sessions.
type TranscodeProvider struct {
	counter SessionCounter
}

// NewTranscodeProvider wires streaming session metrics into diagnostics.
func NewTranscodeProvider(counter SessionCounter) *TranscodeProvider {
	return &TranscodeProvider{counter: counter}
}

func (p *TranscodeProvider) Name() string   { return "transcode" }
func (p *TranscodeProvider) Critical() bool { return false }

func (p *TranscodeProvider) Check(_ context.Context) Check {
	if p.counter == nil {
		return Check{Name: p.Name(), Status: StatusDegraded, Detail: "streaming service unavailable"}
	}
	sw := p.counter.ActiveSessions()
	hw := p.counter.HWActiveSessions()
	return Check{
		Name:   p.Name(),
		Status: StatusOK,
		Detail: fmt.Sprintf("%d active sessions (%d hw)", sw, hw),
	}
}
