package db

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithTx_Commit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d := newMigratedDB(t)

	err := WithTx(ctx, d, func(q Querier) error {
		return NewSettingsRepo(q).Set(ctx, "tx.commit", "yes")
	})
	require.NoError(t, err)

	v, err := NewSettingsRepo(d.Querier()).Get(ctx, "tx.commit")
	require.NoError(t, err)
	require.Equal(t, "yes", v)
}

func TestWithTx_Rollback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d := newMigratedDB(t)
	sentinel := errors.New("boom")

	err := WithTx(ctx, d, func(q Querier) error {
		if err := NewSettingsRepo(q).Set(ctx, "tx.rollback", "should-not-persist"); err != nil {
			return err
		}
		return sentinel
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, sentinel))
	require.True(t, errors.Is(err, ErrTxRollback))

	_, err = NewSettingsRepo(d.Querier()).Get(ctx, "tx.rollback")
	require.True(t, errors.Is(err, ErrSettingNotFound))
}

func TestWithTx_NilHandles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	require.Error(t, WithTx(ctx, nil, func(Querier) error { return nil }))

	var empty DB
	require.Error(t, WithTx(ctx, &empty, func(Querier) error { return nil }))

	d := newMigratedDB(t)
	require.Error(t, WithTx(ctx, d, nil))
}

func TestWithTx_CommitFailsAfterContextCancelled(t *testing.T) {
	t.Parallel()

	parent := context.Background()
	d := newMigratedDB(t)

	ctx, cancel := context.WithCancel(parent)
	err := WithTx(ctx, d, func(q Querier) error {
		if _, err := q.ExecContext(ctx,
			`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, datetime('now'))`,
			"tx.commit.cancel", "x"); err != nil {
			return err
		}
		cancel()
		return nil
	})
	require.Error(t, err)
}

func TestWithTx_BeginFailsOnClosedDB(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	d := newMigratedDB(t)
	require.NoError(t, d.Close())

	err := WithTx(ctx, d, func(Querier) error { return nil })
	require.Error(t, err)
}

func TestWithTx_LibraryAndMedia(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	d := newMigratedDB(t)
	dir := t.TempDir()

	err := WithTx(ctx, d, func(q Querier) error {
		libRepo := NewLibraryRepo(q)
		mediaRepo := NewMediaRepo(q)
		lib, err := libRepo.Create(ctx, "TxLib", LibraryKindMovie, []string{dir}, true)
		if err != nil {
			return err
		}
		itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Title", nil)
		if err != nil {
			return err
		}
		score := 0.5
		if err := mediaRepo.SetMatch(ctx, itemID, MatchMatched, &score); err != nil {
			return err
		}
		return mediaRepo.SetMatch(ctx, itemID, MatchUnmatched, nil)
	})
	require.NoError(t, err)
}
