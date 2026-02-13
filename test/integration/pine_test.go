//go:build integration

package integration

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// pineState is the JSON shape returned by most Pine editor endpoints.
type pineState struct {
	Status       string `json:"status"`
	IsVisible    bool   `json:"is_visible"`
	MonacoReady  bool   `json:"monaco_ready"`
	ScriptName   string `json:"script_name"`
	ScriptSource string `json:"script_source"`
	SourceLength int    `json:"source_length"`
	LineCount    int    `json:"line_count"`
	MatchCount   int    `json:"match_count"`
}

// ensurePineOpen makes sure the Pine editor is visible and Monaco is ready.
// Retries the toggle once if it times out (first open can be slow).
func ensurePineOpen(t *testing.T) {
	t.Helper()
	resp := env.GET(t, "/api/v1/pine/status")
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	if st.IsVisible && st.MonacoReady {
		return
	}

	for attempt := range 2 {
		resp = env.POST(t, "/api/v1/pine/toggle", nil)
		if resp.StatusCode == http.StatusOK {
			toggled := decodeJSON[pineState](t, resp)
			if toggled.IsVisible {
				time.Sleep(1 * time.Second)
				return
			}
		}
		resp.Body.Close()
		if attempt == 0 {
			t.Logf("pine toggle attempt %d failed (status=%d), retrying...", attempt+1, resp.StatusCode)
			time.Sleep(2 * time.Second)
		}
	}
	t.Fatal("failed to open Pine editor after retries")
}

// ensurePineClosed makes sure the Pine editor is closed.
func ensurePineClosed(t *testing.T) {
	t.Helper()
	resp := env.GET(t, "/api/v1/pine/status")
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	if !st.IsVisible {
		return
	}
	resp = env.POST(t, "/api/v1/pine/toggle", nil)
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
	} else {
		resp.Body.Close()
		// Retry once on timeout.
		time.Sleep(2 * time.Second)
		resp = env.POST(t, "/api/v1/pine/toggle", nil)
		resp.Body.Close()
	}
	time.Sleep(1 * time.Second)
}

func TestPineStatus(t *testing.T) {
	resp := env.GET(t, "/api/v1/pine/status")
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	t.Logf("pine status: visible=%v monaco_ready=%v status=%s", st.IsVisible, st.MonacoReady, st.Status)
}

func TestPineToggleOpenClose(t *testing.T) {
	ensurePineClosed(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	// Open the editor.
	ensurePineOpen(t)

	resp := env.GET(t, "/api/v1/pine/status")
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	if !st.IsVisible {
		t.Fatal("expected is_visible=true after opening")
	}
	t.Logf("opened: visible=%v monaco=%v", st.IsVisible, st.MonacoReady)

	// Close the editor.
	ensurePineClosed(t)

	resp = env.GET(t, "/api/v1/pine/status")
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[pineState](t, resp)
	if st.IsVisible {
		t.Fatal("expected is_visible=false after closing")
	}
	t.Logf("closed: visible=%v", st.IsVisible)
}

func TestPineSourceReadWrite(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	const testSource = `//@version=6
indicator("Integration Test Indicator", overlay=true)
plot(close, "Close Price")
`

	// Write source.
	resp := env.PUT(t, "/api/v1/pine/source", map[string]any{
		"source": testSource,
	})
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	requireField(t, st.Status, "set", "status")
	if st.SourceLength == 0 {
		t.Fatal("expected source_length > 0 after set")
	}
	t.Logf("set source: length=%d lines=%d name=%q", st.SourceLength, st.LineCount, st.ScriptName)

	// Read source back.
	resp = env.GET(t, "/api/v1/pine/source")
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[pineState](t, resp)
	if st.ScriptSource == "" {
		t.Fatal("expected non-empty script_source")
	}
	if !strings.Contains(st.ScriptSource, "Integration Test Indicator") {
		t.Fatalf("source does not contain expected text, got: %s", st.ScriptSource[:min(100, len(st.ScriptSource))])
	}
	t.Logf("read source: length=%d lines=%d name=%q", st.SourceLength, st.LineCount, st.ScriptName)
}

func TestPineSetSource_EmptyRejected(t *testing.T) {
	resp := env.PUT(t, "/api/v1/pine/source", map[string]any{
		"source": "",
	})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for empty source")
	}
	resp.Body.Close()
	t.Logf("empty source correctly rejected with status %d", resp.StatusCode)
}

