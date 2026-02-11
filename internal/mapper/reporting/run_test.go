package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBuildsCapabilityMatrixAndSummary(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "research_data")
	dateRoot := filepath.Join(dataDir, "2026-02-11")

	correlations := []capabilityCorrelationRecord{
		{
			Capability:                    "trading",
			PrimaryRecommendedControlPath: "runtime.api_request_path",
			ControlPathRationale:          "fetch/xhr runtime traces dominate and align with API/event anchors",
			ConfidenceScore:               0.9,
			TemporalLinkageTraceIDs:       []string{"tab-1:fetch:fetch-1"},
			EvidenceLinks: []evidenceLink{
				{
					Source:   "runtime-trace",
					Path:     "mapper/runtime-probes/runtime-trace.jsonl",
					RecordID: "tab-1:fetch:fetch-1",
					Field:    "surface",
					Value:    "fetch",
				},
				{
					Source:   "static-graph",
					Path:     "mapper/static-analysis/js-bundle-dependency-graph.jsonl",
					RecordID: "2026-02-11/chart_a/resources/js/tradingPanel.js",
					Field:    "chunk_name",
					Value:    "tradingPanel",
				},
			},
		},
	}
	writeFixtureJSONL(t, filepath.Join(dateRoot, correlationInputRelative), correlations)

	t.Setenv("RESEARCHER_DATA_DIR", dataDir)
	var out bytes.Buffer
	if err := Run(context.Background(), &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := out.String(); got != "reporting: complete\n" {
		t.Fatalf("Run() output = %q, want %q", got, "reporting: complete\n")
	}

	matrixPath := filepath.Join(dateRoot, matrixOutputRelative)
	matrixData, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", matrixPath, err)
	}
	rows := decodeJSONLLines[capabilityMatrixRecord](t, matrixData)
	if len(rows) != 1 {
		t.Fatalf("matrix rows = %d, want 1", len(rows))
	}
	row := rows[0]
	if row.CapabilityID != "trading" {
		t.Fatalf("capability_id = %q, want trading", row.CapabilityID)
	}
	if row.Feature == "" || row.FeatureName == "" {
		t.Fatalf("feature fields must be populated: %+v", row)
	}
	if len(row.CandidateEntryModules) == 0 {
		t.Fatalf("candidate_entry_modules must be populated: %+v", row)
	}
	if row.ModuleRef == "" {
		t.Fatalf("module_ref must be populated: %+v", row)
	}
	if row.TraceID == "" {
		t.Fatalf("trace_id must be populated: %+v", row)
	}
	if row.Confidence <= 0 {
		t.Fatalf("confidence must be > 0: %+v", row)
	}
	if len(row.Preconditions) == 0 {
		t.Fatalf("preconditions must be populated: %+v", row)
	}
	if len(row.RiskNotes) == 0 {
		t.Fatalf("risk_notes must be populated: %+v", row)
	}
	if row.RecommendedControlMethod == "" || row.RecommendedControlPath == "" {
		t.Fatalf("control fields must be populated: %+v", row)
	}
	if len(row.EvidenceIDs) != 2 || row.EvidenceIDs[0] != "TRADING-E001" {
		t.Fatalf("evidence_ids unexpected: %+v", row.EvidenceIDs)
	}

	schemaPath := filepath.Join(dateRoot, matrixSchemaOutputPath)
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", schemaPath, err)
	}
	var schema matrixSchema
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		t.Fatalf("json.Unmarshal(schema) error = %v", err)
	}
	if schema.SchemaVersion != matrixSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", schema.SchemaVersion, matrixSchemaVersion)
	}
	if !contains(schema.Required, "capability_id") || !contains(schema.Required, "recommended_control_method") || !contains(schema.Required, "evidence_ids") {
		t.Fatalf("required fields missing in schema: %+v", schema.Required)
	}

	reportPath := filepath.Join(dateRoot, summaryReportOutputPath)
	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", reportPath, err)
	}
	report := string(reportData)
	if !strings.Contains(report, "[TRADING-E001](#evidence-trading-e001)") {
		t.Fatalf("summary report missing evidence link: %s", report)
	}
	if !strings.Contains(report, "<a id=\"evidence-trading-e001\"></a>") {
		t.Fatalf("summary report missing evidence anchor: %s", report)
	}
}

func TestRunContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out bytes.Buffer
	if err := Run(ctx, &out); err == nil {
		t.Fatal("Run() error = nil, want non-nil")
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

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
