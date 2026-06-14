package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestScanner_IncrementalSkipsUnchanged(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "Movie (2020).mkv")
	require.NoError(t, os.WriteFile(path, []byte("fake"), 0o644))

	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 4})
	defer queue.Close()
	scanner := NewScanner(ScannerConfig{DB: d, Prober: stubProber{}, Progress: NoopProgressPublisher{}})
	svc := NewService(ServiceConfig{DB: d, Queue: queue, Scanner: scanner})
	lib, err := svc.CreateLibrary(ctx, "T", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)

	run1, task1, err := svc.TriggerScan(ctx, lib.ID, db.ScanFull)
	require.NoError(t, err)
	waitScan(t, ctx, queue, svc, task1, run1)

	run2, task2, err := svc.TriggerScan(ctx, lib.ID, db.ScanIncremental)
	require.NoError(t, err)
	waitScan(t, ctx, queue, svc, task2, run2)
}

func TestScanner_HashFiles(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "a.mkv")
	require.NoError(t, os.WriteFile(path, []byte("content"), 0o644))

	scanner := NewScanner(ScannerConfig{
		DB: d, Prober: stubProber{}, Progress: NoopProgressPublisher{}, HashFiles: true,
	})
	libRepo := db.NewLibraryRepo(d.Querier())
	lib, err := libRepo.Create(ctx, "H", db.LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	runID, err := db.NewScanRepo(d.Querier()).CreateRun(ctx, lib.ID, db.ScanFull, "")
	require.NoError(t, err)
	require.NoError(t, scanner.run(ctx, lib.ID, db.ScanFull, runID))
}

func waitScan(t *testing.T, ctx context.Context, queue *jobs.MemoryQueue, svc *Service, task jobs.TaskID, runID int64) {
	t.Helper()
	for i := 0; i < 100; i++ {
		st, err := queue.Status(ctx, task)
		require.NoError(t, err)
		run, err := svc.GetScanRun(ctx, runID)
		require.NoError(t, err)
		if st == jobs.TaskSucceeded && run.Status == db.ScanSucceeded {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("scan did not finish")
}