func TestPineSave(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	// Write a valid script first.
	resp := env.PUT(t, "/api/v1/pine/source", map[string]any{
		"source": `//@version=6
indicator("Pine Save Test", overlay=true)
plot(close)
`,
	})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	time.Sleep(500 * time.Millisecond)

	// Save the script.
	resp = env.POST(t, "/api/v1/pine/save", nil)
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	requireField(t, st.Status, "saved", "status")
	t.Logf("saved: status=%s visible=%v", st.Status, st.IsVisible)
}

func TestPineAddToChart(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	// Write a valid script.
	resp := env.PUT(t, "/api/v1/pine/source", map[string]any{
		"source": `//@version=6
indicator("Pine AddToChart Test", overlay=true)
plot(close, color=color.blue)
`,
	})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	time.Sleep(500 * time.Millisecond)

	// Add to chart.
	resp = env.POST(t, "/api/v1/pine/add-to-chart", nil)
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	requireField(t, st.Status, "added", "status")
	t.Logf("added to chart: status=%s visible=%v", st.Status, st.IsVisible)

	// Clean up: remove the study that was just added.
	time.Sleep(1 * time.Second)
	resp = env.GET(t, env.chartPath("studies"))
	requireStatus(t, resp, http.StatusOK)
	studies := decodeJSON[struct {
		Studies []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"studies"`
	}](t, resp)
	for _, s := range studies.Studies {
		if strings.Contains(s.Name, "Pine AddToChart Test") {
			r := env.DELETE(t, env.chartPath("studies/"+s.ID))
			r.Body.Close()
			break
		}
	}
}

func TestPineUndoRedo(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	// Write source.
	resp := env.PUT(t, "/api/v1/pine/source", map[string]any{
		"source": `//@version=6
indicator("Undo Redo Test", overlay=true)
plot(close)
`,
	})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	// Undo.
	resp = env.POST(t, "/api/v1/pine/undo", nil)
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	t.Logf("undo: status=%s visible=%v", st.Status, st.IsVisible)

	// Redo.
	resp = env.POST(t, "/api/v1/pine/redo", nil)
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[pineState](t, resp)
	t.Logf("redo: status=%s visible=%v", st.Status, st.IsVisible)
}

func TestPineFindReplace(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	// Write source with known text.
	resp := env.PUT(t, "/api/v1/pine/source", map[string]any{
		"source": `//@version=6
indicator("FindReplace Test", overlay=true)
plot(close, "Close Price")
plot(open, "Open Price")
`,
	})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	// Find and replace "Price" with "Value".
	resp = env.POST(t, "/api/v1/pine/find-replace", map[string]any{
		"find":    "Price",
		"replace": "Value",
	})
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	if st.MatchCount < 2 {
		t.Fatalf("expected at least 2 replacements, got %d", st.MatchCount)
	}
	t.Logf("find-replace: status=%s matches=%d", st.Status, st.MatchCount)

	// Verify the replacement took effect.
	resp = env.GET(t, "/api/v1/pine/source")
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[pineState](t, resp)
	if !strings.Contains(st.ScriptSource, "Value") {
		t.Fatal("expected source to contain 'Value' after replacement")
	}
	if strings.Contains(st.ScriptSource, "Price") {
		t.Fatal("expected source to no longer contain 'Price' after replacement")
	}
}

func TestPineGoToLine(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	resp := env.POST(t, "/api/v1/pine/go-to-line", map[string]any{
		"line": 2,
	})
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	t.Logf("go-to-line: status=%s visible=%v", st.Status, st.IsVisible)
}

func TestPineGoToLine_InvalidZero(t *testing.T) {
	resp := env.POST(t, "/api/v1/pine/go-to-line", map[string]any{
		"line": 0,
	})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for line=0")
	}
	resp.Body.Close()
	t.Logf("line=0 correctly rejected with status %d", resp.StatusCode)
}

func TestPineDeleteLine(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	// Write multi-line source.
	resp := env.PUT(t, "/api/v1/pine/source", map[string]any{
		"source": `//@version=6
indicator("DeleteLine Test", overlay=true)
plot(close)
plot(open)
plot(high)
`,
	})
	requireStatus(t, resp, http.StatusOK)
	before := decodeJSON[pineState](t, resp)
	time.Sleep(500 * time.Millisecond)

	// Delete one line.
	resp = env.POST(t, "/api/v1/pine/delete-line", map[string]any{"count": 1})
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	t.Logf("delete-line: status=%s before_lines=%d", st.Status, before.LineCount)
}

func TestPineMoveLine(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	// Move line down.
	resp := env.POST(t, "/api/v1/pine/move-line", map[string]any{
		"direction": "down",
		"count":     1,
	})
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	t.Logf("move-line down: status=%s", st.Status)

	// Move line up.
	resp = env.POST(t, "/api/v1/pine/move-line", map[string]any{
		"direction": "up",
		"count":     1,
	})
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[pineState](t, resp)
	t.Logf("move-line up: status=%s", st.Status)
}

func TestPineToggleComment(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	// Toggle comment on, then off.
	resp := env.POST(t, "/api/v1/pine/toggle-comment", nil)
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	t.Logf("toggle-comment: status=%s", st.Status)

	resp = env.POST(t, "/api/v1/pine/toggle-comment", nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestPineToggleConsole(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	// Toggle console on.
	resp := env.POST(t, "/api/v1/pine/toggle-console", nil)
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	t.Logf("toggle-console: status=%s", st.Status)

	// Toggle console off.
	resp = env.POST(t, "/api/v1/pine/toggle-console", nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestPineInsertLine(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	resp := env.POST(t, "/api/v1/pine/insert-line", nil)
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	t.Logf("insert-line: status=%s", st.Status)
}

func TestPineConsole(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	resp := env.GET(t, "/api/v1/pine/console")
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Messages []struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"messages"`
	}](t, resp)
	t.Logf("console: %d messages", len(result.Messages))
}

