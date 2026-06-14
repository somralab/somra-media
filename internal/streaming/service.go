package streaming

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/somralab/somra-media/internal/platform/db"
)

// ServiceConfig configures the streaming service.
type ServiceConfig struct {
	CacheDir          string
	SessionTTL        time.Duration
	IdleTimeout       time.Duration
	MaxConcurrent     int
	MaxHWConcurrent   int
	MaxTranscodeQueue int
	FFmpegBin         string
	FFprobeBin        string
}

// Service orchestrates playback sessions.
type Service struct {
	cfg     ServiceConfig
	repo    *db.PlaybackRepo
	media   *db.MediaRepo
	decide  *DecisionEngine
	procMgr *ProcessManager
	metrics *Metrics
	logger  *slog.Logger

	hwMu     sync.RWMutex
	hwConfig HWRuntimeConfig
	hwActive int

	queue   []queuedJob
	queueMu sync.Mutex
}

type queuedJob struct {
	sessionID string
	fn        func(context.Context) error
}

// NewService wires streaming dependencies.
func NewService(cfg ServiceConfig, repo *db.PlaybackRepo, media *db.MediaRepo, logger *slog.Logger) *Service {
	if cfg.SessionTTL <= 0 {
		cfg.SessionTTL = 4 * time.Hour
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 15 * time.Minute
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 2
	}
	if cfg.MaxHWConcurrent <= 0 {
		cfg.MaxHWConcurrent = 2
	}
	if cfg.MaxTranscodeQueue <= 0 {
		cfg.MaxTranscodeQueue = 8
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		cfg:     cfg,
		repo:    repo,
		media:   media,
		decide:  NewDecisionEngine(),
		procMgr: NewProcessManager(ProcessManagerConfig{MaxConcurrent: cfg.MaxConcurrent, FFmpegBin: cfg.FFmpegBin}),
		metrics: NewMetrics(),
		logger:  logger,
		hwConfig: HWRuntimeConfig{
			Mode:             HWModeAuto,
			MaxHWSessions:    cfg.MaxHWConcurrent,
			MaxTotalSessions: cfg.MaxConcurrent,
			VAAPIDevice:      vaapiDevice(),
		},
	}
}

// ApplyRuntimeSettings syncs streaming limits from persisted settings.
func (s *Service) ApplyRuntimeSettings(cfg HWRuntimeConfig) {
	s.hwMu.Lock()
	defer s.hwMu.Unlock()
	if cfg.MaxTotalSessions > 0 {
		s.cfg.MaxConcurrent = cfg.MaxTotalSessions
		s.procMgr.SetMaxConcurrent(cfg.MaxTotalSessions)
	}
	if cfg.MaxHWSessions > 0 {
		s.hwConfig.MaxHWSessions = cfg.MaxHWSessions
	}
	if cfg.Mode != "" {
		s.hwConfig.Mode = cfg.Mode
	}
	if cfg.Preferred != "" {
		s.hwConfig.Preferred = cfg.Preferred
	}
	if len(cfg.Available) > 0 {
		s.hwConfig.Available = append([]Accelerator(nil), cfg.Available...)
	}
	if cfg.VAAPIDevice != "" {
		s.hwConfig.VAAPIDevice = cfg.VAAPIDevice
	}
	s.hwConfig.MaxTotalSessions = s.cfg.MaxConcurrent
}

func (s *Service) currentHWConfig() HWRuntimeConfig {
	s.hwMu.RLock()
	defer s.hwMu.RUnlock()
	return s.hwConfig
}

func (s *Service) activeHWCount() int {
	s.hwMu.RLock()
	defer s.hwMu.RUnlock()
	return s.hwActive
}

func (s *Service) incHWActive() {
	s.hwMu.Lock()
	s.hwActive++
	s.hwMu.Unlock()
	s.metrics.incHWActive()
}

func (s *Service) decHWActive() {
	s.hwMu.Lock()
	if s.hwActive > 0 {
		s.hwActive--
	}
	s.hwMu.Unlock()
	s.metrics.decHWActive()
}

// Metrics returns service telemetry.
func (s *Service) Metrics() *Metrics {
	return s.metrics
}

// PlayRequest starts or resumes playback for a media item.
type PlayRequest struct {
	UserID              string
	MediaItemID         int64
	Capabilities        ClientCapabilities
	AudioStreamIndex    *int
	SubtitleStreamIndex *int
	StartPositionMs     int64
}

// PlayResponse describes a started session.
type PlayResponse struct {
	SessionID   string
	Mode        Mode
	ManifestURL string
	ExpiresAt   time.Time
	Decision    Decision
}

