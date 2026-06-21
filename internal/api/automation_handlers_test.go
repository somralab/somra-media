package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	indexersearch "github.com/somralab/somra-media/internal/automation/indexer"
	"github.com/somralab/somra-media/internal/platform/db"
	"github.com/somralab/somra-media/internal/plugin"
	"github.com/somralab/somra-media/internal/plugin/stub"
)

func newAutomationTestRouter(t *testing.T) (http.Handler, *db.DB, string, *plugin.Manager) {
	t.Helper()
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	mgr := plugin.NewManager(newPluginTestStore(d), plugin.ManagerOptions{EncryptionKey: "test-jwt-secret"})
	require.NoError(t, mgr.RegisterFactory(stub.NewIndexerFactory()))
	require.NoError(t, mgr.RegisterFactory(stub.NewDownloadClientFactory()))

	_, err := mgr.Create(t.Context(), plugin.InstanceRecord{
		PluginType:     plugin.PluginTypeIndexer,
		Implementation: stub.Implementation,
		Name:           "auto-idx",
		Config:         []byte("{}"),
		Enabled:        true,
	})
	require.NoError(t, err)

	h := New(Options{
		AuthHandlers:   &AuthHandlers{Service: svc},
		AuthMiddleware: &AuthMiddleware{Service: svc},
		AutomationHandlers: &AutomationHandlers{
			AutoRepo: db.NewAutomationRepo(d.Querier()),
			Search:   &indexersearch.SearchService{Manager: mgr},
		},
	})

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	var tok map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tok))
	return h, d, tok["accessToken"].(string), mgr
}

