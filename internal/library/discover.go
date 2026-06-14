package library

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

var mediaExtensions = map[string]struct{}{
	".mkv": {}, ".mp4": {}, ".avi": {}, ".mov": {}, ".wmv": {},
	".m4v": {}, ".webm": {}, ".ts": {}, ".m2ts": {},
	".mp3": {}, ".flac": {}, ".aac": {}, ".ogg": {}, ".m4a": {}, ".wav": {},
}

// IsMediaFile reports whether path has a supported media extension.
func IsMediaFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := mediaExtensions[ext]
	return ok
}

// DiscoverMedia walks roots and invokes fn for each supported media file.
func DiscoverMedia(ctx context.Context, roots []string, fn func(path string, info fs.DirEntry) error) error {
	for _, root := range roots {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !IsMediaFile(path) {
				return nil
			}
			if err := ValidateMediaPath(path, roots); err != nil {
				return nil
			}
			return fn(path, d)
		})
		if err != nil {
			return fmt.Errorf("discover %q: %w", root, err)
		}
	}
	return nil
}
