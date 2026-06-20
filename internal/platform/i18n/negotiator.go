package i18n

import (
	"strings"

	"golang.org/x/text/language"
)

// Negotiate resolves the best language tag using the priority order
// defined in .plan/i18n-localization.md §3:
//
//  1. userPref       — the authenticated user's persisted preference
//  2. systemDefault  — the operator-defined default
//  3. acceptLang     — the request's Accept-Language header
//  4. en-US fallback — applied implicitly via SourceLanguage
//
// Each parameter is best-effort: empty values are skipped and malformed
// tags fall through to the next source. The function never panics and
// always returns a usable tag.
func Negotiate(acceptLang, userPref, systemDefault string, supported ...language.Tag) language.Tag {
	if len(supported) == 0 {
		supported = []language.Tag{SourceLanguage}
	}
	matcher := language.NewMatcher(supported)

	if tag, ok := matchOne(matcher, supported, userPref); ok {
		return tag
	}
	if tag, ok := matchOne(matcher, supported, systemDefault); ok {
		return tag
	}
	if acceptLang != "" {
		tags, _, err := language.ParseAcceptLanguage(acceptLang)
		if err == nil && len(tags) > 0 {
			if idx, ok := matchIndex(matcher, tags...); ok {
				return supported[idx]
			}
		}
	}
	return SourceLanguage
}

// matchOne canonicalizes a single tag string (e.g. "tr-TR") against the
// supported set, returning the registered supported tag (not the
// matcher's potentially derived result).
func matchOne(matcher language.Matcher, supported []language.Tag, raw string) (language.Tag, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return language.Tag{}, false
	}
	tag, err := language.Parse(raw)
	if err != nil {
		return language.Tag{}, false
	}
	idx, ok := matchIndex(matcher, tag)
	if !ok {
		return language.Tag{}, false
	}
	return supported[idx], true
}

// matchIndex runs the matcher and reports the position of the chosen
// supported tag. We rely on language.Matcher returning the supported
// tag's index in its third return value.
func matchIndex(matcher language.Matcher, tags ...language.Tag) (int, bool) {
	_, idx, conf := matcher.Match(tags...)
	if conf <= language.No {
		return 0, false
	}
	return idx, true
}
