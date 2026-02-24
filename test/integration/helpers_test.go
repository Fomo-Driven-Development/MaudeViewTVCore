//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	// UI settle times.
	testSettleShort  = 200 * time.Millisecond
	testSettleMedium = 500 * time.Millisecond
	testSettleLong   = 1 * time.Second
	testSettleSmall  = 3 * testSettleShort / 2 // 300ms
)

const (
	// Data loading settle times.
	testDataSettleMedium = 2 * testSettleLong
	testDataSettleLong   = 5 * testSettleLong
)

const (
	// Page reload settle times.
	testSettleReload      = 3 * time.Second
	testReloadSettleLongest = 8 * testSettleLong
)

var env *Env

// Env holds shared state for all integration tests.
type Env struct {
	BaseURL          string
	Client           *http.Client
	ChartID          string // discovered from /api/v1/charts
	OriginalLayoutID int    // for switch-back in teardown
	TestLayoutName   string // for deletion in teardown (resolved by name)
	StrategyReady    bool   // true if setup loaded a strategy
	IsMulti          bool   // true when running against tv_multi_controller
}

// discoverChartID fetches /api/v1/charts and sets env.ChartID to the first chart.
// If TV_TEST_CHART_ID is set, that value is used directly without an HTTP call.
func (e *Env) discoverChartID() error {
	if id := os.Getenv("TV_TEST_CHART_ID"); id != "" {
		e.ChartID = id
		return nil
	}
	resp, err := e.Client.Get(e.BaseURL + "/api/v1/charts")
	if err != nil {
		return fmt.Errorf("server not reachable at %s: %w", e.BaseURL, err)
	}
	defer resp.Body.Close()

	var listing struct {
		Charts []struct {
			ChartID string `json:"chart_id"`
		} `json:"charts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return fmt.Errorf("decode charts: %w", err)
	}
	if len(listing.Charts) == 0 {
		return fmt.Errorf("no charts found at %s", e.BaseURL)
	}
	e.ChartID = listing.Charts[0].ChartID
	return nil
}

// layoutInfo mirrors the JSON shape from /api/v1/layouts.
type layoutInfo struct {
	ID   int    `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
}

// layoutStatus mirrors the JSON shape from /api/v1/layout/status.
type layoutStatus struct {
	LayoutName string `json:"layout_name"`
	LayoutID   string `json:"layout_id"`
}

// layoutActionResult mirrors the JSON shape from layout action endpoints.
type layoutActionResult struct {
	Status     string `json:"status"`
	LayoutName string `json:"layout_name"`
	LayoutID   string `json:"layout_id"`
}

// listLayouts fetches all layouts.
func (e *Env) listLayouts() ([]layoutInfo, error) {
	resp, err := e.Client.Get(e.BaseURL + e.featurePath("layouts"))
	if err != nil {
		return nil, fmt.Errorf("list layouts: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list layouts: status %d: %s", resp.StatusCode, body)
	}
	var result struct {
		Layouts []layoutInfo `json:"layouts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode layouts: %w", err)
	}
	return result.Layouts, nil
}

// currentLayoutName returns the name of the currently active layout.
func (e *Env) currentLayoutName() (string, error) {
	resp, err := e.Client.Get(e.BaseURL + e.featurePath("layout/status"))
	if err != nil {
		return "", fmt.Errorf("layout status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("layout status: status %d: %s", resp.StatusCode, body)
	}
	var st layoutStatus
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return "", fmt.Errorf("decode layout status: %w", err)
	}
	return st.LayoutName, nil
}

// resolveLayoutNumericID finds a layout's numeric ID by name.
func (e *Env) resolveLayoutNumericID(name string) (int, error) {
	layouts, err := e.listLayouts()
	if err != nil {
		return 0, err
	}
	for _, l := range layouts {
		if l.Name == name {
			return l.ID, nil
		}
	}
	return 0, fmt.Errorf("layout %q not found", name)
}

// currentLayoutNumericID returns the numeric ID of the currently active layout.
// Matches by both name and layout_id (URL) to handle unsaved/orphaned layouts.
func (e *Env) currentLayoutNumericID() (int, error) {
	resp, err := e.Client.Get(e.BaseURL + e.featurePath("layout/status"))
	if err != nil {
		return 0, fmt.Errorf("layout status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("layout status: status %d: %s", resp.StatusCode, body)
	}
	var st layoutStatus
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return 0, fmt.Errorf("decode layout status: %w", err)
	}

	layouts, err := e.listLayouts()
	if err != nil {
		return 0, err
	}
	// Match by URL first (layout_id from status == url in list), then by name.
	for _, l := range layouts {
		if l.URL == st.LayoutID {
			return l.ID, nil
		}
	}
	for _, l := range layouts {
		if l.Name == st.LayoutName {
			return l.ID, nil
		}
	}
	return 0, fmt.Errorf("current layout %q (id=%s) not found in layout list", st.LayoutName, st.LayoutID)
}

// doJSON performs an HTTP request with a JSON body, returning the response.
func (e *Env) doJSON(method, path string, body any) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, e.BaseURL+path, r)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return e.Client.Do(req)
}

// skipOrFatal calls t.Fatal if strategy setup succeeded (real regression),
// or t.Skip if setup failed (graceful degradation).
func skipOrFatal(t *testing.T, msg string) {
	t.Helper()
	if env.StrategyReady {
		t.Fatal(msg)
	}
	t.Skip(msg)
}

const testStrategySource = `//@version=6
strategy("tv_agent Test Strategy", overlay=true)
if bar_index > 10
    strategy.entry("Long", strategy.long)
if bar_index > 20
    strategy.close("Long")
`

// setupStrategyLayout clones the current layout, adds a test strategy via Pine
// editor, and sets env.StrategyReady on success.
func setupStrategyLayout() error {
	// 1. Record original layout numeric ID.
	origID, err := env.currentLayoutNumericID()
	if err != nil {
		return fmt.Errorf("get original layout ID: %w", err)
	}
	env.OriginalLayoutID = origID
	fmt.Fprintf(os.Stdout, "integration: original layout ID: %d\n", origID)

	// 2. Clone layout with a unique name.
	cloneName := fmt.Sprintf("tv_agent_test_%d", time.Now().Unix())
	resp, err := env.doJSON(http.MethodPost, env.featurePath("layout/clone"), map[string]any{
		"name": cloneName,
	})
	if err != nil {
		return fmt.Errorf("clone layout: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("clone layout: status %d: %s", resp.StatusCode, body)
	}
	var cloneResult layoutActionResult
	if err := json.NewDecoder(resp.Body).Decode(&cloneResult); err != nil {
		return fmt.Errorf("decode clone result: %w", err)
	}
	env.TestLayoutName = cloneName
	fmt.Fprintf(os.Stdout, "integration: cloned layout %q → status=%s id=%s\n", cloneName, cloneResult.Status, cloneResult.LayoutID)

	// 3. Wait for clone to settle (auto-switches + page reload).
	// Clone triggers page reload. Retry until the page is back and we can
	// confirm we're on the cloned layout by checking status name.
	for attempt := range 5 {
		wait := testSettleReload + time.Duration(attempt*2)*testSettleLong
		time.Sleep(wait)

		currentName, nameErr := env.currentLayoutName()
		if nameErr != nil {
			fmt.Fprintf(os.Stdout, "integration: status attempt %d failed: %v\n", attempt+1, nameErr)
			continue
		}
		if currentName == cloneName {
			fmt.Fprintf(os.Stdout, "integration: confirmed on cloned layout %q (attempt %d)\n", cloneName, attempt+1)
			break
		}
		fmt.Fprintf(os.Stdout, "integration: on %q, expected %q (attempt %d)\n", currentName, cloneName, attempt+1)
	}

	// Save the cloned layout so it persists (needed for teardown deletion).
	saveResp, err := env.doJSON(http.MethodPost, env.featurePath("layout/save"), nil)
	if err != nil {
		fmt.Fprintf(os.Stdout, "integration: save warning: %v\n", err)
	} else {
		saveResp.Body.Close()
	}

	// 4. Re-discover chart ID (changes after layout switch).
	if err := env.discoverChartID(); err != nil {
		return fmt.Errorf("discover chart after clone: %w", err)
	}
	fmt.Fprintf(os.Stdout, "integration: chart ID after clone: %s\n", env.ChartID)

	// 5. Open Pine editor → new strategy → write source → add to chart → close.
	if err := addTestStrategy(); err != nil {
		return fmt.Errorf("add test strategy: %w", err)
	}

	// 6. Wait for backtest data to generate.
	time.Sleep(testDataSettleLong)

	env.StrategyReady = true
	fmt.Fprintf(os.Stdout, "integration: strategy setup complete (StrategyReady=true)\n")
	return nil
}

// addTestStrategy opens Pine editor, loads a new strategy template, writes the
// test strategy source, adds it to the chart, and closes the editor.
func addTestStrategy() error {
	// Open Pine editor.
	resp, err := env.doJSON(http.MethodPost, env.featurePath("pine/toggle"), nil)
	if err != nil {
		return fmt.Errorf("open pine: %w", err)
	}
	resp.Body.Close()
	time.Sleep(testDataSettleMedium)

	// Verify it opened.
	resp, err = env.Client.Get(env.BaseURL + env.featurePath("pine/status"))
	if err != nil {
		return fmt.Errorf("pine status: %w", err)
	}
	var st pineState
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		resp.Body.Close()
		return fmt.Errorf("decode pine status: %w", err)
	}
	resp.Body.Close()
	if !st.IsVisible {
		// Retry toggle once.
		time.Sleep(testDataSettleMedium)
		resp, err = env.doJSON(http.MethodPost, env.featurePath("pine/toggle"), nil)
		if err != nil {
			return fmt.Errorf("retry open pine: %w", err)
		}
		resp.Body.Close()
		time.Sleep(testDataSettleMedium)
	}

	// Load new strategy template.
	resp, err = env.doJSON(http.MethodPost, env.featurePath("pine/new-strategy"), nil)
	if err != nil {
		return fmt.Errorf("new-strategy: %w", err)
	}
	resp.Body.Close()
	time.Sleep(testSettleLong)

	// Write test strategy source.
	resp, err = env.doJSON(http.MethodPut, env.featurePath("pine/source"), map[string]any{
		"source": testStrategySource,
	})
	if err != nil {
		return fmt.Errorf("set source: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set source: status %d: %s", resp.StatusCode, body)
	}
	fmt.Fprintf(os.Stdout, "integration: wrote test strategy source\n")

	time.Sleep(testSettleMedium)

	// Add to chart.
	resp2, err := env.doJSON(http.MethodPost, env.featurePath("pine/add-to-chart"), nil)
	if err != nil {
		return fmt.Errorf("add-to-chart: %w", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		return fmt.Errorf("add-to-chart: status %d: %s", resp2.StatusCode, body)
	}
	fmt.Fprintf(os.Stdout, "integration: added strategy to chart\n")

	time.Sleep(testSettleLong)

	// Save the script so it persists as a named study on the chart.
	// Without saving, the strategy is an "unnamed" editor tab that gets removed
	// when subsequent Pine tests load new templates (e.g. Ctrl+K Ctrl+S).
	saveResp, err := env.doJSON(http.MethodPost, env.featurePath("pine/save"), nil)
	if err != nil {
		fmt.Fprintf(os.Stdout, "integration: save strategy warning: %v\n", err)
	} else {
		saveResp.Body.Close()
		fmt.Fprintf(os.Stdout, "integration: saved strategy script\n")
	}

	time.Sleep(testSettleLong)

	// NOTE: Do NOT call PineNewTab (Shift+Alt+T) here. That shortcut pops the
	// Pine editor out to a separate browser tab (tradingview.com/pine/...) rather
	// than creating a new script tab within the panel. The popped-out editor is
	// invisible to DOM queries in the chart page, causing ensurePineOpen() to
	// timeout (~12s) in subsequent Pine tests.
	//
	// If Pine tests overwrite the strategy tab (e.g. Ctrl+K Ctrl+S creates a new
	// strategy template), TestStrategy_Init will re-add the test strategy before
	// strategy tests run.

	// Close Pine editor.
	resp3, err := env.doJSON(http.MethodPost, env.featurePath("pine/toggle"), nil)
	if err != nil {
		return fmt.Errorf("close pine: %w", err)
	}
	resp3.Body.Close()
	time.Sleep(testSettleLong)

	return nil
}

// teardownStrategyLayout switches back to the original layout and deletes the
// test clone. Errors are logged but not fatal (orphaned layouts are harmless).
func teardownStrategyLayout() {
	if env.TestLayoutName == "" {
		return // no clone was created
	}

	fmt.Fprintf(os.Stdout, "integration: teardown — switching back to layout %d\n", env.OriginalLayoutID)

	// Switch back to original layout.
	resp, err := env.doJSON(http.MethodPost, env.featurePath("layout/switch"), map[string]any{
		"id": env.OriginalLayoutID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: teardown switch: %v\n", err)
	} else {
		resp.Body.Close()
	}

	// Wait for layout switch to settle (triggers page reload).
	time.Sleep(testDataSettleLong)

	// Resolve and delete the test layout by name.
	// By now the layout has been saved and should be in the list.
	testLayoutID, err := env.resolveLayoutNumericID(env.TestLayoutName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: teardown resolve %q: %v (orphan layout — delete manually)\n", env.TestLayoutName, err)
		return
	}

	path := env.featurePath(fmt.Sprintf("layout/%d", testLayoutID))
	req, err := http.NewRequest(http.MethodDelete, env.BaseURL+path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: teardown delete request: %v\n", err)
	} else {
		resp, err = env.Client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "integration: teardown delete: %v\n", err)
		} else {
			resp.Body.Close()
			fmt.Fprintf(os.Stdout, "integration: teardown — deleted test layout %d (%s)\n", testLayoutID, env.TestLayoutName)
		}
	}

	// Re-discover chart ID for any subsequent use.
	if err := env.discoverChartID(); err != nil {
		fmt.Fprintf(os.Stderr, "integration: teardown discover chart: %v\n", err)
	}
}

func TestMain(m *testing.M) {
	baseURL := os.Getenv("TV_CONTROLLER_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8188"
	}

	env = &Env{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}

	// Verify server is reachable and discover first chart ID.
	if err := env.discoverChartID(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "integration: using chart %s at %s\n", env.ChartID, env.BaseURL)

	// Detect single vs multi controller mode by probing the session-level layout status path.
	// In multi-controller, /api/v1/layout/status returns 404; chart-scoped paths are used instead.
	probeResp, probeErr := env.Client.Get(env.BaseURL + "/api/v1/layout/status")
	if probeErr == nil {
		probeResp.Body.Close()
		env.IsMulti = probeResp.StatusCode == http.StatusNotFound
	}
	fmt.Fprintf(os.Stdout, "integration: mode=%s\n", map[bool]string{true: "multi", false: "single"}[env.IsMulti])

	// Warmup: set a known timeframe to ensure the chart is focused and interactive.
	warmupURL := fmt.Sprintf("%s/api/v1/chart/%s/timeframe?preset=1Y", env.BaseURL, env.ChartID)
	req, _ := http.NewRequest(http.MethodPut, warmupURL, nil)
	if wr, err := env.Client.Do(req); err == nil {
		wr.Body.Close()
	}
	time.Sleep(testSettleLong)

	// Strategy layout setup: clone layout, add test strategy.
	if err := setupStrategyLayout(); err != nil {
		fmt.Fprintf(os.Stderr, "integration: strategy setup failed (tests will skip): %v\n", err)
	}

	code := m.Run()
	teardownStrategyLayout()
	os.Exit(code)
}

// --- HTTP helpers ---

func (e *Env) GET(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := e.Client.Get(e.BaseURL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

func (e *Env) PUT(t *testing.T, path string, body any) *http.Response {
	t.Helper()
	return e.do(t, http.MethodPut, path, body)
}

func (e *Env) POST(t *testing.T, path string, body any) *http.Response {
	t.Helper()
	return e.do(t, http.MethodPost, path, body)
}

func (e *Env) DELETE(t *testing.T, path string) *http.Response {
	t.Helper()
	return e.do(t, http.MethodDelete, path, nil)
}

func (e *Env) do(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("%s %s: marshal body: %v", method, path, err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, e.BaseURL+path, r)
	if err != nil {
		t.Fatalf("%s %s: new request: %v", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := e.Client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

// --- Assertion helpers ---

func requireStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d; body: %s", resp.StatusCode, want, body)
	}
}

func decodeJSON[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return v
}

func requireField[T comparable](t *testing.T, got, want T, name string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

// --- Chart path helper ---

func (e *Env) chartPath(suffix string) string {
	return fmt.Sprintf("/api/v1/chart/%s/%s", e.ChartID, suffix)
}

// featurePath returns the path for session-level feature namespaces.
// In single-controller mode: /api/v1/{suffix}
// In multi-controller mode: /api/v1/chart/{id}/{suffix}
// Covers: layout, alerts, watchlists, pine, hotlists, study-templates, indicators.
func (e *Env) featurePath(suffix string) string {
	if e.IsMulti {
		return fmt.Sprintf("/api/v1/chart/%s/%s", e.ChartID, suffix)
	}
	return "/api/v1/" + suffix
}

// panePath returns the path for pane navigation actions.
// In single-controller mode: /api/v1/chart/{action}
// In multi-controller mode: /api/v1/chart/{id}/pane/{action}
func (e *Env) panePath(action string) string {
	if e.IsMulti {
		return fmt.Sprintf("/api/v1/chart/%s/pane/%s", e.ChartID, action)
	}
	return "/api/v1/chart/" + action
}
