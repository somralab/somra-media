package releaseparse

import (
	"regexp"
	"strings"

	"github.com/somralab/somra-media/internal/plugin"
)

var (
	reResolution = regexp.MustCompile(`(?i)(2160p|1080p|720p|480p|4k|uhd)`)
	reCodec      = regexp.MustCompile(`(?i)(x265|h265|hevc|x264|h264|av1|vp9)`)
	reSource     = regexp.MustCompile(`(?i)(web-dl|webrip|bluray|blu-ray|remux|hdtv|dvdrip|bdrip)`)
)

// Enrich fills SearchResult quality fields parsed from the release title.
func Enrich(r plugin.SearchResult) plugin.SearchResult {
	title := r.Title
	if title == "" {
		return r
	}
	if r.Resolution == "" {
		r.Resolution = matchGroup(reResolution, title)
	}
	if r.Codec == "" {
		r.Codec = normalizeCodec(matchGroup(reCodec, title))
	}
	if r.Source == "" {
		r.Source = strings.ToUpper(matchGroup(reSource, title))
	}
	return r
}

func matchGroup(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return strings.ToLower(m[1])
}

func normalizeCodec(raw string) string {
	switch strings.ToLower(raw) {
	case "x265", "h265", "hevc":
		return "hevc"
	case "x264", "h264":
		return "h264"
	case "av1":
		return "av1"
	case "vp9":
		return "vp9"
	default:
		return raw
	}
}

// ScoreHint returns a rough quality rank for scoring (higher is better).
func ScoreHint(r plugin.SearchResult) int {
	score := 0
	switch strings.ToLower(r.Resolution) {
	case "2160p", "4k", "uhd":
		score += 40
	case "1080p":
		score += 30
	case "720p":
		score += 20
	case "480p":
		score += 10
	}
	switch strings.ToLower(r.Codec) {
	case "hevc", "x265", "h265":
		score += 5
	case "h264", "x264":
		score += 3
	}
	switch strings.ToUpper(r.Source) {
	case "REMUX", "BLURAY", "BLU-RAY":
		score += 8
	case "WEB-DL", "WEBRIP":
		score += 5
	}
	if r.Seeders != nil {
		score += min(*r.Seeders, 50)
	}
	return score
}
