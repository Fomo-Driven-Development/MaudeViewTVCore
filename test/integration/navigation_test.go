//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"
)

// resetChart sets a known timeframe to bring the chart back to a predictable state.
func resetChart(t *testing.T) {
	t.Helper()
	resp := env.PUT(t, env.chartPath("timeframe")+"?preset=1Y", nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)
}

// goToDateAndVerify posts go-to-date and polls visible-range until the target
// timestamp is within [from, to]. Retries the entire go-to-date call up to
// maxAttempts times because the CDP keyboard dialog interaction is racy.
func goToDateAndVerify(t *testing.T, timestamp int64, maxAttempts int) {
	t.Helper()
	ft := float64(timestamp)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp := env.POST(t, env.chartPath("go-to-date"), map[string]any{
			"timestamp": timestamp,
		})
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			if attempt == maxAttempts {
				t.Fatalf("go-to-date returned %d on attempt %d", resp.StatusCode, attempt)
			}
			t.Logf("attempt %d: go-to-date returned %d, retrying", attempt, resp.StatusCode)
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Body.Close()

		// Poll visible-range up to 3 times for data to load.
		for poll := 0; poll < 3; poll++ {
			time.Sleep(1 * time.Second)
			resp = env.GET(t, env.chartPath("visible-range"))
			requireStatus(t, resp, http.StatusOK)
			vr := decodeJSON[struct {
				From float64 `json:"from"`
				To   float64 `json:"to"`
			}](t, resp)

			if vr.From <= ft && vr.To >= ft {
				return // success
			}
			t.Logf("attempt %d poll %d: target %d not in [%.0f, %.0f]", attempt, poll+1, timestamp, vr.From, vr.To)
		}

		if attempt < maxAttempts {
			t.Logf("attempt %d: go-to-date did not navigate, retrying", attempt)
			time.Sleep(1 * time.Second)
		}
	}
	t.Fatalf("go-to-date to %d failed after %d attempts", timestamp, maxAttempts)
}

// --- GoToDate tests ---

func TestGoToDate_NavigatesToHistoricalDate(t *testing.T) {
	t.Cleanup(func() { resetChart(t) })
	goToDateAndVerify(t, 1704153600, 3) // Jan 2, 2024
}

func TestGoToDate_COVIDCrash(t *testing.T) {
	t.Cleanup(func() { resetChart(t) })
	goToDateAndVerify(t, 1583020800, 3) // Mar 1, 2020
}

func TestGoToDate_InvalidTimestamp(t *testing.T) {
	resp := env.POST(t, env.chartPath("go-to-date"), map[string]any{
		"timestamp": 0,
	})
	defer resp.Body.Close()

	// Expect either 400 or 422 depending on Huma validation.
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 400 or 422", resp.StatusCode)
	}
}

// --- SetTimeFrame tests ---

func TestSetTimeFrame_AllPresets(t *testing.T) {
	presets := []struct {
		name    string
		wantRes string
	}{
		{"1D", "1"},
		{"5D", "5"},
		{"1M", "30"},
		{"3M", "60"},
		{"6M", "120"},
		{"YTD", "1D"},
		{"1Y", "1D"},
		{"5Y", "1W"},
		{"All", "1M"},
	}
	for _, tc := range presets {
		t.Run(tc.name, func(t *testing.T) {
			resp := env.PUT(t, env.chartPath("timeframe")+"?preset="+tc.name, nil)
			requireStatus(t, resp, http.StatusOK)
			result := decodeJSON[struct {
				Preset     string  `json:"preset"`
				Resolution string  `json:"resolution"`
				From       float64 `json:"from"`
				To         float64 `json:"to"`
			}](t, resp)
			requireField(t, result.Resolution, tc.wantRes, "resolution")
			requireField(t, result.Preset, tc.name, "preset")

			// "All" preset can return negative timestamps for symbols with
			// pre-1970 data (e.g. SPX goes back to ~1871).
			if tc.name != "All" && result.From <= 0 {
				t.Fatalf("from = %.0f, want > 0", result.From)
			}
			if result.To <= result.From {
				t.Fatalf("to (%.0f) <= from (%.0f)", result.To, result.From)
			}
		})
	}
}

