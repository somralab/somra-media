package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// RequestMediaKind classifies requested content.
type RequestMediaKind string

const (
	RequestMediaKindMovie RequestMediaKind = "movie"
	RequestMediaKindTV    RequestMediaKind = "tv"
)

// RequestQualityResolution is the preferred release resolution.
type RequestQualityResolution string

const (
	RequestQuality1080p RequestQualityResolution = "1080p"
	RequestQuality720p  RequestQualityResolution = "720p"
	RequestQualityAny   RequestQualityResolution = "any"
)

// RequestStatus tracks lifecycle state for a content request.
type RequestStatus string

const (
	RequestStatusPending   RequestStatus = "pending"
	RequestStatusApproved  RequestStatus = "approved"
	RequestStatusRejected  RequestStatus = "rejected"
	RequestStatusCompleted RequestStatus = "completed"
	RequestStatusCancelled RequestStatus = "cancelled"
)

// Request is a user-submitted content acquisition request.
type Request struct {
	ID                int64                    `json:"id"`
	UserID            string                   `json:"userId"`
	MediaKind         RequestMediaKind         `json:"mediaKind"`
	Provider          string                   `json:"provider"`
	ExternalID        string                   `json:"externalId"`
	Title             string                   `json:"title"`
	PosterURL         string                   `json:"posterUrl,omitempty"`
	QualityResolution RequestQualityResolution `json:"qualityResolution"`
	QualityProfile    string                   `json:"qualityProfile,omitempty"`
	Status            RequestStatus            `json:"status"`
	CollisionFlag     bool                     `json:"collisionFlag"`
	AdminNote         string                   `json:"adminNote,omitempty"`
	CreatedAt         time.Time                `json:"createdAt"`
	UpdatedAt         time.Time                `json:"updatedAt"`
}

// RequestListFilter narrows List results.
type RequestListFilter struct {
	UserID string
	Status RequestStatus
	Limit  int
	Offset int
}

// RequestUpdate carries mutable request fields for PATCH.
type RequestUpdate struct {
	QualityResolution *RequestQualityResolution
	QualityProfile    *string
	AdminNote         *string
	Status            *RequestStatus
	CollisionFlag     *bool
}

// RequestPolicy holds admin-configured request workflow rules.
type RequestPolicy struct {
	AutoApproveRoles  string `json:"autoApproveRoles"`
	UserQuotaPerMonth int    `json:"userQuotaPerMonth"`
	AdminSettings     string `json:"adminSettings"`
	UpdatedAt         time.Time
}

// RequestRepo persists content requests and policies.
type RequestRepo struct {
	q Querier
}

// NewRequestRepo returns a repository bound to q.
func NewRequestRepo(q Querier) *RequestRepo {
	return &RequestRepo{q: q}
}

var (
	ErrRequestNotFound       = errors.New("db request: not found")
	ErrRequestPolicyNotFound = errors.New("db request policy: not found")
)

// Create inserts a new request row.
func (r *RequestRepo) Create(ctx context.Context, req Request) (int64, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return 0, fmt.Errorf("db request create: user id is required")
	}
	if strings.TrimSpace(req.Provider) == "" || strings.TrimSpace(req.ExternalID) == "" {
		return 0, fmt.Errorf("db request create: provider and external id are required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return 0, fmt.Errorf("db request create: title is required")
	}
	status := req.Status
	if status == "" {
		status = RequestStatusPending
	}
	resolution := req.QualityResolution
	if resolution == "" {
		resolution = RequestQualityAny
	}

	res, err := r.q.ExecContext(ctx, `
		INSERT INTO requests (
			user_id, media_kind, provider, external_id, title, poster_url,
			quality_resolution, quality_profile, status, collision_flag, admin_note,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`, req.UserID, req.MediaKind, req.Provider, req.ExternalID, req.Title, nullStr(req.PosterURL),
		resolution, nullStr(req.QualityProfile), status, boolToInt(req.CollisionFlag), nullStr(req.AdminNote))
	if err != nil {
		return 0, fmt.Errorf("db request create: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("db request create id: %w", err)
	}
	return id, nil
}

// GetByID returns a request by primary key.
func (r *RequestRepo) GetByID(ctx context.Context, id int64) (Request, error) {
	var req Request
	var poster, profile, note sql.NullString
	var collision int
	var created, updated string
	err := r.q.QueryRowContext(ctx, `
		SELECT id, user_id, media_kind, provider, external_id, title, poster_url,
			quality_resolution, quality_profile, status, collision_flag, admin_note,
			created_at, updated_at
		FROM requests WHERE id = ?
	`, id).Scan(
		&req.ID, &req.UserID, &req.MediaKind, &req.Provider, &req.ExternalID, &req.Title, &poster,
		&req.QualityResolution, &profile, &req.Status, &collision, &note, &created, &updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Request{}, ErrRequestNotFound
	}
	if err != nil {
		return Request{}, fmt.Errorf("db request get by id: %w", err)
	}
	if poster.Valid {
		req.PosterURL = poster.String
	}
	if profile.Valid {
		req.QualityProfile = profile.String
	}
	if note.Valid {
		req.AdminNote = note.String
	}
	req.CollisionFlag = collision != 0
	req.CreatedAt = parseSQLiteTime(created)
	req.UpdatedAt = parseSQLiteTime(updated)
	return req, nil
}

// List returns requests matching filter, newest first.
func (r *RequestRepo) List(ctx context.Context, f RequestListFilter) ([]Request, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, user_id, media_kind, provider, external_id, title, poster_url,
			quality_resolution, quality_profile, status, collision_flag, admin_note,
			created_at, updated_at
		FROM requests WHERE 1=1
	`
	var args []any
	if f.UserID != "" {
		query += " AND user_id = ?"
		args = append(args, f.UserID)
	}
	if f.Status != "" {
		query += " AND status = ?"
		args = append(args, f.Status)
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, f.Offset)

	rows, err := r.q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db request list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Request
	for rows.Next() {
		req, err := scanRequestRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, req)
	}
	return out, rows.Err()
}

// Update applies mutable fields to an existing request.
func (r *RequestRepo) Update(ctx context.Context, id int64, patch RequestUpdate) error {
	cur, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if patch.QualityResolution != nil {
		cur.QualityResolution = *patch.QualityResolution
	}
	if patch.QualityProfile != nil {
		cur.QualityProfile = *patch.QualityProfile
	}
	if patch.AdminNote != nil {
		cur.AdminNote = *patch.AdminNote
	}
	if patch.Status != nil {
		cur.Status = *patch.Status
	}
	if patch.CollisionFlag != nil {
		cur.CollisionFlag = *patch.CollisionFlag
	}

	res, err := r.q.ExecContext(ctx, `
		UPDATE requests SET
			quality_resolution = ?,
			quality_profile = ?,
			status = ?,
			collision_flag = ?,
			admin_note = ?,
			updated_at = datetime('now')
		WHERE id = ?
	`, cur.QualityResolution, nullStr(cur.QualityProfile), cur.Status, boolToInt(cur.CollisionFlag),
		nullStr(cur.AdminNote), id)
	if err != nil {
		return fmt.Errorf("db request update: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db request update rows: %w", err)
	}
	if n == 0 {
		return ErrRequestNotFound
	}
	return nil
}

// SetStatus updates only the status column.
func (r *RequestRepo) SetStatus(ctx context.Context, id int64, status RequestStatus) error {
	res, err := r.q.ExecContext(ctx, `
		UPDATE requests SET status = ?, updated_at = datetime('now') WHERE id = ?
	`, status, id)
	if err != nil {
		return fmt.Errorf("db request set status: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db request set status rows: %w", err)
	}
	if n == 0 {
		return ErrRequestNotFound
	}
	return nil
}

// CountByUserInMonth returns requests created by userID in the given UTC month (YYYY-MM).
func (r *RequestRepo) CountByUserInMonth(ctx context.Context, userID, yearMonth string) (int64, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(yearMonth) == "" {
		return 0, fmt.Errorf("db request count: user id and year-month are required")
	}
	var n int64
	err := r.q.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM requests
		WHERE user_id = ? AND strftime('%Y-%m', created_at) = ?
	`, userID, yearMonth).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("db request count: %w", err)
	}
	return n, nil
}

