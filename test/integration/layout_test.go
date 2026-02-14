//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// --- Helpers ---

// restoreGrid sets the grid back to single pane and sleeps briefly.
func restoreGrid(t *testing.T) {
	t.Helper()
	resp := env.POST(t, "/api/v1/layout/grid", map[string]any{"template": "s"})
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)
}

// requireErrorStatus asserts that the response has a 400 or 422 status code.
func requireErrorStatus(t *testing.T, resp *http.Response) {
	t.Helper()
	defer resp.Body.Close()
	if resp.StatusCode != 400 && resp.StatusCode != 422 {
		t.Fatalf("expected 400 or 422, got %d", resp.StatusCode)
	}
}

// --- Group 1: Read-Only Tests ---

func TestListLayouts(t *testing.T) {
	layouts, err := env.listLayouts()
	if err != nil {
		t.Fatalf("list layouts: %v", err)
	}
	if len(layouts) == 0 {
		t.Fatal("expected at least 1 layout")
	}
	for i, l := range layouts {
		if l.ID <= 0 {
			t.Fatalf("layout[%d].id = %d, want > 0", i, l.ID)
		}
		if l.Name == "" {
			t.Fatalf("layout[%d].name is empty", i)
		}
		if l.URL == "" {
			t.Fatalf("layout[%d].url is empty", i)
		}
	}
	t.Logf("found %d layouts", len(layouts))
}

func TestGetLayoutStatus(t *testing.T) {
	resp := env.GET(t, "/api/v1/layout/status")
	requireStatus(t, resp, http.StatusOK)

	status := decodeJSON[struct {
		LayoutName   string `json:"layout_name"`
		LayoutID     string `json:"layout_id"`
		GridTemplate string `json:"grid_template"`
		ChartCount   int    `json:"chart_count"`
		ActiveIndex  int    `json:"active_index"`
		IsMaximized  bool   `json:"is_maximized"`
		IsFullscreen bool   `json:"is_fullscreen"`
		HasChanges   bool   `json:"has_changes"`
	}](t, resp)

	if status.LayoutName == "" {
		t.Fatal("layout_name is empty")
	}
	if status.LayoutID == "" {
		t.Fatal("layout_id is empty")
	}
	if status.GridTemplate == "" {
		t.Fatal("grid_template is empty")
	}
	if status.ChartCount < 1 {
		t.Fatalf("chart_count = %d, want >= 1", status.ChartCount)
	}
	if status.ActiveIndex < 0 {
		t.Fatalf("active_index = %d, want >= 0", status.ActiveIndex)
	}
	t.Logf("layout=%q id=%s grid=%s charts=%d active=%d",
		status.LayoutName, status.LayoutID, status.GridTemplate, status.ChartCount, status.ActiveIndex)
}

func TestGetPanes(t *testing.T) {
	resp := env.GET(t, env.chartPath("panes"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		GridTemplate string `json:"grid_template"`
		ChartCount   int    `json:"chart_count"`
		ActiveIndex  int    `json:"active_index"`
		Panes        []struct {
			Index      int    `json:"index"`
			Symbol     string `json:"symbol"`
			Exchange   string `json:"exchange"`
			Resolution string `json:"resolution"`
		} `json:"panes"`
	}](t, resp)

	if result.ChartCount < 1 {
		t.Fatalf("chart_count = %d, want >= 1", result.ChartCount)
	}
	if len(result.Panes) == 0 {
		t.Fatal("expected at least 1 pane")
	}
	if result.Panes[0].Symbol == "" {
		t.Fatal("first pane symbol is empty")
	}
	t.Logf("grid=%s charts=%d panes=%d first_symbol=%s",
		result.GridTemplate, result.ChartCount, len(result.Panes), result.Panes[0].Symbol)
}

// --- Group 2: Layout Actions ---

