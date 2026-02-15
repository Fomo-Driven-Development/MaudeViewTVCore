//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"
)

// replayStatus is the JSON shape returned by GET .../replay/status.
type replayStatus struct {
	IsReplayStarted   bool    `json:"is_replay_started"`
	IsReplayFinished  bool    `json:"is_replay_finished"`
	IsReplayConnected bool    `json:"is_replay_connected"`
	IsAutoplayStarted bool    `json:"is_autoplay_started"`
	ReplayPoint       any     `json:"replay_point"`
	ServerTime        any     `json:"server_time"`
	AutoplayDelay     float64 `json:"autoplay_delay"`
	Depth             any     `json:"depth"`
}

// navStatus is the JSON shape returned by most replay action endpoints.
type navStatus struct {
	ChartID string `json:"chart_id"`
	Status  string `json:"status"`
}

// ensureReplayDeactivated leaves replay mode if currently active.
func ensureReplayDeactivated(t *testing.T) {
	t.Helper()
	resp := env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[replayStatus](t, resp)
	if st.IsReplayStarted || st.IsReplayConnected {
		resp = env.POST(t, env.chartPath("replay/deactivate"), nil)
		requireStatus(t, resp, http.StatusOK)
		resp.Body.Close()
		time.Sleep(testDataSettleMedium)
	}
}

func TestReplayStatus_BeforeActivation(t *testing.T) {
	ensureReplayDeactivated(t)

	resp := env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[replayStatus](t, resp)

	if st.IsReplayStarted {
		t.Fatal("expected is_replay_started=false before activation")
	}
	if st.IsAutoplayStarted {
		t.Fatal("expected is_autoplay_started=false before activation")
	}
	t.Logf("replay status before activation: started=%v connected=%v", st.IsReplayStarted, st.IsReplayConnected)
}

func TestReplayProbe(t *testing.T) {
	resp := env.GET(t, env.chartPath("replay/probe"))
	requireStatus(t, resp, http.StatusOK)

	probe := decodeJSON[struct {
		Found       bool     `json:"found"`
		AccessPaths []string `json:"access_paths"`
		Methods     []string `json:"methods"`
	}](t, resp)

	if !probe.Found {
		t.Fatal("expected replay manager to be found")
	}
	if len(probe.Methods) == 0 {
		t.Fatal("expected replay manager to have methods")
	}
	t.Logf("replay probe: found=%v, %d methods, paths=%v", probe.Found, len(probe.Methods), probe.AccessPaths)
}

func TestReplayProbeDeep(t *testing.T) {
	resp := env.GET(t, env.chartPath("replay/probe/deep"))
	requireStatus(t, resp, http.StatusOK)

	probe := decodeJSON[map[string]any](t, resp)
	if len(probe) == 0 {
		t.Fatal("expected deep probe to return data")
	}
	t.Logf("deep probe keys: %d", len(probe))
}

func TestReplayScan(t *testing.T) {
	resp := env.GET(t, env.chartPath("replay/scan"))
	requireStatus(t, resp, http.StatusOK)

	scan := decodeJSON[map[string]any](t, resp)
	if len(scan) == 0 {
		t.Fatal("expected scan to return data")
	}
	t.Logf("scan result keys: %d", len(scan))
}

func TestReplayActivateAutoAndDeactivate(t *testing.T) {
	ensureReplayDeactivated(t)
	t.Cleanup(func() { ensureReplayDeactivated(t) })

	// Activate at first available date.
	resp := env.POST(t, env.chartPath("replay/activate/auto"), nil)
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[map[string]any](t, resp)
	t.Logf("activate auto result: %v", result)

	time.Sleep(testDataSettleMedium)

	// Status should show replay started.
	resp = env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[replayStatus](t, resp)
	if !st.IsReplayStarted {
		t.Fatal("expected is_replay_started=true after activate/auto")
	}
	t.Logf("after activate/auto: started=%v connected=%v", st.IsReplayStarted, st.IsReplayConnected)

	// Deactivate.
	resp = env.POST(t, env.chartPath("replay/deactivate"), nil)
	requireStatus(t, resp, http.StatusOK)
	nav := decodeJSON[navStatus](t, resp)
	requireField(t, nav.Status, "deactivated", "status")

	time.Sleep(testDataSettleMedium)

	// Status should show replay not started.
	resp = env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[replayStatus](t, resp)
	if st.IsReplayStarted {
		t.Fatal("expected is_replay_started=false after deactivate")
	}
}