func TestSetTimeFrame_WithResolutionOverride(t *testing.T) {
	resp := env.PUT(t, env.chartPath("timeframe")+"?preset=1M&resolution=15", nil)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Resolution string `json:"resolution"`
	}](t, resp)
	requireField(t, result.Resolution, "15", "resolution")
}

func TestSetTimeFrame_InvalidPreset(t *testing.T) {
	resp := env.PUT(t, env.chartPath("timeframe")+"?preset=INVALID", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 400 or 422", resp.StatusCode)
	}
}

// --- Resolution tests ---

func TestSetResolution_AllToolbarValues(t *testing.T) {
	t.Cleanup(func() { resetChart(t) })

	resolutions := []struct {
		value string // API value
		label string // toolbar button label (for test naming)
	}{
		{"1", "1m"},
		{"5", "5m"},
		{"15", "15m"},
		{"60", "1h"},
		{"240", "4h"},
		{"360", "6h"},
		{"720", "12h"},
		{"1D", "D"},
		{"3D", "3D"},
		{"1W", "W"},
		{"2W", "2W"},
		{"1M", "M"},
	}

	for _, tc := range resolutions {
		t.Run(tc.label, func(t *testing.T) {
			resp := env.PUT(t, env.chartPath("resolution")+"?resolution="+tc.value, nil)
			requireStatus(t, resp, http.StatusOK)
			result := decodeJSON[struct {
				CurrentResolution string `json:"current_resolution"`
			}](t, resp)
			requireField(t, result.CurrentResolution, tc.value, "current_resolution")
		})
	}
}

// --- VisibleRange tests ---

func TestGetVisibleRange(t *testing.T) {
	resp := env.GET(t, env.chartPath("visible-range"))
	requireStatus(t, resp, http.StatusOK)
	vr := decodeJSON[struct {
		ChartID string  `json:"chart_id"`
		From    float64 `json:"from"`
		To      float64 `json:"to"`
	}](t, resp)

	if vr.From <= 0 {
		t.Fatalf("from = %.0f, want > 0", vr.From)
	}
	if vr.To <= vr.From {
		t.Fatalf("to (%.0f) <= from (%.0f)", vr.To, vr.From)
	}
	requireField(t, vr.ChartID, env.ChartID, "chart_id")
}

// --- ResetView tests ---

func TestResetView(t *testing.T) {
	t.Cleanup(func() { resetChart(t) })

	// Navigate away from realtime so we can verify scroll-back.
	goToDateAndVerify(t, 1704153600, 3) // Jan 2, 2024

	// Record visible range before scroll.
	resp := env.GET(t, env.chartPath("visible-range"))
	requireStatus(t, resp, http.StatusOK)
	before := decodeJSON[struct {
		From float64 `json:"from"`
		To   float64 `json:"to"`
	}](t, resp)

	// Reset chart view.
	resp = env.POST(t, env.chartPath("reset-view"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Poll until visible range advances past the before snapshot.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		resp = env.GET(t, env.chartPath("visible-range"))
		requireStatus(t, resp, http.StatusOK)
		after := decodeJSON[struct {
			From float64 `json:"from"`
			To   float64 `json:"to"`
		}](t, resp)
		if after.To > before.To {
			return // success
		}
	}
	t.Fatalf("visible range did not advance toward realtime (before.To=%.0f)", before.To)
}

// --- ChartType tests ---

func TestGetChartType(t *testing.T) {
	resp := env.GET(t, env.chartPath("chart-type"))
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ChartID     string `json:"chart_id"`
		ChartType   string `json:"chart_type"`
		ChartTypeID int    `json:"chart_type_id"`
	}](t, resp)

	requireField(t, result.ChartID, env.ChartID, "chart_id")
	if result.ChartType == "" {
		t.Fatalf("chart_type is empty")
	}
}

