package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
)

// BrowseHandlers serves discover, search, paginated browse, and detail endpoints.
type BrowseHandlers struct {
	Browse *db.BrowseRepo
	Locale func(*http.Request) string
}

// Mount registers browse/discover routes (requires auth + library:read).
func (h *BrowseHandlers) Mount(r chi.Router) {
	r.With(RequirePermission(auth.PermLibraryRead)).Get("/discover/home", h.discoverHome)
	r.With(RequirePermission(auth.PermLibraryRead)).Get("/search", h.search)
	r.With(RequirePermission(auth.PermLibraryRead)).Get("/libraries/{libraryID}/items", h.listItemsPaginated)
	r.With(RequirePermission(auth.PermLibraryRead)).Get("/media-items/{itemID}/detail", h.getDetail)
}

func (h *BrowseHandlers) locale(r *http.Request) string {
	if h.Locale != nil {
		return h.Locale(r)
	}
	return "en-US"
}

func (h *BrowseHandlers) discoverHome(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.unauthorized"))
		return
	}
	home, err := h.Browse.DiscoverHome(r.Context(), ac.Claims.UserID, h.locale(r))
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "discover.home.failed"))
		return
	}
	for i := range home.Shelves {
		items := make([]db.MediaItemSummary, len(home.Shelves[i].Items))
		for j, item := range home.Shelves[i].Items {
			items[j] = item
		}
		filtered := filterSummariesByParental(r, home.Shelves[i].Items)
		home.Shelves[i].Items = filtered
	}
	writeJSON(w, http.StatusOK, home)
}

func (h *BrowseHandlers) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit := queryInt(r, "limit", 20)
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	results, err := h.Browse.SearchFTS(r.Context(), q, h.locale(r), limit)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "search.failed"))
		return
	}
	items := make([]db.MediaItem, 0, len(results))
	for _, res := range results {
		items = append(items, res.MediaItem)
	}
	filtered := filterByParental(r, items)
	allowed := make(map[int64]struct{}, len(filtered))
	for _, item := range filtered {
		allowed[item.ID] = struct{}{}
	}
	out := make([]db.SearchResult, 0, len(filtered))
	for _, res := range results {
		if _, ok := allowed[res.ID]; ok {
			out = append(out, res)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": out, "query": q})
}

func (h *BrowseHandlers) listItemsPaginated(w http.ResponseWriter, r *http.Request) {
	libraryID, err := parseID(chi.URLParam(r, "libraryID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "library.id.invalid"))
		return
	}
	f := db.BrowseFilters{
		Offset:      queryInt(r, "offset", 0),
		Limit:       queryInt(r, "limit", 50),
		Sort:        r.URL.Query().Get("sort"),
		Genre:       r.URL.Query().Get("genre"),
		WatchStatus: r.URL.Query().Get("watchStatus"),
	}
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			f.Year = &y
		}
	}
	if ac, ok := auth.FromContext(r.Context()); ok {
		f.UserID = ac.Claims.UserID
	}
	page, err := h.Browse.ListPaginated(r.Context(), libraryID, h.locale(r), f)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "browse.list.failed"))
		return
	}
	page.Items = filterByParental(r, page.Items)
	writeJSON(w, http.StatusOK, page)
}

func (h *BrowseHandlers) getDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "media.id.invalid"))
		return
	}
	userID := ""
	if ac, ok := auth.FromContext(r.Context()); ok {
		userID = ac.Claims.UserID
	}
	detail, err := h.Browse.GetDetail(r.Context(), id, h.locale(r), userID)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusNotFound, platformerrors.CodeNotFound, "media.notFound"))
		return
	}
	if ac, ok := auth.FromContext(r.Context()); ok {
		maxRating := ac.Profile.MaxContentRating
		if ac.Profile.IsChild && maxRating == nil {
			g := "PG"
			maxRating = &g
		}
		if !auth.RatingAllowed(maxRating, detail.ContentRating) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "media.notFound"))
			return
		}
	}
	writeJSON(w, http.StatusOK, detail)
}

func filterSummariesByParental(r *http.Request, items []db.MediaItemSummary) []db.MediaItemSummary {
	plain := make([]db.MediaItem, len(items))
	for i, item := range items {
		plain[i] = item.MediaItem
	}
	filtered := filterByParental(r, plain)
	allowed := make(map[int64]struct{}, len(filtered))
	for _, item := range filtered {
		allowed[item.ID] = struct{}{}
	}
	out := make([]db.MediaItemSummary, 0, len(filtered))
	for _, item := range items {
		if _, ok := allowed[item.ID]; ok {
			out = append(out, item)
		}
	}
	return out
}

func queryInt(r *http.Request, key string, def int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}
