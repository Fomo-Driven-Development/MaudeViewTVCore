//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"
)

// --- Drawing test helpers ---

// clearDrawings removes all drawings from the test chart.
func clearDrawings(t *testing.T) {
	t.Helper()
	resp := env.DELETE(t, env.chartPath("drawings"))
	requireStatus(t, resp, http.StatusNoContent)
	resp.Body.Close()
}

// drawingPoint returns a point map at a fractional offset into the visible range.
// offset 0.0 = range start, 1.0 = range end.
func drawingPoint(t *testing.T, offset float64) map[string]any {
	t.Helper()
	resp := env.GET(t, env.chartPath("visible-range"))
	requireStatus(t, resp, http.StatusOK)
	vr := decodeJSON[struct {
		From float64 `json:"from"`
		To   float64 `json:"to"`
	}](t, resp)

	ts := vr.From + (vr.To-vr.From)*offset
	return map[string]any{"time": ts, "price": 100.0}
}

// makePoints returns N distinct points spread across the visible range.
func makePoints(t *testing.T, n int) []map[string]any {
	t.Helper()
	pts := make([]map[string]any, n)
	for i := range n {
		frac := float64(i+1) / float64(n+1)
		pts[i] = drawingPoint(t, frac)
	}
	return pts
}

// createSinglePoint creates a single-point drawing and asserts success.
func createSinglePoint(t *testing.T, shape string) {
	t.Helper()
	pt := drawingPoint(t, 0.5)
	body := map[string]any{
		"point":   pt,
		"options": map[string]any{"shape": shape},
	}
	resp := env.POST(t, env.chartPath("drawings"), body)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}](t, resp)
	if result.ID == "" {
		t.Fatalf("%s: id is empty", shape)
	}
	requireField(t, result.Status, "created", "status")
}

// createMultipoint creates a multi-point drawing and asserts success.
func createMultipoint(t *testing.T, shape string, numPoints int) {
	t.Helper()
	pts := makePoints(t, numPoints)
	body := map[string]any{
		"points":  pts,
		"options": map[string]any{"shape": shape},
	}
	resp := env.POST(t, env.chartPath("drawings/multipoint"), body)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}](t, resp)
	if result.ID == "" {
		t.Fatalf("%s: id is empty", shape)
	}
	requireField(t, result.Status, "created", "status")
}

// listDrawingCount returns the number of drawings currently on the chart.
func listDrawingCount(t *testing.T) int {
	t.Helper()
	resp := env.GET(t, env.chartPath("drawings"))
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Shapes []struct {
			ID string `json:"id"`
		} `json:"shapes"`
	}](t, resp)
	return len(result.Shapes)
}

// --- Line shape tests (9 shapes) ---

