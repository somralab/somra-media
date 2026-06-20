//go:build integration

package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/somralab/somra-media/internal/automation/download"
	"github.com/somralab/somra-media/internal/automation/worker"
	"github.com/somralab/somra-media/internal/platform/bootstrap"
	"github.com/somralab/somra-media/internal/platform/db"
	indexersearch "github.com/somralab/somra-media/internal/automation/indexer"
)

func TestIntegration_AutomationHandoffWithStubPlugins(t *testing.T) {
	ts := bootstrap.NewTestServer(t)
	ctx := context.Background()

	indexerBody := []byte(`{"pluginType":"indexer","implementation":"stub","name":"auto-indexer","enabled":true}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.Server.URL+"/api/v1/plugins/instances", bytes.NewReader(indexerBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := ts.Server.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	clientBody := []byte(`{"pluginType":"download_client","implementation":"stub","name":"auto-dl","enabled":true}`)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, ts.Server.URL+"/api/v1/plugins/instances", bytes.NewReader(clientBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	createBody, _ := json.Marshal(map[string]any{
		"mediaKind":  "movie",
		"provider":   "tmdb",
		"externalId": "auto-flow-1",
		"title":      "Automation Movie",
	})
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, ts.Server.URL+"/api/v1/requests", bytes.NewReader(createBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()
	requestID := int64(created["id"].(float64))

	autoRepo := db.NewAutomationRepo(ts.Components.DB.Querier())
	_, err = autoRepo.RecordHandoff(ctx, requestID)
	require.NoError(t, err)

	processor := &worker.Processor{
		AutoRepo: autoRepo,
		Requests: db.NewRequestRepo(ts.Components.DB.Querier()),
		Search:   &indexersearch.SearchService{Manager: ts.Plugins.Manager},
		Manager:  ts.Plugins.Manager,
	}
	require.NoError(t, processor.ProcessPending(ctx))

	downloads, err := autoRepo.ListDownloads(ctx, 10)
	require.NoError(t, err)
	require.NotEmpty(t, downloads)

	monitor := &download.Monitor{
		AutoRepo: autoRepo,
		Requests: db.NewRequestRepo(ts.Components.DB.Querier()),
		Manager:  ts.Plugins.Manager,
	}
	require.NoError(t, monitor.Poll(ctx))

	row, err := autoRepo.GetDownloadByID(ctx, downloads[0].ID)
	require.NoError(t, err)
	require.Equal(t, db.AutomationDownloadCompleted, row.Status)

	searchBody := []byte(`{"title":"Automation Movie","mediaKind":"movie"}`)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, ts.Server.URL+"/api/v1/automation/indexers/search", bytes.NewReader(searchBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, ts.Server.URL+"/api/v1/automation/quality-profiles", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+ts.AdminToken)
	resp, err = ts.Server.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
