//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"
)

func TestSearchIndicators(t *testing.T) {
	resp := env.POST(t, env.chartPath("indicators/search"), map[string]any{
		"query": "Volume",
	})
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Status     string `json:"status"`
		Query      string `json:"query"`
		TotalCount int    `json:"total_count"`
		Results    []struct {
			Name  string `json:"name"`
			Index int    `json:"index"`
		} `json:"results"`
	}](t, resp)

	if result.TotalCount == 0 {
		t.Fatal("expected at least one search result for 'Volume'")
	}
	t.Logf("found %d results for 'Volume', first: %s", result.TotalCount, result.Results[0].Name)

	// Allow dialog to fully dismiss before next test
	time.Sleep(1 * time.Second)
}

func TestSearchIndicators_EmptyQuery(t *testing.T) {
	resp := env.POST(t, env.chartPath("indicators/search"), map[string]any{
		"query": "",
	})
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
		resp.Body.Close()
		t.Fatalf("expected 400 or 422 for empty query, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAddIndicatorBySearch(t *testing.T) {
	// First get the current study count
	resp := env.GET(t, env.chartPath("studies"))
	requireStatus(t, resp, http.StatusOK)
	before := decodeJSON[struct {
		Studies []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"studies"`
	}](t, resp)
	beforeCount := len(before.Studies)

	// Add RSI via indicator search
	resp = env.POST(t, env.chartPath("indicators/add"), map[string]any{
		"query": "RSI",
		"index": 0,
	})
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Status string `json:"status"`
		Query  string `json:"query"`
		Name   string `json:"name"`
		Index  int    `json:"index"`
	}](t, resp)
	t.Logf("added indicator: %s (query=%s, index=%d)", result.Name, result.Query, result.Index)

	// Wait for study to appear
	time.Sleep(1 * time.Second)

	// Verify study was added
	resp = env.GET(t, env.chartPath("studies"))
	requireStatus(t, resp, http.StatusOK)
	after := decodeJSON[struct {
		Studies []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"studies"`
	}](t, resp)

	if len(after.Studies) <= beforeCount {
		t.Fatalf("expected study count to increase, before=%d after=%d", beforeCount, len(after.Studies))
	}

	// Cleanup: remove the last added study (DELETE returns 204 No Content)
	lastStudy := after.Studies[len(after.Studies)-1]
	resp = env.DELETE(t, env.chartPath("studies/"+lastStudy.ID))
	requireStatus(t, resp, http.StatusNoContent)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)
}

func TestListFavoriteIndicators(t *testing.T) {
	resp := env.GET(t, env.chartPath("indicators/favorites"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Status   string `json:"status"`
		Category string `json:"category"`
		Results  []struct {
			Name  string `json:"name"`
			Index int    `json:"index"`
		} `json:"results"`
		TotalCount int `json:"total_count"`
	}](t, resp)

	t.Logf("favorites: %d results, category=%s", result.TotalCount, result.Category)

	// Allow dialog to fully dismiss before next test
	time.Sleep(1 * time.Second)
}

func TestProbeIndicatorDialogDOM(t *testing.T) {
	resp := env.GET(t, "/api/v1/indicators/probe-dom")
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}
