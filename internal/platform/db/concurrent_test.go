package db

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConcurrent_WALReadersAndWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent WAL test in -short mode")
	}
	t.Parallel()

	const (
		readers       = 16
		readsPer      = 50
		writes        = 100
		overallBudget = 30 * time.Second
	)

	ctx, cancel := context.WithTimeout(context.Background(), overallBudget)
	defer cancel()

	d := newMigratedDB(t)
	repo := NewSettingsRepo(d.Querier())
	require.NoError(t, repo.Set(ctx, "concurrent.seed", "0"))

	var wg sync.WaitGroup
	errCh := make(chan error, readers+1)

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := NewSettingsRepo(d.Querier())
			for j := 0; j < readsPer; j++ {
				if _, err := r.Get(ctx, "concurrent.seed"); err != nil {
					errCh <- fmt.Errorf("reader %d iter %d: %w", id, j, err)
					return
				}
			}
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		w := NewSettingsRepo(d.Querier())
		for j := 0; j < writes; j++ {
			if err := w.Set(ctx, "concurrent.seed", fmt.Sprintf("v-%d", j)); err != nil {
				errCh <- fmt.Errorf("writer iter %d: %w", j, err)
				return
			}
		}
	}()

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Errorf("goroutine reported: %v", err)
	}

	v, err := repo.Get(ctx, "concurrent.seed")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("v-%d", writes-1), v)
}

// TestConcurrent_SequentialTxWritesPersist documents that consecutive
// write transactions on the same DB pool commit independently. SQLite
// serialises writers in WAL mode, so spawning multiple parallel write
// transactions is unsafe without deadlock-aware retry logic — that is a
// future concern (see plan/sprint-02 jobs) and outside the M1 scope.
func TestConcurrent_SequentialTxWritesPersist(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d := newMigratedDB(t)

	for i := 0; i < 4; i++ {
		i := i
		require.NoError(t, WithTx(ctx, d, func(q Querier) error {
			return NewSettingsRepo(q).Set(ctx, fmt.Sprintf("tx.seq.%d", i), "v")
		}))
	}

	for i := 0; i < 4; i++ {
		v, err := NewSettingsRepo(d.Querier()).Get(ctx, fmt.Sprintf("tx.seq.%d", i))
		require.NoError(t, err)
		require.Equal(t, "v", v)
	}
}
