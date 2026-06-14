package requests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memoryHandoffQueue struct {
	records []Request
}

func (q *memoryHandoffQueue) RecordHandoff(_ context.Context, req Request) error {
	q.records = append(q.records, req)
	return nil
}

func TestNoOpAutomationHandoff(t *testing.T) {
	t.Parallel()
	var h NoOpAutomationHandoff
	require.NoError(t, h.Handoff(context.Background(), Request{ID: 1}))
}

func TestQueuingAutomationHandoff(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	q := &memoryHandoffQueue{}
	h := QueuingAutomationHandoff{Queue: q}
	req := Request{ID: 42, Status: StatusApproved}
	require.NoError(t, h.Handoff(ctx, req))
	require.Len(t, q.records, 1)
	assert.Equal(t, int64(42), q.records[0].ID)

	h = QueuingAutomationHandoff{}
	require.NoError(t, h.Handoff(ctx, req))
}
