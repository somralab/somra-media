package streaming

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateCacheSegmentPath ensures a requested segment stays inside session cache.
func ValidateCacheSegmentPath(cacheRoot, sessionID, rel string) (string, error) {
	cacheRoot = filepath.Clean(cacheRoot)
	sessionDir := filepath.Join(cacheRoot, sessionID)
	clean := filepath.Clean(filepath.Join(sessionDir, rel))
	relPath, err := filepath.Rel(sessionDir, clean)
	if err != nil {
		return "", fmt.Errorf("streaming path %q: %w", rel, err)
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("streaming path %q: traversal", rel)
	}
	base := filepath.Base(clean)
	if base != InitSegmentName && !strings.HasPrefix(base, "seg_") &&
		base != "master.m3u8" && !strings.HasSuffix(base, ".m3u8") &&
		base != "source" {
		return "", fmt.Errorf("streaming path %q: not allowed", rel)
	}
	return clean, nil
}

// SessionCacheDir returns the on-disk cache directory for a session.
func SessionCacheDir(cacheRoot, sessionID string) string {
	return filepath.Join(filepath.Clean(cacheRoot), sessionID)
}
