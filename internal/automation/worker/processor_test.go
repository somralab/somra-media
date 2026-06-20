package worker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/automation/automationtest"
	indexersearch "github.com/somralab/somra-media/internal/automation/indexer"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestProcessor_ProcessPendingEndToEnd(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	automationtest.CreateStubIndexer(t, mgr, "worker-idx")
	automationtest.CreateStubDownloadClient(t, mgr, "worker-dl")

	reqID := automationtest.CreateApprovedRequest(t, d, "Worker Movie")
	autoRepo := db.NewAutomationRepo(d.Querier())
	_, err := autoRepo.RecordHandoff(ctx, reqID)
	require.NoError(t, err)

	p := &Processor{
		AutoRepo: autoRepo,
		Requests: db.NewRequestRepo(d.Querier()),
		Search:   &indexersearch.SearchService{Manager: mgr},
		Manager:  mgr,
	}
	require.NoError(t, p.ProcessPending(ctx))

	downloads, err := autoRepo.ListDownloads(ctx, 10)
	require.NoError(t, err)
	require.NotEmpty(t, downloads)
}

func TestProcessor_ProcessPendingRequiresDeps(t *testing.T) {
	var p *Processor
	require.Error(t, p.ProcessPending(context.Background()))
}

func TestProcessor_NoSearchResults(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	// Indexer intentionally not created.

	reqID := automationtest.CreateApprovedRequest(t, d, "Empty Search Movie")
	autoRepo := db.NewAutomationRepo(d.Querier())
	_, err := autoRepo.RecordHandoff(ctx, reqID)
	require.NoError(t, err)

	p := &Processor{
		AutoRepo: autoRepo,
		Requests: db.NewRequestRepo(d.Querier()),
		Search:   &indexersearch.SearchService{Manager: mgr},
		Manager:  mgr,
	}
	require.NoError(t, p.ProcessPending(ctx))

	handoffs, err := autoRepo.ListPendingHandoffs(ctx, 10)
	require.NoError(t, err)
	require.Empty(t, handoffs)
}

func TestProcessor_NilSearchService(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	automationtest.CreateStubIndexer(t, mgr, "worker-idx-nil-search")
	automationtest.CreateStubDownloadClient(t, mgr, "worker-dl-nil-search")

	reqID := automationtest.CreateApprovedRequest(t, d, "Nil Search Movie")
	autoRepo := db.NewAutomationRepo(d.Querier())
	_, err := autoRepo.RecordHandoff(ctx, reqID)
	require.NoError(t, err)

	p := &Processor{
		AutoRepo: autoRepo,
		Requests: db.NewRequestRepo(d.Querier()),
		Manager:  mgr,
	}
	require.NoError(t, p.ProcessPending(ctx))
}

func TestProcessor_NoDownloadClient(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	automationtest.CreateStubIndexer(t, mgr, "worker-idx-no-dl")

	reqID := automationtest.CreateApprovedRequest(t, d, "No Client Movie")
	autoRepo := db.NewAutomationRepo(d.Querier())
	_, err := autoRepo.RecordHandoff(ctx, reqID)
	require.NoError(t, err)

	p := &Processor{
		AutoRepo: autoRepo,
		Requests: db.NewRequestRepo(d.Querier()),
		Search:   &indexersearch.SearchService{Manager: mgr},
		Manager:  mgr,
	}
	require.NoError(t, p.ProcessPending(ctx))
}

func TestProcessor_InvalidQualityProfile(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	automationtest.CreateStubIndexer(t, mgr, "worker-idx-bad-profile")
	automationtest.CreateStubDownloadClient(t, mgr, "worker-dl-bad-profile")

	reqID := automationtest.CreateApprovedRequest(t, d, "Bad Profile Movie")
	reqRepo := db.NewRequestRepo(d.Querier())
	missing := "does-not-exist"
	require.NoError(t, reqRepo.Update(ctx, reqID, db.RequestUpdate{QualityProfile: &missing}))

	autoRepo := db.NewAutomationRepo(d.Querier())
	_, err := autoRepo.RecordHandoff(ctx, reqID)
	require.NoError(t, err)

	p := &Processor{
		AutoRepo: autoRepo,
		Requests: reqRepo,
		Search:   &indexersearch.SearchService{Manager: mgr},
		Manager:  mgr,
	}
	require.NoError(t, p.ProcessPending(ctx))
}

func TestProcessor_InvalidProfileSpec(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	mgr := automationtest.NewManager(t, d)
	automationtest.CreateStubIndexer(t, mgr, "worker-idx-bad-spec")
	automationtest.CreateStubDownloadClient(t, mgr, "worker-dl-bad-spec")

	autoRepo := db.NewAutomationRepo(d.Querier())
	_, err := autoRepo.CreateQualityProfile(ctx, "bad-spec", "not-json", false)
	require.NoError(t, err)

	reqID := automationtest.CreateApprovedRequest(t, d, "Bad Spec Movie")
	reqRepo := db.NewRequestRepo(d.Querier())
	profile := "bad-spec"
	require.NoError(t, reqRepo.Update(ctx, reqID, db.RequestUpdate{QualityProfile: &profile}))

	_, err = autoRepo.RecordHandoff(ctx, reqID)
	require.NoError(t, err)

	p := &Processor{
		AutoRepo: autoRepo,
		Requests: reqRepo,
		Search:   &indexersearch.SearchService{Manager: mgr},
		Manager:  mgr,
	}
	require.NoError(t, p.ProcessPending(ctx))
}
