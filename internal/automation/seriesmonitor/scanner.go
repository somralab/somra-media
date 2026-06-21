package seriesmonitor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/somralab/somra-media/internal/automation/grab"
	indexersearch "github.com/somralab/somra-media/internal/automation/indexer"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/requests"
)

// Scanner checks enabled series monitors and enqueues approved requests for new episodes.
type Scanner struct {
	AutoRepo *db.AutomationRepo
	Requests *db.RequestRepo
	Search   *indexersearch.SearchService
	Logger   *slog.Logger
}

// ScanEnabledMonitors searches indexers and records handoffs for newly discovered episodes.
func (s *Scanner) ScanEnabledMonitors(ctx context.Context) error {
	if s == nil || s.AutoRepo == nil || s.Requests == nil || s.Search == nil {
		return fmt.Errorf("series monitor scanner: dependencies required")
	}
	monitors, err := s.AutoRepo.ListEnabledMonitors(ctx, 50)
	if err != nil {
		return err
	}
	for _, m := range monitors {
		if err := s.scanOne(ctx, m); err != nil && s.Logger != nil {
			s.Logger.Warn("series monitor scan failed", slog.Int64("monitorId", m.ID), slog.Any("error", err))
		}
	}
	return nil
}

func (s *Scanner) scanOne(ctx context.Context, m db.AutomationMonitor) error {
	resp, err := s.Search.Search(ctx, indexersearch.SearchRequest{
		Query: plugin.SearchQuery{
			Title:     m.Title,
			MediaKind: plugin.MediaKindTV,
		},
	})
	if err != nil {
		return err
	}

	profile, err := s.resolveProfile(ctx, m.QualityProfile)
	if err != nil {
		return err
	}
	spec, err := grab.ParseProfileSpec(profile.Spec)
	if err != nil {
		return err
	}

	lastSeason, lastEpisode := m.LastSeason, m.LastEpisode
	for _, row := range resp.Results {
		season, episode, ok := ParseEpisode(row.Title)
		if !ok || !isNewerEpisode(m.LastSeason, m.LastEpisode, season, episode) {
			continue
		}
		exists, err := s.AutoRepo.MonitorEpisodeExists(ctx, m.ID, season, episode)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		best := grab.PickBest([]plugin.SearchResult{row}, spec, requests.Request{
			MediaKind:         requests.MediaKindTV,
			Title:             fmt.Sprintf("%s S%02dE%02d", m.Title, season, episode),
			QualityResolution: requests.QualityAny,
			QualityProfile:    m.QualityProfile,
		})
		if best == nil {
			continue
		}
		reqID, err := s.Requests.Create(ctx, db.Request{
			UserID:            m.UserID,
			MediaKind:         db.RequestMediaKindTV,
			Provider:          m.Provider,
			ExternalID:        m.ExternalID,
			Title:             fmt.Sprintf("%s S%02dE%02d", m.Title, season, episode),
			QualityResolution: db.RequestQualityAny,
			QualityProfile:    m.QualityProfile,
			Status:            db.RequestStatusApproved,
		})
		if err != nil {
			return err
		}
		if _, err := s.AutoRepo.RecordHandoff(ctx, reqID); err != nil {
			return err
		}
		if err := s.AutoRepo.RecordMonitorEpisode(ctx, m.ID, season, episode, reqID); err != nil {
			return err
		}
		if isNewerEpisode(lastSeason, lastEpisode, season, episode) {
			lastSeason, lastEpisode = season, episode
		}
	}
	return s.AutoRepo.UpdateMonitorProgress(ctx, m.ID, lastSeason, lastEpisode)
}

func (s *Scanner) resolveProfile(ctx context.Context, name string) (db.QualityProfile, error) {
	if name != "" {
		return s.AutoRepo.GetQualityProfileByName(ctx, name)
	}
	return s.AutoRepo.GetDefaultQualityProfile(ctx)
}
