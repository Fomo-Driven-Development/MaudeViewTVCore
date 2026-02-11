package runtimeprobes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/target"
)

func TestFilterProbeTargets(t *testing.T) {
	targets := []*target.Info{
		{TargetID: target.ID("tab-1"), Type: "page", URL: "https://www.tradingview.com/chart"},
		{TargetID: target.ID("tab-2"), Type: "page", URL: "https://example.com"},
		{TargetID: target.ID("tab-3"), Type: "service_worker", URL: "https://www.tradingview.com"},
		{TargetID: target.ID("tab-4"), Type: "page", URL: "https://TRADINGVIEW.com/screener"},
	}

	got := filterProbeTargets(targets, "tradingview.com")
	if len(got) != 2 {
		t.Fatalf("filterProbeTargets() len = %d, want 2", len(got))
	}
	if got[0].ID != target.ID("tab-1") {
		t.Fatalf("first target id = %q, want %q", got[0].ID, "tab-1")
	}
	if got[1].ID != target.ID("tab-4") {
		t.Fatalf("second target id = %q, want %q", got[1].ID, "tab-4")
	}
}

func TestBootstrapTargetsLogsLifecycleWithTabIDs(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo}))

	targets := []probeTarget{
		{ID: target.ID("tab-1"), URL: "https://www.tradingview.com/chart/abc"},
		{ID: target.ID("tab-2"), URL: "https://www.tradingview.com/chart/xyz"},
	}

	runner := func(_ context.Context, tab probeTarget) (probeLifecycle, error) {
		switch tab.ID {
		case target.ID("tab-1"):
			return probeLifecycle{
				Attached: true,
				Result: probeBootstrapResult{
					AlreadyInjected: false,
					URL:             tab.URL,
					Installed:       []string{"fetch", "xhr", "websocket", "event_bus"},
				},
			}, nil
		case target.ID("tab-2"):
			return probeLifecycle{
				Attached: true,
				Result: probeBootstrapResult{
					AlreadyInjected: true,
					URL:             tab.URL,
					Installed:       []string{"fetch", "xhr", "websocket", "event_bus"},
				},
			}, nil
		default:
			return probeLifecycle{}, errors.New("unexpected tab")
		}
	}

	attached, injected, records, err := bootstrapTargets(context.Background(), logger, targets, runner)
	if err != nil {
		t.Fatalf("bootstrapTargets() error = %v", err)
	}
	if attached != 2 {
		t.Fatalf("bootstrapTargets() attached = %d, want 2", attached)
	}
	if injected != 2 {
		t.Fatalf("bootstrapTargets() injected = %d, want 2", injected)
	}
	if len(records) == 0 {
		t.Fatal("bootstrapTargets() records empty, want non-empty")
	}
	seen := map[string]bool{}
	for _, record := range records {
		seen[record.Surface] = true
	}
	for _, surface := range []string{"fetch", "xhr", "websocket", "event_bus"} {
		if !seen[surface] {
			t.Fatalf("records missing surface %q", surface)
		}
	}

	logText := logs.String()
	for _, want := range []string{
		"Runtime probe attach success",
		"Runtime probe inject success",
		"tab_id=tab-1",
		"tab_id=tab-2",
	} {
		if !strings.Contains(logText, want) {
			t.Fatalf("logs missing %q in %q", want, logText)
		}
	}
}

func TestRedactSecrets(t *testing.T) {
	in := map[string]any{
		"url":             "https://api.example.com?token=abc123&session=xyz",
		"Authorization":   "Bearer very.secret.token",
		"nested":          map[string]any{"cookie": "sid=sensitive", "value": "ok"},
		"array_payload":   []any{"jwt=eyJabc123.XYZ987.aaaaaa", map[string]any{"api_key": "hello"}},
		"non_secret":      "safe",
		"authToken":       "plaintext",
		"another_payload": []any{1, true},
	}

	got := redactSecrets(in)

	if got["authToken"] != "[REDACTED]" {
		t.Fatalf("authToken = %v, want [REDACTED]", got["authToken"])
	}
	url, _ := got["url"].(string)
	if strings.Contains(url, "abc123") || strings.Contains(url, "xyz") {
		t.Fatalf("url not redacted: %q", url)
	}
	auth, _ := got["Authorization"].(string)
	if strings.Contains(strings.ToLower(auth), "very.secret.token") {
		t.Fatalf("authorization not redacted: %q", auth)
	}
	nested, _ := got["nested"].(map[string]any)
	if nested["cookie"] != "[REDACTED]" {
		t.Fatalf("nested cookie = %v, want [REDACTED]", nested["cookie"])
	}
}

