package seriesmonitor

import (
	"regexp"
	"strconv"
)

var reEpisode = regexp.MustCompile(`(?i)[.\s_-]S(\d{1,2})E(\d{1,2})`)

// ParseEpisode extracts season and episode numbers from a release title.
func ParseEpisode(title string) (season, episode int, ok bool) {
	m := reEpisode.FindStringSubmatch(title)
	if len(m) < 3 {
		return 0, 0, false
	}
	s, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, 0, false
	}
	e, err := strconv.Atoi(m[2])
	if err != nil {
		return 0, 0, false
	}
	return s, e, true
}

func isNewerEpisode(lastSeason, lastEpisode, season, episode int) bool {
	if season > lastSeason {
		return true
	}
	if season == lastSeason && episode > lastEpisode {
		return true
	}
	return false
}
