package handoff

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/automation/automationtest"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/requests"
)

func TestQueue_RecordHandoff(t *testing.T) {
	ctx := context.Background()
	d := automationtest.OpenDB(t)
	reqID := automationtest.CreateApprovedRequest(t, d, "Handoff Movie")

	q := &Queue{Repo: db.NewAutomationRepo(d.Querier())}
	require.NoError(t, q.RecordHandoff(ctx, requests.Request{ID: reqID}))

	var nilQ *Queue
	require.Error(t, nilQ.RecordHandoff(ctx, requests.Request{ID: reqID}))
}
