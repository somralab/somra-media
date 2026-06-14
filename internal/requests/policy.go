package requests

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/somralab/somra-media/internal/platform/db"
)

// RequestPolicy holds admin-configured request rules (domain view).
type RequestPolicy struct {
	MonthlyQuotaPerUser int
	AutoApproveRoles    []string
}

// PolicyStore loads request_policies configuration.
type PolicyStore interface {
	GetPolicy(ctx context.Context) (RequestPolicy, error)
}

// RequestCounter counts user requests in the current quota window.
type RequestCounter interface {
	CountUserRequestsInMonth(ctx context.Context, userID string, month time.Time) (int, error)
}

// PolicyDecision is the outcome of quota and auto-approve evaluation.
type PolicyDecision struct {
	AutoApprove   bool
	QuotaAllowed  bool
	UsedQuota     int
	MaxQuota      int
	MatchingRoles []string
}

// PolicyService evaluates quotas and role-based auto-approval.
type PolicyService struct {
	Policies PolicyStore
	Counter  RequestCounter
	Now      func() time.Time
}

// Evaluate checks whether userID may create a request and whether it auto-approves.
func (s *PolicyService) Evaluate(ctx context.Context, userID string, roles []string) (PolicyDecision, error) {
	if s == nil {
		return PolicyDecision{QuotaAllowed: true}, nil
	}
	decision := PolicyDecision{QuotaAllowed: true, AutoApprove: false}
	if s.Policies == nil {
		return decision, nil
	}
	policy, err := s.Policies.GetPolicy(ctx)
	if err != nil {
		return PolicyDecision{}, err
	}
	decision.MaxQuota = policy.MonthlyQuotaPerUser
	decision.MatchingRoles = matchingRoles(roles, policy.AutoApproveRoles)
	decision.AutoApprove = len(decision.MatchingRoles) > 0

	if policy.MonthlyQuotaPerUser <= 0 || s.Counter == nil {
		return decision, nil
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	used, err := s.Counter.CountUserRequestsInMonth(ctx, userID, monthStart(now))
	if err != nil {
		return PolicyDecision{}, err
	}
	decision.UsedQuota = used
	decision.QuotaAllowed = used < policy.MonthlyQuotaPerUser
	return decision, nil
}

// ValidateQuota returns ErrQuotaExceeded when the monthly limit is reached.
func (s *PolicyService) ValidateQuota(ctx context.Context, userID string, roles []string) error {
	decision, err := s.Evaluate(ctx, userID, roles)
	if err != nil {
		return err
	}
	if !decision.QuotaAllowed {
		return ErrQuotaExceeded
	}
	return nil
}

func matchingRoles(userRoles, autoRoles []string) []string {
	if len(autoRoles) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(autoRoles))
	for _, r := range autoRoles {
		allowed[r] = struct{}{}
	}
	var matched []string
	for _, r := range userRoles {
		if _, ok := allowed[r]; ok {
			matched = append(matched, r)
		}
	}
	return matched
}

func monthStart(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
}

// DBPolicyStore adapts [db.RequestRepo] policy reads.
type DBPolicyStore struct {
	Repo *db.RequestRepo
}

// GetPolicy implements [PolicyStore].
func (s DBPolicyStore) GetPolicy(ctx context.Context) (RequestPolicy, error) {
	if s.Repo == nil {
		return RequestPolicy{}, fmt.Errorf("requests policy store: repo is nil")
	}
	row, err := s.Repo.GetPolicy(ctx)
	if err != nil {
		return RequestPolicy{}, fmt.Errorf("requests policy store: %w", err)
	}
	var roles []string
	if row.AutoApproveRoles != "" {
		if err := json.Unmarshal([]byte(row.AutoApproveRoles), &roles); err != nil {
			return RequestPolicy{}, fmt.Errorf("requests policy store decode roles: %w", err)
		}
	}
	return RequestPolicy{
		MonthlyQuotaPerUser: row.UserQuotaPerMonth,
		AutoApproveRoles:    roles,
	}, nil
}

// DBRequestCounter adapts monthly quota counting via [db.RequestRepo].
type DBRequestCounter struct {
	Repo *db.RequestRepo
}

// CountUserRequestsInMonth implements [RequestCounter].
func (c DBRequestCounter) CountUserRequestsInMonth(ctx context.Context, userID string, month time.Time) (int, error) {
	if c.Repo == nil {
		return 0, fmt.Errorf("requests counter: repo is nil")
	}
	n, err := c.Repo.CountByUserInMonth(ctx, userID, month.Format("2006-01"))
	if err != nil {
		return 0, fmt.Errorf("requests counter: %w", err)
	}
	return int(n), nil
}
