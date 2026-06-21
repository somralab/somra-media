//go:build integration

package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/db"
)

type rbacCase struct {
	name       string
	method     string
	path       string
	perm       string // empty = authenticated-only
	allowGuest bool
	allowUser  bool
	allowChild bool
}

func TestIntegration_UpgradeFromPenultimateMigration(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()
	cfg := db.Default()
	cfg.DataDir = dataDir

	d, err := db.Open(ctx, cfg)
	require.NoError(t, err)

	versions, err := db.ListEmbeddedVersions()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(versions), 2)
	penultimate := versions[len(versions)-2]

	require.NoError(t, db.MigrateUpTo(ctx, d, penultimate, nil))
	const markerKey = "bootstrap_upgrade_marker"
	_, err = d.SQL().ExecContext(ctx,
		`INSERT INTO settings (key, value, updated_at) VALUES (?, 'persist', datetime('now'))`,
		markerKey,
	)
	require.NoError(t, err)
	dbPath := filepath.Join(dataDir, cfg.DBFile)
	require.NoError(t, d.Close())

	restoredDir := t.TempDir()
	raw, err := os.ReadFile(dbPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(restoredDir, cfg.DBFile), raw, 0o600))

	restoredCfg := db.Default()
	restoredCfg.DataDir = restoredDir
	restoredCfg.DBFile = cfg.DBFile
	upgraded, err := db.Initialize(ctx, restoredCfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = upgraded.Close() })

	var value string
	err = upgraded.SQL().QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, markerKey).Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, "persist", value)

	cur, tgt, err := db.MigrateStatus(ctx, upgraded)
	require.NoError(t, err)
	assert.Equal(t, tgt, cur)
}

func TestIntegration_RBACMatrix(t *testing.T) {
	ts := bootstrap.NewTestServer(t)
	userToken := createRoleToken(t, ts, "plainuser", "UserPass1!", []string{auth.RoleUser})
	childToken := createRoleToken(t, ts, "childuser", "ChildPass1!", []string{auth.RoleChild})

	cases := []rbacCase{
		{name: "health public", method: http.MethodGet, path: "/api/v1/health", allowGuest: true, allowUser: true, allowChild: true},
		{name: "version public", method: http.MethodGet, path: "/api/v1/version", allowGuest: true, allowUser: true, allowChild: true},
		{name: "setup status public", method: http.MethodGet, path: "/api/v1/setup/status", allowGuest: true, allowUser: true, allowChild: true},
		{name: "system detect public", method: http.MethodGet, path: "/api/v1/system/detect", allowGuest: true, allowUser: true, allowChild: true},
		{name: "onboarding status public", method: http.MethodGet, path: "/api/v1/onboarding/status", allowGuest: true, allowUser: true, allowChild: true},

		{name: "profile edit", method: http.MethodGet, path: "/api/v1/profile", perm: auth.PermProfileEdit, allowUser: true},
		{name: "watch state auth only", method: http.MethodGet, path: "/api/v1/watch-state", allowUser: true, allowChild: true},

		{name: "libraries read", method: http.MethodGet, path: "/api/v1/libraries", perm: auth.PermLibraryRead, allowUser: true, allowChild: true},
		{name: "libraries write", method: http.MethodPost, path: "/api/v1/libraries", perm: auth.PermLibraryWrite, allowUser: true},
		{name: "settings admin", method: http.MethodGet, path: "/api/v1/settings", perm: auth.PermUsersManage},

		{name: "plugins manage", method: http.MethodGet, path: "/api/v1/plugins/catalog", perm: auth.PermPluginsManage},
		{name: "automation downloads", method: http.MethodGet, path: "/api/v1/automation/downloads", perm: auth.PermPluginsManage},

		{name: "requests read", method: http.MethodGet, path: "/api/v1/requests", perm: auth.PermRequestsRead, allowUser: true},
		{name: "requests policies manage", method: http.MethodGet, path: "/api/v1/requests/policies", perm: auth.PermRequestsRead, allowUser: true},
		{name: "requests policies patch", method: http.MethodPatch, path: "/api/v1/requests/policies", perm: auth.PermRequestsManage},

		{name: "notifications channels manage", method: http.MethodGet, path: "/api/v1/notifications/channels", perm: auth.PermNotificationsManage},
		{name: "notifications preferences auth", method: http.MethodGet, path: "/api/v1/notifications/preferences", allowUser: true, allowChild: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name+"/guest", func(t *testing.T) {
			code := doRBACRequest(t, ts, tc.method, tc.path, "")
			if tc.allowGuest {
				assertRBACAllowed(t, code)
				return
			}
			assert.Equal(t, http.StatusUnauthorized, code)
		})
		t.Run(tc.name+"/user", func(t *testing.T) {
			code := doRBACRequest(t, ts, tc.method, tc.path, userToken)
			if tc.allowUser {
				assertRBACAllowed(t, code)
				return
			}
			if tc.perm != "" {
				assert.Equal(t, http.StatusForbidden, code)
				return
			}
			assert.Equal(t, http.StatusUnauthorized, code)
		})
		t.Run(tc.name+"/child", func(t *testing.T) {
			code := doRBACRequest(t, ts, tc.method, tc.path, childToken)
			if tc.allowChild {
				assertRBACAllowed(t, code)
				return
			}
			if tc.perm != "" {
				assert.Equal(t, http.StatusForbidden, code)
				return
			}
			assert.Equal(t, http.StatusUnauthorized, code)
		})
		t.Run(tc.name+"/admin", func(t *testing.T) {
			code := doRBACRequest(t, ts, tc.method, tc.path, ts.AdminToken)
			assertRBACAllowed(t, code)
		})
	}
}

func doRBACRequest(t *testing.T, ts *bootstrap.TestServer, method, path, token string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, ts.Server.URL+path, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := ts.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode
}

func assertRBACAllowed(t *testing.T, code int) {
	t.Helper()
	assert.NotEqual(t, http.StatusUnauthorized, code, "expected authenticated access")
	assert.NotEqual(t, http.StatusForbidden, code, "expected permission granted")
}

func createRoleToken(t *testing.T, ts *bootstrap.TestServer, username, password string, roles []string) string {
	t.Helper()
	createBody, _ := json.Marshal(map[string]any{
		"username": username,
		"password": password,
		"roles":    roles,
	})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.Server.URL+"/api/v1/users", bytes.NewReader(createBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := ts.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	loginBody, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, ts.Server.URL+"/api/v1/auth/login", bytes.NewReader(loginBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var tok map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&tok))
	return tok["accessToken"].(string)
}
