package requests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

type stubCounter struct {
	count int
	err   error
}

func (s stubCounter) CountUserRequestsInMonth(_ context.Context, _ string, _ time.Time) (int, error) {
	return s.count, s.err
}

func TestPolicyService_Evaluate_AutoApprove(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &PolicyService{
		Policies: StaticPolicyStore{Policy: RequestPolicy{
			MonthlyQuotaPerUser: 5,
			AutoApproveRoles:    []string{auth.RoleAdmin, auth.RoleUser},
		}},
		Counter: stubCounter{count: 1},
	}

	decision, err := svc.Evaluate(ctx, "user-1", []string{auth.RoleUser})
	require.NoError(t, err)
	assert.True(t, decision.AutoApprove)
	assert.True(t, decision.QuotaAllowed)
	assert.Equal(t, 1, decision.UsedQuota)
	assert.Equal(t, 5, decision.MaxQuota)
	assert.Equal(t, []string{auth.RoleUser}, decision.MatchingRoles)
}

func TestPolicyService_Evaluate_NoAutoApproveForChild(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &PolicyService{
		Policies: StaticPolicyStore{Policy: RequestPolicy{
			AutoApproveRoles: []string{auth.RoleAdmin},
		}},
	}
	decision, err := svc.Evaluate(ctx, "user-2", []string{auth.RoleChild})
	require.NoError(t, err)
	assert.False(t, decision.AutoApprove)
}

func TestPolicyService_Evaluate_QuotaExceeded(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &PolicyService{
		Policies: StaticPolicyStore{Policy: RequestPolicy{MonthlyQuotaPerUser: 3}},
		Counter:  stubCounter{count: 3},
	}
	decision, err := svc.Evaluate(ctx, "user-1", nil)
	require.NoError(t, err)
	assert.False(t, decision.QuotaAllowed)

	err = svc.ValidateQuota(ctx, "user-1", nil)
	require.ErrorIs(t, err, ErrQuotaExceeded)
}

func TestPolicyService_ValidateQuota_Allowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &PolicyService{
		Policies: StaticPolicyStore{Policy: RequestPolicy{MonthlyQuotaPerUser: 5}},
		Counter:  stubCounter{count: 2},
	}
	require.NoError(t, svc.ValidateQuota(ctx, "user-1", nil))
}

func TestPolicyService_UnlimitedQuota(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &PolicyService{
		Policies: StaticPolicyStore{Policy: RequestPolicy{MonthlyQuotaPerUser: 0}},
		Counter:  stubCounter{count: 100},
	}
	decision, err := svc.Evaluate(ctx, "user-1", nil)
	require.NoError(t, err)
	assert.True(t, decision.QuotaAllowed)
}

func TestPolicyService_NilDeps(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var svc *PolicyService
	decision, err := svc.Evaluate(ctx, "user-1", nil)
	require.NoError(t, err)
	assert.True(t, decision.QuotaAllowed)
}

func TestDBPolicyStore_Errors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var store DBPolicyStore
	_, err := store.GetPolicy(ctx)
	require.Error(t, err)

	d := openTestDB(t)
	defer d.Close()
	reqRepo := db.NewRequestRepo(d.Querier())
	_, err = d.Querier().ExecContext(ctx, `UPDATE request_policies SET auto_approve_roles = 'not-json' WHERE id = 1`)
	require.NoError(t, err)
	_, err = DBPolicyStore{Repo: reqRepo}.GetPolicy(ctx)
	require.Error(t, err)
}

func TestDBRequestCounter_Errors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var counter DBRequestCounter
	_, err := counter.CountUserRequestsInMonth(ctx, "u", time.Now())
	require.Error(t, err)
}

func TestPolicyService_NilPoliciesStore(t *testing.T) {
	t.Parallel()
	svc := &PolicyService{Counter: stubCounter{count: 0}}
	decision, err := svc.Evaluate(context.Background(), "u", nil)
	require.NoError(t, err)
	assert.False(t, decision.AutoApprove)
	assert.True(t, decision.QuotaAllowed)
}

func TestMonthStart(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 6, 14, 15, 30, 0, 0, time.UTC)
	start := monthStart(ts)
	assert.Equal(t, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), start)
}
