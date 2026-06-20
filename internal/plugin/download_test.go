package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDownloadClient struct {
	id      string
	items   map[string]DownloadItem
	addErr  error
	statErr error
}

func (m *mockDownloadClient) ID() string              { return m.id }
func (m *mockDownloadClient) Type() PluginType        { return PluginTypeDownloadClient }
func (m *mockDownloadClient) ContractVersion() string { return ContractVersion }

func (m *mockDownloadClient) Add(_ context.Context, req AddRequest) (DownloadItem, error) {
	if m.addErr != nil {
		return DownloadItem{}, m.addErr
	}
	item := DownloadItem{
		DownloadID: "dl-new",
		ClientID:   m.id,
		ReleaseID:  req.ReleaseID,
		Status:     DownloadStatusQueued,
		Progress:   0,
	}
	if m.items == nil {
		m.items = make(map[string]DownloadItem)
	}
	m.items[item.DownloadID] = item
	return item, nil
}

func (m *mockDownloadClient) Status(_ context.Context, downloadID string) (DownloadItem, error) {
	if m.statErr != nil {
		return DownloadItem{}, m.statErr
	}
	item, ok := m.items[downloadID]
	if !ok {
		return DownloadItem{}, ErrUnsupportedCapability
	}
	return item, nil
}

var _ DownloadClient = (*mockDownloadClient)(nil)

func TestDownloadClientAddAndStatus(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &mockDownloadClient{id: "qbittorrent-1"}

	require.NoError(t, ValidateContract(client))

	added, err := client.Add(ctx, AddRequest{
		ReleaseID: "rel-99",
		Title:     "Movie 1080p",
		Category:  "movies",
		Priority:  1,
	})
	require.NoError(t, err)
	assert.Equal(t, DownloadStatusQueued, added.Status)
	assert.Equal(t, "rel-99", added.ReleaseID)
	assert.Equal(t, "dl-new", added.DownloadID)

	added.Status = DownloadStatusDownloading
	added.Progress = 42.5
	added.DownloadedBytes = 1_000_000
	added.TotalBytes = 2_000_000
	client.items[added.DownloadID] = added

	status, err := client.Status(ctx, added.DownloadID)
	require.NoError(t, err)
	assert.Equal(t, DownloadStatusDownloading, status.Status)
	assert.InDelta(t, 42.5, status.Progress, 0.001)
	assert.Equal(t, int64(1_000_000), status.DownloadedBytes)
}

func TestDownloadClientCompletedItem(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	completedAt := time.Date(2024, 7, 1, 18, 0, 0, 0, time.UTC)
	client := &mockDownloadClient{
		id: "sabnzbd-1",
		items: map[string]DownloadItem{
			"dl-done": {
				DownloadID:  "dl-done",
				ClientID:    "sabnzbd-1",
				Status:      DownloadStatusCompleted,
				Progress:    100,
				SavePath:    "/downloads/movie.mkv",
				CompletedAt: &completedAt,
			},
		},
	}

	status, err := client.Status(ctx, "dl-done")
	require.NoError(t, err)
	assert.Equal(t, DownloadStatusCompleted, status.Status)
	assert.Equal(t, "/downloads/movie.mkv", status.SavePath)
	require.NotNil(t, status.CompletedAt)
}

func TestDownloadClientStatusNotFound(t *testing.T) {
	t.Parallel()
	client := &mockDownloadClient{id: "empty"}
	_, err := client.Status(context.Background(), "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedCapability)
}

func TestDownloadStatusConstants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, DownloadStatus("queued"), DownloadStatusQueued)
	assert.Equal(t, DownloadStatus("downloading"), DownloadStatusDownloading)
	assert.Equal(t, DownloadStatus("paused"), DownloadStatusPaused)
	assert.Equal(t, DownloadStatus("completed"), DownloadStatusCompleted)
	assert.Equal(t, DownloadStatus("failed"), DownloadStatusFailed)
}