// StartPlay validates access, decides mode, and creates a session.
func (s *Service) StartPlay(ctx context.Context, req PlayRequest) (PlayResponse, error) {
	if req.UserID == "" || req.MediaItemID <= 0 {
		return PlayResponse{}, fmt.Errorf("streaming play: invalid request")
	}
	caps := req.Capabilities
	if len(caps.VideoCodecs) == 0 {
		caps = DefaultBrowserCapabilities()
	}

	file, err := s.media.GetPrimaryFileByItemID(ctx, req.MediaItemID)
	if err != nil {
		return PlayResponse{}, err
	}
	tech, err := s.media.GetTechnicalByFileID(ctx, file.ID)
	if err != nil {
		return PlayResponse{}, err
	}

	probe := MediaProbe{
		Container: tech.Container, VideoCodec: tech.VideoCodec,
		VideoWidth: tech.VideoWidth, VideoHeight: tech.VideoHeight,
		AudioCodec: tech.AudioCodec, AudioChannels: tech.AudioChannels,
		DurationMs: tech.DurationMs,
	}
	decision := s.decide.Decide(caps, probe)

	sessionID := uuid.NewString()
	expires := time.Now().UTC().Add(s.cfg.SessionTTL)
	cachePath := SessionCacheDir(s.cfg.CacheDir, sessionID)

	status := db.PlaybackActive
	if decision.Mode == ModeTranscode {
		n, err := s.repo.CountActiveTranscodes(ctx)
		if err != nil {
			return PlayResponse{}, err
		}
		if n >= s.cfg.MaxConcurrent {
			status = db.PlaybackQueued
		}
	}

	rec := db.PlaybackSession{
		ID: sessionID, UserID: req.UserID, MediaItemID: req.MediaItemID,
		MediaFileID: file.ID, Mode: db.PlaybackMode(decision.Mode), Status: status,
		CachePath: cachePath, StartPositionMs: req.StartPositionMs,
		AudioStreamIndex: req.AudioStreamIndex, SubtitleStreamIndex: req.SubtitleStreamIndex,
		ExpiresAt: expires,
	}
	if err := s.repo.Create(ctx, rec); err != nil {
		return PlayResponse{}, err
	}
	s.metrics.incActive()

	packFn := func(runCtx context.Context) error {
		if err := EnsureOutputDir(cachePath); err != nil {
			return err
		}
		if decision.Mode == ModeDirectPlay {
			if err := WriteDirectPlayManifest(cachePath); err != nil {
				return err
			}
			return osLinkSource(runCtx, file.Path, cachePath)
		}
		tiers := BuildLadder(probe.VideoWidth, probe.VideoHeight)
		if err := s.procMgr.Acquire(runCtx); err != nil {
			return err
		}
		s.metrics.incStarts()
		hwCfg := s.currentHWConfig()
		tpath := SelectTranscodePath(hwCfg, probe, s.activeHWCount())
		opts := PackagerOptions{
			SourcePath: file.Path, OutputDir: cachePath, Mode: decision.Mode,
			StartPositionMs: req.StartPositionMs, Tiers: tiers,
			AudioStreamIndex: req.AudioStreamIndex, SubtitleStreamIndex: req.SubtitleStreamIndex,
			TranscodePath: tpath, HWConfig: hwCfg,
		}
		if err := s.startTranscodeWithFallback(runCtx, sessionID, opts, tpath); err != nil {
			s.metrics.incErrors()
			return err
		}
		if err := s.repo.UpdateStatus(runCtx, sessionID, db.PlaybackActive, ""); err != nil {
			return err
		}
		return nil
	}

	if status == db.PlaybackQueued {
		s.enqueue(sessionID, packFn)
		s.updateQueueMetric()
	} else {
		go s.runPackaging(sessionID, packFn)
	}

	manifestURL := fmt.Sprintf("/api/v1/streaming/sessions/%s/master.m3u8", sessionID)
	return PlayResponse{
		SessionID: sessionID, Mode: decision.Mode,
		ManifestURL: manifestURL, ExpiresAt: expires, Decision: decision,
	}, nil
}

func osLinkSource(_ context.Context, src, cacheDir string) error {
	dst := filepathJoin(cacheDir, "source")
	if _, err := osStat(dst); err == nil {
		return nil
	}
	return osSymlinkOrCopy(src, dst)
}

