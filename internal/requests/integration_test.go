package requests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestDBPendingRequestLookup(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "requester", "hash", []string{"user"})
	require.NoError(t, err)

	reqRepo := db.NewRequestRepo(d.Querier())
	id, err := reqRepo.Create(ctx, db.Request{
		UserID:     userID,
		MediaKind:  db.RequestMediaKindMovie,
		Provider:   "tmdb",
		ExternalID: "999",
		Title:      "Pending Title",
	})
	require.NoError(t, err)

	lookup := &DBPendingRequestLookup{Q: d.Querier()}
	found, gotID, err := lookup.HasPendingByProviderID(ctx, "tmdb", "999")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, id, gotID)

	found, _, err = lookup.HasPendingByProviderID(ctx, "tmdb", "missing")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestCollisionChecker_DBIntegration(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "requester", "hash", []string{"user"})
	require.NoError(t, err)

	libRepo := db.NewLibraryRepo(d.Querier())
	mediaRepo := db.NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Movies", db.LibraryKindMovie, []string{dir}, true)
	require.NoError(t, err)

	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Owned", nil)
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetProviderID(ctx, itemID, "tmdb", "111"))

	reqRepo := db.NewRequestRepo(d.Querier())
	_, err = reqRepo.Create(ctx, db.Request{
		UserID: userID, MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "222", Title: "Waiting",
	})
	require.NoError(t, err)

	checker := &CollisionChecker{
		Library:  &DBLibraryLookup{Q: d.Querier()},
		Requests: &DBPendingRequestLookup{Q: d.Querier()},
	}

	err = checker.ValidateCreation(ctx, "tmdb", "111")
	require.ErrorIs(t, err, ErrCollisionInLibrary)

	err = checker.ValidateCreation(ctx, "tmdb", "222")
	require.ErrorIs(t, err, ErrCollisionDuplicatePending)

	require.NoError(t, checker.ValidateCreation(ctx, "tmdb", "333"))
}

func TestDBPolicyStore_AndCounter(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	defer d.Close()

	users := db.NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "requester", "hash", []string{auth.RoleAdmin})
	require.NoError(t, err)

	reqRepo := db.NewRequestRepo(d.Querier())
	_, err = reqRepo.Create(ctx, db.Request{
		UserID: userID, MediaKind: db.RequestMediaKindTV,
		Provider: "tmdb", ExternalID: "1", Title: "Show",
	})
	require.NoError(t, err)

	svc := &PolicyService{
		Policies: DBPolicyStore{Repo: reqRepo},
		Counter:  DBRequestCounter{Repo: reqRepo},
	}
	decision, err := svc.Evaluate(ctx, userID, []string{auth.RoleAdmin})
	require.NoError(t, err)
	assert.True(t, decision.AutoApprove)
	assert.True(t, decision.QuotaAllowed)
	assert.Equal(t, 1, decision.UsedQuota)
	assert.Equal(t, 10, decision.MaxQuota)
}

func TestFromDBRequest_RoundTrip(t *testing.T) {
	row := db.Request{
		ID: 1, UserID: "u", MediaKind: db.RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "9", Title: "T",
		QualityResolution: db.RequestQuality1080p, Status: db.RequestStatusPending,
	}
	domain := FromDBRequest(row)
	back := ToDBRequest(domain)
	assert.Equal(t, row.UserID, back.UserID)
	assert.Equal(t, row.QualityResolution, back.QualityResolution)
}

func TestCollisionChecker_NilChecker(t *testing.T) {
	var c *CollisionChecker
	_, err := c.Check(context.Background(), "tmdb", "1")
	require.Error(t, err)
}

func TestDBLibraryLookup_NilQuerier(t *testing.T) {
	var l *DBLibraryLookup
	_, _, err := l.ExistsByProviderID(context.Background(), "tmdb", "1")
	require.Error(t, err)
}
