package download

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/somralab/somra-media/internal/automation/importsvc"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
)

// Monitor polls active downloads and triggers import on completion.
type Monitor struct {
	AutoRepo *db.AutomationRepo
	Requests *db.RequestRepo
	Manager  *plugin.Manager
	Import   *importsvc.Service
	Logger   *slog.Logger
}

// Poll updates download rows from client status.
func (m *Monitor) Poll(ctx context.Context) error {
	if m == nil || m.AutoRepo == nil || m.Manager == nil {
		return fmt.Errorf("automation download monitor: dependencies required")
	}
	rows, err := m.AutoRepo.ListActiveDownloads(ctx, 50)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := m.pollOne(ctx, row); err != nil && m.Logger != nil {
			m.Logger.Warn("download poll failed", slog.Int64("downloadId", row.ID), slog.Any("error", err))
		}
	}
	return nil
}

func (m *Monitor) pollOne(ctx context.Context, row db.AutomationDownload) error {
	client, err := m.Manager.DownloadClient(ctx, row.ClientInstanceID)
	if err != nil {
		return err
	}
	item, err := client.Status(ctx, row.ClientDownloadID)
	if err != nil {
		return err
	}
	status := mapStatus(item.Status)
	if err := m.AutoRepo.UpdateDownloadProgress(ctx, row.ID, status, item.Progress, item.SavePath, item.ErrorDetail); err != nil {
		return err
	}
	if status != db.AutomationDownloadCompleted {
		return nil
	}
	if m.Import != nil && item.SavePath != "" {
		if err := m.Import.ImportCompleted(ctx, item.SavePath); err != nil && m.Logger != nil {
			m.Logger.Warn("automation import failed", slog.Any("error", err))
		}
	}
	if row.HandoffID != nil {
		_ = m.AutoRepo.UpdateHandoffStatus(ctx, *row.HandoffID, db.HandoffCompleted, "")
	}
	if row.RequestID != nil && m.Requests != nil {
		st := db.RequestStatusCompleted
		_ = m.Requests.Update(ctx, *row.RequestID, db.RequestUpdate{Status: &st})
	}
	return nil
}

func mapStatus(s plugin.DownloadStatus) db.AutomationDownloadStatus {
	switch s {
	case plugin.DownloadStatusDownloading:
		return db.AutomationDownloadDownloading
	case plugin.DownloadStatusPaused:
		return db.AutomationDownloadPaused
	case plugin.DownloadStatusCompleted:
		return db.AutomationDownloadCompleted
	case plugin.DownloadStatusFailed:
		return db.AutomationDownloadFailed
	default:
		return db.AutomationDownloadQueued
	}
}
