package worker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/automation/automationtest"
	indexersearch "github.com/somralab/somra-media/internal/automation/indexer"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestProcessor_ProcessPendingRequiresDeps(t *testing.T) {
	var p *Processor
	require.Error(t, p.ProcessPending(context.Background()))
}

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
