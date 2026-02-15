//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestHealth(t *testing.T) {
	resp := env.GET(t, "/health")
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "ok", "status")
}

func TestDeepHealth(t *testing.T) {
	resp := env.GET(t, "/api/v1/health/deep")
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[map[string]any](t, resp)
	if len(result) == 0 {
		t.Fatal("expected deep health to return data")
	}
	t.Logf("deep health keys: %d", len(result))
}

func TestBrowserScreenshot(t *testing.T) {
	resp := env.POST(t, "/api/v1/browser_screenshot", map[string]any{
		"format": "png",
	})
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Snapshot map[string]any `json:"snapshot"`
		URL      string         `json:"url"`
	}](t, resp)

	if result.URL == "" {
		t.Fatal("expected non-empty snapshot url")
	}
	if result.Snapshot == nil {
		t.Fatal("expected snapshot metadata")
	}
	t.Logf("browser screenshot: url=%s", result.URL)
}

func TestGetActiveChart(t *testing.T) {
	resp := env.GET(t, "/api/v1/charts/active")
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		ChartID    string `json:"chart_id"`
		TargetID   string `json:"target_id"`
		URL        string `json:"url"`
		Title      string `json:"title"`
		ChartIndex int    `json:"chart_index"`
		ChartCount int    `json:"chart_count"`
	}](t, resp)

	if result.ChartID == "" {
		t.Fatal("expected non-empty chart_id")
	}
	requireField(t, result.ChartID, env.ChartID, "chart_id")
	if result.ChartCount < 1 {
		t.Fatalf("chart_count = %d, want >= 1", result.ChartCount)
	}
	t.Logf("active chart: id=%s index=%d count=%d title=%s", result.ChartID, result.ChartIndex, result.ChartCount, result.Title)
}

// TestPageReload is deliberately placed last and skipped by default because
// it reloads the TradingView page, disrupting all subsequent tests.
// Run with: go test -run TestPageReload -count=1
func TestPageReload(t *testing.T) {
	t.Skip("skipped by default: page reload disrupts other tests")

	resp := env.POST(t, "/api/v1/page/reload", map[string]any{
		"mode": "normal",
	})
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[map[string]any](t, resp)
	t.Logf("page reload result: %v", result)
}
