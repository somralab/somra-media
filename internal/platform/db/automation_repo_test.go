package db

import (
	"context"
	"fmt"
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

func TestAutomationRepo_Monitors(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "monitor-user", "hash", []string{"user"})
	require.NoError(t, err)

	repo := NewAutomationRepo(d.Querier())
	id, err := repo.CreateMonitor(ctx, AutomationMonitor{
		UserID:     userID,
		Title:      "Monitor Show",
		Provider:   "tmdb",
		ExternalID: "mon-1",
		Enabled:    true,
	})
	require.NoError(t, err)
	require.Positive(t, id)

	_, err = repo.CreateMonitor(ctx, AutomationMonitor{
		UserID:     userID,
		Title:      "Monitor Show 2",
		Provider:   "tmdb",
		ExternalID: "mon-1",
		Enabled:    true,
	})
	require.ErrorIs(t, err, ErrAutomationMonitorDuplicate)

	got, err := repo.GetMonitorByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "Monitor Show", got.Title)

	all, err := repo.ListMonitors(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)

	enabled, err := repo.ListEnabledMonitors(ctx, 10)
	require.NoError(t, err)
	require.Len(t, enabled, 1)

	disabled := false
	require.NoError(t, repo.PatchMonitor(ctx, id, nil, nil, &disabled))
	got, err = repo.GetMonitorByID(ctx, id)
	require.NoError(t, err)
	require.False(t, got.Enabled)

	require.NoError(t, repo.UpdateMonitorProgress(ctx, id, 1, 3))
	got, err = repo.GetMonitorByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, got.LastCheckedAt)

	allWithCheck, err := repo.ListMonitors(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, allWithCheck)

	reqRepo := NewRequestRepo(d.Querier())
	requestID, err := reqRepo.Create(ctx, Request{
		UserID:            userID,
		MediaKind:         RequestMediaKindTV,
		Provider:          "tmdb",
		ExternalID:        "mon-1",
		Title:             "Monitor Show S01E03",
		QualityResolution: RequestQualityAny,
		Status:            RequestStatusApproved,
	})
	require.NoError(t, err)
	require.NoError(t, repo.RecordMonitorEpisode(ctx, id, 1, 3, requestID))
	exists, err := repo.MonitorEpisodeExists(ctx, id, 1, 3)
	require.NoError(t, err)
	require.True(t, exists)

	profileID, err := repo.GetQualityProfileByID(ctx, 1)
	require.NoError(t, err)
	require.NotEmpty(t, profileID.Name)

	require.NoError(t, repo.UpdateQualityProfile(ctx, profileID.ID, "default-renamed", profileID.Spec, nil))
	updated, err := repo.GetQualityProfileByID(ctx, profileID.ID)
	require.NoError(t, err)
	require.Equal(t, "default-renamed", updated.Name)

	require.NoError(t, repo.DeleteMonitor(ctx, id))
	_, err = repo.GetMonitorByID(ctx, id)
	require.ErrorIs(t, err, ErrAutomationMonitorNotFound)
}

func TestAutomationRepo_UpdateQualityProfileBranches(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewAutomationRepo(d.Querier())
	profile, err := repo.GetDefaultQualityProfile(ctx)
	require.NoError(t, err)

	isDefault := true
	require.NoError(t, repo.UpdateQualityProfile(ctx, profile.ID, profile.Name, profile.Spec, &isDefault))

	isDefaultFalse := false
	require.NoError(t, repo.UpdateQualityProfile(ctx, profile.ID, "default-patched", profile.Spec, &isDefaultFalse))

	require.NoError(t, repo.UpdateQualityProfile(ctx, profile.ID, "default-patched-2", profile.Spec, nil))

	require.Error(t, repo.UpdateQualityProfile(ctx, profile.ID, "   ", profile.Spec, nil))
	require.ErrorIs(t, repo.UpdateQualityProfile(ctx, 999999, "missing", "{}", nil), ErrQualityProfileNotFound)

	otherID, err := repo.CreateQualityProfile(ctx, "other-profile", "{}", false)
	require.NoError(t, err)
	thirdID, err := repo.CreateQualityProfile(ctx, "third-profile", "{}", false)
	require.NoError(t, err)
	require.ErrorIs(t, repo.UpdateQualityProfile(ctx, thirdID, "other-profile", "{}", nil), ErrQualityProfileDuplicate)
	_ = otherID
}

func TestAutomationRepo_PatchMonitorBranches(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "patch-user", "hash", []string{"user"})
	require.NoError(t, err)

	repo := NewAutomationRepo(d.Querier())
	id, err := repo.CreateMonitor(ctx, AutomationMonitor{
		UserID:     userID,
		Title:      "Patch Show",
		Provider:   "tmdb",
		ExternalID: "patch-1",
		Enabled:    true,
	})
	require.NoError(t, err)

	require.NoError(t, repo.PatchMonitor(ctx, id, nil, nil, nil))

	newTitle := "Patch Show Renamed"
	newProfile := "default"
	require.NoError(t, repo.PatchMonitor(ctx, id, &newTitle, &newProfile, nil))

	empty := "   "
	require.Error(t, repo.PatchMonitor(ctx, id, &empty, nil, nil))
	require.ErrorIs(t, repo.PatchMonitor(ctx, 999999, &newTitle, nil, nil), ErrAutomationMonitorNotFound)
}

func TestAutomationRepo_ListEnabledMonitorsDefaultLimit(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewAutomationRepo(d.Querier())
	rows, err := repo.ListEnabledMonitors(ctx, 0)
	require.NoError(t, err)
	require.Empty(t, rows)
}

func TestAutomationRepo_GetQualityProfileByIDNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewAutomationRepo(d.Querier())
	_, err := repo.GetQualityProfileByID(ctx, 999999)
	require.ErrorIs(t, err, ErrQualityProfileNotFound)
}

func TestAutomationRepo_DeleteMonitorNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewAutomationRepo(d.Querier())
	require.ErrorIs(t, repo.DeleteMonitor(ctx, 999999), ErrAutomationMonitorNotFound)
}

func TestAutomationRepo_CreateMonitorValidation(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewAutomationRepo(d.Querier())
	_, err := repo.CreateMonitor(ctx, AutomationMonitor{})
	require.Error(t, err)
}

func TestAutomationRepo_ListMonitorsMultiple(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "list-user", "hash", []string{"user"})
	require.NoError(t, err)

	repo := NewAutomationRepo(d.Querier())
	for i := range 2 {
		_, err = repo.CreateMonitor(ctx, AutomationMonitor{
			UserID:     userID,
			Title:      fmt.Sprintf("Show %d", i),
			Provider:   "tmdb",
			ExternalID: fmt.Sprintf("list-%d", i),
			Enabled:    true,
		})
		require.NoError(t, err)
	}

	rows, err := repo.ListMonitors(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	enabled, err := repo.ListEnabledMonitors(ctx, 0)
	require.NoError(t, err)
	require.Len(t, enabled, 2)
}

func TestAutomationRepo_ListEnabledMonitorsClosedDB(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := NewAutomationRepo(d.Querier())
	require.NoError(t, d.Close())

	_, err := repo.ListEnabledMonitors(ctx, 10)
	require.Error(t, err)
}
