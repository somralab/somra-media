package handoff

import (
	"context"
	"fmt"

	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/requests"
)

// Queue persists approved requests for automation processing.
type Queue struct {
	Repo *db.AutomationRepo
}

// RecordHandoff implements requests.HandoffQueue.
func (q *Queue) RecordHandoff(ctx context.Context, req requests.Request) error {
	if q == nil || q.Repo == nil {
		return fmt.Errorf("automation handoff queue: repo required")
	}
	_, err := q.Repo.RecordHandoff(ctx, req.ID)
	return err
}
