package library

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ParsedName holds filename/folder pre-parse results.
type ParsedName struct {
	Title   string
	Year    *int
	Season  *int
	Episode *int
}

var (
	movieYearRe   = regexp.MustCompile(`(?i)^(.+?)[\.\s_-]+(?:\(?(\d{4})\)?)[\.\s_-]*`)
	tvSeasonEpRe  = regexp.MustCompile(`(?i)[Ss](\d{1,2})[Ee](\d{1,3})`)
	tvSeasonEpAlt = regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,3})`)
	yearOnlyRe    = regexp.MustCompile(`\((\d{4})\)`)
)

// ParseFileName extracts title/year or season/episode from common naming patterns.
func ParseFileName(name string) ParsedName {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	base = strings.ReplaceAll(base, ".", " ")
	base = strings.TrimSpace(base)

	if m := tvSeasonEpRe.FindStringSubmatch(base); len(m) == 3 {
		s, _ := strconv.Atoi(m[1])
		e, _ := strconv.Atoi(m[2])
		title := strings.TrimSpace(tvSeasonEpRe.Split(base, -1)[0])
		return ParsedName{Title: cleanTitle(title), Season: &s, Episode: &e}
	}
	if m := tvSeasonEpAlt.FindStringSubmatch(base); len(m) == 3 {
		s, _ := strconv.Atoi(m[1])
		e, _ := strconv.Atoi(m[2])
		title := strings.TrimSpace(tvSeasonEpAlt.Split(base, -1)[0])
		return ParsedName{Title: cleanTitle(title), Season: &s, Episode: &e}
	}
	if m := movieYearRe.FindStringSubmatch(base); len(m) >= 3 {
		y, _ := strconv.Atoi(m[2])
		return ParsedName{Title: cleanTitle(m[1]), Year: &y}
	}
	if m := yearOnlyRe.FindStringSubmatch(base); len(m) == 2 {
		y, _ := strconv.Atoi(m[1])
		title := strings.TrimSpace(yearOnlyRe.ReplaceAllString(base, ""))
		return ParsedName{Title: cleanTitle(title), Year: &y}
	}
	return ParsedName{Title: cleanTitle(base)}
}

func cleanTitle(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "-_ ")
	replacer := strings.NewReplacer("_", " ", ".", " ")
	return strings.TrimSpace(replacer.Replace(s))
}
