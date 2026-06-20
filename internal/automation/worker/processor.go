package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/somralab/somra-media/internal/automation/grab"
	indexersearch "github.com/somralab/somra-media/internal/automation/indexer"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/requests"
)

// Processor handles pending automation handoffs.
type Processor struct {
	AutoRepo *db.AutomationRepo
	Requests *db.RequestRepo
	Search   *indexersearch.SearchService
	Manager  *plugin.Manager
	Logger   *slog.Logger
}

// ProcessPending grabs releases for pending handoffs.
func (p *Processor) ProcessPending(ctx context.Context) error {
	if p == nil || p.AutoRepo == nil || p.Requests == nil || p.Manager == nil {
		return fmt.Errorf("automation processor: dependencies required")
	}
	handoffs, err := p.AutoRepo.ListPendingHandoffs(ctx, 10)
	if err != nil {
		return err
	}
	for _, h := range handoffs {
		if err := p.processOne(ctx, h); err != nil && p.Logger != nil {
			p.Logger.Warn("automation handoff failed", slog.Int64("handoffId", h.ID), slog.Any("error", err))
		}
	}
	return nil
}

func (p *Processor) processOne(ctx context.Context, h db.AutomationHandoff) error {
	_ = p.AutoRepo.UpdateHandoffStatus(ctx, h.ID, db.HandoffProcessing, "")
	reqRow, err := p.Requests.GetByID(ctx, h.RequestID)
	if err != nil {
		_ = p.AutoRepo.UpdateHandoffStatus(ctx, h.ID, db.HandoffFailed, err.Error())
		return err
	}
	req := requests.Request{
		ID:                reqRow.ID,
		Status:            requests.Status(reqRow.Status),
		MediaKind:         requests.MediaKind(reqRow.MediaKind),
		Title:             reqRow.Title,
		QualityResolution: requests.QualityResolution(reqRow.QualityResolution),
		QualityProfile:    reqRow.QualityProfile,
	}
	profile, err := p.resolveProfile(ctx, req.QualityProfile)
	if err != nil {
		_ = p.AutoRepo.UpdateHandoffStatus(ctx, h.ID, db.HandoffFailed, err.Error())
		return err
	}
	spec, err := grab.ParseProfileSpec(profile.Spec)
	if err != nil {
		_ = p.AutoRepo.UpdateHandoffStatus(ctx, h.ID, db.HandoffFailed, err.Error())
		return err
	}
	searchResp, err := p.search(ctx, req)
	if err != nil {
		_ = p.AutoRepo.UpdateHandoffStatus(ctx, h.ID, db.HandoffFailed, err.Error())
		return err
	}
	best := grab.PickBest(searchResp.Results, spec, req)
	if best == nil {
		_ = p.AutoRepo.UpdateHandoffStatus(ctx, h.ID, db.HandoffFailed, "no suitable release")
		return fmt.Errorf("no suitable release for request %d", req.ID)
	}
	client, instanceID, err := p.pickClient(best.Protocol)
	if err != nil {
		_ = p.AutoRepo.UpdateHandoffStatus(ctx, h.ID, db.HandoffFailed, err.Error())
		return err
	}
	item, err := client.Add(ctx, plugin.AddRequest{
		ReleaseID: best.ReleaseID,
		Title:     best.Title,
		Labels:    []string{fmt.Sprintf("request-%d", req.ID)},
	})
	if err != nil {
		_ = p.AutoRepo.UpdateHandoffStatus(ctx, h.ID, db.HandoffFailed, err.Error())
		return err
	}
	reqID := req.ID
	handoffID := h.ID
	_, err = p.AutoRepo.CreateDownload(ctx, db.AutomationDownload{
		RequestID:        &reqID,
		HandoffID:        &handoffID,
		ClientInstanceID: instanceID,
		ClientDownloadID: item.DownloadID,
		ReleaseID:        best.ReleaseID,
		Title:            best.Title,
		Protocol:         string(best.Protocol),
		Status:           db.AutomationDownloadQueued,
		Progress:         item.Progress,
		SavePath:         item.SavePath,
	})
	if err != nil {
		_ = p.AutoRepo.UpdateHandoffStatus(ctx, h.ID, db.HandoffFailed, err.Error())
		return err
	}
	return p.AutoRepo.UpdateHandoffStatus(ctx, h.ID, db.HandoffGrabbed, "")
}

func (p *Processor) search(ctx context.Context, req requests.Request) (indexersearch.SearchResponse, error) {
	if p.Search == nil {
		return indexersearch.SearchResponse{}, fmt.Errorf("search service unavailable")
	}
	q := plugin.SearchQuery{
		Title:     req.Title,
		MediaKind: plugin.MediaKind(req.MediaKind),
	}
	return p.Search.Search(ctx, indexersearch.SearchRequest{Query: q})
}

func (p *Processor) resolveProfile(ctx context.Context, name string) (db.QualityProfile, error) {
	if name != "" {
		return p.AutoRepo.GetQualityProfileByName(ctx, name)
	}
	return p.AutoRepo.GetDefaultQualityProfile(ctx)
}

func (p *Processor) pickClient(proto plugin.Protocol) (plugin.DownloadClient, int64, error) {
	clients := p.Manager.EnabledDownloadClients()
	for _, c := range clients {
		id, err := strconv.ParseInt(c.ID(), 10, 64)
		if err != nil {
			continue
		}
		rec, err := p.Manager.Get(context.Background(), id)
		if err != nil {
			continue
		}
		switch proto {
		case plugin.ProtocolTorrent:
			if rec.Implementation == "qbittorrent" || rec.Implementation == "stub" {
				return c, id, nil
			}
		case plugin.ProtocolUsenet:
			if rec.Implementation == "sabnzbd" || rec.Implementation == "stub" {
				return c, id, nil
			}
		}
	}
	return nil, 0, fmt.Errorf("no enabled download client for protocol %s", proto)
}
