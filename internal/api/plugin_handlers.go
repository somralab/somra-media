package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/somralab/somra-media/internal/auth"
	platformerrors "github.com/somralab/somra-media/internal/platform/errors"
	"github.com/somralab/somra-media/internal/plugin"
)

// PluginHandlers serves plugin catalog and instance management endpoints.
type PluginHandlers struct {
	Manager *plugin.Manager
}

// Mount registers /plugins routes.
func (h *PluginHandlers) Mount(r chi.Router) {
	r.Route("/plugins", func(r chi.Router) {
		r.With(RequirePermission(auth.PermPluginsManage)).Get("/catalog", h.listCatalog)
		r.With(RequirePermission(auth.PermPluginsManage)).Get("/instances", h.listInstances)
		r.With(RequirePermission(auth.PermPluginsManage)).Post("/instances", h.createInstance)
		r.With(RequirePermission(auth.PermPluginsManage)).Get("/instances/{instanceId}", h.getInstance)
		r.With(RequirePermission(auth.PermPluginsManage)).Patch("/instances/{instanceId}", h.patchInstance)
		r.With(RequirePermission(auth.PermPluginsManage)).Delete("/instances/{instanceId}", h.deleteInstance)
		r.With(RequirePermission(auth.PermPluginsManage)).Post("/instances/{instanceId}/test", h.testInstance)
	})
}

func (h *PluginHandlers) listCatalog(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "plugins.catalog.unavailable"))
		return
	}
	entries := h.Manager.Catalog()
	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]any{
			"pluginType":      e.PluginType,
			"implementation":  e.Implementation,
			"contractVersion": e.ContractVersion,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"catalog": out})
}

func (h *PluginHandlers) listInstances(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "plugins.instances.list.failed"))
		return
	}
	rows, err := h.Manager.List(r.Context())
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.list.failed"))
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, rec := range rows {
		item, err := h.instanceResponse(r.Context(), rec)
		if err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.list.failed"))
			return
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"instances": out})
}

func (h *PluginHandlers) getInstance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "instanceId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.id.invalid"))
		return
	}
	rec, err := h.Manager.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, plugin.ErrPluginNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "plugins.instances.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.get.failed"))
		return
	}
	item, err := h.instanceResponse(r.Context(), rec)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.get.failed"))
		return
	}
	writeJSON(w, http.StatusOK, item)
}

type pluginInstanceInput struct {
	PluginType     string         `json:"pluginType"`
	Implementation string         `json:"implementation"`
	Name           string         `json:"name"`
	Config         map[string]any `json:"config"`
	Enabled        *bool          `json:"enabled"`
}

func (h *PluginHandlers) createInstance(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil {
		writeError(w, r, platformerrors.New(http.StatusServiceUnavailable, platformerrors.CodeInternal, "plugins.instances.create.failed"))
		return
	}
	var in pluginInstanceInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.PluginType == "" || in.Implementation == "" {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.create.invalid"))
		return
	}
	cfgBytes, err := json.Marshal(in.Config)
	if err != nil || in.Config == nil {
		if in.Config == nil {
			cfgBytes = []byte("{}")
		} else {
			writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.create.invalid"))
			return
		}
	}
	enabled := false
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	id, err := h.Manager.Create(r.Context(), plugin.InstanceRecord{
		PluginType:     plugin.PluginType(in.PluginType),
		Implementation: in.Implementation,
		Name:           in.Name,
		Config:         cfgBytes,
		Enabled:        enabled,
	})
	if err != nil {
		if errors.Is(err, plugin.ErrFactoryNotFound) || errors.Is(err, plugin.ErrDuplicateInstance) {
			writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.create.invalid"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.create.failed"))
		return
	}
	rec, err := h.Manager.Get(r.Context(), id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.create.failed"))
		return
	}
	item, err := h.instanceResponse(r.Context(), rec)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.create.failed"))
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

type pluginInstancePatch struct {
	Name    *string        `json:"name"`
	Config  map[string]any `json:"config"`
	Enabled *bool          `json:"enabled"`
}

func (h *PluginHandlers) patchInstance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "instanceId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.id.invalid"))
		return
	}
	rec, err := h.Manager.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, plugin.ErrPluginNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "plugins.instances.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.patch.failed"))
		return
	}
	var patch pluginInstancePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.patch.invalid"))
		return
	}
	ctx := r.Context()
	if patch.Name != nil {
		if err := h.Manager.UpdateName(ctx, id, *patch.Name); err != nil {
			if errors.Is(err, plugin.ErrDuplicateInstance) {
				writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.patch.invalid"))
				return
			}
			writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.patch.failed"))
			return
		}
	}
	if patch.Config != nil {
		cfgBytes, err := json.Marshal(patch.Config)
		if err != nil {
			writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.patch.invalid"))
			return
		}
		if err := h.Manager.PatchConfig(ctx, id, cfgBytes); err != nil {
			writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.patch.failed"))
			return
		}
	}
	if patch.Enabled != nil {
		if *patch.Enabled && !rec.Enabled {
			if err := h.Manager.Enable(ctx, id); err != nil {
				writeError(w, r, platformerrors.Wrap(err, http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.patch.failed"))
				return
			}
		} else if !*patch.Enabled && rec.Enabled {
			if err := h.Manager.Disable(ctx, id); err != nil {
				writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.patch.failed"))
				return
			}
		}
	}
	rec, err = h.Manager.Get(ctx, id)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.patch.failed"))
		return
	}
	item, err := h.instanceResponse(ctx, rec)
	if err != nil {
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.patch.failed"))
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *PluginHandlers) deleteInstance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "instanceId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.id.invalid"))
		return
	}
	if err := h.Manager.Delete(r.Context(), id); err != nil {
		if errors.Is(err, plugin.ErrPluginNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "plugins.instances.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.delete.failed"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PluginHandlers) testInstance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "instanceId"))
	if err != nil {
		writeError(w, r, platformerrors.New(http.StatusBadRequest, platformerrors.CodeValidation, "plugins.instances.id.invalid"))
		return
	}
	result, err := h.Manager.Test(r.Context(), id)
	if err != nil {
		if errors.Is(err, plugin.ErrPluginNotFound) {
			writeError(w, r, platformerrors.New(http.StatusNotFound, platformerrors.CodeNotFound, "plugins.instances.not_found"))
			return
		}
		writeError(w, r, platformerrors.Wrap(err, http.StatusInternalServerError, platformerrors.CodeInternal, "plugins.instances.test.failed"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success":    result.Success,
		"messageKey": result.MessageKey,
		"details":    result.Details,
	})
}

func (h *PluginHandlers) instanceResponse(ctx context.Context, rec plugin.InstanceRecord) (map[string]any, error) {
	config, err := h.Manager.PublicConfig(ctx, rec.ID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		config = map[string]any{}
	}
	return map[string]any{
		"id":             rec.ID,
		"pluginType":     rec.PluginType,
		"implementation": rec.Implementation,
		"name":           rec.Name,
		"config":         config,
		"enabled":        rec.Enabled,
		"createdAt":      rec.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":      rec.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}
