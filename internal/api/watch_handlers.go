package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/db"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
)

// WatchHandlers serves watch state, favorites, and watchlist.
type WatchHandlers struct {
	Watch *db.WatchRepo
}

type watchStateRequest struct {
	PositionMs int64 `json:"positionMs"`
	Completed  bool  `json:"completed"`
}

type watchStateResponse struct {
	MediaItemID int64 `json:"mediaItemId"`
	PositionMs  int64 `json:"positionMs"`
	Completed   bool  `json:"completed"`
}

// Mount registers watch routes (requires auth).
func (h *WatchHandlers) Mount(r chi.Router) {
	r.Route("/watch-state", func(r chi.Router) {
		r.Get("/", h.listWatchState)
		r.Route("/{mediaItemID}", func(r chi.Router) {
			r.Get("/", h.getWatchState)
			r.Put("/", h.upsertWatchState)
		})
	})
	r.Route("/favorites", func(r chi.Router) {
		r.Get("/", h.listFavorites)
		r.Post("/{mediaItemID}", h.addFavorite)
		r.Delete("/{mediaItemID}", h.removeFavorite)
	})
	r.Route("/watchlist", func(r chi.Router) {
		r.Get("/", h.listWatchlist)
		r.Post("/{mediaItemID}", h.addWatchlist)
		r.Delete("/{mediaItemID}", h.removeWatchlist)
	})
}

func (h *WatchHandlers) listWatchState(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	states, err := h.Watch.ListWatchStates(r.Context(), ac.Claims.UserID)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.watch.list_failed"))
		return
	}
	out := make([]watchStateResponse, 0, len(states))
	for _, s := range states {
		out = append(out, watchStateResponse{MediaItemID: s.MediaItemID, PositionMs: s.PositionMs, Completed: s.Completed})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *WatchHandlers) getWatchState(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.watch.invalid_item"))
		return
	}
	ws, err := h.Watch.GetWatchState(r.Context(), ac.Claims.UserID, itemID)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "auth.watch.not_found"))
		return
	}
	writeJSON(w, http.StatusOK, watchStateResponse{MediaItemID: ws.MediaItemID, PositionMs: ws.PositionMs, Completed: ws.Completed})
}

func (h *WatchHandlers) upsertWatchState(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.watch.invalid_item"))
		return
	}
	var req watchStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.watch.invalid"))
		return
	}
	ws := db.WatchState{UserID: ac.Claims.UserID, MediaItemID: itemID, PositionMs: req.PositionMs, Completed: req.Completed}
	if err := h.Watch.UpsertWatchState(r.Context(), ws); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.watch.update_failed"))
		return
	}
	writeJSON(w, http.StatusOK, watchStateResponse{MediaItemID: itemID, PositionMs: req.PositionMs, Completed: req.Completed})
}

func (h *WatchHandlers) listFavorites(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	ids, err := h.Watch.ListFavorites(r.Context(), ac.Claims.UserID)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.favorites.list_failed"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"mediaItemIds": ids})
}

func (h *WatchHandlers) addFavorite(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.favorites.invalid_item"))
		return
	}
	if err := h.Watch.AddFavorite(r.Context(), ac.Claims.UserID, itemID); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.favorites.add_failed"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WatchHandlers) removeFavorite(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.favorites.invalid_item"))
		return
	}
	if err := h.Watch.RemoveFavorite(r.Context(), ac.Claims.UserID, itemID); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.favorites.remove_failed"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WatchHandlers) listWatchlist(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	ids, err := h.Watch.ListWatchlist(r.Context(), ac.Claims.UserID)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.watchlist.list_failed"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"mediaItemIds": ids})
}

func (h *WatchHandlers) addWatchlist(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.watchlist.invalid_item"))
		return
	}
	if err := h.Watch.AddWatchlist(r.Context(), ac.Claims.UserID, itemID); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.watchlist.add_failed"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WatchHandlers) removeWatchlist(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, r, platformerrors.New(http.StatusUnauthorized, platformerrors.CodeUnauthorized, "auth.errors.unauthorized"))
		return
	}
	itemID, err := parseItemID(r)
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "auth.watchlist.invalid_item"))
		return
	}
	if err := h.Watch.RemoveWatchlist(r.Context(), ac.Claims.UserID, itemID); err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "auth.watchlist.remove_failed"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseItemID(r *http.Request) (int64, error) {
	raw := chi.URLParam(r, "mediaItemID")
	if raw == "" {
		raw = chi.URLParam(r, "itemId")
	}
	return strconv.ParseInt(raw, 10, 64)
}
