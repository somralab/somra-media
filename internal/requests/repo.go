package requests

import (
	"context"
	"time"

	"github.com/somralab/somra-media/internal/platform/db"
)

// Repository persists request CRUD (implemented by Wave 1A [db.RequestRepo]).
type Repository interface {
	Create(ctx context.Context, req Request) (Request, error)
	GetByID(ctx context.Context, id int64) (Request, error)
	Update(ctx context.Context, req Request) (Request, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]Request, error)
	HasPendingByProviderID(ctx context.Context, provider, externalID string) (bool, int64, error)
	CountUserRequestsInMonth(ctx context.Context, userID string, month time.Time) (int, error)
}

// StaticPolicyStore returns a fixed policy (tests and bootstrap defaults).
type StaticPolicyStore struct {
	Policy RequestPolicy
}

// GetPolicy implements [PolicyStore].
func (s StaticPolicyStore) GetPolicy(_ context.Context) (RequestPolicy, error) {
	return s.Policy, nil
}

// FromDBRequest maps a persistence row to the domain model.
func FromDBRequest(row db.Request) Request {
	return Request{
		ID:                row.ID,
		UserID:            row.UserID,
		Status:            Status(row.Status),
		MediaKind:         MediaKind(row.MediaKind),
		Provider:          row.Provider,
		ExternalID:        row.ExternalID,
		Title:             row.Title,
		PosterURL:         row.PosterURL,
		QualityResolution: QualityResolution(row.QualityResolution),
		QualityProfile:    row.QualityProfile,
		CollisionFlag:     row.CollisionFlag,
		AdminNote:         row.AdminNote,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

// ToDBRequest maps the domain model to a persistence row for insert.
func ToDBRequest(req Request) db.Request {
	return db.Request{
		ID:                req.ID,
		UserID:            req.UserID,
		MediaKind:         db.RequestMediaKind(req.MediaKind),
		Provider:          req.Provider,
		ExternalID:        req.ExternalID,
		Title:             req.Title,
		PosterURL:         req.PosterURL,
		QualityResolution: db.RequestQualityResolution(req.QualityResolution),
		QualityProfile:    req.QualityProfile,
		Status:            db.RequestStatus(req.Status),
		CollisionFlag:     req.CollisionFlag,
		AdminNote:         req.AdminNote,
		CreatedAt:         req.CreatedAt,
		UpdatedAt:         req.UpdatedAt,
	}
}