// HasActiveByProviderExternal reports whether a non-terminal request exists for provider/external_id.
func (r *RequestRepo) HasActiveByProviderExternal(ctx context.Context, provider, externalID string) (bool, error) {
	var n int
	err := r.q.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM requests
		WHERE provider = ? AND external_id = ?
			AND status NOT IN ('rejected', 'cancelled', 'completed')
	`, provider, externalID).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("db request active lookup: %w", err)
	}
	return n > 0, nil
}

// GetPolicy returns the singleton request policy row.
func (r *RequestRepo) GetPolicy(ctx context.Context) (RequestPolicy, error) {
	var p RequestPolicy
	var updated string
	err := r.q.QueryRowContext(ctx, `
		SELECT auto_approve_roles, user_quota_per_month, admin_settings, updated_at
		FROM request_policies WHERE id = 1
	`).Scan(&p.AutoApproveRoles, &p.UserQuotaPerMonth, &p.AdminSettings, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return RequestPolicy{}, ErrRequestPolicyNotFound
	}
	if err != nil {
		return RequestPolicy{}, fmt.Errorf("db request policy get: %w", err)
	}
	p.UpdatedAt = parseSQLiteTime(updated)
	return p, nil
}

// UpsertPolicy replaces the singleton request policy row.
func (r *RequestRepo) UpsertPolicy(ctx context.Context, p RequestPolicy) error {
	if p.AutoApproveRoles == "" {
		p.AutoApproveRoles = "[]"
	}
	if p.AdminSettings == "" {
		p.AdminSettings = "{}"
	}
	res, err := r.q.ExecContext(ctx, `
		INSERT INTO request_policies (id, auto_approve_roles, user_quota_per_month, admin_settings, updated_at)
		VALUES (1, ?, ?, ?, datetime('now'))
		ON CONFLICT(id) DO UPDATE SET
			auto_approve_roles = excluded.auto_approve_roles,
			user_quota_per_month = excluded.user_quota_per_month,
			admin_settings = excluded.admin_settings,
			updated_at = datetime('now')
	`, p.AutoApproveRoles, p.UserQuotaPerMonth, p.AdminSettings)
	if err != nil {
		return fmt.Errorf("db request policy upsert: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db request policy upsert rows: %w", err)
	}
	if n == 0 {
		return ErrRequestPolicyNotFound
	}
	return nil
}

func scanRequestRow(rows *sql.Rows) (Request, error) {
	var req Request
	var poster, profile, note sql.NullString
	var collision int
	var created, updated string
	if err := rows.Scan(
		&req.ID, &req.UserID, &req.MediaKind, &req.Provider, &req.ExternalID, &req.Title, &poster,
		&req.QualityResolution, &profile, &req.Status, &collision, &note, &created, &updated,
	); err != nil {
		return Request{}, fmt.Errorf("db request scan: %w", err)
	}
	if poster.Valid {
		req.PosterURL = poster.String
	}
	if profile.Valid {
		req.QualityProfile = profile.String
	}
	if note.Valid {
		req.AdminNote = note.String
	}
	req.CollisionFlag = collision != 0
	req.CreatedAt = parseSQLiteTime(created)
	req.UpdatedAt = parseSQLiteTime(updated)
	return req, nil
}
