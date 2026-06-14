package subtitles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Storage writes subtitle bytes to a dedicated cache directory.
type Storage struct {
	Root string
}

// Save writes content to root/subtitles/{mediaItemID}/{language}.srt.
func (s *Storage) Save(mediaItemID int64, language string, content []byte) (string, error) {
	if s.Root == "" {
		return "", fmt.Errorf("subtitles storage: root not configured")
	}
	lang := sanitizeLang(language)
	dir := filepath.Join(s.Root, "subtitles", fmt.Sprintf("%d", mediaItemID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("subtitles storage mkdir: %w", err)
	}
	path := filepath.Join(dir, lang+".srt")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", fmt.Errorf("subtitles storage write: %w", err)
	}
	return path, nil
}

func sanitizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	lang = strings.ReplaceAll(lang, "/", "")
	lang = strings.ReplaceAll(lang, "..", "")
	if lang == "" {
		return "und"
	}
	return lang
}
