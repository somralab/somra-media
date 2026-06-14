package api

import (
	"context"
	"net/http"

	"github.com/somralab/somra-media/internal/auth"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
)

type acceptLanguageCtxKey struct{}

// AuthMiddleware validates Bearer tokens and attaches AuthContext.
type AuthMiddleware struct {
	Service *auth.Service
}

// Middleware returns chi-compatible auth middleware.
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, ok := auth.BearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
			return
		}
		claims, err := m.Service.TokenService().Validate(r.Context(), raw)
		if err != nil {
			writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
			return
		}
		ac, err := m.Service.ResolveAuth(r.Context(), claims)
		if err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.errors.internal"))
			return
		}
		ctx := auth.WithAuthContext(r.Context(), ac)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission returns middleware that checks a permission.
func RequirePermission(perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ac, ok := auth.FromContext(r.Context())
			if !ok || !auth.HasPermission(ac, perm) {
				writeError(w, r, platformerrors.New(http.StatusForbidden, platformerrors.CodeForbidden, "auth.errors.forbidden"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole returns middleware that checks a role.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ac, ok := auth.FromContext(r.Context())
			if !ok || !auth.HasRole(ac, role) {
				writeError(w, r, platformerrors.New(http.StatusForbidden, platformerrors.CodeForbidden, "auth.errors.forbidden"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ProfileLocaleMiddleware overrides Accept-Language with profile locale when authenticated.
func ProfileLocaleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ac, ok := auth.FromContext(r.Context()); ok && ac.Profile.Locale != "" {
			r = r.WithContext(withAcceptLanguage(r.Context(), ac.Profile.Locale))
			r.Header.Set("Accept-Language", ac.Profile.Locale)
		}
		next.ServeHTTP(w, r)
	})
}

func withAcceptLanguage(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, acceptLanguageCtxKey{}, locale)
}

// AcceptLanguageFromContext returns negotiated locale override when set.
func AcceptLanguageFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	v, ok := ctx.Value(acceptLanguageCtxKey{}).(string)
	return v, ok
}
