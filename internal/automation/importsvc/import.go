package importsvc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/somralab/somra-media/internal/library"
	"github.com/somralab/somra-media/internal/platform/db"
)

// Service moves completed downloads and triggers library scan.
type Service struct {
	Library     *library.Service
	ImportRoot  string
	TargetLibID int64
}

// ImportCompleted moves savePath content into import root and enqueues scan.
func (s *Service) ImportCompleted(ctx context.Context, savePath string) error {
	if s == nil || s.Library == nil {
		return fmt.Errorf("automation import: library service required")
	}
	if savePath == "" {
		return fmt.Errorf("automation import: empty save path")
	}
	scanPath := savePath
	if s.ImportRoot != "" {
		dest := filepath.Join(s.ImportRoot, filepath.Base(savePath))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("automation import mkdir: %w", err)
		}
		if err := os.Rename(savePath, dest); err == nil {
			scanPath = dest
		}
	}
	_ = scanPath
	libID := s.TargetLibID
	if libID <= 0 {
		libs, err := s.Library.ListLibraries(ctx)
		if err != nil || len(libs) == 0 {
			return fmt.Errorf("automation import: no target library")
		}
		libID = libs[0].ID
	}
	_, _, err := s.Library.TriggerScan(ctx, libID, db.ScanIncremental)
	return err
}
