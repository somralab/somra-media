package plugin

import (
	"context"
	"time"
)

// DownloadStatus is the lifecycle state of a download item.
type DownloadStatus string

const (
	DownloadStatusQueued      DownloadStatus = "queued"
	DownloadStatusDownloading DownloadStatus = "downloading"
	DownloadStatusPaused      DownloadStatus = "paused"
	DownloadStatusCompleted   DownloadStatus = "completed"
	DownloadStatusFailed      DownloadStatus = "failed"
)

// AddRequest asks a download client to enqueue a release from indexer search.
type AddRequest struct {
	ReleaseID string   `json:"releaseId"`
	Title     string   `json:"title,omitempty"`
	Category  string   `json:"category,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Priority  int      `json:"priority,omitempty"`
}

// DownloadItem is the normalized status of a single download.
type DownloadItem struct {
	DownloadID      string         `json:"downloadId"`
	ClientID        string         `json:"clientId"`
	ReleaseID       string         `json:"releaseId,omitempty"`
	Status          DownloadStatus `json:"status"`
	Progress        float64        `json:"progress"`
	TotalBytes      int64          `json:"totalBytes,omitempty"`
	DownloadedBytes int64          `json:"downloadedBytes,omitempty"`
	SavePath        string         `json:"savePath,omitempty"`
	ErrorDetail     string         `json:"errorDetail,omitempty"`
	CompletedAt     *time.Time     `json:"completedAt,omitempty"`
}

// DownloadClient enqueues releases and reports per-item download status.
type DownloadClient interface {
	Plugin
	Add(ctx context.Context, req AddRequest) (DownloadItem, error)
	Status(ctx context.Context, downloadID string) (DownloadItem, error)
}
