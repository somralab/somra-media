package i18n

import (
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Localizer is a thin wrapper around *goi18n.Localizer that exposes a
// small, opinionated API focused on the patterns Somra actually uses:
// translating an error-message key with optional template data.
type Localizer struct {
	inner *goi18n.Localizer
	tag   language.Tag
}

// Localize returns a Localizer for the given language tag, using the
// bundle's source language as the implicit fallback.
func (b *Bundle) Localize(tag language.Tag) *Localizer {
	loc := goi18n.NewLocalizer(b.inner, tag.String(), SourceLanguage.String())
	return &Localizer{inner: loc, tag: tag}
}

// Tag returns the language tag this Localizer was built for.
func (l *Localizer) Tag() language.Tag { return l.tag }

// Message resolves key against the bundle. If the key is missing in all
// loaded locales, the key itself is returned so callers always have a
// non-empty string to render. templateData may be nil.
func (l *Localizer) Message(key string, templateData map[string]any) string {
	cfg := &goi18n.LocalizeConfig{MessageID: key}
	if len(templateData) > 0 {
		cfg.TemplateData = templateData
	}
	msg, err := l.inner.Localize(cfg)
	if err != nil || msg == "" {
		return key
	}
	return msg
}

// Inner exposes the underlying go-i18n localizer for callers that need
// advanced features (plurals, complex template data).
func (l *Localizer) Inner() *goi18n.Localizer { return l.inner }
