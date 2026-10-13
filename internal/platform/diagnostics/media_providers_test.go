package diagnostics_test

import (
	"context"
	"os"
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

// fakeBinary creates an executable file so the ffmpeg check does not depend
// on what is installed on the machine running the tests.
func fakeBinary(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755))
	return path
}

func TestFFmpegProvider(t *testing.T) {
	ffmpeg := fakeBinary(t, "ffmpeg")
	ffprobe := fakeBinary(t, "ffprobe")
	missing := filepath.Join(t.TempDir(), "nope")

	cases := []struct {
		name            string
		ffmpeg, ffprobe string
		status          diagnostics.Status
		detail          string
	}{
		{"both present", ffmpeg, ffprobe, diagnostics.StatusOK, "ffmpeg and ffprobe available"},
		{"ffmpeg missing", missing, ffprobe, diagnostics.StatusDegraded, "ffmpeg: "},
		{"ffprobe missing", ffmpeg, missing, diagnostics.StatusDegraded, "ffprobe: "},
		{"not configured", "", ffprobe, diagnostics.StatusDegraded, "path not configured"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := diagnostics.NewFFmpegProvider(tc.ffmpeg, tc.ffprobe).Check(context.Background())
			assert.Equal(t, "ffmpeg", c.Name)
			assert.Equal(t, tc.status, c.Status)
			assert.Contains(t, c.Detail, tc.detail)
		})
	}
}

func TestMediaProviders_DegradeButNeverTakeSystemDown(t *testing.T) {
	reg := diagnostics.NewRegistry()
	reg.Register(diagnostics.NewDiskSpaceProvider(""))
	reg.Register(diagnostics.NewFFmpegProvider("", ""))
	reg.Register(diagnostics.NewTranscodeProvider(nil))

	snap := reg.Run(context.Background())

	require.Equal(t, diagnostics.StatusDegraded, snap.Overall,
		"a missing media toolchain degrades the server; it must not report it as down")
	names := make([]string, 0, len(snap.Checks))
	for _, c := range snap.Checks {
		names = append(names, c.Name)
		assert.False(t, c.Critical, "%s must be non-critical", c.Name)
		assert.Equal(t, diagnostics.StatusDegraded, c.Status, c.Name)
	}
	assert.Equal(t, []string{"disk", "ffmpeg", "transcode"}, names)
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
