// Package i18n is the backend message catalog and locale-negotiation
// layer for Somra. It wraps nicksnyder/go-i18n/v2 with embedded
// resources and a Chi-friendly middleware so HTTP handlers can produce
// localized error envelopes without reaching for global state.
//
// Resource files live under locales/ and are embedded at build time.
// The fallback/source language is en-US per .plan/i18n-localization.md §1.
package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.toml
var localeFS embed.FS

// SourceLanguage is the canonical en-US tag used as the bundle default
// and the ultimate fallback during locale negotiation.
var SourceLanguage = language.AmericanEnglish

// Bundle wraps *goi18n.Bundle with the set of languages it has loaded.
// It is safe for concurrent use after construction.
type Bundle struct {
	inner *goi18n.Bundle
	tags  []language.Tag
}

// NewBundle constructs a Bundle from the embedded locales/*.toml files.
// It returns an error if any resource fails to parse.
func NewBundle() (*Bundle, error) {
	b := goi18n.NewBundle(SourceLanguage)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	tags := []language.Tag{}
	entries, err := fs.ReadDir(localeFS, "locales")
	if err != nil {
		return nil, fmt.Errorf("i18n: read embedded locales: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".toml") {
			continue
		}
		full := path.Join("locales", name)
		data, err := localeFS.ReadFile(full)
		if err != nil {
			return nil, fmt.Errorf("i18n: read %s: %w", full, err)
		}
		mf, err := b.ParseMessageFileBytes(data, name)
		if err != nil {
			return nil, fmt.Errorf("i18n: parse %s: %w", full, err)
		}
		tags = append(tags, mf.Tag)
	}
	return &Bundle{inner: b, tags: tags}, nil
}

// NewBundleFromFS builds a Bundle from an alternative fs (used in tests).
// The directory passed must contain *.toml files using the go-i18n
// active.<lang>.toml convention.
func NewBundleFromFS(fsys fs.FS, dir string) (*Bundle, error) {
	b := goi18n.NewBundle(SourceLanguage)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	tags := []language.Tag{}
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("i18n: read %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		full := path.Join(dir, e.Name())
		data, err := fs.ReadFile(fsys, full)
		if err != nil {
			return nil, fmt.Errorf("i18n: read %s: %w", full, err)
		}
		mf, err := b.ParseMessageFileBytes(data, e.Name())
		if err != nil {
			return nil, fmt.Errorf("i18n: parse %s: %w", full, err)
		}
		tags = append(tags, mf.Tag)
	}
	return &Bundle{inner: b, tags: tags}, nil
}

// Tags returns the language tags loaded into the bundle. The returned
// slice is shared and should be treated as read-only.
func (b *Bundle) Tags() []language.Tag { return b.tags }

// Matcher returns a language.Matcher backed by the loaded tags.
// The matcher is cached on first use.
func (b *Bundle) Matcher() language.Matcher {
	return language.NewMatcher(b.tags)
}

// Inner exposes the underlying go-i18n bundle for advanced callers
// (e.g. tests that need to register messages dynamically).
func (b *Bundle) Inner() *goi18n.Bundle { return b.inner }