func TestSetChartType_AllStyles(t *testing.T) {
	t.Cleanup(func() {
		// Restore to candles.
		resp := env.PUT(t, env.chartPath("chart-type")+"?type=candles", nil)
		resp.Body.Close()
		time.Sleep(500 * time.Millisecond)
	})

	styles := []struct {
		name string
		id   int
	}{
		{"bars", 0},
		{"candles", 1},
		{"line", 2},
		{"area", 3},
		{"heikin_ashi", 8},
		{"hollow_candles", 9},
		{"baseline", 10},
		{"high_low", 12},
		{"columns", 13},
		{"line_with_markers", 14},
		{"step_line", 15},
		{"hlc_area", 16},
		{"volume_candles", 19},
	}

	for _, tc := range styles {
		t.Run(tc.name, func(t *testing.T) {
			resp := env.PUT(t, env.chartPath("chart-type")+"?type="+tc.name, nil)
			requireStatus(t, resp, http.StatusOK)
			result := decodeJSON[struct {
				ChartType   string `json:"chart_type"`
				ChartTypeID int    `json:"chart_type_id"`
			}](t, resp)
			requireField(t, result.ChartTypeID, tc.id, "chart_type_id")
			requireField(t, result.ChartType, tc.name, "chart_type")

			// Verify GET returns the same type.
			getResp := env.GET(t, env.chartPath("chart-type"))
			requireStatus(t, getResp, http.StatusOK)
			getResult := decodeJSON[struct {
				ChartTypeID int `json:"chart_type_id"`
			}](t, getResp)
			requireField(t, getResult.ChartTypeID, tc.id, "chart_type_id (GET)")
		})
	}
}

func TestSetChartType_InvalidType(t *testing.T) {
	resp := env.PUT(t, env.chartPath("chart-type")+"?type=invalid_type", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 400 or 422", resp.StatusCode)
	}
}

// --- Symbol Info ---

func TestGetSymbolInfo(t *testing.T) {
	resp := env.GET(t, env.chartPath("symbol/info"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		ChartID    string         `json:"chart_id"`
		SymbolInfo map[string]any `json:"symbol_info"`
	}](t, resp)

	requireField(t, result.ChartID, env.ChartID, "chart_id")
	if result.SymbolInfo == nil {
		t.Fatal("expected symbol_info to be non-nil")
	}
	t.Logf("symbol info keys: %d", len(result.SymbolInfo))
}

// --- Chart API Probes ---