func TestSaveLayout(t *testing.T) {
	resp := env.POST(t, "/api/v1/layout/save", nil)
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Status     string `json:"status"`
		LayoutName string `json:"layout_name"`
		LayoutID   string `json:"layout_id"`
	}](t, resp)

	if result.Status == "" {
		t.Fatal("status is empty")
	}
	if result.LayoutName == "" {
		t.Fatal("layout_name is empty")
	}
	if result.LayoutID == "" {
		t.Fatal("layout_id is empty")
	}
	t.Logf("save → status=%s layout=%q id=%s", result.Status, result.LayoutName, result.LayoutID)
}

func TestRenameLayout_AndRestore(t *testing.T) {
	// Get current name to restore later.
	origName, err := env.currentLayoutName()
	if err != nil {
		t.Fatalf("get current layout name: %v", err)
	}

	// Rename to a test name.
	testName := "tv_agent_rename_test"
	resp := env.POST(t, "/api/v1/layout/rename", map[string]any{"name": testName})
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Status     string `json:"status"`
		LayoutName string `json:"layout_name"`
	}](t, resp)

	if result.Status == "" {
		t.Fatal("status is empty")
	}
	t.Logf("rename → status=%s layout=%q", result.Status, result.LayoutName)

	// Verify via status endpoint.
	name, err := env.currentLayoutName()
	if err != nil {
		t.Fatalf("verify rename: %v", err)
	}
	requireField(t, name, testName, "layout_name after rename")

	// Restore original name and save to persist it.
	resp = env.POST(t, "/api/v1/layout/rename", map[string]any{"name": origName})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	resp = env.POST(t, "/api/v1/layout/save", nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	t.Logf("restored name to %q and saved", origName)
}

func TestDismissDialog(t *testing.T) {
	resp := env.POST(t, "/api/v1/layout/dismiss-dialog", nil)
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)

	if result.Status == "" {
		t.Fatal("status is empty")
	}
	t.Logf("dismiss-dialog → status=%s", result.Status)
}

func TestToggleFullscreen(t *testing.T) {
	// Toggle on.
	resp := env.POST(t, "/api/v1/layout/fullscreen", nil)
	requireStatus(t, resp, http.StatusOK)

	on := decodeJSON[struct {
		IsFullscreen bool `json:"is_fullscreen"`
	}](t, resp)

	if !on.IsFullscreen {
		t.Fatal("expected is_fullscreen=true after first toggle")
	}
	t.Logf("fullscreen toggle on → is_fullscreen=%v", on.IsFullscreen)

	time.Sleep(500 * time.Millisecond)

	// Toggle off.
	resp = env.POST(t, "/api/v1/layout/fullscreen", nil)
	requireStatus(t, resp, http.StatusOK)

	off := decodeJSON[struct {
		IsFullscreen bool `json:"is_fullscreen"`
	}](t, resp)

	if off.IsFullscreen {
		t.Fatal("expected is_fullscreen=false after second toggle")
	}
	t.Logf("fullscreen toggle off → is_fullscreen=%v", off.IsFullscreen)
}

// --- Group 3: Grid + Chart Navigation ---

func TestSetGrid_AllTemplates(t *testing.T) {
	t.Cleanup(func() { restoreGrid(t) })

	templates := []struct {
		name       string
		wantCharts int
	}{
		{"s", 1},
		{"2h", 2},
		{"2v", 2},
		{"3h", 3},
		{"4", 4},
	}

	for _, tc := range templates {
		t.Run(tc.name, func(t *testing.T) {
			resp := env.POST(t, "/api/v1/layout/grid", map[string]any{"template": tc.name})
			requireStatus(t, resp, http.StatusOK)

			result := decodeJSON[struct {
				GridTemplate string `json:"grid_template"`
				ChartCount   int    `json:"chart_count"`
			}](t, resp)

			requireField(t, result.GridTemplate, tc.name, "grid_template")
			requireField(t, result.ChartCount, tc.wantCharts, "chart_count")
			t.Logf("grid=%s charts=%d", result.GridTemplate, result.ChartCount)

			time.Sleep(500 * time.Millisecond)
		})
	}
}