func TestLines(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	singlePoint := []string{"horizontal_line", "horizontal_ray", "vertical_line", "cross_line"}
	multiPoint := []struct {
		shape  string
		points int
	}{
		{"trend_line", 2},
		{"ray", 2},
		{"info_line", 2},
		{"extended", 2},
		{"trend_angle", 2},
	}

	for _, shape := range singlePoint {
		t.Run(shape, func(t *testing.T) {
			createSinglePoint(t, shape)
		})
	}
	for _, tc := range multiPoint {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	wantCount := len(singlePoint) + len(multiPoint)
	if count < wantCount {
		t.Fatalf("drawing count = %d, want >= %d", count, wantCount)
	}
}

// --- Channel shape tests (4 shapes) ---

func TestChannels(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []struct {
		shape  string
		points int
	}{
		{"parallel_channel", 3},
		{"regression_trend", 2},
		{"flat_bottom", 3},
		{"disjoint_angle", 3},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Pitchfork shape tests (4 shapes) ---

func TestPitchforks(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []struct {
		shape  string
		points int
	}{
		{"pitchfork", 3},
		{"schiff_pitchfork", 3},
		{"schiff_pitchfork_modified", 3},
		{"inside_pitchfork", 3},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Fibonacci shape tests (11 shapes) ---

func TestFibonacci(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []struct {
		shape  string
		points int
	}{
		{"fib_retracement", 2},
		{"fib_trend_ext", 3},
		{"fib_channel", 3},
		{"fib_timezone", 2},
		{"fib_speed_resist_fan", 2},
		{"fib_trend_time", 3},
		{"fib_circles", 2},
		{"fib_spiral", 2},
		{"fib_speed_resist_arcs", 2},
		{"fib_wedge", 3},
		{"pitchfan", 3},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Gann shape tests (4 shapes) ---

func TestGann(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []struct {
		shape  string
		points int
	}{
		{"gannbox_square", 2},
		{"gannbox_fixed", 2},
		{"gannbox", 2},
		{"gannbox_fan", 2},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Pattern shape tests (6 shapes) ---

func TestPatterns(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []struct {
		shape  string
		points int
	}{
		{"xabcd_pattern", 5},
		{"cypher_pattern", 5},
		{"head_and_shoulders", 7},
		{"abcd_pattern", 4},
		{"triangle_pattern", 4},
		{"3divers_pattern", 7},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Elliott Wave shape tests (5 shapes) ---

func TestElliottWaves(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []struct {
		shape  string
		points int
	}{
		{"elliott_impulse_wave", 6},
		{"elliott_correction", 4},
		{"elliott_triangle_wave", 6},
		{"elliott_double_combo", 4},
		{"elliott_triple_combo", 6},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Cycle shape tests (3 shapes) ---

func TestCycles(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []struct {
		shape  string
		points int
	}{
		{"cyclic_lines", 2},
		{"time_cycles", 2},
		{"sine_line", 2},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Projection shape tests (6 shapes) ---

func TestProjection(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []struct {
		shape  string
		points int
	}{
		{"long_position", 2},
		{"short_position", 2},
		{"forecast", 2},
		{"bars_pattern", 2},
		{"ghost_feed", 5},
		{"projection", 3},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Volume-based shape tests (3 shapes) ---

func TestVolumeBased(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	singlePoint := []string{"anchored_vwap", "anchored_volume_profile"}
	multiPoint := []struct {
		shape  string
		points int
	}{
		{"fixed_range_volume_profile", 2},
	}

	for _, shape := range singlePoint {
		t.Run(shape, func(t *testing.T) {
			createSinglePoint(t, shape)
		})
	}
	for _, tc := range multiPoint {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	wantCount := len(singlePoint) + len(multiPoint)
	if count < wantCount {
		t.Fatalf("drawing count = %d, want >= %d", count, wantCount)
	}
}

// --- Measurer shape tests (3 shapes) ---

func TestMeasurer(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []struct {
		shape  string
		points int
	}{
		{"price_range", 2},
		{"date_range", 2},
		{"date_and_price_range", 2},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Brush shape tests (2 shapes) ---

func TestBrushes(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	// Variable-point shapes (Points=-1); use 3 points for testing.
	shapes := []struct {
		shape  string
		points int
	}{
		{"brush", 3},
		{"highlighter", 3},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Arrow shape tests (4 shapes) ---

func TestArrows(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	singlePoint := []string{"arrow_up", "arrow_down"}
	multiPoint := []struct {
		shape  string
		points int
	}{
		{"arrow_marker", 2},
		{"arrow", 2},
	}

	for _, shape := range singlePoint {
		t.Run(shape, func(t *testing.T) {
			createSinglePoint(t, shape)
		})
	}
	for _, tc := range multiPoint {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	wantCount := len(singlePoint) + len(multiPoint)
	if count < wantCount {
		t.Fatalf("drawing count = %d, want >= %d", count, wantCount)
	}
}

// --- Shape shape tests (10 shapes) ---

func TestShapes(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []struct {
		shape  string
		points int
	}{
		{"rectangle", 2},
		{"rotated_rectangle", 3},
		{"path", 3},
		{"circle", 2},
		{"ellipse", 3},
		{"polyline", 3},
		{"triangle", 3},
		{"arc", 3},
		{"curve", 2},
		{"double_curve", 2},
	}

	for _, tc := range shapes {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Text & Notes shape tests (13 shapes) ---

func TestTextAndNotes(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	singlePoint := []string{
		"text", "anchored_text", "note", "anchored_note",
		"balloon", "table", "comment", "price_label", "signpost", "flag",
	}
	multiPoint := []struct {
		shape  string
		points int
	}{
		{"text_note", 2},
		{"price_note", 2},
		{"callout", 2},
	}

	for _, shape := range singlePoint {
		t.Run(shape, func(t *testing.T) {
			createSinglePoint(t, shape)
		})
	}
	for _, tc := range multiPoint {
		t.Run(tc.shape, func(t *testing.T) {
			createMultipoint(t, tc.shape, tc.points)
		})
	}

	count := listDrawingCount(t)
	wantCount := len(singlePoint) + len(multiPoint)
	if count < wantCount {
		t.Fatalf("drawing count = %d, want >= %d", count, wantCount)
	}
}

// --- Content shape tests (3 shapes) ---

func TestContent(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []string{"image", "tweet", "idea"}

	for _, shape := range shapes {
		t.Run(shape, func(t *testing.T) {
			createSinglePoint(t, shape)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Tweet drawing tests (dedicated endpoint) ---

func TestTweetDrawing(t *testing.T) {
	t.Skip("skipped by default: TradingView backend rate-limits tweet data fetches; run with -run TestTweetDrawing")
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	body := map[string]any{
		"tweet_url": "https://x.com/DGNSREKT/status/2023511775363174708",
	}
	resp := env.POST(t, env.chartPath("drawings/tweet"), body)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ChartID string `json:"chart_id"`
		ID      string `json:"id"`
		Status  string `json:"status"`
		URL     string `json:"url"`
	}](t, resp)
	if result.ID == "" {
		t.Fatal("tweet drawing: id is empty")
	}
	requireField(t, result.Status, "created", "status")
	t.Logf("tweet drawing created: id=%s url=%s", result.ID, result.URL)

	count := listDrawingCount(t)
	if count < 1 {
		t.Fatalf("expected at least 1 drawing after tweet creation, got %d", count)
	}
}

func TestTweetDrawing_DeleteByID(t *testing.T) {
	t.Skip("skipped by default: TradingView backend rate-limits tweet data fetches; run with -run TestTweetDrawing_DeleteByID")
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	// Brief pause to avoid TradingView backend rate-limiting on tweet data fetch.
	time.Sleep(2 * time.Second)

	// Create a tweet drawing.
	body := map[string]any{
		"tweet_url": "https://x.com/DGNSREKT/status/2023511775363174708",
	}
	resp := env.POST(t, env.chartPath("drawings/tweet"), body)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ID string `json:"id"`
	}](t, resp)
	if result.ID == "" {
		t.Fatal("tweet drawing: id is empty")
	}

	// Verify it exists.
	before := listDrawingCount(t)
	if before < 1 {
		t.Fatalf("expected at least 1 drawing, got %d", before)
	}

	// Delete by ID.
	resp = env.DELETE(t, env.chartPath("drawings/"+result.ID))
	requireStatus(t, resp, http.StatusNoContent)
	resp.Body.Close()

	// Verify it was removed.
	after := listDrawingCount(t)
	if after >= before {
		t.Fatalf("expected drawing count to decrease, before=%d after=%d", before, after)
	}
	t.Logf("tweet drawing %s deleted, count %d -> %d", result.ID, before, after)
}

func TestTweetDrawing_Validation(t *testing.T) {
	t.Run("empty_url", func(t *testing.T) {
		body := map[string]any{"tweet_url": ""}
		resp := env.POST(t, env.chartPath("drawings/tweet"), body)
		defer resp.Body.Close()
		requireStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("invalid_domain", func(t *testing.T) {
		body := map[string]any{"tweet_url": "https://example.com/not-a-tweet"}
		resp := env.POST(t, env.chartPath("drawings/tweet"), body)
		defer resp.Body.Close()
		requireStatus(t, resp, http.StatusBadRequest)
	})
}

// --- Icon shape tests (3 shapes) ---

func TestIcons(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	shapes := []string{"icon", "emoji", "sticker"}

	for _, shape := range shapes {
		t.Run(shape, func(t *testing.T) {
			createSinglePoint(t, shape)
		})
	}

	count := listDrawingCount(t)
	if count < len(shapes) {
		t.Fatalf("drawing count = %d, want >= %d", count, len(shapes))
	}
}

// --- Tool activation tests ---

func TestToolActivation(t *testing.T) {
	// Ensure we return to cursor at the end.
	t.Cleanup(func() {
		resp := env.POST(t, env.chartPath("tools/cursor"), nil)
		resp.Body.Close()
	})

	tools := []string{"measure", "zoom", "eraser", "cursor"}
	for _, tool := range tools {
		t.Run(tool, func(t *testing.T) {
			// Activate the tool.
			resp := env.POST(t, env.chartPath("tools/"+tool), nil)
			requireStatus(t, resp, http.StatusOK)
			result := decodeJSON[struct {
				ChartID string `json:"chart_id"`
				Status  string `json:"status"`
			}](t, resp)
			requireField(t, result.Status, "activated", "status")

			// Verify via GET drawing tool.
			resp = env.GET(t, env.chartPath("drawings/tool"))
			requireStatus(t, resp, http.StatusOK)
			toolResult := decodeJSON[struct {
				Tool string `json:"tool"`
			}](t, resp)
			requireField(t, toolResult.Tool, tool, "tool")
		})
	}
}

// --- Validation tests ---

func TestDrawings_Validation(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })

	t.Run("wrong_point_count", func(t *testing.T) {
		pts := makePoints(t, 2)
		body := map[string]any{
			"points":  pts,
			"options": map[string]any{"shape": "pitchfork"},
		}
		resp := env.POST(t, env.chartPath("drawings/multipoint"), body)
		defer resp.Body.Close()
		requireStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("wrong_endpoint", func(t *testing.T) {
		pt := drawingPoint(t, 0.5)
		body := map[string]any{
			"point":   pt,
			"options": map[string]any{"shape": "trend_line"},
		}
		resp := env.POST(t, env.chartPath("drawings"), body)
		defer resp.Body.Close()
		requireStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("missing_shape", func(t *testing.T) {
		pt := drawingPoint(t, 0.5)
		body := map[string]any{
			"point":   pt,
			"options": map[string]any{},
		}
		resp := env.POST(t, env.chartPath("drawings"), body)
		defer resp.Body.Close()
		requireStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("unknown_shape_passthrough", func(t *testing.T) {
		pt := drawingPoint(t, 0.5)
		body := map[string]any{
			"point":   pt,
			"options": map[string]any{"shape": "some_future_shape"},
		}
		resp := env.POST(t, env.chartPath("drawings"), body)
		requireStatus(t, resp, http.StatusOK)
		result := decodeJSON[struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}](t, resp)
		if result.ID == "" {
			t.Fatal("id is empty for unknown shape passthrough")
		}
		requireField(t, result.Status, "created", "status")
	})
}

// --- List, Get, Delete, Clone, Z-Order ---

func TestDrawings_ListAndGetByID(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	// Create a drawing to list and get.
	createSinglePoint(t, "horizontal_line")

	// List drawings.
	resp := env.GET(t, env.chartPath("drawings"))
	requireStatus(t, resp, http.StatusOK)
	listing := decodeJSON[struct {
		ChartID string `json:"chart_id"`
		Shapes  []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"shapes"`
	}](t, resp)

	requireField(t, listing.ChartID, env.ChartID, "chart_id")
	if len(listing.Shapes) == 0 {
		t.Fatal("expected at least 1 drawing in list")
	}
	shapeID := listing.Shapes[0].ID
	t.Logf("listed %d drawings, first id=%s", len(listing.Shapes), shapeID)

	// Get drawing by ID.
	resp = env.GET(t, env.chartPath("drawings/"+shapeID))
	requireStatus(t, resp, http.StatusOK)
	detail := decodeJSON[map[string]any](t, resp)
	if len(detail) == 0 {
		t.Fatal("expected drawing detail data")
	}
	t.Logf("get drawing %s: %d keys", shapeID, len(detail))
}

func TestDrawings_DeleteByID(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	// Create two drawings.
	createSinglePoint(t, "horizontal_line")
	createSinglePoint(t, "arrow_up")

	// Get the IDs.
	resp := env.GET(t, env.chartPath("drawings"))
	requireStatus(t, resp, http.StatusOK)
	listing := decodeJSON[struct {
		Shapes []struct {
			ID string `json:"id"`
		} `json:"shapes"`
	}](t, resp)
	if len(listing.Shapes) < 2 {
		t.Fatalf("expected at least 2 drawings, got %d", len(listing.Shapes))
	}
	targetID := listing.Shapes[0].ID

	// Delete one by ID.
	resp = env.DELETE(t, env.chartPath("drawings/"+targetID))
	requireStatus(t, resp, http.StatusNoContent)
	resp.Body.Close()

	// Verify count decreased.
	afterCount := listDrawingCount(t)
	if afterCount >= len(listing.Shapes) {
		t.Fatalf("expected drawing count to decrease, before=%d after=%d", len(listing.Shapes), afterCount)
	}
	t.Logf("deleted drawing %s, count %d → %d", targetID, len(listing.Shapes), afterCount)
}

func TestDrawings_Clone(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	// Create a drawing to clone.
	createSinglePoint(t, "flag")

	// Get its ID.
	resp := env.GET(t, env.chartPath("drawings"))
	requireStatus(t, resp, http.StatusOK)
	listing := decodeJSON[struct {
		Shapes []struct {
			ID string `json:"id"`
		} `json:"shapes"`
	}](t, resp)
	if len(listing.Shapes) == 0 {
		t.Fatal("expected at least 1 drawing")
	}
	originalID := listing.Shapes[0].ID
	beforeCount := len(listing.Shapes)

	// Clone it.
	resp = env.POST(t, env.chartPath("drawings/"+originalID+"/clone"), nil)
	requireStatus(t, resp, http.StatusOK)
	cloneResult := decodeJSON[struct {
		ChartID string `json:"chart_id"`
		ID      string `json:"id"`
		Status  string `json:"status"`
	}](t, resp)
	requireField(t, cloneResult.Status, "cloned", "status")
	if cloneResult.ID == "" {
		t.Fatal("expected cloned drawing ID")
	}
	t.Logf("cloned %s → %s", originalID, cloneResult.ID)

	// Verify count increased.
	afterCount := listDrawingCount(t)
	if afterCount <= beforeCount {
		t.Fatalf("expected drawing count to increase, before=%d after=%d", beforeCount, afterCount)
	}
}

func TestDrawings_ZOrder(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	// Create a drawing.
	createSinglePoint(t, "horizontal_line")

	resp := env.GET(t, env.chartPath("drawings"))
	requireStatus(t, resp, http.StatusOK)
	listing := decodeJSON[struct {
		Shapes []struct {
			ID string `json:"id"`
		} `json:"shapes"`
	}](t, resp)
	if len(listing.Shapes) == 0 {
		t.Fatal("expected at least 1 drawing")
	}
	shapeID := listing.Shapes[0].ID

	actions := []string{"bring_to_front", "send_to_back", "bring_forward", "send_backward"}
	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			resp := env.POST(t, env.chartPath("drawings/"+shapeID+"/z-order"), map[string]any{
				"action": action,
			})
			requireStatus(t, resp, http.StatusOK)
			result := decodeJSON[struct {
				ChartID string `json:"chart_id"`
				Status  string `json:"status"`
			}](t, resp)
			requireField(t, result.Status, "executed", "status")
		})
	}
}

// --- Toggles ---

func TestDrawings_GetToggles(t *testing.T) {
	resp := env.GET(t, env.chartPath("drawings/toggles"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		ChartID string         `json:"chart_id"`
		Toggles map[string]any `json:"toggles"`
	}](t, resp)

	requireField(t, result.ChartID, env.ChartID, "chart_id")
	if result.Toggles == nil {
		t.Fatal("expected toggles data")
	}
	t.Logf("drawing toggles: %v", result.Toggles)
}

func TestDrawings_HideShow(t *testing.T) {
	// Hide all drawings.
	resp := env.PUT(t, env.chartPath("drawings/toggles/hide"), map[string]any{
		"value": true,
	})
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "set", "status")

	// Show all drawings.
	resp = env.PUT(t, env.chartPath("drawings/toggles/hide"), map[string]any{
		"value": false,
	})
	requireStatus(t, resp, http.StatusOK)
	result = decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "set", "status")
	t.Log("hide/show toggle: OK")
}

func TestDrawings_LockUnlock(t *testing.T) {
	// Lock all drawings.
	resp := env.PUT(t, env.chartPath("drawings/toggles/lock"), map[string]any{
		"value": true,
	})
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "set", "status")

	// Unlock all drawings.
	resp = env.PUT(t, env.chartPath("drawings/toggles/lock"), map[string]any{
		"value": false,
	})
	requireStatus(t, resp, http.StatusOK)
	result = decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "set", "status")
	t.Log("lock/unlock toggle: OK")
}

func TestDrawings_Magnet(t *testing.T) {
	// Enable magnet.
	resp := env.PUT(t, env.chartPath("drawings/toggles/magnet"), map[string]any{
		"enabled": true,
	})
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "set", "status")

	// Disable magnet.
	resp = env.PUT(t, env.chartPath("drawings/toggles/magnet"), map[string]any{
		"enabled": false,
	})
	requireStatus(t, resp, http.StatusOK)
	result = decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "set", "status")
	t.Log("magnet toggle: OK")
}

// --- Drawing Visibility ---

func TestDrawings_Visibility(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	createSinglePoint(t, "horizontal_line")

	resp := env.GET(t, env.chartPath("drawings"))
	requireStatus(t, resp, http.StatusOK)
	listing := decodeJSON[struct {
		Shapes []struct {
			ID string `json:"id"`
		} `json:"shapes"`
	}](t, resp)
	if len(listing.Shapes) == 0 {
		t.Fatal("expected at least 1 drawing")
	}
	shapeID := listing.Shapes[0].ID

	// Hide individual drawing.
	resp = env.PUT(t, env.chartPath("drawings/"+shapeID+"/visibility"), map[string]any{
		"visible": false,
	})
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "set", "status")

	// Show it again.
	resp = env.PUT(t, env.chartPath("drawings/"+shapeID+"/visibility"), map[string]any{
		"visible": true,
	})
	requireStatus(t, resp, http.StatusOK)
	result = decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "set", "status")
	t.Logf("visibility toggle on %s: OK", shapeID)
}

// --- State Export/Import ---

func TestDrawings_StateExportImport(t *testing.T) {
	t.Cleanup(func() { clearDrawings(t) })
	clearDrawings(t)

	// Create some drawings.
	createSinglePoint(t, "horizontal_line")
	createSinglePoint(t, "arrow_up")

	// Export state.
	resp := env.GET(t, env.chartPath("drawings/state"))
	requireStatus(t, resp, http.StatusOK)
	exported := decodeJSON[struct {
		ChartID string `json:"chart_id"`
		State   any    `json:"state"`
	}](t, resp)
	requireField(t, exported.ChartID, env.ChartID, "chart_id")
	if exported.State == nil {
		t.Fatal("expected state data from export")
	}
	t.Logf("exported drawings state")

	// Clear drawings.
	clearDrawings(t)
	if listDrawingCount(t) != 0 {
		t.Fatal("expected 0 drawings after clear")
	}

	// Import state.
	resp = env.PUT(t, env.chartPath("drawings/state"), map[string]any{
		"state": exported.State,
	})
	requireStatus(t, resp, http.StatusOK)
	importResult := decodeJSON[struct {
		ChartID string `json:"chart_id"`
		Status  string `json:"status"`
	}](t, resp)
	requireField(t, importResult.Status, "imported", "status")
	t.Logf("imported drawings state → %s", importResult.Status)
}

// --- Discovery endpoint test ---

func TestDrawings_DiscoveryEndpoint(t *testing.T) {
	resp := env.GET(t, "/api/v1/drawings/shapes")
	requireStatus(t, resp, http.StatusOK)

	type shapeInfo struct {
		Name   string `json:"name"`
		Label  string `json:"label"`
		Points int    `json:"points"`
	}
	type shapeGroup struct {
		Category string      `json:"category"`
		Label    string      `json:"label"`
		Shapes   []shapeInfo `json:"shapes"`
	}
	result := decodeJSON[struct {
		Groups []shapeGroup `json:"groups"`
	}](t, resp)

	if len(result.Groups) != 17 {
		t.Fatalf("group count = %d, want 17", len(result.Groups))
	}

	wantGroups := []struct {
		category   string
		shapeCount int
	}{
		{"lines", 9},
		{"channels", 4},
		{"pitchforks", 4},
		{"fibonacci", 11},
		{"gann", 4},
		{"patterns", 6},
		{"elliott_waves", 5},
		{"cycles", 3},
		{"projection", 6},
		{"volume_based", 3},
		{"measurer", 3},
		{"brushes", 2},
		{"arrows", 4},
		{"shapes", 10},
		{"text_and_notes", 13},
		{"content", 3},
		{"icons", 3},
	}
	for i, wg := range wantGroups {
		g := result.Groups[i]
		requireField(t, g.Category, wg.category, "category")
		if len(g.Shapes) != wg.shapeCount {
			t.Fatalf("group %q: shape count = %d, want %d", g.Category, len(g.Shapes), wg.shapeCount)
		}
	}
}
