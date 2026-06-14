package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecommendDefaultsEdgeCases(t *testing.T) {
	def := RecommendDefaults(1, "en-US")
	assert.Equal(t, 1, def.MaxConcurrentTranscodes)

	def = RecommendDefaults(8, "")
	assert.Equal(t, "en-US", def.DefaultLocale)
}

func TestDetectSystemEmptyPaths(t *testing.T) {
	profile := DetectSystem(nil)
	assert.Greater(t, profile.CPUCores, 0)
	assert.Empty(t, profile.Paths)
}

func TestValidatePathsSkipsBlank(t *testing.T) {
	assert.Empty(t, ValidatePaths([]string{"", "  "}))
}