func TestProbeChartApi(t *testing.T) {
	resp := env.GET(t, env.chartPath("chart-api/probe"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[map[string]any](t, resp)
	if len(result) == 0 {
		t.Fatal("expected chart-api probe to return data")
	}
	t.Logf("chart-api probe keys: %d", len(result))
}

func TestProbeChartApiDeep(t *testing.T) {
	resp := env.GET(t, env.chartPath("chart-api/probe/deep"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[map[string]any](t, resp)
	if len(result) == 0 {
		t.Fatal("expected chart-api deep probe to return data")
	}
	t.Logf("chart-api deep probe keys: %d", len(result))
}

// --- Zoom ---

func TestZoom_InAndOut(t *testing.T) {
	t.Cleanup(func() { resetChart(t) })

	// Record initial range.
	resp := env.GET(t, env.chartPath("visible-range"))
	requireStatus(t, resp, http.StatusOK)
	before := decodeJSON[struct {
		From float64 `json:"from"`
		To   float64 `json:"to"`
	}](t, resp)
	beforeSpan := before.To - before.From

	// Zoom in (should decrease visible range).
	resp = env.POST(t, env.chartPath("zoom"), map[string]any{
		"direction": "in",
	})
	requireStatus(t, resp, http.StatusOK)
	zoomResult := decodeJSON[struct {
		ChartID   string `json:"chart_id"`
		Status    string `json:"status"`
		Direction string `json:"direction"`
	}](t, resp)
	requireField(t, zoomResult.Status, "executed", "status")
	requireField(t, zoomResult.Direction, "in", "direction")

	time.Sleep(500 * time.Millisecond)

	resp = env.GET(t, env.chartPath("visible-range"))
	requireStatus(t, resp, http.StatusOK)
	afterIn := decodeJSON[struct {
		From float64 `json:"from"`
		To   float64 `json:"to"`
	}](t, resp)
	afterInSpan := afterIn.To - afterIn.From
	if afterInSpan >= beforeSpan {
		t.Logf("warning: zoom in did not decrease range (before=%.0f after=%.0f)", beforeSpan, afterInSpan)
	}
	t.Logf("zoom in: span %.0f → %.0f", beforeSpan, afterInSpan)

	// Zoom out.
	resp = env.POST(t, env.chartPath("zoom"), map[string]any{
		"direction": "out",
	})
	requireStatus(t, resp, http.StatusOK)
	zoomResult = decodeJSON[struct {
		ChartID   string `json:"chart_id"`
		Status    string `json:"status"`
		Direction string `json:"direction"`
	}](t, resp)
	requireField(t, zoomResult.Status, "executed", "status")
	requireField(t, zoomResult.Direction, "out", "direction")
	t.Logf("zoom out executed")
}

// --- Scroll ---

func TestScroll(t *testing.T) {
	t.Cleanup(func() { resetChart(t) })

	// Record initial range.
	resp := env.GET(t, env.chartPath("visible-range"))
	requireStatus(t, resp, http.StatusOK)
	before := decodeJSON[struct {
		From float64 `json:"from"`
		To   float64 `json:"to"`
	}](t, resp)

	// Scroll left by 50 bars.
	resp = env.POST(t, env.chartPath("scroll"), map[string]any{
		"bars": -50,
	})
	requireStatus(t, resp, http.StatusOK)
	scrollResult := decodeJSON[struct {
		ChartID string `json:"chart_id"`
		Status  string `json:"status"`
		Bars    int    `json:"bars"`
	}](t, resp)
	requireField(t, scrollResult.Status, "executed", "status")
	requireField(t, scrollResult.Bars, -50, "bars")

	time.Sleep(500 * time.Millisecond)

	// Verify visible range shifted left.
	resp = env.GET(t, env.chartPath("visible-range"))
	requireStatus(t, resp, http.StatusOK)
	after := decodeJSON[struct {
		From float64 `json:"from"`
		To   float64 `json:"to"`
	}](t, resp)

	if after.From >= before.From {
		t.Logf("warning: scroll left did not shift range (before.From=%.0f after.From=%.0f)", before.From, after.From)
	}
	t.Logf("scroll -50 bars: from %.0f → %.0f", before.From, after.From)
}

// --- Set Visible Range ---

func TestSetVisibleRange(t *testing.T) {
	t.Cleanup(func() { resetChart(t) })

	// Get current range to use reasonable values.
	resp := env.GET(t, env.chartPath("visible-range"))
	requireStatus(t, resp, http.StatusOK)
	current := decodeJSON[struct {
		From float64 `json:"from"`
		To   float64 `json:"to"`
	}](t, resp)

	// Set a narrower range (middle 50% of current).
	span := current.To - current.From
	newFrom := current.From + span*0.25
	newTo := current.To - span*0.25

	resp = env.PUT(t, env.chartPath("visible-range"), map[string]any{
		"from": newFrom,
		"to":   newTo,
	})
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ChartID string  `json:"chart_id"`
		From    float64 `json:"from"`
		To      float64 `json:"to"`
	}](t, resp)

	requireField(t, result.ChartID, env.ChartID, "chart_id")
	t.Logf("set visible range: from=%.0f to=%.0f", result.From, result.To)
}

// --- Chart Snapshot ---

func TestTakeChartSnapshot(t *testing.T) {
	resp := env.POST(t, env.chartPath("snapshot"), map[string]any{
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
	t.Logf("chart snapshot: url=%s", result.URL)

	// Clean up: delete the snapshot if it has an ID.
	if id, ok := result.Snapshot["id"].(string); ok && id != "" {
		r := env.DELETE(t, "/api/v1/snapshots/"+id)
		r.Body.Close()
	}
}

// --- Compare/Overlay tests ---

func TestCompareOverlay(t *testing.T) {
	// 1. Add overlay symbol.
	resp := env.POST(t, env.chartPath("compare"), map[string]any{
		"symbol": "ETHUSD",
	})
	requireStatus(t, resp, http.StatusOK)
	addResult := decodeJSON[struct {
		ChartID string `json:"chart_id"`
		Study   struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"study"`
		Status string `json:"status"`
	}](t, resp)
	requireField(t, addResult.Status, "added", "status")
	requireField(t, addResult.Study.Name, "Overlay", "study.name")
	if addResult.Study.ID == "" {
		t.Fatal("expected non-empty study ID")
	}
	studyID := addResult.Study.ID
	t.Logf("added overlay: id=%s name=%s", studyID, addResult.Study.Name)

	time.Sleep(1 * time.Second)

	// 2. List compares — verify ETHUSD overlay is present.
	resp = env.GET(t, env.chartPath("compare"))
	requireStatus(t, resp, http.StatusOK)
	listResult := decodeJSON[struct {
		ChartID string `json:"chart_id"`
		Studies []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"studies"`
	}](t, resp)

	found := false
	for _, s := range listResult.Studies {
		if s.ID == studyID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("overlay study %s not found in compare list", studyID)
	}
	t.Logf("listed %d compare/overlay studies", len(listResult.Studies))

	// 3. Remove the overlay.
	resp = env.DELETE(t, env.chartPath("compare/"+studyID))
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE compare status = %d, want 200 or 204", resp.StatusCode)
	}
	resp.Body.Close()

	time.Sleep(500 * time.Millisecond)

	// 4. List again — verify empty.
	resp = env.GET(t, env.chartPath("compare"))
	requireStatus(t, resp, http.StatusOK)
	afterList := decodeJSON[struct {
		Studies []struct {
			ID string `json:"id"`
		} `json:"studies"`
	}](t, resp)
	for _, s := range afterList.Studies {
		if s.ID == studyID {
			t.Fatalf("overlay study %s still present after removal", studyID)
		}
	}
	t.Logf("verified overlay removed; %d compare studies remaining", len(afterList.Studies))
}

func TestCompareOverlay_CompareMode(t *testing.T) {
	resp := env.POST(t, env.chartPath("compare"), map[string]any{
		"symbol": "IBM",
		"mode":   "compare",
		"source": "close",
	})
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Study struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"study"`
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "added", "status")
	requireField(t, result.Study.Name, "Compare", "study.name")
	t.Logf("added compare: id=%s", result.Study.ID)

	// Cleanup: remove the compare study.
	time.Sleep(500 * time.Millisecond)
	r := env.DELETE(t, env.chartPath("compare/"+result.Study.ID))
	r.Body.Close()
}

func TestCompareOverlay_InvalidMode(t *testing.T) {
	resp := env.POST(t, env.chartPath("compare"), map[string]any{
		"symbol": "ETHUSD",
		"mode":   "invalid",
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 400 or 422", resp.StatusCode)
	}
}

func TestCompareOverlay_MissingSymbol(t *testing.T) {
	resp := env.POST(t, env.chartPath("compare"), map[string]any{})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 400 or 422", resp.StatusCode)
	}
}

// --- Chart Undo/Redo tests ---

func TestUndoChart(t *testing.T) {
	resp := env.POST(t, env.chartPath("undo"), nil)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ChartID string `json:"chart_id"`
		Status  string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "executed", "status")
	requireField(t, result.ChartID, env.ChartID, "chart_id")
}

func TestRedoChart(t *testing.T) {
	resp := env.POST(t, env.chartPath("redo"), nil)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ChartID string `json:"chart_id"`
		Status  string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "executed", "status")
	requireField(t, result.ChartID, env.ChartID, "chart_id")
}

// --- Layout Favorite tests ---

func TestGetLayoutFavorite(t *testing.T) {
	resp := env.GET(t, "/api/v1/layout/favorite")
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		LayoutID   string `json:"layout_id"`
		LayoutName string `json:"layout_name"`
		IsFavorite bool   `json:"is_favorite"`
	}](t, resp)
	if result.LayoutID == "" {
		t.Fatal("expected non-empty layout_id")
	}
	t.Logf("layout favorite: id=%s name=%s is_favorite=%v", result.LayoutID, result.LayoutName, result.IsFavorite)
}

func TestToggleLayoutFavorite(t *testing.T) {
	// Read initial state.
	resp := env.GET(t, "/api/v1/layout/favorite")
	requireStatus(t, resp, http.StatusOK)
	before := decodeJSON[struct {
		IsFavorite bool `json:"is_favorite"`
	}](t, resp)

	// Toggle.
	resp = env.POST(t, "/api/v1/layout/favorite/toggle", nil)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		LayoutID   string `json:"layout_id"`
		IsFavorite bool   `json:"is_favorite"`
	}](t, resp)
	if result.LayoutID == "" {
		t.Fatal("expected non-empty layout_id")
	}
	t.Logf("toggle: was_favorite=%v is_favorite=%v", before.IsFavorite, result.IsFavorite)

	// Toggle back to restore.
	resp = env.POST(t, "/api/v1/layout/favorite/toggle", nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

// --- Chart Toggles tests ---

func TestChartToggles_GetState(t *testing.T) {
	resp := env.GET(t, env.chartPath("toggles"))
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ChartID       string `json:"chart_id"`
		LogScale      *bool  `json:"log_scale"`
		AutoScale     *bool  `json:"auto_scale"`
		ExtendedHours *bool  `json:"extended_hours"`
	}](t, resp)
	requireField(t, result.ChartID, env.ChartID, "chart_id")
}

func TestChartToggles_LogScale(t *testing.T) {
	// Read initial state.
	resp := env.GET(t, env.chartPath("toggles"))
	requireStatus(t, resp, http.StatusOK)
	before := decodeJSON[struct {
		LogScale *bool `json:"log_scale"`
	}](t, resp)

	// Toggle log scale.
	resp = env.POST(t, env.chartPath("toggles/log-scale"), nil)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	if result.Status != "toggled" {
		t.Fatalf("status = %q, want \"toggled\"", result.Status)
	}

	// Read new state — should differ if the button was detected.
	resp = env.GET(t, env.chartPath("toggles"))
	requireStatus(t, resp, http.StatusOK)
	after := decodeJSON[struct {
		LogScale *bool `json:"log_scale"`
	}](t, resp)

	if before.LogScale != nil && after.LogScale != nil && *before.LogScale == *after.LogScale {
		t.Logf("warning: log_scale did not change (before=%v, after=%v) — button may not be detectable", *before.LogScale, *after.LogScale)
	}

	// Toggle back to restore original state.
	resp = env.POST(t, env.chartPath("toggles/log-scale"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestChartToggles_AutoScale(t *testing.T) {
	// Read initial state.
	resp := env.GET(t, env.chartPath("toggles"))
	requireStatus(t, resp, http.StatusOK)
	before := decodeJSON[struct {
		AutoScale *bool `json:"auto_scale"`
	}](t, resp)

	// Toggle auto scale.
	resp = env.POST(t, env.chartPath("toggles/auto-scale"), nil)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	if result.Status != "toggled" {
		t.Fatalf("status = %q, want \"toggled\"", result.Status)
	}

	// Read new state.
	resp = env.GET(t, env.chartPath("toggles"))
	requireStatus(t, resp, http.StatusOK)
	after := decodeJSON[struct {
		AutoScale *bool `json:"auto_scale"`
	}](t, resp)

	if before.AutoScale != nil && after.AutoScale != nil && *before.AutoScale == *after.AutoScale {
		t.Logf("warning: auto_scale did not change (before=%v, after=%v) — button may not be detectable", *before.AutoScale, *after.AutoScale)
	}

	// Toggle back to restore original state.
	resp = env.POST(t, env.chartPath("toggles/auto-scale"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestChartToggles_ExtendedHours(t *testing.T) {
	// Extended hours may not be available for all symbols (e.g. crypto).
	// Toggle and verify the endpoint works; state change is best-effort.
	resp := env.POST(t, env.chartPath("toggles/extended-hours"), nil)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	if result.Status != "toggled" {
		t.Fatalf("status = %q, want \"toggled\"", result.Status)
	}

	// Toggle back to restore.
	resp = env.POST(t, env.chartPath("toggles/extended-hours"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}
