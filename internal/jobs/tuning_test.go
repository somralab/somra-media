package jobs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/somralab/somra-media/internal/jobs"
)

func TestRecommendQueueTuning(t *testing.T) {
	two := jobs.RecommendQueueTuning(2)
	assert.Equal(t, 1, two.Workers)
	assert.Equal(t, 8, two.Buffer)

	four := jobs.RecommendQueueTuning(4)
	assert.Equal(t, 2, four.Workers)
	assert.Equal(t, 16, four.Buffer)

	eight := jobs.RecommendQueueTuning(8)
	assert.Equal(t, 3, eight.Workers)
	assert.Equal(t, 32, eight.Buffer)
}
