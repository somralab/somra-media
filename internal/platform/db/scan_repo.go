package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ScanType identifies full vs incremental scans.
type ScanType string

const (
	ScanFull        ScanType = "full"
	ScanIncremental ScanType = "incremental"
)

// ScanStatus tracks scan run lifecycle.
type ScanStatus string

const (
	ScanPending   ScanStatus = "pending"
	ScanRunning   ScanStatus = "running"
	ScanSucceeded ScanStatus = "succeeded"
	ScanFailed    ScanStatus = "failed"
	ScanCancelled ScanStatus = "cancelled"
)

// ScanRun records a library scan execution.
type ScanRun struct {
	ID         int64      `json:"id"`
	LibraryID  int64      `json:"libraryId"`
	TaskID     string     `json:"taskId,omitempty"`
	ScanType   ScanType   `json:"scanType"`
	Status     ScanStatus `json:"status"`
	FilesTotal int        `json:"filesTotal"`
	FilesDone  int        `json:"filesDone"`
	Error      string     `json:"error,omitempty"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt,omitempty"`
}

// ScanRepo persists scan history and progress.
type ScanRepo struct {
	q Querier
}

// NewScanRepo returns a repository bound to q.
func NewScanRepo(q Querier) *ScanRepo {
	return &ScanRepo{q: q}
}

var ErrScanRunNotFound = errors.New("db scan: not found")

// CreateRun inserts a new scan run.
func (r *ScanRepo) CreateRun(ctx context.Context, libraryID int64, scanType ScanType, taskID string) (int64, error) {
	res, err := r.q.ExecContext(ctx, `
		INSERT INTO scan_run (library_id, task_id, scan_type, status, created_at)
		VALUES (?, ?, ?, 'pending', datetime('now'))
	`, libraryID, nullStr(taskID), scanType)
	if err != nil {
		return 0, fmt.Errorf("db scan create: %w", err)
	}
	return res.LastInsertId()
}

// MarkRunning transitions a run to running.
func (r *ScanRepo) MarkRunning(ctx context.Context, id int64) error {
	_, err := r.q.ExecContext(ctx, `
		UPDATE scan_run SET status = 'running', started_at = datetime('now') WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("db scan mark running %d: %w", id, err)
	}
	return nil
}

// UpdateProgress updates file counters.
func (r *ScanRepo) UpdateProgress(ctx context.Context, id int64, total, done int) error {
	_, err := r.q.ExecContext(ctx, `
		UPDATE scan_run SET files_total = ?, files_done = ? WHERE id = ?
	`, total, done, id)
	if err != nil {
		return fmt.Errorf("db scan progress %d: %w", id, err)
	}
	return nil
}

// Finish marks a run complete.
func (r *ScanRepo) Finish(ctx context.Context, id int64, status ScanStatus, errMsg string) error {
	_, err := r.q.ExecContext(ctx, `
		UPDATE scan_run SET status = ?, error = ?, finished_at = datetime('now') WHERE id = ?
	`, status, nullStr(errMsg), id)
	if err != nil {
		return fmt.Errorf("db scan finish %d: %w", id, err)
	}
	return nil
}

// GetByID returns a scan run.
func (r *ScanRepo) GetByID(ctx context.Context, id int64) (ScanRun, error) {
	return r.scanOne(ctx, `SELECT id, library_id, task_id, scan_type, status,
		files_total, files_done, error, started_at, finished_at, created_at
		FROM scan_run WHERE id = ?`, id)
}

// ListByLibrary returns recent scan runs.
func (r *ScanRepo) ListByLibrary(ctx context.Context, libraryID int64, limit int) ([]ScanRun, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, library_id, task_id, scan_type, status,
		       files_total, files_done, error, started_at, finished_at, created_at
		FROM scan_run WHERE library_id = ?
		ORDER BY created_at DESC LIMIT ?
	`, libraryID, limit)
	if err != nil {
		return nil, fmt.Errorf("db scan list %d: %w", libraryID, err)
	}
	defer func() { _ = rows.Close() }()
	return scanRuns(rows)
}

func (r *ScanRepo) scanOne(ctx context.Context, q string, id int64) (ScanRun, error) {
	rows, err := r.q.QueryContext(ctx, q, id)
	if err != nil {
		return ScanRun{}, fmt.Errorf("db scan get %d: %w", id, err)
	}
	defer func() { _ = rows.Close() }()
	runs, err := scanRuns(rows)
	if err != nil {
		return ScanRun{}, err
	}
	if len(runs) == 0 {
		return ScanRun{}, fmt.Errorf("db scan get %d: %w", id, ErrScanRunNotFound)
	}
	return runs[0], nil
}

func scanRuns(rows *sql.Rows) ([]ScanRun, error) {
	var out []ScanRun
	for rows.Next() {
		var sr ScanRun
		var taskID, errMsg sql.NullString
		var started, finished, created sql.NullString
		if err := rows.Scan(
			&sr.ID, &sr.LibraryID, &taskID, &sr.ScanType, &sr.Status,
			&sr.FilesTotal, &sr.FilesDone, &errMsg, &started, &finished, &created,
		); err != nil {
			return nil, fmt.Errorf("db scan scan row: %w", err)
		}
		sr.TaskID = taskID.String
		sr.Error = errMsg.String
		sr.StartedAt = parseTimePtr(started)
		sr.FinishedAt = parseTimePtr(finished)
		if created.Valid {
			sr.CreatedAt, _ = time.Parse(time.RFC3339, created.String)
		}
		out = append(out, sr)
	}
	return out, rows.Err()
}

func parseTimePtr(v sql.NullString) *time.Time {
	if !v.Valid || v.String == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, v.String)
	if err != nil {
		return nil
	}
	return &t
}