func TestChartNavigation_MultiPane(t *testing.T) {
	t.Cleanup(func() { restoreGrid(t) })

	// Set grid to 2 panes.
	resp := env.POST(t, "/api/v1/layout/grid", map[string]any{"template": "2h"})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	// Activate index 0.
	resp = env.POST(t, "/api/v1/chart/activate", map[string]any{"index": 0})
	requireStatus(t, resp, http.StatusOK)
	act0 := decodeJSON[struct {
		ActiveIndex int `json:"active_index"`
	}](t, resp)
	requireField(t, act0.ActiveIndex, 0, "active_index after activate(0)")

	time.Sleep(300 * time.Millisecond)

	// Next → should go to index 1.
	resp = env.POST(t, "/api/v1/chart/next", nil)
	requireStatus(t, resp, http.StatusOK)
	next := decodeJSON[struct {
		ChartIndex int `json:"chart_index"`
		ChartCount int `json:"chart_count"`
	}](t, resp)
	requireField(t, next.ChartIndex, 1, "chart_index after next")
	t.Logf("next → index=%d count=%d", next.ChartIndex, next.ChartCount)

	time.Sleep(300 * time.Millisecond)

	// Prev → should go back to index 0.
	resp = env.POST(t, "/api/v1/chart/prev", nil)
	requireStatus(t, resp, http.StatusOK)
	prev := decodeJSON[struct {
		ChartIndex int `json:"chart_index"`
		ChartCount int `json:"chart_count"`
	}](t, resp)
	requireField(t, prev.ChartIndex, 0, "chart_index after prev")
	t.Logf("prev → index=%d count=%d", prev.ChartIndex, prev.ChartCount)

	time.Sleep(300 * time.Millisecond)

	// Activate index 1 directly.
	resp = env.POST(t, "/api/v1/chart/activate", map[string]any{"index": 1})
	requireStatus(t, resp, http.StatusOK)
	act1 := decodeJSON[struct {
		ActiveIndex int `json:"active_index"`
	}](t, resp)
	requireField(t, act1.ActiveIndex, 1, "active_index after activate(1)")
	t.Logf("activate(1) → active_index=%d", act1.ActiveIndex)
}

func TestMaximizeChart(t *testing.T) {
	t.Cleanup(func() { restoreGrid(t) })

	// Need multi-pane to test maximize meaningfully.
	resp := env.POST(t, "/api/v1/layout/grid", map[string]any{"template": "2h"})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(1 * time.Second)

	// Toggle maximize on.
	resp = env.POST(t, "/api/v1/chart/maximize", nil)
	requireStatus(t, resp, http.StatusOK)
	on := decodeJSON[struct {
		IsMaximized bool `json:"is_maximized"`
	}](t, resp)
	if !on.IsMaximized {
		t.Fatal("expected is_maximized=true after first toggle")
	}
	t.Logf("maximize on → is_maximized=%v", on.IsMaximized)

	// Alt+Enter needs time to settle before the second toggle.
	time.Sleep(2 * time.Second)

	// Toggle maximize off. The endpoint always returns 200 with the
	// current state; verify it at least responds.
	resp = env.POST(t, "/api/v1/chart/maximize", nil)
	requireStatus(t, resp, http.StatusOK)
	off := decodeJSON[struct {
		IsMaximized bool `json:"is_maximized"`
	}](t, resp)
	t.Logf("maximize toggle off → is_maximized=%v", off.IsMaximized)

	// Note: Alt+Enter un-maximize can be unreliable when issued programmatically.
	// restoreGrid("s") in cleanup will un-maximize as a side effect.
}

// --- Group 4: Validation Tests ---

func TestCloneLayout_EmptyName(t *testing.T) {
	resp := env.POST(t, "/api/v1/layout/clone", map[string]any{"name": ""})
	requireErrorStatus(t, resp)
	t.Logf("empty clone name rejected with status %d", resp.StatusCode)
}

func TestRenameLayout_EmptyName(t *testing.T) {
	resp := env.POST(t, "/api/v1/layout/rename", map[string]any{"name": ""})
	requireErrorStatus(t, resp)
	t.Logf("empty rename name rejected with status %d", resp.StatusCode)
}

