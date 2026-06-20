package download

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/automation/automationtest"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
)

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

	autoRepo := db.NewAutomationRepo(d.Querier())
	reqID := automationtest.CreateApprovedRequest(t, d, "Monitor Movie")
	handoffID, err := autoRepo.RecordHandoff(ctx, reqID)
	require.NoError(t, err)

	client, err := mgr.DownloadClient(ctx, clientID)
	require.NoError(t, err)
	item, err := client.Add(ctx, plugin.AddRequest{ReleaseID: "rel-1", Title: "Monitor Movie"})
	require.NoError(t, err)

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
		Requests: db.NewRequestRepo(d.Querier()),
		Manager:  mgr,
	}
	require.NoError(t, m.Poll(ctx))

	row, err := autoRepo.GetDownloadByID(ctx, dlID)
	require.NoError(t, err)
	require.Equal(t, db.AutomationDownloadCompleted, row.Status)
}
