package requests

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/somralab/somra-media/internal/platform/db"
)

// LibraryLookup answers whether a provider identifier is already in the library.
type LibraryLookup interface {
	ExistsByProviderID(ctx context.Context, provider, externalID string) (bool, int64, error)
}

// RequestLookup answers whether a pending request exists for a provider id.
type RequestLookup interface {
	HasPendingByProviderID(ctx context.Context, provider, externalID string) (found bool, requestID int64, err error)
}

// CollisionOutcome describes overlap with library or in-flight requests.
type CollisionOutcome struct {
	Blocked            bool
	InLibrary          bool
	LibraryMediaItemID int64
	DuplicatePending   bool
	ExistingRequestID  int64
}

// CollisionChecker validates new requests against library and pending duplicates.
type CollisionChecker struct {
	Library  LibraryLookup
	Requests RequestLookup
}

// Check inspects provider+externalID and returns a collision outcome.
func (c *CollisionChecker) Check(ctx context.Context, provider, externalID string) (CollisionOutcome, error) {
	if c == nil {
		return CollisionOutcome{}, fmt.Errorf("requests collision: checker is nil")
	}
	var out CollisionOutcome
	if c.Library != nil {
		found, itemID, err := c.Library.ExistsByProviderID(ctx, provider, externalID)
		if err != nil {
			return CollisionOutcome{}, fmt.Errorf("requests collision library lookup: %w", err)
		}
		out.InLibrary = found
		out.LibraryMediaItemID = itemID
	}
	if c.Requests != nil {
		found, reqID, err := c.Requests.HasPendingByProviderID(ctx, provider, externalID)
		if err != nil {
			return CollisionOutcome{}, fmt.Errorf("requests collision pending lookup: %w", err)
		}
		out.DuplicatePending = found
		out.ExistingRequestID = reqID
	}
	out.Blocked = out.InLibrary || out.DuplicatePending
	return out, nil
}

// ValidateCreation returns an error when creation should be blocked.
func (c *CollisionChecker) ValidateCreation(ctx context.Context, provider, externalID string) error {
	out, err := c.Check(ctx, provider, externalID)
	if err != nil {
		return err
	}
	if out.InLibrary {
		return fmt.Errorf("requests validate provider %q external %q: %w", provider, externalID, ErrCollisionInLibrary)
	}
	if out.DuplicatePending {
		return fmt.Errorf("requests validate provider %q external %q: %w", provider, externalID, ErrCollisionDuplicatePending)
	}
	return nil
}

// DBLibraryLookup resolves provider_id rows via the shared db.Querier abstraction.
type DBLibraryLookup struct {
	Q db.Querier
}

// ExistsByProviderID implements [LibraryLookup].
func (l *DBLibraryLookup) ExistsByProviderID(ctx context.Context, provider, externalID string) (bool, int64, error) {
	if l == nil || l.Q == nil {
		return false, 0, fmt.Errorf("requests library lookup: querier is nil")
	}
	var itemID int64
	err := l.Q.QueryRowContext(ctx, `
		SELECT media_item_id FROM provider_id
		WHERE provider = ? AND external_id = ?
		LIMIT 1
	`, provider, externalID).Scan(&itemID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("requests library lookup %q/%q: %w", provider, externalID, err)
	}
	return true, itemID, nil
}

// DBPendingRequestLookup finds pending requests by provider and external id.
type DBPendingRequestLookup struct {
	Q db.Querier
}

// HasPendingByProviderID implements [RequestLookup].
func (l *DBPendingRequestLookup) HasPendingByProviderID(ctx context.Context, provider, externalID string) (bool, int64, error) {
	if l == nil || l.Q == nil {
		return false, 0, fmt.Errorf("requests pending lookup: querier is nil")
	}
	var id int64
	err := l.Q.QueryRowContext(ctx, `
		SELECT id FROM requests
		WHERE provider = ? AND external_id = ? AND status = 'pending'
		LIMIT 1
	`, provider, externalID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("requests pending lookup %q/%q: %w", provider, externalID, err)
	}
	return true, id, nil
}
