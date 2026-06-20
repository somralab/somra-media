package importsvc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/automation/automationtest"
	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/library"
	"github.com/somralab/somra-media/internal/platform/db"
)

type testProber struct{}

func (testProber) Probe(context.Context, string) (library.ProbeResult, error) {
	return library.ProbeResult{}, nil
}

func TestImportCompleted_Validation(t *testing.T) {
	ctx := context.Background()
	var nilSvc *Service
	require.Error(t, nilSvc.ImportCompleted(ctx, "/path"))

	svc := &Service{}
	require.Error(t, svc.ImportCompleted(ctx, ""))
}

func TestImportCompleted_TriggersScan(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)

	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 4})
	t.Cleanup(func() { queue.Close() })

	libSvc := library.NewService(library.ServiceConfig{
		DB:      d,
		Queue:   queue,
		Scanner: library.NewScanner(library.ScannerConfig{DB: d, Prober: testProber{}, Progress: library.NoopProgressPublisher{}}),
	})
	libDir := t.TempDir()
	_, err := libSvc.CreateLibrary(ctx, "Import Lib", db.LibraryKindMovie, []string{libDir}, false)
	require.NoError(t, err)

	savePath := filepath.Join(t.TempDir(), "movie.mkv")
	require.NoError(t, os.WriteFile(savePath, []byte("x"), 0o644))

	svc := &Service{Library: libSvc, ImportRoot: t.TempDir()}
	require.NoError(t, svc.ImportCompleted(ctx, savePath))
}
