//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestGetSymbol(t *testing.T) {
	resp := env.GET(t, env.chartPath("symbol"))
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ChartID       string `json:"chart_id"`
		CurrentSymbol string `json:"current_symbol"`
	}](t, resp)

	requireField(t, result.ChartID, env.ChartID, "chart_id")
	if result.CurrentSymbol == "" {
		t.Fatal("current_symbol is empty")
	}
}

func TestSetSymbol_AndRestore(t *testing.T) {
	// Read current symbol so we can restore it.
	resp := env.GET(t, env.chartPath("symbol"))
	requireStatus(t, resp, http.StatusOK)
	original := decodeJSON[struct {
		CurrentSymbol string `json:"current_symbol"`
	}](t, resp)

	t.Cleanup(func() {
		r := env.PUT(t, env.chartPath("symbol")+"?symbol="+original.CurrentSymbol, nil)
		r.Body.Close()
	})

	// Set to AAPL.
	resp = env.PUT(t, env.chartPath("symbol")+"?symbol=AAPL", nil)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		RequestedSymbol string `json:"requested_symbol"`
		CurrentSymbol   string `json:"current_symbol"`
	}](t, resp)

	requireField(t, result.RequestedSymbol, "AAPL", "requested_symbol")
	if result.CurrentSymbol == "" {
		t.Fatal("current_symbol is empty after set")
	}
}

func TestGetResolution(t *testing.T) {
	resp := env.GET(t, env.chartPath("resolution"))
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ChartID           string `json:"chart_id"`
		CurrentResolution string `json:"current_resolution"`
	}](t, resp)

	requireField(t, result.ChartID, env.ChartID, "chart_id")
	if result.CurrentResolution == "" {
		t.Fatal("current_resolution is empty")
	}
}

func TestSetResolution_AndRestore(t *testing.T) {
	// Read current resolution so we can restore it.
	resp := env.GET(t, env.chartPath("resolution"))
	requireStatus(t, resp, http.StatusOK)
	original := decodeJSON[struct {
		CurrentResolution string `json:"current_resolution"`
	}](t, resp)

	t.Cleanup(func() {
		r := env.PUT(t, env.chartPath("resolution")+"?resolution="+original.CurrentResolution, nil)
		r.Body.Close()
	})

	// Set to 15-minute.
	resp = env.PUT(t, env.chartPath("resolution")+"?resolution=15", nil)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		RequestedResolution string `json:"requested_resolution"`
		CurrentResolution   string `json:"current_resolution"`
	}](t, resp)

	requireField(t, result.RequestedResolution, "15", "requested_resolution")
	if result.CurrentResolution == "" {
		t.Fatal("current_resolution is empty after set")
	}
}
