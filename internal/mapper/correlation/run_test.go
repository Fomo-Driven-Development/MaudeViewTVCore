package correlation

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunBuildsCapabilityCorrelations(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "research_data")
	date := "2026-02-11"
	dateRoot := filepath.Join(dataDir, date)

	staticAnalysisRecords := []staticAnalysisRecord{
		{
			PrimaryKey: "2026-02-11/chart_a/resources/js/tradingPanel.js",
			FilePath:   "2026-02-11/chart_a/resources/js/tradingPanel.js",
			ChunkName:  "tradingPanel",
			SignalAnchors: []staticSignalAnchor{
				{Type: "api_route", Value: "/api/v1/orders"},
				{Type: "action_event", Value: "order:create"},
			},
		},
	}
	staticGraphNodes := []staticGraphNode{
		{
			PrimaryKey: "2026-02-11/chart_a/resources/js/tradingPanel.js",
			FilePath:   "2026-02-11/chart_a/resources/js/tradingPanel.js",
			ChunkName:  "tradingPanel",
			DomainHints: []staticDomainHint{
				{Domain: "trading", Rationale: "matched keyword \"trading\" in chunk_name \"tradingPanel\""},
			},
		},
	}

	start := time.Date(2026, 2, 11, 15, 0, 0, 0, time.UTC)
	traces := []runtimeTraceRecord{
		{
			Timestamp: start.Add(2 * time.Second),
			TraceID:   "tab-1:fetch:fetch-1",
			TabID:     "tab-1",
			TabURL:    "https://www.tradingview.com/chart/abc",
			Surface:   "fetch",
			EventType: "request",
			Payload: map[string]any{
				"url":   "/api/v1/orders",
				"stack": "at submitOrder (tradingPanel.js:12:5)",
			},
		},
		{
			Timestamp: start.Add(4 * time.Second),
			TraceID:   "tab-1:xhr:xhr-1",
			TabID:     "tab-1",
			TabURL:    "https://www.tradingview.com/chart/abc",
			Surface:   "xhr",
			EventType: "request",
			Payload: map[string]any{
				"endpoint": "/api/v1/orders",
				"file":     "tradingPanel.js",
			},
		},
		{
			Timestamp: start.Add(5 * time.Second),
			TraceID:   "tab-1:event_bus:event-1",
			TabID:     "tab-1",
			TabURL:    "https://www.tradingview.com/chart/abc",
			Surface:   "event_bus",
			EventType: "bus_call",
			Payload: map[string]any{
				"eventName": "order:create",
			},
		},
	}
	sessions := []runtimeTraceSession{
		{
			SessionID:   "tab-1:trading_panel_interactions:1",
			ProfileName: "trading_panel_interactions",
			TabID:       "tab-1",
			StartedAt:   start,
			EndedAt:     start.Add(10 * time.Second),
		},
	}

	writeFixtureJSONL(t, filepath.Join(dateRoot, staticAnalysisRelativePath), staticAnalysisRecords)
	writeFixtureJSONL(t, filepath.Join(dateRoot, staticGraphRelativePath), staticGraphNodes)
	writeFixtureJSONL(t, filepath.Join(dateRoot, runtimeTraceRelativePath), traces)
	writeFixtureJSONL(t, filepath.Join(dateRoot, traceSessionsRelativePath), sessions)

	t.Setenv("RESEARCHER_DATA_DIR", dataDir)
	var out bytes.Buffer
	if err := Run(context.Background(), &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := out.String(); got != "correlation: complete\n" {
		t.Fatalf("Run() output = %q, want %q", got, "correlation: complete\n")
	}

	outputPath := filepath.Join(dateRoot, correlationOutputRelative)
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", outputPath, err)
	}
	records := decodeJSONLLines[capabilityCorrelationRecord](t, data)
	if len(records) != 1 {
		t.Fatalf("correlation records = %d, want 1", len(records))
	}

	got := records[0]
	if got.Capability != "trading" {
		t.Fatalf("capability = %q, want trading", got.Capability)
	}
	if got.PrimaryRecommendedControlPath != "runtime.api_request_path" {
		t.Fatalf("primary_recommended_control_path = %q, want runtime.api_request_path", got.PrimaryRecommendedControlPath)
	}
	if got.ConfidenceScore <= 0.75 {
		t.Fatalf("confidence_score = %f, want > 0.75", got.ConfidenceScore)
	}
	if !containsAny(got.ConfidenceRationale, "anchor correlation") {
		t.Fatalf("confidence_rationale missing anchor entry: %+v", got.ConfidenceRationale)
	}
	if !containsAny(got.ConfidenceRationale, "temporal linkage") {
		t.Fatalf("confidence_rationale missing temporal entry: %+v", got.ConfidenceRationale)
	}
	if len(got.TemporalLinkageTraceIDs) != 3 {
		t.Fatalf("temporal_linkage_trace_ids = %d, want 3", len(got.TemporalLinkageTraceIDs))
	}
	if !hasEvidenceSource(got.EvidenceLinks, "runtime-trace") {
		t.Fatalf("missing runtime-trace evidence link: %+v", got.EvidenceLinks)
	}
	if !hasEvidenceSource(got.EvidenceLinks, "static-analysis") {
		t.Fatalf("missing static-analysis evidence link: %+v", got.EvidenceLinks)
	}
	if !hasEvidenceSource(got.EvidenceLinks, "runtime-session") {
		t.Fatalf("missing runtime-session evidence link: %+v", got.EvidenceLinks)
	}
}

func TestRunContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out bytes.Buffer
	if err := Run(ctx, &out); err == nil {
		t.Fatal("Run() error = nil, want context cancellation error")
	}
}

func writeFixtureJSONL[T any](t *testing.T, path string, records []T) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create(%q) error = %v", path, err)
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	for _, rec := range records {
		if err := enc.Encode(rec); err != nil {
			t.Fatalf("Encode(%q) error = %v", path, err)
		}
	}
}

func decodeJSONLLines[T any](t *testing.T, data []byte) []T {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	out := make([]T, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec T
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("json.Unmarshal(%q) error = %v", line, err)
		}
		out = append(out, rec)
	}
	return out
}

func containsAny(values []string, needle string) bool {
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func hasEvidenceSource(links []evidenceLink, source string) bool {
	for _, link := range links {
		if link.Source == source {
			return true
		}
	}
	return false
}
