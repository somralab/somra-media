package library

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/platform/db"
)

type stubProber struct{}

func (stubProber) Probe(_ context.Context, _ string) (ProbeResult, error) {
	return ProbeResult{DurationMs: 1000, Container: "matroska", VideoCodec: "h264"}, nil
}

func TestService_CreateAndScan(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Movie (2020).mkv"), []byte("fake"), 0o644))

	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 4})
	defer queue.Close()

	scanner := NewScanner(ScannerConfig{DB: d, Prober: stubProber{}, Progress: NoopProgressPublisher{}})
	svc := NewService(ServiceConfig{DB: d, Queue: queue, Scanner: scanner})

	lib, err := svc.CreateLibrary(ctx, "Test", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	runID, taskID, err := svc.TriggerScan(ctx, lib.ID, db.ScanFull)
	require.NoError(t, err)
	assert.NotEmpty(t, taskID)
	assert.Positive(t, runID)

	var run db.ScanRun
	for i := 0; i < 100; i++ {
		st, err := queue.Status(ctx, taskID)
		require.NoError(t, err)
		run, err = svc.GetScanRun(ctx, runID)
		require.NoError(t, err)
		if st == jobs.TaskSucceeded && run.Status == db.ScanSucceeded {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, db.ScanSucceeded, run.Status)
	assert.Equal(t, 1, run.FilesDone)
}

func TestDiscoverMedia_FindsSupportedFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.mkv"), []byte("x"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("x"), 0o644))

	var found []string
	err := DiscoverMedia(context.Background(), []string{dir}, func(path string, _ fs.DirEntry) error {
		found = append(found, path)
		return nil
	})
	require.NoError(t, err)
	assert.Len(t, found, 1)
}

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	cfg := db.Default()
	cfg.DataDir = t.TempDir()
	d, err := db.Initialize(context.Background(), cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })
	return d
}
