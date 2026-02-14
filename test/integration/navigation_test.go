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
