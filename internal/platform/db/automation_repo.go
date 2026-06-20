package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// HandoffStatus tracks automation handoff lifecycle.
type HandoffStatus string

const (
	HandoffPending    HandoffStatus = "pending"
	HandoffProcessing HandoffStatus = "processing"
	HandoffGrabbed    HandoffStatus = "grabbed"
	HandoffCompleted  HandoffStatus = "completed"
	HandoffFailed     HandoffStatus = "failed"
)

// AutomationHandoff records an approved request awaiting acquisition.
type AutomationHandoff struct {
	ID        int64         `json:"id"`
	RequestID int64         `json:"requestId"`
	Status    HandoffStatus `json:"status"`
	Error     string        `json:"error,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// AutomationDownloadStatus mirrors plugin download lifecycle in DB.
type AutomationDownloadStatus string

const (
	AutomationDownloadQueued      AutomationDownloadStatus = "queued"
	AutomationDownloadDownloading AutomationDownloadStatus = "downloading"
	AutomationDownloadPaused      AutomationDownloadStatus = "paused"
	AutomationDownloadCompleted   AutomationDownloadStatus = "completed"
	AutomationDownloadFailed      AutomationDownloadStatus = "failed"
)

// AutomationDownload tracks an active/completed client download.
type AutomationDownload struct {
	ID               int64                    `json:"id"`
	RequestID        *int64                   `json:"requestId,omitempty"`
	HandoffID        *int64                   `json:"handoffId,omitempty"`
	ClientInstanceID int64                    `json:"clientInstanceId"`
	ClientDownloadID string                   `json:"clientDownloadId"`
	ReleaseID        string                   `json:"releaseId"`
	Title            string                   `json:"title"`
	Protocol         string                   `json:"protocol"`
	Status           AutomationDownloadStatus `json:"status"`
	Progress         float64                  `json:"progress"`
	SavePath         string                   `json:"savePath,omitempty"`
	Error            string                   `json:"error,omitempty"`
	CreatedAt        time.Time                `json:"createdAt"`
	UpdatedAt        time.Time                `json:"updatedAt"`
}

// QualityProfile stores grab scoring preferences.
type QualityProfile struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Spec      string    `json:"spec"`
	IsDefault bool      `json:"isDefault"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AutomationRepo persists automation workflow state.
type AutomationRepo struct {
	q Querier
}

// NewAutomationRepo returns a repository bound to q.
func NewAutomationRepo(q Querier) *AutomationRepo {
	return &AutomationRepo{q: q}
}

var (
	ErrAutomationHandoffNotFound = errors.New("db automation handoff: not found")
	ErrQualityProfileNotFound    = errors.New("db quality profile: not found")
	ErrQualityProfileDuplicate   = errors.New("db quality profile: duplicate name")
)

// RecordHandoff inserts a pending handoff for requestID.
func (r *AutomationRepo) RecordHandoff(ctx context.Context, requestID int64) (int64, error) {
	res, err := r.q.ExecContext(ctx, `
INSERT INTO automation_handoffs (request_id, status)
VALUES (?, 'pending')
ON CONFLICT(request_id) DO UPDATE SET status='pending', error=NULL, updated_at=datetime('now')`,
		requestID)
	if err != nil {
		return 0, fmt.Errorf("db automation record handoff: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil || id == 0 {
		var existing int64
		err = r.q.QueryRowContext(ctx, `SELECT id FROM automation_handoffs WHERE request_id = ?`, requestID).Scan(&existing)
		if err != nil {
			return 0, fmt.Errorf("db automation record handoff id: %w", err)
		}
		return existing, nil
	}
	return id, nil
}

// ListPendingHandoffs returns handoffs awaiting processing.
func (r *AutomationRepo) ListPendingHandoffs(ctx context.Context, limit int) ([]AutomationHandoff, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.q.QueryContext(ctx, `
SELECT id, request_id, status, COALESCE(error,''), created_at, updated_at
FROM automation_handoffs
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanHandoffs(rows)
}

// UpdateHandoffStatus sets handoff status and optional error message.
func (r *AutomationRepo) UpdateHandoffStatus(ctx context.Context, id int64, status HandoffStatus, errMsg string) error {
	res, err := r.q.ExecContext(ctx, `
UPDATE automation_handoffs SET status = ?, error = ?, updated_at = datetime('now') WHERE id = ?`,
		status, nullIfEmpty(errMsg), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrAutomationHandoffNotFound
	}
	return nil
}

// CreateDownload inserts a download tracking row.
func (r *AutomationRepo) CreateDownload(ctx context.Context, d AutomationDownload) (int64, error) {
	res, err := r.q.ExecContext(ctx, `
INSERT INTO automation_downloads (
  request_id, handoff_id, client_instance_id, client_download_id,
  release_id, title, protocol, status, progress, save_path
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		nullableInt64(d.RequestID), nullableInt64(d.HandoffID), d.ClientInstanceID,
		d.ClientDownloadID, d.ReleaseID, d.Title, d.Protocol, d.Status, d.Progress, d.SavePath)
	if err != nil {
		return 0, fmt.Errorf("db automation create download: %w", err)
	}
	return res.LastInsertId()
}

// ListActiveDownloads returns non-terminal downloads.
func (r *AutomationRepo) ListActiveDownloads(ctx context.Context, limit int) ([]AutomationDownload, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.q.QueryContext(ctx, `
SELECT id, request_id, handoff_id, client_instance_id, client_download_id,
       release_id, title, protocol, status, progress, COALESCE(save_path,''), COALESCE(error,''),
       created_at, updated_at
FROM automation_downloads
WHERE status NOT IN ('completed', 'failed')
ORDER BY updated_at ASC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanDownloads(rows)
}

// ListDownloads returns recent downloads.
func (r *AutomationRepo) ListDownloads(ctx context.Context, limit int) ([]AutomationDownload, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.q.QueryContext(ctx, `
SELECT id, request_id, handoff_id, client_instance_id, client_download_id,
       release_id, title, protocol, status, progress, COALESCE(save_path,''), COALESCE(error,''),
       created_at, updated_at
FROM automation_downloads
ORDER BY updated_at DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanDownloads(rows)
}

// GetDownloadByID returns one download row.
func (r *AutomationRepo) GetDownloadByID(ctx context.Context, id int64) (AutomationDownload, error) {
	row := r.q.QueryRowContext(ctx, `
SELECT id, request_id, handoff_id, client_instance_id, client_download_id,
       release_id, title, protocol, status, progress, COALESCE(save_path,''), COALESCE(error,''),
       created_at, updated_at
FROM automation_downloads WHERE id = ?`, id)
	var d AutomationDownload
	var reqID, handoffID sql.NullInt64
	var created, updated string
	if err := row.Scan(&d.ID, &reqID, &handoffID, &d.ClientInstanceID, &d.ClientDownloadID,
		&d.ReleaseID, &d.Title, &d.Protocol, &d.Status, &d.Progress, &d.SavePath, &d.Error,
		&created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AutomationDownload{}, fmt.Errorf("db automation download: not found")
		}
		return AutomationDownload{}, err
	}
	if reqID.Valid {
		v := reqID.Int64
		d.RequestID = &v
	}
	if handoffID.Valid {
		v := handoffID.Int64
		d.HandoffID = &v
	}
	d.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	d.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
	return d, nil
}

// UpdateDownloadProgress updates status fields from client poll.
func (r *AutomationRepo) UpdateDownloadProgress(ctx context.Context, id int64, status AutomationDownloadStatus, progress float64, savePath, errMsg string) error {
	_, err := r.q.ExecContext(ctx, `
UPDATE automation_downloads
SET status = ?, progress = ?, save_path = ?, error = ?, updated_at = datetime('now')
WHERE id = ?`, status, progress, savePath, nullIfEmpty(errMsg), id)
	return err
}

// ListQualityProfiles returns all profiles.
func (r *AutomationRepo) ListQualityProfiles(ctx context.Context) ([]QualityProfile, error) {
	rows, err := r.q.QueryContext(ctx, `
SELECT id, name, spec, is_default, created_at, updated_at FROM quality_profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []QualityProfile
	for rows.Next() {
		var p QualityProfile
		var isDef int
		var created, updated string
		if err := rows.Scan(&p.ID, &p.Name, &p.Spec, &isDef, &created, &updated); err != nil {
			return nil, err
		}
		p.IsDefault = isDef != 0
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetQualityProfileByName returns a profile by name.
func (r *AutomationRepo) GetQualityProfileByName(ctx context.Context, name string) (QualityProfile, error) {
	row := r.q.QueryRowContext(ctx, `
SELECT id, name, spec, is_default, created_at, updated_at FROM quality_profiles WHERE name = ?`, name)
	var p QualityProfile
	var isDef int
	var created, updated string
	if err := row.Scan(&p.ID, &p.Name, &p.Spec, &isDef, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return QualityProfile{}, ErrQualityProfileNotFound
		}
		return QualityProfile{}, err
	}
	p.IsDefault = isDef != 0
	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
	return p, nil
}

// GetDefaultQualityProfile returns the default profile.
func (r *AutomationRepo) GetDefaultQualityProfile(ctx context.Context) (QualityProfile, error) {
	row := r.q.QueryRowContext(ctx, `
SELECT id, name, spec, is_default, created_at, updated_at FROM quality_profiles WHERE is_default = 1 LIMIT 1`)
	var p QualityProfile
	var isDef int
	var created, updated string
	if err := row.Scan(&p.ID, &p.Name, &p.Spec, &isDef, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return QualityProfile{}, ErrQualityProfileNotFound
		}
		return QualityProfile{}, err
	}
	p.IsDefault = isDef != 0
	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
	return p, nil
}

// CreateQualityProfile inserts a profile.
func (r *AutomationRepo) CreateQualityProfile(ctx context.Context, name, spec string, isDefault bool) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("db quality profile: name required")
	}
	if spec == "" {
		spec = "{}"
	}
	def := 0
	if isDefault {
		def = 1
	}
	res, err := r.q.ExecContext(ctx, `
INSERT INTO quality_profiles (name, spec, is_default) VALUES (?, ?, ?)`, name, spec, def)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, ErrQualityProfileDuplicate
		}
		return 0, err
	}
	return res.LastInsertId()
}

func scanHandoffs(rows *sql.Rows) ([]AutomationHandoff, error) {
	var out []AutomationHandoff
	for rows.Next() {
		var h AutomationHandoff
		var created, updated string
		if err := rows.Scan(&h.ID, &h.RequestID, &h.Status, &h.Error, &created, &updated); err != nil {
			return nil, err
		}
		h.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		h.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
		out = append(out, h)
	}
	return out, rows.Err()
}

func scanDownloads(rows *sql.Rows) ([]AutomationDownload, error) {
	var out []AutomationDownload
	for rows.Next() {
		var d AutomationDownload
		var reqID, handoffID sql.NullInt64
		var created, updated string
		if err := rows.Scan(&d.ID, &reqID, &handoffID, &d.ClientInstanceID, &d.ClientDownloadID,
			&d.ReleaseID, &d.Title, &d.Protocol, &d.Status, &d.Progress, &d.SavePath, &d.Error,
			&created, &updated); err != nil {
			return nil, err
		}
		if reqID.Valid {
			v := reqID.Int64
			d.RequestID = &v
		}
		if handoffID.Valid {
			v := handoffID.Int64
			d.HandoffID = &v
		}
		d.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		d.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
		out = append(out, d)
	}
	return out, rows.Err()
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableInt64(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}
