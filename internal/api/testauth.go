package api

import (
	"net/http"
	"time"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

// testAuthMiddleware injects admin auth context for handler tests.
func testAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac := auth.AuthContext{
			Claims: auth.Claims{
				Subject: auth.Subject{
					UserID:   "test-user",
					Username: "testadmin",
					Roles:    []string{auth.RoleAdmin},
				},
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
				SessionID: "test-session",
			},
			Permissions: []string{
				auth.PermLibraryRead,
				auth.PermLibraryWrite,
				auth.PermUsersManage,
				auth.PermProfileEdit,
				auth.PermRequestsCreate,
				auth.PermRequestsRead,
				auth.PermRequestsManage,
				auth.PermNotificationsManage,
				auth.PermPluginsManage,
			},
			Profile: db.UserProfile{UserID: "test-user", Locale: "en-US", Theme: "cinematic"},
		}
		next.ServeHTTP(w, r.WithContext(auth.WithAuthContext(r.Context(), ac)))
	})
}
