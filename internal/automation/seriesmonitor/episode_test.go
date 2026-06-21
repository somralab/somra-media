package seriesmonitor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEpisode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title   string
		season  int
		episode int
		ok      bool
	}{
		{"Show.Name.S01E05.1080p", 1, 5, true},
		{"Show S02E10 WEB-DL", 2, 10, true},
		{"Show.S12E03", 12, 3, true},
		{"Movie.2024.1080p", 0, 0, false},
		{"", 0, 0, false},
	}
	for _, tc := range cases {
		season, episode, ok := ParseEpisode(tc.title)
		require.Equal(t, tc.ok, ok, tc.title)
		if ok {
			require.Equal(t, tc.season, season)
			require.Equal(t, tc.episode, episode)
		}
	}
}

func TestIsNewerEpisode(t *testing.T) {
	t.Parallel()
	require.True(t, isNewerEpisode(1, 5, 1, 6))
	require.True(t, isNewerEpisode(1, 5, 2, 1))
	require.False(t, isNewerEpisode(2, 1, 1, 9))
	require.False(t, isNewerEpisode(1, 5, 1, 5))
}