func (s *Service) startTranscodeWithFallback(ctx context.Context, sessionID string, opts PackagerOptions, tpath TranscodePath) error {
	if tpath.UseHW {
		s.metrics.incHWStarts()
		s.incHWActive()
		defer s.decHWActive()
		s.logger.Info("streaming hw transcode start",
			slog.String("session", sessionID),
			slog.String("hw_accelerator", string(tpath.Accelerator)),
			slog.String("hw_encoder", tpath.VideoEncoder),
			slog.String("selection", tpath.SelectionNote),
		)
		if err := StartPackaging(ctx, s.procMgr, opts); err == nil {
			if werr := WaitForManifest(opts.OutputDir, 30*time.Second); werr == nil {
				return nil
			}
		} else {
			s.metrics.incHWErrors()
			s.logger.Warn("streaming hw transcode failed",
				slog.String("session", sessionID),
				slog.String("hw_accelerator", string(tpath.Accelerator)),
				slog.Bool("hw_fallback", true),
			)
		}
		s.metrics.incHWFallbacks()
	}

	swPath := TranscodePath{UseHW: false, VideoEncoder: "libx264", SelectionNote: "sw_fallback"}
	opts.TranscodePath = swPath
	s.logger.Info("streaming sw fallback transcode",
		slog.String("session", sessionID),
		slog.Bool("hw_fallback", true),
	)
	if err := StartPackaging(ctx, s.procMgr, opts); err != nil {
		return err
	}
	return WaitForManifest(opts.OutputDir, 30*time.Second)
}

// StopSession terminates a session and removes cache.
func (s *Service) StopSession(ctx context.Context, sessionID, userID string) error {
	sess, err := s.repo.GetByIDForUser(ctx, sessionID, userID)
	if err != nil {
		return err
	}
	s.procMgr.Stop(sessionID)
	if err := RemoveDir(sess.CachePath); err != nil {
		s.logger.Warn("streaming cache cleanup", slog.String("session", sessionID), slog.Any("error", err))
	}
	if err := s.repo.Stop(ctx, sessionID); err != nil {
		return err
	}
	s.metrics.decActive()
	return nil
}

// GetSession returns a session owned by the user.
func (s *Service) GetSession(ctx context.Context, sessionID, userID string) (db.PlaybackSession, error) {
	return s.repo.GetByIDForUser(ctx, sessionID, userID)
}

// TouchSession updates last access time.
func (s *Service) TouchSession(ctx context.Context, sessionID string) error {
	return s.repo.TouchLastAccess(ctx, sessionID)
}

// ReapIdle stops sessions idle beyond configured timeout.
func (s *Service) ReapIdle(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-s.cfg.IdleTimeout)
	sessions, err := s.repo.ListIdleSessions(ctx, cutoff)
	if err != nil {
		return err
	}
	for _, sess := range sessions {
		s.procMgr.Stop(sess.ID)
		_ = RemoveDir(sess.CachePath)
		_ = s.repo.UpdateStatus(ctx, sess.ID, db.PlaybackExpired, "idle_timeout")
		s.metrics.decActive()
	}
	expired, err := s.repo.ListExpired(ctx, time.Now().UTC())
	if err != nil {
		return err
	}
	for _, sess := range expired {
		s.procMgr.Stop(sess.ID)
		_ = RemoveDir(sess.CachePath)
		_ = s.repo.UpdateStatus(ctx, sess.ID, db.PlaybackExpired, "expired")
		s.metrics.decActive()
	}
	return nil
}

func (s *Service) enqueue(sessionID string, fn func(context.Context) error) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()
	if len(s.queue) >= s.cfg.MaxTranscodeQueue {
		return
	}
	s.queue = append(s.queue, queuedJob{sessionID: sessionID, fn: fn})
}

func (s *Service) updateQueueMetric() {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()
	s.metrics.setQueue(int64(len(s.queue)))
}

func (s *Service) runPackaging(sessionID string, fn func(context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := fn(ctx); err != nil {
		s.logger.Error("streaming packaging failed", slog.String("session", sessionID), slog.Any("error", err))
		_ = s.repo.UpdateStatus(context.Background(), sessionID, db.PlaybackFailed, err.Error())
		s.metrics.incErrors()
		return
	}
	s.drainQueue()
}

func (s *Service) drainQueue() {
	s.queueMu.Lock()
	if len(s.queue) == 0 {
		s.queueMu.Unlock()
		s.updateQueueMetric()
		return
	}
	job := s.queue[0]
	s.queue = s.queue[1:]
	s.queueMu.Unlock()
	s.updateQueueMetric()
	go s.runPackaging(job.sessionID, job.fn)
}

// ErrQueueFull indicates transcode queue capacity reached.
var ErrQueueFull = errors.New("streaming: transcode queue full")

// os helpers extracted for testing.
var (
	osStat          = osStatDefault
	osSymlinkOrCopy = symlinkOrCopyDefault
	filepathJoin    = filepath.Join
)

func osStatDefault(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func symlinkOrCopyDefault(src, dst string) error {
	if err := os.Symlink(src, dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("streaming link source: %w", err)
	}
	return os.WriteFile(dst, data, 0o640)
}
