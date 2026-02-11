package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCompliantArtifacts(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "research_data")
	dateRoot := filepath.Join(dataDir, "2026-02-11")
	writeCompliantArtifacts(t, dateRoot)

	t.Setenv("RESEARCHER_DATA_DIR", dataDir)
	var out bytes.Buffer
	if err := Run(context.Background(), &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := out.String(); got != "validation: complete\n" {
		t.Fatalf("Run() output = %q, want %q", got, "validation: complete\n")
	}
}

func TestRunFailsOnSchemaDrift(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "research_data")
	dateRoot := filepath.Join(dataDir, "2026-02-11")
	writeCompliantArtifacts(t, dateRoot)

	matrixPath := filepath.Join(dateRoot, "mapper/reporting/capability-matrix.jsonl")
	rewriteJSONL(t, matrixPath, []map[string]any{
		{
			"capability_id":              "trading",
			"feature_name":               "Trading",
			"candidate_entry_modules":    []string{"tradingPanel"},
			"module_ref":                 "tradingPanel",
			"trace_id":                   "tab-1:fetch:fetch-1",
			"confidence":                 0.87,
			"preconditions":              []string{"session initialized"},
			"risk_notes":                 []string{"low"},
			"recommended_control_method": "network_request_intercept",
			"recommended_control_path":   "runtime.api_request_path",
			"evidence_ids":               []string{"TRADING-E001"},
			// "feature" intentionally removed to simulate drift
		},
	})

	t.Setenv("RESEARCHER_DATA_DIR", dataDir)
	var out bytes.Buffer
	err := Run(context.Background(), &out)
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), `missing required field "feature"`) {
		t.Fatalf("Run() error = %v, want missing required field feature", err)
	}
}

func TestRunFailsOnMissingEvidence(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "research_data")
	dateRoot := filepath.Join(dataDir, "2026-02-11")
	writeCompliantArtifacts(t, dateRoot)

	correlationPath := filepath.Join(dateRoot, "mapper/correlation/capability-correlations.jsonl")
	rewriteJSONL(t, correlationPath, []map[string]any{
		{
			"capability":                       "trading",
			"primary_recommended_control_path": "runtime.api_request_path",
			"control_path_rationale":           "fetch traces dominate",
			"confidence_score":                 0.9,
			"temporal_linkage_trace_ids":       []string{"tab-1:fetch:fetch-1"},
			"evidence_links":                   []map[string]any{},
		},
	})

	t.Setenv("RESEARCHER_DATA_DIR", dataDir)
	var out bytes.Buffer
	err := Run(context.Background(), &out)
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), `"evidence_links" must include at least one evidence entry`) {
		t.Fatalf("Run() error = %v, want missing evidence_links error", err)
	}
}

func writeCompliantArtifacts(t *testing.T, dateRoot string) {
	t.Helper()
	rewriteJSONL(t, filepath.Join(dateRoot, "mapper/static-analysis/js-bundle-index.jsonl"), []map[string]any{
		{
			"primary_key": "2026-02-11/chart_a/resources/js/tradingPanel.js",
			"file_path":   "2026-02-11/chart_a/resources/js/tradingPanel.js",
			"size_bytes":  512,
			"sha256":      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"chunk_name":  "tradingPanel",
		},
	})
	rewriteJSONL(t, filepath.Join(dateRoot, "mapper/static-analysis/js-bundle-analysis.jsonl"), []map[string]any{
		{
			"primary_key":   "2026-02-11/chart_a/resources/js/tradingPanel.js",
			"file_path":     "2026-02-11/chart_a/resources/js/tradingPanel.js",
			"chunk_name":    "tradingPanel",
			"functions":     []string{"submitOrder"},
			"classes":       []string{"TradingPanel"},
			"exports":       []string{"submitOrder"},
			"import_edges":  []string{"./api"},
			"require_edges": []string{},
			"signal_anchors": []map[string]any{
				{"type": "api_route", "value": "/api/v1/orders"},
			},
		},
	})
	rewriteJSONL(t, filepath.Join(dateRoot, "mapper/runtime-probes/runtime-trace.jsonl"), []map[string]any{
		{
			"timestamp":  "2026-02-11T15:00:02Z",
			"trace_id":   "tab-1:fetch:fetch-1",
			"tab_id":     "tab-1",
			"tab_url":    "https://www.tradingview.com/chart/abc",
			"surface":    "fetch",
			"event_type": "request",
			"payload": map[string]any{
				"url": "/api/v1/orders",
			},
		},
	})
	rewriteJSONL(t, filepath.Join(dateRoot, "mapper/correlation/capability-correlations.jsonl"), []map[string]any{
		{
			"capability":                       "trading",
			"primary_recommended_control_path": "runtime.api_request_path",
			"control_path_rationale":           "fetch traces dominate",
			"confidence_score":                 0.9,
			"temporal_linkage_trace_ids":       []string{"tab-1:fetch:fetch-1"},
			"evidence_links": []map[string]any{
				{
					"source":    "runtime-trace",
					"path":      "mapper/runtime-probes/runtime-trace.jsonl",
					"record_id": "tab-1:fetch:fetch-1",
					"field":     "surface",
					"value":     "fetch",
				},
			},
		},
	})
	rewriteJSONL(t, filepath.Join(dateRoot, "mapper/reporting/capability-matrix.jsonl"), []map[string]any{
		{
			"capability_id":              "trading",
			"feature":                    "Trading",
			"feature_name":               "Trading",
			"candidate_entry_modules":    []string{"tradingPanel"},
			"module_ref":                 "tradingPanel",
			"trace_id":                   "tab-1:fetch:fetch-1",
			"confidence":                 0.87,
			"preconditions":              []string{"session initialized"},
			"risk_notes":                 []string{"low"},
			"recommended_control_method": "network_request_intercept",
			"recommended_control_path":   "runtime.api_request_path",
			"evidence_ids":               []string{"TRADING-E001"},
		},
	})
}

func rewriteJSONL(t *testing.T, path string, rows []map[string]any) {
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
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			t.Fatalf("Encode(%q) error = %v", path, err)
		}
	}
}
