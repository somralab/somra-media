package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/platform/db"
)

type captureProgress struct {
	events []ProgressEvent
}

func (c *captureProgress) PublishScanProgress(_ context.Context, ev ProgressEvent) {
	c.events = append(c.events, ev)
}

func TestScanner_PublishesProgress(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "Movie (2020).mkv")
	require.NoError(t, os.WriteFile(path, []byte("fake"), 0o644))

	cap := &captureProgress{}
	scanner := NewScanner(ScannerConfig{DB: d, Prober: stubProber{}, Progress: cap})
	libRepo := db.NewLibraryRepo(d.Querier())
	lib, err := libRepo.Create(ctx, "P", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	runID, err := db.NewScanRepo(d.Querier()).CreateRun(ctx, lib.ID, db.ScanFull, "")
	require.NoError(t, err)
	require.NoError(t, scanner.run(ctx, lib.ID, db.ScanFull, runID))
	require.NotEmpty(t, cap.events)
	assert.Equal(t, "succeeded", cap.events[len(cap.events)-1].Status)
}

func TestProber_ParseJSON(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "ffprobe")
	payload := `{"format":{"format_name":"matroska","duration":"90.5"},"streams":[{"codec_type":"video","codec_name":"h264","width":1280,"height":720},{"codec_type":"audio","codec_name":"aac","channels":2},{"codec_type":"subtitle","codec_name":"subrip"}]}`
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho '"+payload+"'\n"), 0o755))

	media := filepath.Join(dir, "clip.mkv")
	require.NoError(t, os.WriteFile(media, []byte("x"), 0o644))

	p := NewProber(script)
	result, err := p.Probe(context.Background(), media)
	require.NoError(t, err)
	assert.Equal(t, int64(90500), result.DurationMs)
	assert.Equal(t, "h264", result.VideoCodec)
	assert.Equal(t, 1280, result.VideoWidth)
	assert.Equal(t, "aac", result.AudioCodec)
	assert.Equal(t, 2, result.AudioChannels)
	assert.Equal(t, 1, result.SubtitleCount)
}

func TestProber_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "ffprobe")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho not-json\n"), 0o755))
	media := filepath.Join(dir, "clip.mkv")
	require.NoError(t, os.WriteFile(media, []byte("x"), 0o644))

	p := NewProber(script)
	_, err := p.Probe(context.Background(), media)
	require.Error(t, err)
}

func TestNewProber_DefaultBinary(t *testing.T) {
	p := NewProber("")
	assert.Equal(t, "ffprobe", p.Binary)
}

func TestValidateRootPath_EmptyAndFile(t *testing.T) {
	_, err := ValidateRootPath("   ")
	require.Error(t, err)

	file := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))
	_, err = ValidateRootPath(file)
	require.Error(t, err)
}

func TestValidatePaths_Dedupes(t *testing.T) {
	dir := t.TempDir()
	got, err := validatePaths([]string{dir, dir})
	require.NoError(t, err)
	require.Len(t, got, 1)
}

func TestValidatePaths_RejectsEmpty(t *testing.T) {
	_, err := validatePaths(nil)
	require.Error(t, err)
}

func TestService_TriggerScanUnknownLibrary(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	defer queue.Close()
	scanner := NewScanner(ScannerConfig{DB: d, Prober: stubProber{}})
	svc := NewService(ServiceConfig{DB: d, Queue: queue, Scanner: scanner})

	_, _, err := svc.TriggerScan(ctx, 9999, db.ScanFull)
	require.Error(t, err)
}

func TestNewScanner_DefaultDependencies(t *testing.T) {
	d := openTestDB(t)
	scanner := NewScanner(ScannerConfig{DB: d})
	require.NotNil(t, scanner.prober)
	require.NotNil(t, scanner.progress)
	require.NotNil(t, scanner.logger)
}