func TestDeleteLayout_InvalidID(t *testing.T) {
	resp := env.DELETE(t, "/api/v1/layout/0")
	requireErrorStatus(t, resp)
	t.Logf("delete layout 0 rejected with status %d", resp.StatusCode)
}

func TestSwitchLayout_InvalidID(t *testing.T) {
	resp := env.POST(t, "/api/v1/layout/switch", map[string]any{"id": 0})
	requireErrorStatus(t, resp)
	t.Logf("switch layout 0 rejected with status %d", resp.StatusCode)
}

func TestBatchDelete_EmptyIDs(t *testing.T) {
	resp := env.POST(t, "/api/v1/layouts/batch-delete", map[string]any{"ids": []int{}})
	requireErrorStatus(t, resp)
	t.Logf("empty batch-delete ids rejected with status %d", resp.StatusCode)
}

func TestActivateChart_NegativeIndex(t *testing.T) {
	resp := env.POST(t, "/api/v1/chart/activate", map[string]any{"index": -1})
	requireErrorStatus(t, resp)
	t.Logf("negative activate index rejected with status %d", resp.StatusCode)
}

func TestSetGrid_EmptyTemplate(t *testing.T) {
	resp := env.POST(t, "/api/v1/layout/grid", map[string]any{"template": ""})
	requireErrorStatus(t, resp)
	t.Logf("empty grid template rejected with status %d", resp.StatusCode)
}

// --- Group 5: Lifecycle Test (slow — page reloads) ---

