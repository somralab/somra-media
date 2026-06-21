package library_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/somralab/somra-media/internal/library"
)

func TestRecommendScanProgressBatch(t *testing.T) {
	assert.Equal(t, 10, library.RecommendScanProgressBatch(2))
	assert.Equal(t, 25, library.RecommendScanProgressBatch(4))
	assert.Equal(t, 50, library.RecommendScanProgressBatch(8))
}
