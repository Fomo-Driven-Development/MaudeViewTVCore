//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// strategyPath returns the chart-scoped path for strategy endpoints.
func strategyPath(suffix string) string {
	return env.chartPath("strategy/" + suffix)
}

// getActiveStrategy returns the active strategy data or calls t.Skip if none exists.
// The endpoint returns {strategy, inputs, meta, status} even with no strategy loaded,
// but all values will be null. We check the "strategy" field to detect a real strategy.
func getActiveStrategy(t *testing.T) map[string]any {
	t.Helper()
	resp := env.GET(t, strategyPath("active"))
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		skipOrFatal(t, "no active strategy on chart (non-200 response)")
	}
	result := decodeJSON[map[string]any](t, resp)
	if len(result) == 0 {
		skipOrFatal(t, "no active strategy on chart (empty response)")
	}
	// The endpoint always returns the shape {strategy, inputs, meta, status} but
	// values are null when no strategy is loaded.
	if result["strategy"] == nil {
		skipOrFatal(t, "no active strategy on chart (strategy field is null)")
	}
	return result
}

// TestStrategy_Init runs first (tests within a file execute in declaration order)
// and re-adds the test strategy if a prior test group (e.g. Pine editor tests with
// Ctrl+K Ctrl+S "New strategy") auto-applied a different strategy to the chart,
// replacing ours. TradingView only allows one strategy at a time.
func TestStrategy_Init(t *testing.T) {
	if !env.StrategyReady {
		t.Skip("strategy not set up")
		return
	}

	resp := env.GET(t, strategyPath("active"))
	if resp.StatusCode == http.StatusOK {
		result := decodeJSON[map[string]any](t, resp)
		if result["strategy"] != nil {
			t.Log("strategy already present on chart — no re-add needed")
			return
		}
	} else {
		resp.Body.Close()
	}

	t.Log("strategy disappeared (likely replaced by Pine tests), re-adding...")
	if err := addTestStrategy(); err != nil {
		t.Fatalf("re-add strategy: %v", err)
	}
	time.Sleep(5 * time.Second)

	// Verify the strategy is now active.
	resp = env.GET(t, strategyPath("active"))
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[map[string]any](t, resp)
	if result["strategy"] == nil {
		t.Fatal("strategy still not active after re-add")
	}
	t.Log("strategy re-added and verified")
}

// --- Probing & Discovery ---

func TestStrategyProbe(t *testing.T) {
	resp := env.GET(t, strategyPath("probe"))
	requireStatus(t, resp, http.StatusOK)

	probe := decodeJSON[struct {
		Found       bool     `json:"found"`
		AccessPaths []string `json:"access_paths"`
		Methods     []string `json:"methods"`
	}](t, resp)

	if !probe.Found {
		t.Fatal("expected strategy API to be found")
	}
	if len(probe.Methods) == 0 {
		t.Fatal("expected strategy API to have methods")
	}
	t.Logf("strategy probe: found=%v, %d methods, paths=%v", probe.Found, len(probe.Methods), probe.AccessPaths)
}

// --- Read Operations ---

func TestStrategyList(t *testing.T) {
	resp := env.GET(t, strategyPath("list"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Strategies any `json:"strategies"`
	}](t, resp)

	t.Logf("strategies result: %v", result.Strategies)
}

func TestStrategyGetActive(t *testing.T) {
	active := getActiveStrategy(t)
	t.Logf("active strategy keys: %d", len(active))
}

func TestStrategyReport(t *testing.T) {
	_ = getActiveStrategy(t)

	resp := env.GET(t, strategyPath("report"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[map[string]any](t, resp)
	if len(result) == 0 {
		t.Fatal("expected report data")
	}
	t.Logf("report keys: %d", len(result))
}

func TestStrategyDateRange(t *testing.T) {
	_ = getActiveStrategy(t)

	resp := env.GET(t, strategyPath("date-range"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		DateRange any `json:"date_range"`
	}](t, resp)

	if result.DateRange == nil {
		skipOrFatal(t, "date_range is nil (strategy may not have backtest data)")
	}
	t.Logf("date range: %v", result.DateRange)
}

// --- Validation Tests ---

func TestSetActiveStrategy_EmptyID(t *testing.T) {
	resp := env.PUT(t, strategyPath("active"), map[string]any{
		"strategy_id": "",
	})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for empty strategy_id")
	}
	resp.Body.Close()
	t.Logf("empty strategy_id correctly rejected with status %d", resp.StatusCode)
}

func TestSetStrategyInput_EmptyName(t *testing.T) {
	resp := env.PUT(t, strategyPath("input"), map[string]any{
		"name":  "",
		"value": 42,
	})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for empty name")
	}
	resp.Body.Close()
	t.Logf("empty name correctly rejected with status %d", resp.StatusCode)
}

// --- Stateful Operations ---

