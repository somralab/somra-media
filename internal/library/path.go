package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateRootPath ensures path is safe for use as a library root.
// It rejects traversal segments and requires an existing directory.
func ValidateRootPath(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("library path: empty")
	}
	clean := filepath.Clean(raw)
	if strings.Contains(clean, "..") {
		return "", fmt.Errorf("library path %q: traversal not allowed", raw)
	}
	abs, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("library path %q: %w", raw, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("library path %q: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("library path %q: not a directory", abs)
	}
	return abs, nil
}

// ValidateMediaPath ensures a discovered file stays under one of the library roots.
func ValidateMediaPath(filePath string, roots []string) error {
	absFile, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("media path %q: %w", filePath, err)
	}
	for _, root := range roots {
		absRoot, err := filepath.Abs(filepath.Clean(root))
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(absRoot, absFile)
		if err != nil {
			continue
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			continue
		}
		return nil
	}
	return fmt.Errorf("media path %q: outside library roots", absFile)
}
