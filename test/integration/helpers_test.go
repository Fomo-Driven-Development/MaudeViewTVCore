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

var env *Env

// Env holds shared state for all integration tests.
type Env struct {
	BaseURL string
	Client  *http.Client
	ChartID string // discovered from /api/v1/charts
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
	resp, err := env.Client.Get(env.BaseURL + "/api/v1/charts")
	if err != nil {
		fmt.Fprintf(os.Stderr, "server not reachable at %s: %v\n", env.BaseURL, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var listing struct {
		Charts []struct {
			ChartID string `json:"chart_id"`
		} `json:"charts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil || len(listing.Charts) == 0 {
		fmt.Fprintf(os.Stderr, "no charts found at %s (decode err: %v)\n", env.BaseURL, err)
		os.Exit(1)
	}
	env.ChartID = listing.Charts[0].ChartID
	fmt.Fprintf(os.Stdout, "integration: using chart %s at %s\n", env.ChartID, env.BaseURL)

	// Warmup: set a known timeframe to ensure the chart is focused and interactive.
	// This prevents flaky first-test failures where CDP keyboard shortcuts need chart focus.
	warmupURL := fmt.Sprintf("%s/api/v1/chart/%s/timeframe?preset=1Y", env.BaseURL, env.ChartID)
	req, _ := http.NewRequest(http.MethodPut, warmupURL, nil)
	if wr, err := env.Client.Do(req); err == nil {
		wr.Body.Close()
	}
	time.Sleep(1 * time.Second)

	os.Exit(m.Run())
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
