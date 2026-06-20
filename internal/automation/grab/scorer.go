package grab

import (
	"encoding/json"
	"strings"

	"github.com/somralab/somra-media/internal/automation/releaseparse"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/requests"
)

// ProfileSpec drives release scoring.
type ProfileSpec struct {
	PreferredResolutions []string `json:"preferredResolutions"`
	PreferredCodecs      []string `json:"preferredCodecs"`
	MaxSizeBytes         int64    `json:"maxSizeBytes"`
	PreferredTerms       []string `json:"preferredTerms,omitempty"`
	IgnoredTerms         []string `json:"ignoredTerms,omitempty"`
}

// ParseProfileSpec unmarshals JSON spec from quality_profiles.
func ParseProfileSpec(raw string) (ProfileSpec, error) {
	var spec ProfileSpec
	if raw == "" {
		raw = "{}"
	}
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		return ProfileSpec{}, err
	}
	if len(spec.PreferredResolutions) == 0 {
		spec.PreferredResolutions = []string{"1080p", "720p", "any"}
	}
	if len(spec.PreferredCodecs) == 0 {
		spec.PreferredCodecs = []string{"hevc", "h264"}
	}
	return spec, nil
}

// Score ranks a release; higher is better. -1 means reject.
func Score(r plugin.SearchResult, spec ProfileSpec, req requests.Request) int {
	title := strings.ToLower(r.Title)
	for _, term := range spec.IgnoredTerms {
		if term != "" && strings.Contains(title, strings.ToLower(term)) {
			return -1
		}
	}
	if spec.MaxSizeBytes > 0 && r.SizeBytes > spec.MaxSizeBytes {
		return -1
	}
	score := releaseparse.ScoreHint(r)
	score += resolutionBonus(r.Resolution, spec, req)
	score += codecBonus(r.Codec, spec)
	for _, term := range spec.PreferredTerms {
		if term != "" && strings.Contains(title, strings.ToLower(term)) {
			score += 5
		}
	}
	return score
}

// PickBest returns the highest-scoring release or nil.
func PickBest(results []plugin.SearchResult, spec ProfileSpec, req requests.Request) *plugin.SearchResult {
	var best *plugin.SearchResult
	bestScore := -1
	for i := range results {
		s := Score(results[i], spec, req)
		if s < 0 {
			continue
		}
		if s > bestScore {
			bestScore = s
			best = &results[i]
		}
	}
	return best
}

func resolutionBonus(res string, spec ProfileSpec, req requests.Request) int {
	res = strings.ToLower(res)
	want := string(req.QualityResolution)
	if want != "" && want != "any" && res != "" && !strings.Contains(res, strings.ToLower(want)) {
		return -10
	}
	for i, pref := range spec.PreferredResolutions {
		if strings.EqualFold(pref, res) || (pref == "any" && res != "") {
			return (len(spec.PreferredResolutions) - i) * 3
		}
	}
	return 0
}

func codecBonus(codec string, spec ProfileSpec) int {
	codec = strings.ToLower(codec)
	for i, pref := range spec.PreferredCodecs {
		if strings.EqualFold(pref, codec) {
			return (len(spec.PreferredCodecs) - i) * 2
		}
	}
	return 0
}
