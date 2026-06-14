package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
)

func TestBrowseHandlers_ValidationAndParental(t *testing.T) {
	h, _, d, access := newSprint03Router(t)
	ctx := context.Background()

	lib := db.NewLibraryRepo(d.Querier())
	media := db.NewMediaRepo(d.Querier())
	libRec, err := lib.Create(ctx, "Browse", db.LibraryKindMovie, []string{t.TempDir()}, false)
	require.NoError(t, err)

	itemID, err := media.CreateItem(ctx, libRec.ID, db.LibraryKindMovie, "Restricted", nil)
	require.NoError(t, err)
	_, err = d.SQL().ExecContext(ctx, `UPDATE media_item SET content_rating = ? WHERE id = ?`, "R", itemID)
	require.NoError(t, err)
	require.NoError(t, media.SetLocalizedText(ctx, itemID, "en-US", "title", "Restricted"))

	req := authRequest(http.MethodGet, "/api/v1/libraries/not-a-number/items", access, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/media-items/bad-id/detail", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/media-items/99999/detail", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)

	listPath := fmt.Sprintf("/api/v1/libraries/%d/items?sort=year&year=2024&limit=10&offset=0", libRec.ID)
	req = authRequest(http.MethodGet, listPath, access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/search?q=&limit=999", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	pg := "PG"
	profileBody, _ := json.Marshal(map[string]any{
		"locale": "en-US", "theme": "cinematic", "isChild": true, "maxContentRating": pg,
	})
	req = authRequest(http.MethodPut, "/api/v1/profile", access, profileBody)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodGet, fmt.Sprintf("/api/v1/media-items/%d/detail", itemID), access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestFilterSummariesByParental(t *testing.T) {
	rating := "R"
	items := []db.MediaItemSummary{
		{MediaItem: db.MediaItem{ID: 1, ContentRating: &rating}},
		{MediaItem: db.MediaItem{ID: 2}},
	}
	pg := "PG"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(auth.WithAuthContext(req.Context(), auth.AuthContext{
		Profile: db.UserProfile{IsChild: true, MaxContentRating: &pg},
	}))

	filtered := filterSummariesByParental(req, items)
	require.Len(t, filtered, 1)
	assert.Equal(t, int64(2), filtered[0].ID)

	req = req.WithContext(context.Background())
	assert.Len(t, filterSummariesByParental(req, items), 2)
}

func TestBrowseHandlers_DiscoverHomeWithParentalFilter(t *testing.T) {
	h, _, d, access := newSprint03Router(t)
	ctx := context.Background()

	users := db.NewUserRepo(d.Querier())
	all, err := users.List(ctx)
	require.NoError(t, err)
	userID := all[0].ID

	lib := db.NewLibraryRepo(d.Querier())
	media := db.NewMediaRepo(d.Querier())
	watch := db.NewWatchRepo(d.Querier())
	libRec, err := lib.Create(ctx, "Home", db.LibraryKindMovie, []string{t.TempDir()}, false)
	require.NoError(t, err)

	itemID, err := media.CreateItem(ctx, libRec.ID, db.LibraryKindMovie, "Adult Film", nil)
	require.NoError(t, err)
	_, err = d.SQL().ExecContext(ctx, `UPDATE media_item SET content_rating = ? WHERE id = ?`, "R", itemID)
	require.NoError(t, err)
	require.NoError(t, watch.UpsertWatchState(ctx, db.WatchState{
		UserID: userID, MediaItemID: itemID, PositionMs: 1000, Completed: false,
	}))

	pg := "PG"
	profileBody, _ := json.Marshal(map[string]any{
		"locale": "en-US", "theme": "cinematic", "isChild": true, "maxContentRating": pg,
	})
	req := authRequest(http.MethodPut, "/api/v1/profile", access, profileBody)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	req = authRequest(http.MethodGet, "/api/v1/discover/home", access, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var home map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &home))
	shelves, ok := home["shelves"].([]any)
	require.True(t, ok)
	for _, shelf := range shelves {
		row, ok := shelf.(map[string]any)
		require.True(t, ok)
		if row["id"] == "continueWatching" {
			items, ok := row["items"].([]any)
			require.True(t, ok)
			assert.Empty(t, items)
		}
	}
}

func TestBrowseHandlers_localeDefault(t *testing.T) {
	h := &BrowseHandlers{Browse: db.NewBrowseRepo(openTestDB(t).Querier())}
	assert.Equal(t, "en-US", h.locale(httptest.NewRequest(http.MethodGet, "/", nil)))
}
