package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
}

func TestCheckBackend_OK(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "active.en-US.toml"), `
["errors.internal"]
other = "x"

["errors.bad_request"]
other = "y"
`)
	writeFile(t, filepath.Join(dir, "active.tr-TR.toml"), `
["errors.internal"]
other = "x"

["errors.bad_request"]
other = "y"
`)
	require.NoError(t, checkBackend(dir, "en-US"))
}

func TestCheckBackend_MissingKeys(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "active.en-US.toml"), `["errors.a"]
other = "x"

["errors.b"]
other = "y"
`)
	writeFile(t, filepath.Join(dir, "active.tr-TR.toml"), `["errors.a"]
other = "x"
`)
	err := checkBackend(dir, "en-US")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
	assert.Contains(t, err.Error(), "errors.b")
}

func TestCheckBackend_ExtraKeys(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "active.en-US.toml"), `["errors.a"]
other = "x"
`)
	writeFile(t, filepath.Join(dir, "active.tr-TR.toml"), `["errors.a"]
other = "x"

["errors.unexpected"]
other = "z"
`)
	err := checkBackend(dir, "en-US")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "extra")
}

func TestCheckBackend_MissingDirIsSkipped(t *testing.T) {
	require.NoError(t, checkBackend(filepath.Join(t.TempDir(), "does-not-exist"), "en-US"))
}

func TestCheckFrontend_OK(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en-US", "common.json"), `{"a":{"b":"1"},"c":"2"}`)
	writeFile(t, filepath.Join(dir, "tr-TR", "common.json"), `{"a":{"b":"x"},"c":"y"}`)
	require.NoError(t, checkFrontend(dir, "en-US"))
}

func TestCheckFrontend_DetectsMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en-US", "common.json"), `{"a":"1","b":"2"}`)
	writeFile(t, filepath.Join(dir, "tr-TR", "common.json"), `{"a":"x"}`)
	err := checkFrontend(dir, "en-US")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "common.b")
}

func TestCheckFrontend_DetectsExtra(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en-US", "common.json"), `{"a":"1"}`)
	writeFile(t, filepath.Join(dir, "tr-TR", "common.json"), `{"a":"x","unused":"z"}`)
	err := checkFrontend(dir, "en-US")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "common.unused")
}

func TestCompareKeys_Equal(t *testing.T) {
	a := map[string]struct{}{"x": {}, "y": {}}
	b := map[string]struct{}{"y": {}, "x": {}}
	assert.Equal(t, "", compareKeys(a, b))
}

func TestFlatten_NestedObjects(t *testing.T) {
	out := map[string]struct{}{}
	flatten(map[string]any{"a": map[string]any{"b": "v", "c": map[string]any{"d": "v2"}}}, "ns", out)
	keys := []string{}
	for k := range out {
		keys = append(keys, k)
	}
	assert.Contains(t, keys, "ns.a.b")
	assert.Contains(t, keys, "ns.a.c.d")
}

func TestSortedKeysHelper(t *testing.T) {
	got := sortedKeys(map[string]int{"b": 1, "a": 2, "c": 3})
	assert.Equal(t, []string{"a", "b", "c"}, got)
}

func TestCompareKeys_FormatsDiff(t *testing.T) {
	a := map[string]struct{}{"x": {}, "y": {}}
	b := map[string]struct{}{"x": {}, "z": {}}
	diff := compareKeys(a, b)
	require.NotEmpty(t, diff)
	assert.True(t, strings.Contains(diff, "y") && strings.Contains(diff, "z"))
}

func TestCheckFrontend_MissingDirIsSkipped(t *testing.T) {
	require.NoError(t, checkFrontend(filepath.Join(t.TempDir(), "missing"), "en-US"))
}

func TestCheckBackend_NoLocaleFiles(t *testing.T) {
	require.NoError(t, checkBackend(t.TempDir(), "en-US"))
}

func TestCheckFrontend_NoLocaleDirs(t *testing.T) {
	require.NoError(t, checkFrontend(t.TempDir(), "en-US"))
}

func TestCheckBackend_SourceLocaleMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "active.tr-TR.toml"), `["k.a"]
other = "x"
`)
	err := checkBackend(dir, "en-US")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source locale")
}

func TestCheckFrontend_SourceLocaleMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "tr-TR", "common.json"), `{"a":"x"}`)
	err := checkFrontend(dir, "en-US")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source locale")
}

func TestCheckBackend_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "active.en-US.toml"), "not valid = = toml")
	err := checkBackend(dir, "en-US")
	require.Error(t, err)
}

func TestCheckFrontend_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en-US", "common.json"), `{ not valid`)
	err := checkFrontend(dir, "en-US")
	require.Error(t, err)
}

func TestFlatten_EmptyObject(t *testing.T) {
	out := map[string]struct{}{}
	flatten(map[string]any{}, "x", out)
	_, ok := out["x"]
	assert.True(t, ok)
}

func TestFlatten_RootLevelLeaf(t *testing.T) {
	out := map[string]struct{}{}
	flatten("scalar", "", out)
	_, ok := out[""]
	assert.True(t, ok)
}
