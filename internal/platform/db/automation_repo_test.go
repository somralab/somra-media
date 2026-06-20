package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAutomationRepo_HandoffAndDownload(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "auto-user", "hash", []string{"user"})
	require.NoError(t, err)

	reqRepo := NewRequestRepo(d.Querier())
	requestID, err := reqRepo.Create(ctx, Request{
		UserID:            userID,
		MediaKind:         RequestMediaKindMovie,
		Provider:          "tmdb",
		ExternalID:        "auto-1",
		Title:             "Auto Movie",
		QualityResolution: RequestQualityAny,
		Status:            RequestStatusApproved,
	})
	require.NoError(t, err)

	repo := NewAutomationRepo(d.Querier())
	handoffID, err := repo.RecordHandoff(ctx, requestID)
	require.NoError(t, err)
	require.Positive(t, handoffID)

	pending, err := repo.ListPendingHandoffs(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pending, 1)

	hID := handoffID
	dlID, err := repo.CreateDownload(ctx, AutomationDownload{
		RequestID:        &requestID,
		HandoffID:        &hID,
		ClientInstanceID: 1,
		ClientDownloadID: "stub-dl-1",
		ReleaseID:        "rel-1",
		Title:            "Test",
		Protocol:         "torrent",
		Status:           AutomationDownloadQueued,
	})
	require.NoError(t, err)
	require.Positive(t, dlID)

	require.NoError(t, repo.UpdateDownloadProgress(ctx, dlID, AutomationDownloadCompleted, 1, "/downloads/x", ""))
	require.NoError(t, repo.UpdateHandoffStatus(ctx, handoffID, HandoffCompleted, ""))

	profiles, err := repo.ListQualityProfiles(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, profiles)
}
