//go:build integration

package integration

import (
	"net/http"
	"testing"
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

	if len(result.Groups) != 3 {
		t.Fatalf("group count = %d, want 3", len(result.Groups))
	}

	wantGroups := []struct {
		category   string
		shapeCount int
	}{
		{"lines", 9},
		{"channels", 4},
		{"pitchforks", 4},
	}
	for i, wg := range wantGroups {
		g := result.Groups[i]
		requireField(t, g.Category, wg.category, "category")
		if len(g.Shapes) != wg.shapeCount {
			t.Fatalf("group %q: shape count = %d, want %d", g.Category, len(g.Shapes), wg.shapeCount)
		}
	}
}