func TestStrategy_SetActiveAndRestore(t *testing.T) {
	active := getActiveStrategy(t)

	// Extract the strategy entity ID from the active response.
	strategyData, ok := active["strategy"].(map[string]any)
	if !ok {
		skipOrFatal(t, "active strategy response has no 'strategy' object")
	}
	id, ok := strategyData["id"]
	if !ok {
		skipOrFatal(t, "active strategy has no 'id' field")
	}
	var strategyID string
	switch v := id.(type) {
	case string:
		strategyID = v
	case float64:
		strategyID = fmt.Sprintf("%.0f", v)
	default:
		skipOrFatal(t, fmt.Sprintf("unexpected strategy id type %T", id))
	}
	if strategyID == "" {
		skipOrFatal(t, "active strategy id is empty")
	}
	t.Logf("current active strategy id: %s", strategyID)

	// Set the same strategy as active (idempotent).
	resp := env.PUT(t, strategyPath("active"), map[string]any{
		"strategy_id": strategyID,
	})
	requireStatus(t, resp, http.StatusOK)
	setResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, setResult.Status, "set", "status")
	t.Logf("set active strategy %s → status=%s", strategyID, setResult.Status)

	time.Sleep(1 * time.Second)

	// Verify it's still active.
	afterActive := getActiveStrategy(t)
	afterStrategy, ok := afterActive["strategy"].(map[string]any)
	if !ok {
		t.Fatal("expected 'strategy' in response after set")
	}
	afterID, _ := afterStrategy["id"]
	t.Logf("active strategy after set: %v", afterID)
}

func TestStrategy_GotoDate(t *testing.T) {
	_ = getActiveStrategy(t)

	// Get date range to find a valid timestamp.
	resp := env.GET(t, strategyPath("date-range"))
	requireStatus(t, resp, http.StatusOK)
	dateRangeResp := decodeJSON[struct {
		DateRange any `json:"date_range"`
	}](t, resp)

	if dateRangeResp.DateRange == nil {
		skipOrFatal(t, "no date range available")
	}

	// Extract a timestamp from date_range. It may be a map with from/to or an array.
	var ts float64
	switch dr := dateRangeResp.DateRange.(type) {
	case map[string]any:
		if from, ok := dr["from"].(float64); ok {
			ts = from
		} else if to, ok := dr["to"].(float64); ok {
			ts = to
		}
	case []any:
		if len(dr) > 0 {
			if v, ok := dr[0].(float64); ok {
				ts = v
			}
		}
	}
	if ts == 0 {
		// Fallback: use a reasonable timestamp (Jan 15 2024).
		ts = 1705276800.0
	}
	t.Logf("navigating to timestamp: %.0f", ts)

	resp = env.POST(t, strategyPath("goto"), map[string]any{
		"timestamp": ts,
		"below_bar": false,
	})
	requireStatus(t, resp, http.StatusOK)
	gotoResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, gotoResult.Status, "navigated", "status")
	t.Logf("goto date → status=%s", gotoResult.Status)
}

// --- Full Lifecycle ---

func TestStrategyFullLifecycle(t *testing.T) {
	// 1. Probe strategy API.
	resp := env.GET(t, strategyPath("probe"))
	requireStatus(t, resp, http.StatusOK)
	probe := decodeJSON[struct {
		Found   bool     `json:"found"`
		Methods []string `json:"methods"`
	}](t, resp)
	if !probe.Found {
		t.Fatal("expected strategy API to be found")
	}
	t.Logf("1/6 probe: found=%v, %d methods", probe.Found, len(probe.Methods))

	// 2. List strategies.
	resp = env.GET(t, strategyPath("list"))
	requireStatus(t, resp, http.StatusOK)
	listResult := decodeJSON[struct {
		Strategies any `json:"strategies"`
	}](t, resp)
	t.Logf("2/6 list: strategies=%v", listResult.Strategies)

	// 3. Get active strategy (skip remainder if none).
	active := getActiveStrategy(t)
	t.Logf("3/6 active: %d keys", len(active))

	// 4. Get report.
	resp = env.GET(t, strategyPath("report"))
	requireStatus(t, resp, http.StatusOK)
	report := decodeJSON[map[string]any](t, resp)
	t.Logf("4/6 report: %d keys", len(report))

	// 5. Get date range.
	resp = env.GET(t, strategyPath("date-range"))
	requireStatus(t, resp, http.StatusOK)
	dateRangeResp := decodeJSON[struct {
		DateRange any `json:"date_range"`
	}](t, resp)
	t.Logf("5/6 date-range: %v", dateRangeResp.DateRange)

	// 6. Goto a date (use a timestamp from date range or fallback).
	var ts float64
	switch dr := dateRangeResp.DateRange.(type) {
	case map[string]any:
		if from, ok := dr["from"].(float64); ok {
			ts = from
		}
	case []any:
		if len(dr) > 0 {
			if v, ok := dr[0].(float64); ok {
				ts = v
			}
		}
	}
	if ts == 0 {
		ts = 1705276800.0
	}
	resp = env.POST(t, strategyPath("goto"), map[string]any{
		"timestamp": ts,
		"below_bar": false,
	})
	requireStatus(t, resp, http.StatusOK)
	gotoResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, gotoResult.Status, "navigated", "status")
	t.Logf("6/6 goto: timestamp=%.0f → %s", ts, gotoResult.Status)

	t.Logf("full strategy lifecycle completed successfully")
}