func TestPersistRuntimeTraceRecordsWritesJSONL(t *testing.T) {
	baseDir := t.TempDir()
	now := time.Date(2026, 2, 11, 15, 0, 0, 0, time.UTC)
	records := []runtimeTraceRecord{
		{
			Timestamp:     now,
			TraceID:       "tab-1:fetch:fetch-1",
			TabID:         "tab-1",
			TabURL:        "https://www.tradingview.com/chart",
			Surface:       "fetch",
			EventType:     "request",
			Sequence:      1,
			CorrelationID: "fetch-1",
			Payload: map[string]any{
				"url": "https://api.example.com?token=abc123",
			},
		},
	}

	if err := persistRuntimeTraceRecords(baseDir, now, records); err != nil {
		t.Fatalf("persistRuntimeTraceRecords() error = %v", err)
	}

	path := filepath.Join(baseDir, "2026-02-11", runtimeTraceRelativeOutput)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("line count = %d, want 1", len(lines))
	}

	var got runtimeTraceRecord
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !strings.Contains(got.TraceID, "tab-1:fetch") {
		t.Fatalf("trace_id = %q, want tab-1:fetch:*", got.TraceID)
	}
	if strings.Contains(got.Payload["url"].(string), "abc123") {
		t.Fatalf("payload url not redacted: %q", got.Payload["url"])
	}
}

func TestBuildTraceSessionArtifactsIncludesAllProfiles(t *testing.T) {
	now := time.Date(2026, 2, 11, 16, 0, 0, 0, time.UTC)
	tab := probeTarget{
		ID:  target.ID("tab-1"),
		URL: "https://www.tradingview.com/chart/abc?symbol=NASDAQ%3AAAPL",
	}
	records := append([]runtimeTraceRecord{}, buildProfileSeedRecords(tab, runtimeTraceProfiles)...)
	records = append(records, runtimeTraceRecord{
		Timestamp:     now.Add(2 * time.Second),
		TraceID:       "tab-1:event_bus:trading-1",
		TabID:         "tab-1",
		TabURL:        tab.URL,
		Surface:       "event_bus",
		EventType:     "bus_call",
		Sequence:      10,
		CorrelationID: "trading-1",
		Payload: map[string]any{
			"eventName":       "order:create",
			"active_symbol":   "NASDAQ:MSFT",
			"panel_component": "trading_panel",
		},
	})

	sessions := buildTraceSessionArtifacts(records, runtimeTraceProfiles)
	if len(sessions) != len(runtimeTraceProfiles) {
		t.Fatalf("session count = %d, want %d", len(sessions), len(runtimeTraceProfiles))
	}

	byProfile := make(map[string]traceSessionArtifact, len(sessions))
	for _, session := range sessions {
		byProfile[session.ProfileName] = session
		if session.TraceCount < 1 {
			t.Fatalf("profile %q trace_count = %d, want >= 1", session.ProfileName, session.TraceCount)
		}
		if session.SessionID == "" {
			t.Fatalf("profile %q has empty session_id", session.ProfileName)
		}
		if session.TabContext != "chart" {
			t.Fatalf("profile %q tab_context = %q, want chart", session.ProfileName, session.TabContext)
		}
		if session.StartedAt.IsZero() || session.EndedAt.IsZero() {
			t.Fatalf("profile %q has zero timestamps", session.ProfileName)
		}
		if session.EndedAt.Before(session.StartedAt) {
			t.Fatalf("profile %q ended_at before started_at", session.ProfileName)
		}
	}

	for _, profile := range runtimeTraceProfiles {
		if _, ok := byProfile[profile.Name]; !ok {
			t.Fatalf("missing session artifact for profile %q", profile.Name)
		}
	}

	tradingSession := byProfile["trading_panel_interactions"]
	if tradingSession.ActiveSymbol != "NASDAQ:MSFT" {
		t.Fatalf("trading session active_symbol = %q, want NASDAQ:MSFT", tradingSession.ActiveSymbol)
	}
}

func TestPersistTraceSessionArtifactsWritesJSONL(t *testing.T) {
	baseDir := t.TempDir()
	now := time.Date(2026, 2, 11, 18, 0, 0, 0, time.UTC)
	sessions := []traceSessionArtifact{
		{
			SessionID:    "tab-1:chart_interaction:1",
			ProfileName:  "chart_interaction",
			TabID:        "tab-1",
			TabURL:       "https://www.tradingview.com/chart/abc?symbol=NASDAQ%3AAAPL",
			TabContext:   "chart",
			ActiveSymbol: "NASDAQ:AAPL",
			StartedAt:    now,
			EndedAt:      now.Add(5 * time.Second),
			TraceCount:   2,
			TraceIDs:     []string{"tab-1:trace_profile:chart-1", "tab-1:fetch:fetch-1"},
		},
	}

	if err := persistTraceSessionArtifacts(baseDir, now, sessions); err != nil {
		t.Fatalf("persistTraceSessionArtifacts() error = %v", err)
	}

	path := filepath.Join(baseDir, "2026-02-11", runtimeTraceSessionRelativeOutput)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("line count = %d, want 1", len(lines))
	}

	var got traceSessionArtifact
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got.ProfileName != "chart_interaction" {
		t.Fatalf("profile_name = %q, want chart_interaction", got.ProfileName)
	}
	if got.TabContext != "chart" {
		t.Fatalf("tab_context = %q, want chart", got.TabContext)
	}
	if got.ActiveSymbol != "NASDAQ:AAPL" {
		t.Fatalf("active_symbol = %q, want NASDAQ:AAPL", got.ActiveSymbol)
	}
}