func TestAutomationHandlers_Search(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	searchBody := []byte(`{"title":"Demo Movie","mediaKind":"movie"}`)
	req := authRequest(http.MethodPost, "/api/v1/automation/indexers/search", token, searchBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	results := out["results"].([]any)
	require.NotEmpty(t, results)
}

func TestAutomationHandlers_QualityProfiles(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	req := authRequest(http.MethodGet, "/api/v1/automation/quality-profiles", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	createBody := []byte(`{"name":"1080p-prefer","spec":"{\"preferredResolutions\":[\"1080p\"]}"}`)
	req = authRequest(http.MethodPost, "/api/v1/automation/quality-profiles", token, createBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestAutomationHandlers_ListDownloadsEmpty(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	req := authRequest(http.MethodGet, "/api/v1/automation/downloads", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAutomationHandlers_GetDownloadNotFound(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	req := authRequest(http.MethodGet, "/api/v1/automation/downloads/99999", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAutomationHandlers_SearchInvalid(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	req := authRequest(http.MethodPost, "/api/v1/automation/indexers/search", token, []byte(`{"mediaKind":"movie"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAutomationHandlers_GetDownloadSuccess(t *testing.T) {
	h, d, token, _ := newAutomationTestRouter(t)
	repo := db.NewAutomationRepo(d.Querier())
	id, err := repo.CreateDownload(t.Context(), db.AutomationDownload{
		ClientInstanceID: 1,
		ClientDownloadID: "x",
		ReleaseID:        "r",
		Title:            "t",
		Protocol:         "torrent",
		Status:           db.AutomationDownloadQueued,
	})
	require.NoError(t, err)

	req := authRequest(http.MethodGet, "/api/v1/automation/downloads/"+strconv.FormatInt(id, 10), token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAutomationHandlers_CreateProfileDuplicate(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	body := []byte(`{"name":"default","spec":"{}"}`)
	req := authRequest(http.MethodPost, "/api/v1/automation/quality-profiles", token, body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAutomationHandlers_CreateProfileInvalid(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	req := authRequest(http.MethodPost, "/api/v1/automation/quality-profiles", token, []byte(`{}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAutomationHandlers_QualityProfileGetAndPatch(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	req := authRequest(http.MethodGet, "/api/v1/automation/quality-profiles/1", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	patchBody := []byte(`{"name":"default-updated","spec":"{}"}`)
	req = authRequest(http.MethodPatch, "/api/v1/automation/quality-profiles/1", token, patchBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAutomationHandlers_MonitorsCRUD(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	createBody := []byte(`{"title":"Demo Series","externalId":"tv-123","provider":"tmdb","qualityProfile":"default"}`)
	req := authRequest(http.MethodPost, "/api/v1/automation/monitors", token, createBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	id := int64(created["id"].(float64))

	req = authRequest(http.MethodGet, "/api/v1/automation/monitors", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/automation/monitors/"+strconv.FormatInt(id, 10), token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	patchBody := []byte(`{"enabled":false}`)
	req = authRequest(http.MethodPatch, "/api/v1/automation/monitors/"+strconv.FormatInt(id, 10), token, patchBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodDelete, "/api/v1/automation/monitors/"+strconv.FormatInt(id, 10), token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAutomationHandlers_QualityProfileErrors(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)

	req := authRequest(http.MethodGet, "/api/v1/automation/quality-profiles/not-a-id", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/automation/quality-profiles/99999", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)

	req = authRequest(http.MethodPatch, "/api/v1/automation/quality-profiles/99999", token, []byte(`{"name":"x"}`))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)

	req = authRequest(http.MethodPatch, "/api/v1/automation/quality-profiles/bad", token, []byte(`{"name":"x"}`))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAutomationHandlers_MonitorErrors(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)

	req := authRequest(http.MethodPost, "/api/v1/automation/monitors", token, []byte(`{"title":""}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/automation/monitors/99999", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)

	req = authRequest(http.MethodPatch, "/api/v1/automation/monitors/99999", token, []byte(`{"enabled":true}`))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)

	req = authRequest(http.MethodDelete, "/api/v1/automation/monitors/99999", token, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)

	createBody := []byte(`{"title":"Dup Series","externalId":"dup-1","provider":"tmdb"}`)
	req = authRequest(http.MethodPost, "/api/v1/automation/monitors", token, createBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	req = authRequest(http.MethodPost, "/api/v1/automation/monitors", token, createBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAutomationHandlers_PatchMonitorFields(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	createBody := []byte(`{"title":"Patch Show","externalId":"patch-1","provider":"tmdb","qualityProfile":"default"}`)
	req := authRequest(http.MethodPost, "/api/v1/automation/monitors", token, createBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	id := int64(created["id"].(float64))

	patchBody := []byte(`{"title":"Patch Show Updated","qualityProfile":"default","enabled":true}`)
	req = authRequest(http.MethodPatch, "/api/v1/automation/monitors/"+strconv.FormatInt(id, 10), token, patchBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAutomationHandlers_PatchQualityProfileDefault(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	createBody := []byte(`{"name":"720p-alt","spec":"{}"}`)
	req := authRequest(http.MethodPost, "/api/v1/automation/quality-profiles", token, createBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	id := int64(created["id"].(float64))

	patchBody := []byte(`{"isDefault":true}`)
	req = authRequest(http.MethodPatch, "/api/v1/automation/quality-profiles/"+strconv.FormatInt(id, 10), token, patchBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAutomationHandlers_NilAutoRepo(t *testing.T) {
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	h := New(Options{
		AuthHandlers:   &AuthHandlers{Service: svc},
		AuthMiddleware: &AuthMiddleware{Service: svc},
		AutomationHandlers: &AutomationHandlers{
			AutoRepo: nil,
			Search:   nil,
		},
	})

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	token := jsonDecode(t, rec.Body.Bytes())["accessToken"].(string)

	cases := []struct {
		method string
		path   string
		body   []byte
	}{
		{http.MethodGet, "/api/v1/automation/downloads", nil},
		{http.MethodGet, "/api/v1/automation/quality-profiles", nil},
		{http.MethodGet, "/api/v1/automation/monitors", nil},
		{http.MethodPost, "/api/v1/automation/quality-profiles", []byte(`{"name":"x","spec":"{}"}`)},
	}
	for _, tc := range cases {
		req = authRequest(tc.method, tc.path, token, tc.body)
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusServiceUnavailable, rec.Code, tc.path)
	}
}

func TestAutomationHandlers_PatchMonitorInvalidJSON(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	req := authRequest(http.MethodPatch, "/api/v1/automation/monitors/1", token, []byte(`not-json`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAutomationHandlers_PatchQualityProfileInvalidJSON(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	req := authRequest(http.MethodPatch, "/api/v1/automation/quality-profiles/1", token, []byte(`{`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAutomationHandlers_CreateMonitorWithDefaultProvider(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	body := []byte(`{"title":"Provider Default","externalId":"prov-1"}`)
	req := authRequest(http.MethodPost, "/api/v1/automation/monitors", token, body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestAutomationHandlers_InvalidResourceIDs(t *testing.T) {
	h, _, token, _ := newAutomationTestRouter(t)
	paths := []string{
		"/api/v1/automation/downloads/not-id",
		"/api/v1/automation/quality-profiles/x",
		"/api/v1/automation/monitors/abc",
	}
	for _, path := range paths {
		req := authRequest(http.MethodGet, path, token, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code, path)
	}

	req := authRequest(http.MethodDelete, "/api/v1/automation/monitors/bad-id", token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func jsonDecode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal(b, &out))
	return out
}
