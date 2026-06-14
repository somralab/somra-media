package i18n_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"testing/fstest"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"

	"github.com/somralab/somra-media/internal/platform/i18n"
)

func mustEmbed(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(rel)
	require.NoError(t, err)
	return data
}

func TestNewBundleLoadsEmbeddedLocales(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)

	tags := b.Tags()
	require.NotEmpty(t, tags)
	require.Contains(t, tags, language.AmericanEnglish)
	require.Contains(t, tags, language.MustParse("tr-TR"))
}

func TestLocalizerLooksUpEnglish(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)

	loc := b.Localize(language.AmericanEnglish)
	require.Equal(t, "An internal error occurred.", loc.Message("errors.internal", nil))
	require.Equal(t, "Field email is required.",
		loc.Message("errors.validation.required", map[string]any{"Field": "email"}))
}

func TestLocalizerLooksUpTurkish(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)

	loc := b.Localize(language.MustParse("tr-TR"))
	require.Equal(t, "Beklenmeyen bir sunucu hatası oluştu.", loc.Message("errors.internal", nil))
	require.Equal(t, "email alanı zorunludur.",
		loc.Message("errors.validation.required", map[string]any{"Field": "email"}))
}

func TestLocalizerFallsBackToSource(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)

	loc := b.Localize(language.MustParse("fr-FR"))
	require.Equal(t, "An internal error occurred.", loc.Message("errors.internal", nil))
}

func TestLocalizerMissingKeyReturnsKey(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)
	loc := b.Localize(language.AmericanEnglish)
	require.Equal(t, "errors.no_such_key", loc.Message("errors.no_such_key", nil))
}

// TestLocaleParity ensures en-US and tr-TR define the exact same set of
// message IDs; this guards plan/i18n-localization.md §5 anti-drift rule.
func TestLocaleParity(t *testing.T) {
	t.Parallel()
	keysFor := func(t *testing.T, file string) []string {
		t.Helper()
		var data map[string]map[string]any
		// Read embedded file via the bundle's inner FS by re-parsing here.
		raw := mustEmbed(t, file)
		require.NoError(t, toml.Unmarshal(raw, &data))
		out := make([]string, 0, len(data))
		for k := range data {
			out = append(out, k)
		}
		sort.Strings(out)
		return out
	}

	en := keysFor(t, "locales/active.en-US.toml")
	tr := keysFor(t, "locales/active.tr-TR.toml")
	require.Equal(t, en, tr, "en-US and tr-TR must define the same keys")
}

func TestBundleAccessors(t *testing.T) {
	t.Parallel()
	b, err := i18n.NewBundle()
	require.NoError(t, err)
	require.NotNil(t, b.Inner())
	require.NotNil(t, b.Matcher())

	loc := b.Localize(language.AmericanEnglish)
	require.Equal(t, language.AmericanEnglish, loc.Tag())
	require.NotNil(t, loc.Inner())
}

func TestNewBundleFromFS(t *testing.T) {
	t.Parallel()
	en, err := os.ReadFile(filepath.Join("locales", "active.en-US.toml"))
	require.NoError(t, err)
	fsys := fstest.MapFS{
		"l/active.en-US.toml": &fstest.MapFile{Data: en},
		"l/ignored.txt":       &fstest.MapFile{Data: []byte("nope")},
	}
	b, err := i18n.NewBundleFromFS(fsys, "l")
	require.NoError(t, err)
	require.NotEmpty(t, b.Tags())
}

func TestNewBundleFromFSErrors(t *testing.T) {
	t.Parallel()
	_, err := i18n.NewBundleFromFS(fstest.MapFS{}, "missing")
	require.Error(t, err)

	fsys := fstest.MapFS{
		"l/active.en-US.toml": &fstest.MapFile{Data: []byte("not valid toml ===")},
	}
	_, err = i18n.NewBundleFromFS(fsys, "l")
	require.Error(t, err)
}