func TestPineNewIndicator(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	resp := env.POST(t, "/api/v1/pine/new-indicator", nil)
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	t.Logf("new-indicator: status=%s visible=%v", st.Status, st.IsVisible)

	time.Sleep(500 * time.Millisecond)
}

func TestPineNewStrategy(t *testing.T) {
	ensurePineOpen(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	resp := env.POST(t, "/api/v1/pine/new-strategy", nil)
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	t.Logf("new-strategy: status=%s visible=%v", st.Status, st.IsVisible)

	time.Sleep(500 * time.Millisecond)
}

func TestPineFullLifecycle(t *testing.T) {
	ensurePineClosed(t)
	t.Cleanup(func() { ensurePineClosed(t) })

	// 1. Open editor.
	ensurePineOpen(t)

	// 2. Write a script.
	const script = `//@version=6
indicator("Lifecycle Test", overlay=true)
plot(close, "Close")
plot(open, "Open")
`
	resp := env.PUT(t, "/api/v1/pine/source", map[string]any{"source": script})
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[pineState](t, resp)
	requireField(t, st.Status, "set", "status")
	time.Sleep(500 * time.Millisecond)

	// 3. Read source back.
	resp = env.GET(t, "/api/v1/pine/source")
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[pineState](t, resp)
	if !strings.Contains(st.ScriptSource, "Lifecycle Test") {
		t.Fatal("source does not contain expected script name")
	}

	// 4. Find and replace.
	resp = env.POST(t, "/api/v1/pine/find-replace", map[string]any{
		"find":    "Close",
		"replace": "LastClose",
	})
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[pineState](t, resp)
	if st.MatchCount == 0 {
		t.Fatal("expected at least one replacement")
	}

	// 5. Save.
	resp = env.POST(t, "/api/v1/pine/save", nil)
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[pineState](t, resp)
	requireField(t, st.Status, "saved", "status")

	// 6. Add to chart.
	resp = env.POST(t, "/api/v1/pine/add-to-chart", nil)
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[pineState](t, resp)
	requireField(t, st.Status, "added", "status")

	// 7. Check console.
	resp = env.GET(t, "/api/v1/pine/console")
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// 8. Clean up: remove the study.
	time.Sleep(1 * time.Second)
	resp = env.GET(t, env.chartPath("studies"))
	requireStatus(t, resp, http.StatusOK)
	studies := decodeJSON[struct {
		Studies []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"studies"`
	}](t, resp)
	for _, s := range studies.Studies {
		if strings.Contains(s.Name, "Lifecycle Test") {
			r := env.DELETE(t, env.chartPath("studies/"+s.ID))
			r.Body.Close()
			break
		}
	}

	// 9. Close editor.
	ensurePineClosed(t)

	resp = env.GET(t, "/api/v1/pine/status")
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[pineState](t, resp)
	if st.IsVisible {
		t.Fatal("expected editor closed after toggle")
	}
	t.Logf("full pine lifecycle completed successfully")
}
