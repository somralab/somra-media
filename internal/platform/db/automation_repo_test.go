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

	handoffAgain, err := repo.RecordHandoff(ctx, requestID)
	require.NoError(t, err)
	require.Equal(t, handoffID, handoffAgain)

	pending, err := repo.ListPendingHandoffs(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pending, 1)

	pendingDefaultLimit, err := repo.ListPendingHandoffs(ctx, 0)
	require.NoError(t, err)
	require.Len(t, pendingDefaultLimit, 1)

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

	orphanID, err := repo.CreateDownload(ctx, AutomationDownload{
		ClientInstanceID: 2,
		ClientDownloadID: "orphan-dl",
		ReleaseID:        "rel-2",
		Title:            "Orphan",
		Protocol:         "torrent",
		Status:           AutomationDownloadQueued,
	})
	require.NoError(t, err)
	require.Positive(t, orphanID)

	allDownloads, err := repo.ListDownloads(ctx, 0)
	require.NoError(t, err)
	require.NotEmpty(t, allDownloads)

	activeBeforeComplete, err := repo.ListActiveDownloads(ctx, 0)
	require.NoError(t, err)
	require.NotEmpty(t, activeBeforeComplete)

	require.NoError(t, repo.UpdateDownloadProgress(ctx, dlID, AutomationDownloadCompleted, 1, "/downloads/x", "done"))
	require.NoError(t, repo.UpdateHandoffStatus(ctx, handoffID, HandoffCompleted, ""))

	active, err := repo.ListActiveDownloads(ctx, 10)
	require.NoError(t, err)
	require.Len(t, active, 1)
	require.Equal(t, orphanID, active[0].ID)

	profiles, err := repo.ListQualityProfiles(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, profiles)

	defaultProfile, err := repo.GetDefaultQualityProfile(ctx)
	require.NoError(t, err)
	require.True(t, defaultProfile.IsDefault)

	byName, err := repo.GetQualityProfileByName(ctx, defaultProfile.Name)
	require.NoError(t, err)
	require.Equal(t, defaultProfile.ID, byName.ID)

	customID, err := repo.CreateQualityProfile(ctx, "custom-test", `{"preferredResolutions":["720p"]}`, false)
	require.NoError(t, err)
	require.Positive(t, customID)

	emptySpecID, err := repo.CreateQualityProfile(ctx, "empty-spec", "", false)
	require.NoError(t, err)
	require.Positive(t, emptySpecID)

	got, err := repo.GetDownloadByID(ctx, dlID)
	require.NoError(t, err)
	require.Equal(t, dlID, got.ID)

	_, err = repo.GetDownloadByID(ctx, 999999)
	require.Error(t, err)

	_, err = repo.GetQualityProfileByName(ctx, "missing-profile")
	require.Error(t, err)

	_, err = repo.CreateQualityProfile(ctx, "", "{}", false)
	require.Error(t, err)

	_, err = repo.CreateQualityProfile(ctx, defaultProfile.Name, "{}", false)
	require.ErrorIs(t, err, ErrQualityProfileDuplicate)

	_, err = repo.CreateQualityProfile(ctx, "default-alt", "{}", true)
	require.NoError(t, err)

	_, err = repo.RecordHandoff(ctx, requestID)
	require.NoError(t, err)

	require.Error(t, repo.UpdateHandoffStatus(ctx, 999999, HandoffFailed, "missing"))
}

func TestAutomationRepo_GetDefaultQualityProfileNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	_, err := d.Querier().ExecContext(ctx, `DELETE FROM quality_profiles WHERE is_default = 1`)
	require.NoError(t, err)

	repo := NewAutomationRepo(d.Querier())
	_, err = repo.GetDefaultQualityProfile(ctx)
	require.ErrorIs(t, err, ErrQualityProfileNotFound)
}

func TestAutomationRepo_RecordHandoffInvalidRequest(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewAutomationRepo(d.Querier())
	_, err := repo.RecordHandoff(ctx, 99999999)
	require.Error(t, err)
}

func TestAutomationRepo_CreateDownloadOnClosedDB(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := NewAutomationRepo(d.Querier())
	require.NoError(t, d.Close())

	_, err := repo.CreateDownload(ctx, AutomationDownload{
		ClientInstanceID: 1,
		ClientDownloadID: "x",
		ReleaseID:        "r",
		Title:            "t",
		Protocol:         "torrent",
		Status:           AutomationDownloadQueued,
	})
	require.Error(t, err)
}

func TestAutomationRepo_ListQualityProfilesClosedDB(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := NewAutomationRepo(d.Querier())
	require.NoError(t, d.Close())

	_, err := repo.ListQualityProfiles(ctx)
	require.Error(t, err)
}

func TestAutomationRepo_ListDownloadsClosedDB(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := NewAutomationRepo(d.Querier())
	require.NoError(t, d.Close())

	_, err := repo.ListDownloads(ctx, 10)
	require.Error(t, err)
}
