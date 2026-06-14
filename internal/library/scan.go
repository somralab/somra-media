package library

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/somralab/somra-media/internal/jobs"
	"github.com/somralab/somra-media/internal/platform/db"
)

// ProgressEvent is emitted during scans for SSE consumers.
type ProgressEvent struct {
	LibraryID  int64  `json:"libraryId"`
	ScanRunID  int64  `json:"scanRunId"`
	FilesTotal int    `json:"filesTotal"`
	FilesDone  int    `json:"filesDone"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
}

// ProgressPublisher receives scan progress updates.
type ProgressPublisher interface {
	PublishScanProgress(ctx context.Context, ev ProgressEvent)
}

// NoopProgressPublisher discards progress events.
type NoopProgressPublisher struct{}

func (NoopProgressPublisher) PublishScanProgress(context.Context, ProgressEvent) {}

// FileProber extracts technical metadata from media files.
type FileProber interface {
	Probe(ctx context.Context, path string) (ProbeResult, error)
}

// Scanner orchestrates library file scans.
type Scanner struct {
	logger    *slog.Logger
	db        *db.DB
	prober    FileProber
	progress  ProgressPublisher
	hashFiles bool
}

// ScannerConfig configures a Scanner.
type ScannerConfig struct {
	Logger    *slog.Logger
	DB        *db.DB
	Prober    FileProber
	Progress  ProgressPublisher
	HashFiles bool
}

// NewScanner builds a Scanner.
func NewScanner(cfg ScannerConfig) *Scanner {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Prober == nil {
		cfg.Prober = NewProber("")
	}
	if cfg.Progress == nil {
		cfg.Progress = NoopProgressPublisher{}
	}
	return &Scanner{
		logger:    cfg.Logger,
		db:        cfg.DB,
		prober:    cfg.Prober,
		progress:  cfg.Progress,
		hashFiles: cfg.HashFiles,
	}
}

// ScanJob implements jobs.Job for queued scans.
type ScanJob struct {
	Scanner   *Scanner
	LibraryID int64
	ScanType  db.ScanType
	RunID     int64
}

// Run executes the scan.
func (j *ScanJob) Run(ctx context.Context) error {
	return j.Scanner.run(ctx, j.LibraryID, j.ScanType, j.RunID)
}

func (s *Scanner) run(ctx context.Context, libraryID int64, scanType db.ScanType, runID int64) error {
	libRepo := db.NewLibraryRepo(s.db.Querier())
	mediaRepo := db.NewMediaRepo(s.db.Querier())
	scanRepo := db.NewScanRepo(s.db.Querier())

	lib, err := libRepo.GetByID(ctx, libraryID)
	if err != nil {
		return fmt.Errorf("scan library %d: %w", libraryID, err)
	}

	if err := scanRepo.MarkRunning(ctx, runID); err != nil {
		return err
	}
	s.publish(ctx, ProgressEvent{LibraryID: libraryID, ScanRunID: runID, Status: "running"})

	var paths []string
	if err := DiscoverMedia(ctx, lib.Paths, func(path string, _ fs.DirEntry) error {
		paths = append(paths, path)
		return nil
	}); err != nil {
		_ = scanRepo.Finish(ctx, runID, db.ScanFailed, err.Error())
		return err
	}

	total := len(paths)
	_ = scanRepo.UpdateProgress(ctx, runID, total, 0)

	done := 0
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			_ = scanRepo.Finish(ctx, runID, db.ScanCancelled, err.Error())
			return err
		}

		if scanType == db.ScanIncremental {
			skip, err := s.shouldSkipIncremental(ctx, mediaRepo, path)
			if err != nil {
				s.logger.Warn("scan incremental check", slog.String("path", path), slog.Any("error", err))
			}
			if skip {
				done++
				continue
			}
		}

		if err := s.processFile(ctx, lib, mediaRepo, path); err != nil {
			s.logger.Warn("scan file failed", slog.String("path", path), slog.Any("error", err))
		}
		done++
		_ = scanRepo.UpdateProgress(ctx, runID, total, done)
		s.publish(ctx, ProgressEvent{
			LibraryID: libraryID, ScanRunID: runID,
			FilesTotal: total, FilesDone: done, Status: "running",
		})
	}

	_ = scanRepo.Finish(ctx, runID, db.ScanSucceeded, "")
	s.publish(ctx, ProgressEvent{
		LibraryID: libraryID, ScanRunID: runID,
		FilesTotal: total, FilesDone: done, Status: "succeeded",
	})
	return nil
}

func (s *Scanner) shouldSkipIncremental(ctx context.Context, repo *db.MediaRepo, path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	existing, err := repo.GetFileByPath(ctx, path)
	if err != nil {
		return false, nil
	}
	if existing.SizeBytes != info.Size() || existing.MtimeNs != info.ModTime().UnixNano() {
		return false, nil
	}
	return true, nil
}

func (s *Scanner) processFile(ctx context.Context, lib db.Library, repo *db.MediaRepo, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	parsed := ParseFileName(path)
	hash := ""
	if s.hashFiles {
		hash, _ = fileHash(path)
	}

	file := db.MediaFile{
		LibraryID:     lib.ID,
		Path:          path,
		FileName:      filepath.Base(path),
		SizeBytes:     info.Size(),
		MtimeNs:       info.ModTime().UnixNano(),
		ContentHash:   hash,
		ParsedTitle:   parsed.Title,
		ParsedYear:    parsed.Year,
		ParsedSeason:  parsed.Season,
		ParsedEpisode: parsed.Episode,
	}

	fileID, err := repo.UpsertFile(ctx, file)
	if err != nil {
		return err
	}

	itemID, err := s.ensureMediaItem(ctx, lib, repo, parsed)
	if err != nil {
		return err
	}
	if itemID > 0 {
		_, err = repo.UpsertFile(ctx, db.MediaFile{
			LibraryID: lib.ID, MediaItemID: &itemID,
			Path: path, FileName: file.FileName,
			SizeBytes: file.SizeBytes, MtimeNs: file.MtimeNs,
			ContentHash: hash, ParsedTitle: parsed.Title,
			ParsedYear: parsed.Year, ParsedSeason: parsed.Season, ParsedEpisode: parsed.Episode,
		})
		if err != nil {
			return err
		}
	}

	probe, err := s.prober.Probe(ctx, path)
	if err != nil {
		return fmt.Errorf("probe %q: %w", path, err)
	}
	return repo.UpsertTechnical(ctx, fileID,
		probe.DurationMs, probe.Container, probe.VideoCodec,
		probe.VideoWidth, probe.VideoHeight,
		probe.AudioCodec, probe.AudioChannels, probe.SubtitleCount,
		probe.RawJSON,
	)
}

func (s *Scanner) ensureMediaItem(ctx context.Context, lib db.Library, repo *db.MediaRepo, parsed ParsedName) (int64, error) {
	title := parsed.Title
	if title == "" {
		title = "Unknown"
	}
	itemID, err := repo.CreateItem(ctx, lib.ID, lib.Kind, title, parsed.Year)
	if err != nil {
		return 0, err
	}
	_ = repo.IndexFTS(ctx, itemID, title)
	return itemID, nil
}

func (s *Scanner) publish(ctx context.Context, ev ProgressEvent) {
	s.progress.PublishScanProgress(ctx, ev)
}

func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, io.LimitReader(f, 1<<20)); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Service exposes library CRUD and scan triggering.
type Service struct {
	logger   *slog.Logger
	db       *db.DB
	queue    jobs.JobQueue
	scanner  *Scanner
	watchMu  sync.Mutex
	watching map[int64]*Watcher
}

// ServiceConfig configures Service.
type ServiceConfig struct {
	Logger   *slog.Logger
	DB       *db.DB
	Queue    jobs.JobQueue
	Scanner  *Scanner
	Debounce time.Duration
}

// NewService builds a library service.
func NewService(cfg ServiceConfig) *Service {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	s := &Service{
		logger:   cfg.Logger,
		db:       cfg.DB,
		queue:    cfg.Queue,
		scanner:  cfg.Scanner,
		watching: make(map[int64]*Watcher),
	}
	if cfg.Debounce <= 0 {
		cfg.Debounce = 2 * time.Second
	}
	return s
}

// CreateLibrary validates paths and persists a library.
func (s *Service) CreateLibrary(ctx context.Context, name string, kind db.LibraryKind, paths []string, watchEnabled bool) (db.Library, error) {
	validated, err := validatePaths(paths)
	if err != nil {
		return db.Library{}, err
	}
	repo := db.NewLibraryRepo(s.db.Querier())
	lib, err := repo.Create(ctx, name, kind, validated, watchEnabled)
	if err != nil {
		return db.Library{}, err
	}
	if watchEnabled {
		s.startWatch(lib)
	}
	return lib, nil
}

// ListLibraries returns all libraries.
func (s *Service) ListLibraries(ctx context.Context) ([]db.Library, error) {
	return db.NewLibraryRepo(s.db.Querier()).List(ctx)
}

// GetLibrary returns one library.
func (s *Service) GetLibrary(ctx context.Context, id int64) (db.Library, error) {
	return db.NewLibraryRepo(s.db.Querier()).GetByID(ctx, id)
}

// UpdateLibrary updates a library.
func (s *Service) UpdateLibrary(ctx context.Context, id int64, name string, paths []string, watchEnabled bool) (db.Library, error) {
	validated, err := validatePaths(paths)
	if err != nil {
		return db.Library{}, err
	}
	repo := db.NewLibraryRepo(s.db.Querier())
	lib, err := repo.Update(ctx, id, name, validated, watchEnabled)
	if err != nil {
		return db.Library{}, err
	}
	s.restartWatch(lib)
	return lib, nil
}

// DeleteLibrary removes a library.
func (s *Service) DeleteLibrary(ctx context.Context, id int64) error {
	s.stopWatch(id)
	return db.NewLibraryRepo(s.db.Querier()).Delete(ctx, id)
}

// TriggerScan enqueues a scan job.
func (s *Service) TriggerScan(ctx context.Context, libraryID int64, scanType db.ScanType) (int64, jobs.TaskID, error) {
	if _, err := s.GetLibrary(ctx, libraryID); err != nil {
		return 0, "", err
	}
	runID, err := db.NewScanRepo(s.db.Querier()).CreateRun(ctx, libraryID, scanType, "")
	if err != nil {
		return 0, "", err
	}
	taskID, err := s.queue.Enqueue(ctx, &ScanJob{
		Scanner:   s.scanner,
		LibraryID: libraryID,
		ScanType:  scanType,
		RunID:     runID,
	}, jobs.WithName(fmt.Sprintf("library-scan-%d", libraryID)))
	if err != nil {
		return 0, "", err
	}
	return runID, taskID, nil
}

// ListScanHistory returns scan runs for a library.
func (s *Service) ListScanHistory(ctx context.Context, libraryID int64, limit int) ([]db.ScanRun, error) {
	return db.NewScanRepo(s.db.Querier()).ListByLibrary(ctx, libraryID, limit)
}

// GetScanRun returns a scan run by id.
func (s *Service) GetScanRun(ctx context.Context, id int64) (db.ScanRun, error) {
	return db.NewScanRepo(s.db.Querier()).GetByID(ctx, id)
}

func validatePaths(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("at least one path is required")
	}
	out := make([]string, 0, len(paths))
	seen := make(map[string]struct{})
	for _, p := range paths {
		v, err := ValidateRootPath(p)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out, nil
}

func (s *Service) startWatch(lib db.Library) {
	if !lib.WatchEnabled {
		return
	}
	s.watchMu.Lock()
	defer s.watchMu.Unlock()
	if _, ok := s.watching[lib.ID]; ok {
		return
	}
	w := NewWatcher(s.logger, lib, s.queue, s.scanner, 2*time.Second)
	s.watching[lib.ID] = w
	w.Start(context.Background())
}

func (s *Service) stopWatch(id int64) {
	s.watchMu.Lock()
	defer s.watchMu.Unlock()
	if w, ok := s.watching[id]; ok {
		w.Stop()
		delete(s.watching, id)
	}
}

func (s *Service) restartWatch(lib db.Library) {
	s.stopWatch(lib.ID)
	s.startWatch(lib)
}
