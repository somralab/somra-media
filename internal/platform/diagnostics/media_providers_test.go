package diagnostics_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/platform/diagnostics"
)

type fakeSessions struct {
	sw, hw int64
}

func (f fakeSessions) ActiveSessions() int64   { return f.sw }
func (f fakeSessions) HWActiveSessions() int64 { return f.hw }

func TestDiskSpaceProvider(t *testing.T) {
	dir := t.TempDir()
	p := diagnostics.NewDiskSpaceProvider(dir)
	c := p.Check(context.Background())
	assert.Equal(t, "disk", c.Name)
	assert.Equal(t, diagnostics.StatusOK, c.Status)
	assert.Contains(t, c.Detail, "GiB free")
}

func TestDiskSpaceProvider_EmptyDir(t *testing.T) {
	p := diagnostics.NewDiskSpaceProvider("")
	c := p.Check(context.Background())
	assert.Equal(t, diagnostics.StatusDegraded, c.Status)
}

func TestFFmpegProvider(t *testing.T) {
	p := diagnostics.NewFFmpegProvider("ffmpeg", "ffprobe")
	c := p.Check(context.Background())
	if c.Status == diagnostics.StatusOK {
		assert.Contains(t, c.Detail, "available")
		return
	}
	assert.Equal(t, diagnostics.StatusDegraded, c.Status)
}

func TestTranscodeProvider(t *testing.T) {
	p := diagnostics.NewTranscodeProvider(fakeSessions{sw: 2, hw: 1})
	c := p.Check(context.Background())
	assert.Equal(t, diagnostics.StatusOK, c.Status)
	assert.Contains(t, c.Detail, "2 active")
}

func TestTranscodeProvider_Nil(t *testing.T) {
	p := diagnostics.NewTranscodeProvider(nil)
	c := p.Check(context.Background())
	assert.Equal(t, diagnostics.StatusDegraded, c.Status)
}

func TestDiskSpaceProvider_InvalidPath(t *testing.T) {
	p := diagnostics.NewDiskSpaceProvider(filepath.Join(t.TempDir(), "missing", "nested"))
	c := p.Check(context.Background())
	require.NotEmpty(t, c.Detail)
}
