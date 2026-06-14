package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// UserProfile holds per-user preferences and parental settings.
type UserProfile struct {
	UserID           string
	Locale           string
	Theme            string
	AvatarURL        string
	MaxContentRating *string
	IsChild          bool
}

// ProfileRepo persists user profiles.
type ProfileRepo struct {
	q Querier
}

// NewProfileRepo returns a repository bound to q.
func NewProfileRepo(q Querier) *ProfileRepo {
	return &ProfileRepo{q: q}
}

var ErrProfileNotFound = errors.New("db profile: not found")

// Get returns the profile for userID.
func (r *ProfileRepo) Get(ctx context.Context, userID string) (UserProfile, error) {
	var p UserProfile
	var avatar sql.NullString
	var maxRating sql.NullString
	var isChild int
	err := r.q.QueryRowContext(ctx, `
		SELECT user_id, locale, theme, avatar_url, max_content_rating, is_child
		FROM user_profile WHERE user_id = ?
	`, userID).Scan(&p.UserID, &p.Locale, &p.Theme, &avatar, &maxRating, &isChild)
	if errors.Is(err, sql.ErrNoRows) {
		return UserProfile{}, ErrProfileNotFound
	}
	if err != nil {
		return UserProfile{}, fmt.Errorf("db profile get: %w", err)
	}
	if avatar.Valid {
		p.AvatarURL = avatar.String
	}
	if maxRating.Valid {
		s := maxRating.String
		p.MaxContentRating = &s
	}
	p.IsChild = isChild != 0
	return p, nil
}

// Update replaces mutable profile fields.
func (r *ProfileRepo) Update(ctx context.Context, p UserProfile) error {
	res, err := r.q.ExecContext(ctx, `
		UPDATE user_profile SET
			locale = ?,
			theme = ?,
			avatar_url = ?,
			max_content_rating = ?,
			is_child = ?
		WHERE user_id = ?
	`, p.Locale, p.Theme, nullString(p.AvatarURL), nullStringPtr(p.MaxContentRating), boolToInt(p.IsChild), p.UserID)
	if err != nil {
		return fmt.Errorf("db profile update: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db profile update rows: %w", err)
	}
	if n == 0 {
		return ErrProfileNotFound
	}
	return nil
}

func nullStringPtr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