func TestReplayActivateAtDate(t *testing.T) {
	ensureReplayDeactivated(t)
	t.Cleanup(func() { ensureReplayDeactivated(t) })

	// Activate replay at 2024-01-15 (unix timestamp).
	const jan15_2024 = 1705276800.0
	resp := env.POST(t, env.chartPath("replay/activate"), map[string]any{
		"date": jan15_2024,
	})
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[map[string]any](t, resp)
	t.Logf("activate at date result: %v", result)

	time.Sleep(testDataSettleMedium)

	resp = env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[replayStatus](t, resp)
	if !st.IsReplayStarted {
		t.Fatal("expected is_replay_started=true after activate at date")
	}
	t.Logf("after activate at date: started=%v connected=%v point=%v", st.IsReplayStarted, st.IsReplayConnected, st.ReplayPoint)
}

func TestReplayStep(t *testing.T) {
	ensureReplayDeactivated(t)
	t.Cleanup(func() { ensureReplayDeactivated(t) })

	// Activate replay first.
	resp := env.POST(t, env.chartPath("replay/activate/auto"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(testDataSettleMedium)

	// Step forward 1 bar.
	resp = env.POST(t, env.chartPath("replay/step"), map[string]any{"count": 1})
	requireStatus(t, resp, http.StatusOK)
	nav := decodeJSON[navStatus](t, resp)
	requireField(t, nav.Status, "stepped 1", "status")
	t.Logf("step 1: %s", nav.Status)

	// Step forward 5 bars.
	resp = env.POST(t, env.chartPath("replay/step"), map[string]any{
		"count": 5,
	})
	requireStatus(t, resp, http.StatusOK)
	nav = decodeJSON[navStatus](t, resp)
	requireField(t, nav.Status, "stepped 5", "status")
	t.Logf("step 5: %s", nav.Status)
}

func TestReplayAutoplay(t *testing.T) {
	ensureReplayDeactivated(t)
	t.Cleanup(func() { ensureReplayDeactivated(t) })

	// Activate replay.
	resp := env.POST(t, env.chartPath("replay/activate/auto"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(testDataSettleMedium)

	// Start autoplay.
	resp = env.POST(t, env.chartPath("replay/autoplay/start"), nil)
	requireStatus(t, resp, http.StatusOK)
	nav := decodeJSON[navStatus](t, resp)
	requireField(t, nav.Status, "autoplay_started", "status")

	time.Sleep(testSettleLong)

	// Verify autoplay is running.
	resp = env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[replayStatus](t, resp)
	if !st.IsAutoplayStarted {
		t.Fatal("expected is_autoplay_started=true after start-autoplay")
	}
	t.Logf("autoplay running: delay=%.1f", st.AutoplayDelay)

	// Stop autoplay.
	resp = env.POST(t, env.chartPath("replay/autoplay/stop"), nil)
	requireStatus(t, resp, http.StatusOK)
	nav = decodeJSON[navStatus](t, resp)
	requireField(t, nav.Status, "autoplay_stopped", "status")

	time.Sleep(testSettleMedium)

	// Verify autoplay stopped.
	resp = env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[replayStatus](t, resp)
	if st.IsAutoplayStarted {
		t.Fatal("expected is_autoplay_started=false after stop-autoplay")
	}
}

func TestReplayAutoplayDelay(t *testing.T) {
	ensureReplayDeactivated(t)
	t.Cleanup(func() { ensureReplayDeactivated(t) })

	// Activate replay.
	resp := env.POST(t, env.chartPath("replay/activate/auto"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(testDataSettleMedium)

	// Change autoplay delay.
	resp = env.PUT(t, env.chartPath("replay/autoplay/delay"), map[string]any{
		"delay": 500.0,
	})
	requireStatus(t, resp, http.StatusOK)
	delayResult := decodeJSON[struct {
		ChartID string  `json:"chart_id"`
		Status  string  `json:"status"`
		Delay   float64 `json:"delay"`
	}](t, resp)
	requireField(t, delayResult.Status, "changed", "status")
	t.Logf("changed autoplay delay: %.0f", delayResult.Delay)
}

func TestReplayAutoplayDelay_InvalidZero(t *testing.T) {
	resp := env.PUT(t, env.chartPath("replay/autoplay/delay"), map[string]any{
		"delay": 0,
	})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for delay=0")
	}
	resp.Body.Close()
	t.Logf("delay=0 correctly rejected with status %d", resp.StatusCode)
}

func TestReplayReset(t *testing.T) {
	ensureReplayDeactivated(t)
	t.Cleanup(func() { ensureReplayDeactivated(t) })

	// Activate replay and step a few bars.
	resp := env.POST(t, env.chartPath("replay/activate/auto"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(testDataSettleMedium)

	resp = env.POST(t, env.chartPath("replay/step"), map[string]any{"count": 10})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Reset replay.
	resp = env.POST(t, env.chartPath("replay/reset"), nil)
	requireStatus(t, resp, http.StatusOK)
	nav := decodeJSON[navStatus](t, resp)
	requireField(t, nav.Status, "reset", "status")
	t.Logf("replay reset successful")
}

func TestReplayStartAndStop(t *testing.T) {
	ensureReplayDeactivated(t)
	t.Cleanup(func() { ensureReplayDeactivated(t) })

	// Activate replay first (start requires active replay mode).
	resp := env.POST(t, env.chartPath("replay/activate/auto"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(testDataSettleMedium)

	// Start replay at a specific point.
	const jan15_2024 = 1705276800.0
	resp = env.POST(t, env.chartPath("replay/start"), map[string]any{
		"point": jan15_2024,
	})
	requireStatus(t, resp, http.StatusOK)
	startResult := decodeJSON[navStatus](t, resp)
	requireField(t, startResult.Status, "started", "status")
	t.Logf("replay started: chart_id=%s", startResult.ChartID)

	time.Sleep(testDataSettleMedium)

	// Verify replay is running.
	resp = env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[replayStatus](t, resp)
	if !st.IsReplayStarted {
		t.Fatal("expected is_replay_started=true after start")
	}

	// Stop replay.
	resp = env.POST(t, env.chartPath("replay/stop"), nil)
	requireStatus(t, resp, http.StatusOK)
	stopResult := decodeJSON[navStatus](t, resp)
	requireField(t, stopResult.Status, "stopped", "status")
	t.Logf("replay stopped: chart_id=%s", stopResult.ChartID)

	time.Sleep(testSettleLong)

	// Verify replay stopped (still in replay mode, but playback stopped).
	resp = env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[replayStatus](t, resp)
	t.Logf("after stop: started=%v connected=%v", st.IsReplayStarted, st.IsReplayConnected)
}

func TestReplayFullLifecycle(t *testing.T) {
	ensureReplayDeactivated(t)
	t.Cleanup(func() { ensureReplayDeactivated(t) })

	// 1. Verify not active.
	resp := env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st := decodeJSON[replayStatus](t, resp)
	if st.IsReplayStarted {
		t.Fatal("expected not started at beginning")
	}

	// 2. Activate auto.
	resp = env.POST(t, env.chartPath("replay/activate/auto"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(testDataSettleMedium)

	// 3. Verify started.
	resp = env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[replayStatus](t, resp)
	if !st.IsReplayStarted {
		t.Fatal("expected started after activate")
	}

	// 4. Step forward.
	resp = env.POST(t, env.chartPath("replay/step"), map[string]any{"count": 3})
	requireStatus(t, resp, http.StatusOK)
	nav := decodeJSON[navStatus](t, resp)
	requireField(t, nav.Status, "stepped 3", "status")

	// 5. Start autoplay.
	resp = env.POST(t, env.chartPath("replay/autoplay/start"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(testSettleLong)

	// 6. Stop autoplay.
	resp = env.POST(t, env.chartPath("replay/autoplay/stop"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// 7. Reset replay.
	resp = env.POST(t, env.chartPath("replay/reset"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(testSettleLong)

	// 8. Deactivate.
	resp = env.POST(t, env.chartPath("replay/deactivate"), nil)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
	time.Sleep(testDataSettleMedium)

	// 9. Verify deactivated.
	resp = env.GET(t, env.chartPath("replay/status"))
	requireStatus(t, resp, http.StatusOK)
	st = decodeJSON[replayStatus](t, resp)
	if st.IsReplayStarted {
		t.Fatal("expected not started after full lifecycle")
	}
	t.Logf("full replay lifecycle completed successfully")
}
