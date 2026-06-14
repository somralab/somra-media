package i18n_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"

	"github.com/somralab/somra-media/internal/platform/i18n"
)

func TestNegotiateUserPreferenceWins(t *testing.T) {
	t.Parallel()
	tag := i18n.Negotiate("tr-TR", "en-US", "tr-TR",
		language.AmericanEnglish, language.MustParse("tr-TR"))
	require.Equal(t, language.AmericanEnglish, tag)
}

func TestNegotiateSystemDefaultUsedWhenNoUserPref(t *testing.T) {
	t.Parallel()
	tag := i18n.Negotiate("en-US", "", "tr-TR",
		language.AmericanEnglish, language.MustParse("tr-TR"))
	require.Equal(t, language.MustParse("tr-TR"), tag)
}

func TestNegotiateAcceptLanguage(t *testing.T) {
	t.Parallel()
	tag := i18n.Negotiate("tr-TR,en;q=0.9", "", "",
		language.AmericanEnglish, language.MustParse("tr-TR"))
	require.Equal(t, language.MustParse("tr-TR"), tag)
}

func TestNegotiateFallback(t *testing.T) {
	t.Parallel()
	tag := i18n.Negotiate("", "", "",
		language.AmericanEnglish, language.MustParse("tr-TR"))
	require.Equal(t, i18n.SourceLanguage, tag)
}

func TestNegotiateUnknownTagsFallThrough(t *testing.T) {
	t.Parallel()
	tag := i18n.Negotiate("xx-YY,!!!", "garbage", "also-bad",
		language.AmericanEnglish, language.MustParse("tr-TR"))
	require.Equal(t, i18n.SourceLanguage, tag)
}

func TestNegotiateAcceptLanguageWithSubsetMatch(t *testing.T) {
	t.Parallel()
	tag := i18n.Negotiate("tr", "", "",
		language.AmericanEnglish, language.MustParse("tr-TR"))
	require.Equal(t, language.MustParse("tr-TR"), tag)
}

func TestNegotiateDefaultsToSourceWhenNoSupported(t *testing.T) {
	t.Parallel()
	tag := i18n.Negotiate("tr-TR", "", "")
	require.Equal(t, i18n.SourceLanguage, tag)
}
