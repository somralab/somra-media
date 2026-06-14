package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "requester", "hash", []string{"user"})
	require.NoError(t, err)

	repo := NewRequestRepo(d.Querier())

	policy, err := repo.GetPolicy(ctx)
	require.NoError(t, err)
	assert.Equal(t, 10, policy.UserQuotaPerMonth)
	assert.Contains(t, policy.AutoApproveRoles, "admin")

	id, err := repo.Create(ctx, Request{
		UserID:            userID,
		MediaKind:         RequestMediaKindMovie,
		Provider:          "tmdb",
		ExternalID:        "12345",
		Title:             "Example Movie",
		QualityResolution: RequestQuality1080p,
	})
	require.NoError(t, err)
	require.Greater(t, id, int64(0))

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, RequestStatusPending, got.Status)
	assert.Equal(t, "Example Movie", got.Title)
	assert.False(t, got.CreatedAt.IsZero())

	active, err := repo.HasActiveByProviderExternal(ctx, "tmdb", "12345")
	require.NoError(t, err)
	assert.True(t, active)

	note := "approved by admin"
	status := RequestStatusApproved
	require.NoError(t, repo.Update(ctx, id, RequestUpdate{
		Status:    &status,
		AdminNote: &note,
	}))

	got, err = repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, RequestStatusApproved, got.Status)
	assert.Equal(t, note, got.AdminNote)

	list, err := repo.List(ctx, RequestListFilter{UserID: userID})
	require.NoError(t, err)
	require.Len(t, list, 1)

	count, err := repo.CountByUserInMonth(ctx, userID, got.CreatedAt.Format("2006-01"))
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	require.NoError(t, repo.SetStatus(ctx, id, RequestStatusCompleted))
	got, err = repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, RequestStatusCompleted, got.Status)

	active, err = repo.HasActiveByProviderExternal(ctx, "tmdb", "12345")
	require.NoError(t, err)
	assert.False(t, active)

	require.NoError(t, repo.UpsertPolicy(ctx, RequestPolicy{
		AutoApproveRoles:  `["admin","user"]`,
		UserQuotaPerMonth: 5,
		AdminSettings:     `{"notifyOnCreate":true}`,
	}))
	policy, err = repo.GetPolicy(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, policy.UserQuotaPerMonth)
	assert.Contains(t, policy.AdminSettings, "notifyOnCreate")

	_, err = repo.GetByID(ctx, 99999)
	require.ErrorIs(t, err, ErrRequestNotFound)
}

func TestRequestRepo_Validation(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewRequestRepo(d.Querier())

	_, err := repo.Create(ctx, Request{})
	require.Error(t, err)

	_, err = repo.CountByUserInMonth(ctx, "", "2026-06")
	require.Error(t, err)
}

func TestRequestRepo_CreateValidation(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewRequestRepo(d.Querier())

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "val-user", "hash", []string{"user"})
	require.NoError(t, err)

	_, err = repo.Create(ctx, Request{UserID: userID})
	require.Error(t, err)
	_, err = repo.Create(ctx, Request{UserID: userID, Provider: "tmdb", ExternalID: "1"})
	require.Error(t, err)

	id, err := repo.Create(ctx, Request{
		UserID: userID, MediaKind: RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "defaults", Title: "Defaults",
	})
	require.NoError(t, err)
	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, RequestStatusPending, got.Status)
	assert.Equal(t, RequestQualityAny, got.QualityResolution)
}

func TestRequestRepo_UpsertPolicyRoundTrip(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewRequestRepo(d.Querier())

	require.NoError(t, repo.UpsertPolicy(ctx, RequestPolicy{
		AutoApproveRoles:  `["user"]`,
		UserQuotaPerMonth: 2,
		AdminSettings:     `{"notifyOnApprove":false}`,
	}))
	policy, err := repo.GetPolicy(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, policy.UserQuotaPerMonth)
	assert.Contains(t, policy.AutoApproveRoles, "user")
}

func TestRequestRepo_UpdateQualityFields(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	users := NewUserRepo(d.Querier())
	userID := uuid.NewString()
	_, err := users.Create(ctx, userID, "patch-user", "hash", []string{"user"})
	require.NoError(t, err)

	repo := NewRequestRepo(d.Querier())
	id, err := repo.Create(ctx, Request{
		UserID: userID, MediaKind: RequestMediaKindMovie,
		Provider: "tmdb", ExternalID: "patch-1", Title: "Patch Me",
	})
	require.NoError(t, err)

	res := RequestQuality1080p
	profile := "hd"
	collision := true
	require.NoError(t, repo.Update(ctx, id, RequestUpdate{
		QualityResolution: &res,
		QualityProfile:    &profile,
		CollisionFlag:     &collision,
	}))

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, RequestQuality1080p, got.QualityResolution)
	assert.Equal(t, profile, got.QualityProfile)
	assert.True(t, got.CollisionFlag)

	pending, err := repo.List(ctx, RequestListFilter{Status: RequestStatusPending, UserID: userID})
	require.NoError(t, err)
	require.Len(t, pending, 1)

	require.Error(t, repo.SetStatus(ctx, 99999, RequestStatusApproved))

	status := RequestStatusApproved
	require.ErrorIs(t, repo.Update(ctx, 99999, RequestUpdate{Status: &status}), ErrRequestNotFound)
}

func TestRequestRepo_ListLimitDefault(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewRequestRepo(d.Querier())
	list, err := repo.List(ctx, RequestListFilter{Limit: 0})
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestRequestRepo_UpsertPolicyDefaults(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewRequestRepo(d.Querier())

	require.NoError(t, repo.UpsertPolicy(ctx, RequestPolicy{}))
	policy, err := repo.GetPolicy(ctx)
	require.NoError(t, err)
	assert.Equal(t, "[]", policy.AutoApproveRoles)
	assert.Equal(t, "{}", policy.AdminSettings)
}

func TestRequestRepo_HasActiveFalse(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewRequestRepo(d.Querier())
	active, err := repo.HasActiveByProviderExternal(ctx, "tmdb", "missing")
	require.NoError(t, err)
	assert.False(t, active)
}

func TestRequestRepo_ListAfterClose(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	repo := NewRequestRepo(d.Querier())
	require.NoError(t, d.Close())
	_, err := repo.List(ctx, RequestListFilter{})
	require.Error(t, err)
}
