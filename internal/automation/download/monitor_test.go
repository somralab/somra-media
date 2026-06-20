package download

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/automation/automationtest"
	"github.com/somralab/somra-media/internal/automation/importsvc"
	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/library"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
)

type monitorTestProber struct{}

func (monitorTestProber) Probe(context.Context, string) (library.ProbeResult, error) {
	return library.ProbeResult{}, nil
}

func TestMapStatus(t *testing.T) {
	require.Equal(t, db.AutomationDownloadCompleted, mapStatus(plugin.DownloadStatusCompleted))
	require.Equal(t, db.AutomationDownloadDownloading, mapStatus(plugin.DownloadStatusDownloading))
	require.Equal(t, db.AutomationDownloadQueued, mapStatus(plugin.DownloadStatusQueued))
	require.Equal(t, db.AutomationDownloadPaused, mapStatus(plugin.DownloadStatusPaused))
	require.Equal(t, db.AutomationDownloadFailed, mapStatus(plugin.DownloadStatusFailed))
	require.Equal(t, db.AutomationDownloadQueued, mapStatus(plugin.DownloadStatus("unknown")))
}

func TestMonitor_PollRequiresDeps(t *testing.T) {
	var m *Monitor
	require.Error(t, m.Poll(context.Background()))
}

func TestMonitor_PollCompletesDownload(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	clientID := automationtest.CreateStubDownloadClient(t, mgr, "monitor-dl")

	queue := jobs.NewMemoryQueue(jobs.MemoryQueueConfig{Workers: 1, Buffer: 4})
	t.Cleanup(func() { queue.Close() })
	libSvc := library.NewService(library.ServiceConfig{
		DB:      d,
		Queue:   queue,
		Scanner: library.NewScanner(library.ScannerConfig{DB: d, Prober: monitorTestProber{}, Progress: library.NoopProgressPublisher{}}),
	})
	libDir := t.TempDir()
	_, err := libSvc.CreateLibrary(ctx, "Monitor Lib", db.LibraryKindMovie, []string{libDir}, false)
	require.NoError(t, err)

	autoRepo := db.NewAutomationRepo(d.Querier())
	reqRepo := db.NewRequestRepo(d.Querier())
	reqID := automationtest.CreateApprovedRequest(t, d, "Monitor Movie")
	handoffID, err := autoRepo.RecordHandoff(ctx, reqID)
	require.NoError(t, err)

	client, err := mgr.DownloadClient(ctx, clientID)
	require.NoError(t, err)
	savePath := filepath.Join(t.TempDir(), "done.mkv")
	require.NoError(t, os.WriteFile(savePath, []byte("x"), 0o644))
	item, err := client.Add(ctx, plugin.AddRequest{ReleaseID: "rel-1", Title: "Monitor Movie"})
	require.NoError(t, err)
	// Stub stores save path from prefix; override via second add with known path in title is not supported.
	// Status returns completed with stored save path from Add.
	_ = savePath

	dlID, err := autoRepo.CreateDownload(ctx, db.AutomationDownload{
		RequestID:        &reqID,
		HandoffID:        &handoffID,
		ClientInstanceID: clientID,
		ClientDownloadID: item.DownloadID,
		ReleaseID:        "rel-1",
		Title:            "Monitor Movie",
		Protocol:         "torrent",
		Status:           db.AutomationDownloadQueued,
	})
	require.NoError(t, err)

	m := &Monitor{
		AutoRepo: autoRepo,
		Requests: reqRepo,
		Manager:  mgr,
		Import:   &importsvc.Service{Library: libSvc},
	}
	require.NoError(t, m.Poll(ctx))

	row, err := autoRepo.GetDownloadByID(ctx, dlID)
	require.NoError(t, err)
	require.Equal(t, db.AutomationDownloadCompleted, row.Status)
}
