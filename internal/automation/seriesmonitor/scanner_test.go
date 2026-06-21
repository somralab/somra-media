package seriesmonitor

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/automation/automationtest"
	indexersearch "github.com/somralab/somra-media/internal/automation/indexer"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/plugin/stub"
)

func TestScanner_RequiresDeps(t *testing.T) {
	t.Parallel()
	var s *Scanner
	require.Error(t, s.ScanEnabledMonitors(context.Background()))
}

func TestScanner_ScanEnabledMonitorsEmpty(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	automationtest.CreateStubIndexer(t, mgr, "scan-idx")

	s := &Scanner{
		AutoRepo: db.NewAutomationRepo(d.Querier()),
		Requests: db.NewRequestRepo(d.Querier()),
		Search:   &indexersearch.SearchService{Manager: mgr},
	}
	require.NoError(t, s.ScanEnabledMonitors(ctx))
}

func TestScanner_QueuesNewEpisode(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	_, err := mgr.Create(ctx, plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeIndexer,
		Implementation: stub.Implementation,
		Name:           "tv-idx",
		Config:         []byte(`{"prefix":"MonitorShow.S01E06.1080p"}`),
		Enabled:        true,
	})
	require.NoError(t, err)

	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err = users.Create(ctx, userID, "scan-user", "hash", []string{"user"})
	require.NoError(t, err)

	autoRepo := db.NewAutomationRepo(d.Querier())
	monitorID, err := autoRepo.CreateMonitor(ctx, db.AutomationMonitor{
		UserID:         userID,
		Title:          "MonitorShow",
		Provider:       "tmdb",
		ExternalID:     "show-1",
		QualityProfile: "default",
		Enabled:        true,
	})
	require.NoError(t, err)

	s := &Scanner{
		AutoRepo: autoRepo,
		Requests: db.NewRequestRepo(d.Querier()),
		Search:   &indexersearch.SearchService{Manager: mgr},
		Logger:   slog.Default(),
	}
	require.NoError(t, s.ScanEnabledMonitors(ctx))

	got, err := autoRepo.GetMonitorByID(ctx, monitorID)
	require.NoError(t, err)
	require.Equal(t, 1, got.LastSeason)
	require.Equal(t, 6, got.LastEpisode)

	exists, err := autoRepo.MonitorEpisodeExists(ctx, monitorID, 1, 6)
	require.NoError(t, err)
	require.True(t, exists)

	pending, err := autoRepo.ListPendingHandoffs(ctx, 10)
	require.NoError(t, err)
	require.NotEmpty(t, pending)
}

func TestScanner_SkipsAlreadyTrackedEpisode(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	_, err := mgr.Create(ctx, plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeIndexer,
		Implementation: stub.Implementation,
		Name:           "tv-idx-dup",
		Config:         []byte(`{"prefix":"MonitorShow.S01E06.1080p"}`),
		Enabled:        true,
	})
	require.NoError(t, err)

	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err = users.Create(ctx, userID, "scan-user-2", "hash", []string{"user"})
	require.NoError(t, err)

	autoRepo := db.NewAutomationRepo(d.Querier())
	monitorID, err := autoRepo.CreateMonitor(ctx, db.AutomationMonitor{
		UserID:         userID,
		Title:          "MonitorShow",
		Provider:       "tmdb",
		ExternalID:     "show-2",
		QualityProfile: "default",
		Enabled:        true,
		LastSeason:     1,
		LastEpisode:    6,
	})
	require.NoError(t, err)

	reqRepo := db.NewRequestRepo(d.Querier())
	requestID, err := reqRepo.Create(ctx, db.Request{
		UserID:            userID,
		MediaKind:         db.RequestMediaKindTV,
		Provider:          "tmdb",
		ExternalID:        "show-2",
		Title:             "MonitorShow S01E06",
		QualityResolution: db.RequestQualityAny,
		Status:            db.RequestStatusApproved,
	})
	require.NoError(t, err)
	require.NoError(t, autoRepo.RecordMonitorEpisode(ctx, monitorID, 1, 6, requestID))

	s := &Scanner{
		AutoRepo: autoRepo,
		Requests: reqRepo,
		Search:   &indexersearch.SearchService{Manager: mgr},
	}
	require.NoError(t, s.ScanEnabledMonitors(ctx))

	pending, err := autoRepo.ListPendingHandoffs(ctx, 10)
	require.NoError(t, err)
	require.Empty(t, pending)
}

func TestScanner_UsesDefaultQualityProfile(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	_, err := mgr.Create(ctx, plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeIndexer,
		Implementation: stub.Implementation,
		Name:           "tv-idx-default",
		Config:         []byte(`{"prefix":"DefaultShow.S02E01.720p"}`),
		Enabled:        true,
	})
	require.NoError(t, err)

	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err = users.Create(ctx, userID, "scan-user-3", "hash", []string{"user"})
	require.NoError(t, err)

	autoRepo := db.NewAutomationRepo(d.Querier())
	_, err = autoRepo.CreateMonitor(ctx, db.AutomationMonitor{
		UserID:     userID,
		Title:      "DefaultShow",
		Provider:   "tmdb",
		ExternalID: "show-3",
		Enabled:    true,
	})
	require.NoError(t, err)

	s := &Scanner{
		AutoRepo: autoRepo,
		Requests: db.NewRequestRepo(d.Querier()),
		Search:   &indexersearch.SearchService{Manager: mgr},
	}
	require.NoError(t, s.ScanEnabledMonitors(ctx))
}

func TestScanner_ContinuesAfterMonitorScanError(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	automationtest.CreateStubIndexer(t, mgr, "scan-idx-err")

	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "scan-user-4", "hash", []string{"user"})
	require.NoError(t, err)

	autoRepo := db.NewAutomationRepo(d.Querier())
	_, err = autoRepo.CreateMonitor(ctx, db.AutomationMonitor{
		UserID:         userID,
		Title:          "Bad Profile Show",
		Provider:       "tmdb",
		ExternalID:     "show-4",
		QualityProfile: "missing-profile",
		Enabled:        true,
	})
	require.NoError(t, err)

	s := &Scanner{
		AutoRepo: autoRepo,
		Requests: db.NewRequestRepo(d.Querier()),
		Search:   &indexersearch.SearchService{Manager: mgr},
		Logger:   slog.Default(),
	}
	require.NoError(t, s.ScanEnabledMonitors(ctx))
}
