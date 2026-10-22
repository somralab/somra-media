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

func TestCheckOverflow_FlagsTranslationsOverRatio(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en-US", "nav.json"), `{"home":"Home","menu":{"settings":"Settings"},"count":3}`)
	writeFile(t, filepath.Join(dir, "tr-TR", "nav.json"), `{"home":"Ana sayfanıza dönün","menu":{"settings":"Ayarlar"},"count":3}`)

	err := checkOverflow(dir, "en-US", "tr-TR", 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nav.home: source=4 target=19")
	assert.NotContains(t, err.Error(), "nav.menu.settings")
	assert.NotContains(t, err.Error(), "nav.count")
}

func TestCheckOverflow_CountsRunesNotBytes(t *testing.T) {
	dir := t.TempDir()
	// "Şişeği" is 6 runes but 9 bytes; a byte count would exceed 1.5 × 5.
	writeFile(t, filepath.Join(dir, "en-US", "a.json"), `{"k":"Glass"}`)
	writeFile(t, filepath.Join(dir, "tr-TR", "a.json"), `{"k":"Şişeği"}`)

	require.NoError(t, checkOverflow(dir, "en-US", "tr-TR", 1.5))
}

func TestCheckOverflow_SkipsEmptySourceAndUntranslatedKeys(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en-US", "a.json"), `{"empty":"","only_source":"Hi"}`)
	writeFile(t, filepath.Join(dir, "de-DE", "a.json"), `{"empty":"Diese Übersetzung ist in der Quelle leer"}`)

	require.NoError(t, checkOverflow(dir, "en-US", "de-DE", 1.5))
}

func TestCheckOverflow_MissingLocaleDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en-US", "a.json"), `{"k":"v"}`)

	err := checkOverflow(dir, "en-US", "tr-TR", 1.5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overflow frontend target")

	err = checkOverflow(t.TempDir(), "en-US", "tr-TR", 1.5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overflow frontend source")
}

func TestCheckOverflow_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "en-US", "a.json"), `{ not valid`)
	writeFile(t, filepath.Join(dir, "tr-TR", "a.json"), `{}`)

	require.Error(t, checkOverflow(dir, "en-US", "tr-TR", 1.5))
}

func TestCheckOverflowBackend_FlagsTranslationsOverRatio(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "active.en-US.toml"), `["errors.short"]
other = "Not found"

["errors.ok"]
other = "Forbidden"
`)
	writeFile(t, filepath.Join(dir, "active.tr-TR.toml"), `["errors.short"]
other = "İstediğiniz kaynak bu sunucuda bulunamadı"

["errors.ok"]
other = "Yasak"
`)

	err := checkOverflowBackend(dir, "en-US", "tr-TR", 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "errors.short: source=9 target=41")
	assert.NotContains(t, err.Error(), "errors.ok")
}

func TestCheckOverflowBackend_IgnoresEntriesWithoutOther(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "active.en-US.toml"), `["items"]
one = "item"
`)
	writeFile(t, filepath.Join(dir, "active.de-DE.toml"), `["items"]
one = "ein sehr, sehr langer Elementname"
`)

	require.NoError(t, checkOverflowBackend(dir, "en-US", "de-DE", 1.5))
}

func TestCheckOverflowBackend_MissingFiles(t *testing.T) {
	require.NoError(t, checkOverflowBackend(t.TempDir(), "en-US", "tr-TR", 1.5),
		"no source bundle means there is nothing to compare")

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "active.en-US.toml"), `["k"]
other = "v"
`)
	err := checkOverflowBackend(dir, "en-US", "tr-TR", 1.5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overflow backend target")

	writeFile(t, filepath.Join(dir, "active.en-US.toml"), "not valid = = toml")
	err = checkOverflowBackend(dir, "en-US", "tr-TR", 1.5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overflow backend source")
}

// newLocaleTree writes a minimal, valid backend + frontend locale layout and
// returns the two directories.
func newLocaleTree(t *testing.T, trHome string) (backendDir, frontendDir string) {
	t.Helper()
	root := t.TempDir()
	backendDir = filepath.Join(root, "backend")
	frontendDir = filepath.Join(root, "frontend")
	writeFile(t, filepath.Join(backendDir, "active.en-US.toml"), `["k"]
other = "Home"
`)
	writeFile(t, filepath.Join(backendDir, "active.tr-TR.toml"), `["k"]
other = "Ev"
`)
	writeFile(t, filepath.Join(frontendDir, "en-US", "nav.json"), `{"home":"Home"}`)
	writeFile(t, filepath.Join(frontendDir, "tr-TR", "nav.json"), `{"home":"`+trHome+`"}`)
	return backendDir, frontendDir
}

func TestRun_ExitCodes(t *testing.T) {
	backendDir, frontendDir := newLocaleTree(t, "Ana sayfa")
	base := []string{"--backend-dir", backendDir, "--frontend-dir", frontendDir}

	var stdout, stderr strings.Builder
	require.Equal(t, 0, run(append(base, "--overflow-ratio", "3.5"), &stdout, &stderr), stderr.String())
	assert.Equal(t, "i18n-check: ok\n", stdout.String())

	stdout.Reset()
	stderr.Reset()
	require.Equal(t, 1, run(append(base, "--overflow-ratio", "1.5"), &stdout, &stderr))
	assert.Contains(t, stderr.String(), "frontend overflow check")
	assert.Empty(t, stdout.String())

	require.Equal(t, 0, run(append(base, "--overflow-ratio", "0"), &stdout, &stderr),
		"a ratio of 0 disables the overflow check")

	stderr.Reset()
	require.Equal(t, 2, run([]string{"--no-such-flag"}, &stdout, &stderr))
	assert.Contains(t, stderr.String(), "no-such-flag")
}

func TestRun_ReportsKeyMismatch(t *testing.T) {
	backendDir, frontendDir := newLocaleTree(t, "Ana sayfa")
	writeFile(t, filepath.Join(frontendDir, "tr-TR", "extra.json"), `{"x":"y"}`)

	var stdout, stderr strings.Builder
	code := run([]string{"--backend-dir", backendDir, "--frontend-dir", frontendDir, "--overflow-ratio", "0"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "extra")
}
