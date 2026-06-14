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

	"github.com/somralab/somra-media/internal/platform/db"
)

func TestBrowseHandlers_DiscoverSearchDetail(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	svc := newTestAuthService(t, d)
	mediaRepo := db.NewMediaRepo(d.Querier())

	libRepo := db.NewLibraryRepo(d.Querier())
	lib, err := libRepo.Create(ctx, "Browse", db.LibraryKindMovie, []string{"/tmp"}, true)
	require.NoError(t, err)

	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, db.LibraryKindMovie, "Search Me", nil)
	require.NoError(t, err)
	require.NoError(t, mediaRepo.SetLocalizedText(ctx, itemID, "en-US", "title", "Search Me"))
	require.NoError(t, mediaRepo.IndexFTS(ctx, itemID, "Search Me"))

	h := New(Options{
		AuthHandlers:   &AuthHandlers{Service: svc},
		AuthMiddleware: &AuthMiddleware{Service: svc},
		BrowseHandlers: &BrowseHandlers{
			Browse: db.NewBrowseRepo(d.Querier()),
			Locale: func(*http.Request) string { return "en-US" },
		},
	})

	setupBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "AdminPass1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(setupBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var tok map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tok))
	access := tok["accessToken"].(string)

	req = authRequest(http.MethodGet, "/api/v1/discover/home", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/search?q=Search", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var searchResp struct {
		Results []db.SearchResult `json:"results"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &searchResp))
	assert.NotEmpty(t, searchResp.Results)

	path := "/api/v1/libraries/" + jsonNumber(lib.ID) + "/items?limit=10"
	req = authRequest(http.MethodGet, path, access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/media-items/"+jsonNumber(itemID)+"/detail", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var detailBody map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &detailBody))
	assert.Equal(t, json.RawMessage("[]"), detailBody["cast"])
	assert.Equal(t, json.RawMessage("[]"), detailBody["genres"])
}
