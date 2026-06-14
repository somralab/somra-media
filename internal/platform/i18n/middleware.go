package i18n

import (
	"context"
	"net/http"

	"golang.org/x/text/language"
)

// ctxKey is the unexported request-context key used to stash the
// localizer derived during request middleware.
type ctxKey struct{}

// ContextValue is the payload stored on the request context.
type ContextValue struct {
	Localizer *Localizer
	Tag       language.Tag
}

// Middleware returns a Chi-compatible middleware that:
//   - reads the Accept-Language header,
//   - negotiates a tag against the bundle's loaded locales,
//   - stores a Localizer for downstream handlers to use.
//
// The Sprint 03 user-preference middleware can later wrap this one and
// supply its userPref/systemDefault values via WithLocaleOverrides.
func (b *Bundle) Middleware() func(http.Handler) http.Handler {
	supported := b.Tags()
	if len(supported) == 0 {
		supported = []language.Tag{SourceLanguage}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accept := r.Header.Get("Accept-Language")
			userPref, systemDefault := overridesFromContext(r.Context())
			tag := Negotiate(accept, userPref, systemDefault, supported...)
			value := &ContextValue{Localizer: b.Localize(tag), Tag: tag}
			ctx := context.WithValue(r.Context(), ctxKey{}, value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext returns the Localizer that Middleware attached to ctx, or
// nil if the middleware did not run on this request.
func FromContext(ctx context.Context) *Localizer {
	v, ok := ctx.Value(ctxKey{}).(*ContextValue)
	if !ok || v == nil {
		return nil
	}
	return v.Localizer
}

// TagFromContext returns the negotiated language tag, falling back to
// SourceLanguage when the middleware did not run.
func TagFromContext(ctx context.Context) language.Tag {
	v, ok := ctx.Value(ctxKey{}).(*ContextValue)
	if !ok || v == nil {
		return SourceLanguage
	}
	return v.Tag
}

// overrideKey carries user/system preferences pushed by upstream
// middleware (e.g. Sprint 03 will populate this when the request is
// authenticated).
type overrideKey struct{}

type overrideValue struct {
	UserPref      string
	SystemDefault string
}

// WithLocaleOverrides returns a context carrying user/system locale
// preferences that the Middleware will honour during negotiation. This
// is a hook for future packets (Sprint 03 / Sprint 06) — the current
// behaviour falls back to Accept-Language alone.
func WithLocaleOverrides(parent context.Context, userPref, systemDefault string) context.Context {
	return context.WithValue(parent, overrideKey{}, overrideValue{
		UserPref:      userPref,
		SystemDefault: systemDefault,
	})
}

func overridesFromContext(ctx context.Context) (userPref string, systemDefault string) {
	v, ok := ctx.Value(overrideKey{}).(overrideValue)
	if !ok {
		return "", ""
	}
	return v.UserPref, v.SystemDefault
}
