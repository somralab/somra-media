package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestAuthHandlers_SetupAndLogin(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	h := testRouterWithAuth(New(Options{
		AuthHandlers: &AuthHandlers{Service: svc},
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var tok map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tok))
	assert.NotEmpty(t, tok["accessToken"])
}

func TestRBACMatrix(t *testing.T) {
	cases := []struct {
		name       string
		roles      []string
		perm       string
		wantAccess bool
	}{
		{"admin library read", []string{auth.RoleAdmin}, auth.PermLibraryRead, true},
		{"admin users manage", []string{auth.RoleAdmin}, auth.PermUsersManage, true},
		{"user library write", []string{auth.RoleUser}, auth.PermLibraryWrite, true},
		{"user users manage", []string{auth.RoleUser}, auth.PermUsersManage, false},
		{"child library read", []string{auth.RoleChild}, auth.PermLibraryRead, true},
		{"child library write", []string{auth.RoleChild}, auth.PermLibraryWrite, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := openTestDB(t)
			perms := permissionsForRoles(t, d, tc.roles)
			ac := auth.AuthContext{Permissions: perms}
			assert.Equal(t, tc.wantAccess, auth.HasPermission(ac, tc.perm))
		})
	}
}

func permissionsForRoles(t *testing.T, d *db.DB, roles []string) []string {
	t.Helper()
	ctx := context.Background()
	repo := db.NewUserRepo(d.Querier())
	id := "test-" + roles[0]
	_, err := repo.Create(ctx, id, "user-"+roles[0], "hash", roles)
	require.NoError(t, err)
	perms, err := repo.PermissionsForUser(ctx, id)
	require.NoError(t, err)
	return perms
}

func newTestAuthService(t *testing.T, d *db.DB) *auth.Service {
	t.Helper()
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	jwtCfg := auth.DefaultJWTConfig(secret)
	q := d.Querier()
	return auth.NewService(auth.ServiceConfig{
		Users:    db.NewUserRepo(q),
		Sessions: db.NewSessionRepo(q),
		Profiles: db.NewProfileRepo(q),
		Tokens:   auth.NewJWTService(jwtCfg),
		Refresh:  auth.NewSQLiteRefreshStore(db.NewSessionRepo(q), secret, jwtCfg.RefreshTTL),
		Hasher:   auth.NewPasswordHasher(auth.DefaultPasswordPolicy()),
		Lockout:  auth.NewLoginLockout(db.NewLoginAttemptRepo(q), auth.DefaultLockoutConfig()),
		JWT:      jwtCfg,
	})
}