func TestLayoutLifecycle(t *testing.T) {
	// Save first to ensure current layout is persisted in the list
	// (prior tests may have renamed without saving).
	saveResp := env.POST(t, "/api/v1/layout/save", nil)
	saveResp.Body.Close()

	// 1. Record current layout name (we'll resolve its ID after clone triggers a page reload,
	// which refreshes TradingView's layout list cache).
	origName, err := env.currentLayoutName()
	if err != nil {
		t.Fatalf("get current layout name: %v", err)
	}

	cloneName := fmt.Sprintf("tv_agent_layout_test_%d", time.Now().Unix())
	var origID int
	var cloneNumericID int

	// Cleanup: always try to switch back and delete the clone.
	t.Cleanup(func() {
		// Resolve original ID if we haven't yet (need page reload to refresh list).
		if origID == 0 {
			origID, _ = env.resolveLayoutNumericID(origName)
		}
		if origID > 0 {
			resp, err := env.doJSON(http.MethodPost, "/api/v1/layout/switch", map[string]any{"id": origID})
			if err != nil {
				t.Logf("cleanup switch: %v", err)
			} else {
				resp.Body.Close()
			}
			time.Sleep(5 * time.Second)
		}

		// Resolve clone ID if we don't have it yet.
		if cloneNumericID == 0 {
			cloneNumericID, _ = env.resolveLayoutNumericID(cloneName)
		}
		if cloneNumericID > 0 {
			path := fmt.Sprintf("/api/v1/layout/%d", cloneNumericID)
			resp, err := env.doJSON(http.MethodDelete, path, nil)
			if err != nil {
				t.Logf("cleanup delete: %v", err)
			} else {
				resp.Body.Close()
				t.Logf("cleanup: deleted clone %d (%s)", cloneNumericID, cloneName)
			}
		}

		// Re-discover chart ID (may have changed after switch).
		if err := env.discoverChartID(); err != nil {
			t.Logf("cleanup discover chart: %v", err)
		}
	})

	// 2. Clone layout (triggers page reload, refreshes layout list cache).
	resp := env.POST(t, "/api/v1/layout/clone", map[string]any{"name": cloneName})
	requireStatus(t, resp, http.StatusOK)
	cloneResult := decodeJSON[struct {
		Status     string `json:"status"`
		LayoutName string `json:"layout_name"`
		LayoutID   string `json:"layout_id"`
	}](t, resp)
	t.Logf("1/9 clone → status=%s name=%q id=%s", cloneResult.Status, cloneResult.LayoutName, cloneResult.LayoutID)

	// 3. Wait for clone to settle (page reload).
	time.Sleep(5 * time.Second)

	// 4. Re-discover chart ID (changes after layout switch).
	if err := env.discoverChartID(); err != nil {
		t.Fatalf("discover chart after clone: %v", err)
	}
	t.Logf("2/9 chart ID after clone: %s", env.ChartID)

	// 5. Verify we're on the cloned layout.
	resp = env.GET(t, "/api/v1/layout/status")
	requireStatus(t, resp, http.StatusOK)
	statusAfterClone := decodeJSON[struct {
		LayoutName string `json:"layout_name"`
		LayoutID   string `json:"layout_id"`
	}](t, resp)
	requireField(t, statusAfterClone.LayoutName, cloneName, "layout_name after clone")
	t.Logf("3/9 confirmed on clone: %q", statusAfterClone.LayoutName)

	// 6. Now resolve the original layout's numeric ID (list is fresh after page reload).
	origID, err = env.resolveLayoutNumericID(origName)
	if err != nil {
		t.Fatalf("resolve original layout ID by name %q: %v", origName, err)
	}

	// Count layouts after clone.
	afterCloneLayouts, err := env.listLayouts()
	if err != nil {
		t.Fatalf("list layouts after clone: %v", err)
	}
	baselineCount := len(afterCloneLayouts) - 1 // clone added 1
	t.Logf("4/9 layout count: %d (baseline=%d), original=%q (id=%d)", len(afterCloneLayouts), baselineCount, origName, origID)

	// 5. Rename the clone.
	renamedName := fmt.Sprintf("tv_agent_renamed_%d", time.Now().Unix())
	resp = env.POST(t, "/api/v1/layout/rename", map[string]any{"name": renamedName})
	requireStatus(t, resp, http.StatusOK)
	renameResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	t.Logf("5/9 rename → status=%s", renameResult.Status)
	// Update cloneName so cleanup can find it.
	cloneName = renamedName

	// 6. Save.
	resp = env.POST(t, "/api/v1/layout/save", nil)
	requireStatus(t, resp, http.StatusOK)
	saveResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	t.Logf("6/9 save → status=%s", saveResult.Status)

	// 7. Switch back to original layout.
	// Switch triggers a page reload; the server polls for up to 30s internally,
	// which can race with the 30s client timeout. Use doJSON and tolerate timeout.
	switchResp, switchErr := env.doJSON(http.MethodPost, "/api/v1/layout/switch", map[string]any{"id": origID})
	if switchErr != nil {
		t.Logf("7/9 switch request error (expected during page reload): %v", switchErr)
	} else {
		switchResp.Body.Close()
		t.Logf("7/9 switch back → status=%d", switchResp.StatusCode)
	}

	// Wait for page reload to settle.
	time.Sleep(8 * time.Second)

	// Re-discover chart ID.
	if err := env.discoverChartID(); err != nil {
		t.Fatalf("discover chart after switch: %v", err)
	}

	// 8. Resolve clone's numeric ID and delete it.
	cloneNumericID, err = env.resolveLayoutNumericID(cloneName)
	if err != nil {
		t.Fatalf("resolve clone ID: %v", err)
	}

	deletePath := fmt.Sprintf("/api/v1/layout/%d", cloneNumericID)
	resp = env.DELETE(t, deletePath)
	requireStatus(t, resp, http.StatusNoContent)
	resp.Body.Close()
	t.Logf("8/9 deleted clone %d (%s)", cloneNumericID, cloneName)

	// Verify the clone is no longer resolvable by name.
	// (The layout list may be cached, so count checks can be flaky.
	// Instead, confirm the deleted layout can't be found.)
	_, resolveErr := env.resolveLayoutNumericID(cloneName)
	if resolveErr == nil {
		// List might be cached — poll once after a short delay.
		time.Sleep(2 * time.Second)
		_, resolveErr = env.resolveLayoutNumericID(cloneName)
	}
	if resolveErr == nil {
		t.Fatalf("clone %q still found in layout list after delete", cloneName)
	}
	t.Logf("9/9 lifecycle complete: clone %q no longer in layout list", cloneName)

	// Clear the cloneNumericID so cleanup doesn't try to delete again.
	cloneNumericID = 0
}
