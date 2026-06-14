package library

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRootPath_RejectsTraversal(t *testing.T) {
	_, err := ValidateRootPath("../etc")
	assert.Error(t, err)
}

func TestValidateRootPath_AcceptsDirectory(t *testing.T) {
	dir := t.TempDir()
	got, err := ValidateRootPath(dir)
	require.NoError(t, err)
	assert.Equal(t, dir, got)
}

func TestValidateMediaPath_InsideRoot(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "movie.mkv")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))
	assert.NoError(t, ValidateMediaPath(file, []string{root}))
}

func TestValidateMediaPath_OutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	file := filepath.Join(outside, "movie.mkv")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))
	assert.Error(t, ValidateMediaPath(file, []string{root}))
}
