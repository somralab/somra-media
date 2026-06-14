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

func TestService_CRUDAndHistory(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	dir := t.TempDir()
	dir2 := t.TempDir()

	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	defer queue.Close()
	scanner := NewScanner(ScannerConfig{DB: d, Prober: stubProber{}, Progress: NoopProgressPublisher{}})
	svc := NewService(ServiceConfig{DB: d, Queue: queue, Scanner: scanner})

	lib, err := svc.CreateLibrary(ctx, "A", db.LibraryKindTV, []string{dir}, true)
	require.NoError(t, err)

	libs, err := svc.ListLibraries(ctx)
	require.NoError(t, err)
	require.Len(t, libs, 1)

	lib, err = svc.UpdateLibrary(ctx, lib.ID, "B", []string{dir, dir2}, false)
	require.NoError(t, err)
	require.Equal(t, "B", lib.Name)

	runID, _, err := svc.TriggerScan(ctx, lib.ID, db.ScanIncremental)
	require.NoError(t, err)
	history, err := svc.ListScanHistory(ctx, lib.ID, 5)
	require.NoError(t, err)
	require.NotEmpty(t, history)

	require.NoError(t, svc.DeleteLibrary(ctx, lib.ID))
	_, err = svc.GetLibrary(ctx, lib.ID)
	require.Error(t, err)
	_ = runID
}

func TestWatcher_StartStop(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	dir := t.TempDir()
	libRepo := db.NewLibraryRepo(d.Querier())
	created, err := libRepo.Create(ctx, "W", db.LibraryKindMovie, []string{dir}, true)
	require.NoError(t, err)

	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 2})
	defer queue.Close()
	scanner := NewScanner(ScannerConfig{DB: d, Prober: stubProber{}})
	svc := NewService(ServiceConfig{DB: d, Queue: queue, Scanner: scanner})

	_, err = svc.CreateLibrary(ctx, "Other", db.LibraryKindMovie, []string{t.TempDir()}, true)
	require.NoError(t, err)

	w := NewWatcher(nil, created, queue, scanner, 5*time.Millisecond)
	w.Start(ctx)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "watch.mkv"), []byte("x"), 0o644))
	time.Sleep(50 * time.Millisecond)
	w.Stop()
}

func TestProber_ProbeMissingFile(t *testing.T) {
	p := NewProber("")
	_, err := p.Probe(context.Background(), filepath.Join(t.TempDir(), "missing.mkv"))
	require.Error(t, err)
}

func TestIsMediaFile(t *testing.T) {
	require.True(t, IsMediaFile("a.mkv"))
	require.False(t, IsMediaFile("a.txt"))
}
