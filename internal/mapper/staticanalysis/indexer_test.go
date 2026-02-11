package staticanalysis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestIndexJSBundlesWritesDeterministicMetadata(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "research_data")

	jsFiles := map[string]string{
		"2026-02-11/chart_a/resources/js/12345.a1b2c3.js":       "alpha",
		"2026-02-11/chart_b/resources/js/main.chunk.8899aa.js":  "beta",
		"2026-02-12/chart_c/resources/js/runtime.fedcba9876.js": "gamma",
	}

	for relPath, content := range jsFiles {
		fullPath := filepath.Join(dataDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}
	ignoredPath := filepath.Join(dataDir, "2026-02-11/chart_a/resources/js/not-a-bundle.txt")
	if err := os.WriteFile(ignoredPath, []byte("ignore"), 0o644); err != nil {
		t.Fatalf("WriteFile() ignored file error = %v", err)
	}

	ctx := context.Background()
	if err := indexJSBundles(ctx, dataDir); err != nil {
		t.Fatalf("indexJSBundles() first run error = %v", err)
	}

	firstRunByDate := map[string][]byte{
		"2026-02-11": mustReadFile(t, filepath.Join(dataDir, "2026-02-11", indexRelativeOutput)),
		"2026-02-12": mustReadFile(t, filepath.Join(dataDir, "2026-02-12", indexRelativeOutput)),
	}

	if err := indexJSBundles(ctx, dataDir); err != nil {
		t.Fatalf("indexJSBundles() second run error = %v", err)
	}

	secondRunByDate := map[string][]byte{
		"2026-02-11": mustReadFile(t, filepath.Join(dataDir, "2026-02-11", indexRelativeOutput)),
		"2026-02-12": mustReadFile(t, filepath.Join(dataDir, "2026-02-12", indexRelativeOutput)),
	}

	for date := range firstRunByDate {
		if string(firstRunByDate[date]) != string(secondRunByDate[date]) {
			t.Fatalf("index output for %s changed between runs", date)
		}
	}

	recordsByDate := map[string][]jsBundleRecord{
		"2026-02-11": parseJSONLRecords(t, secondRunByDate["2026-02-11"]),
		"2026-02-12": parseJSONLRecords(t, secondRunByDate["2026-02-12"]),
	}

	if got := len(recordsByDate["2026-02-11"]); got != 2 {
		t.Fatalf("2026-02-11 record count = %d, want 2", got)
	}
	if got := len(recordsByDate["2026-02-12"]); got != 1 {
		t.Fatalf("2026-02-12 record count = %d, want 1", got)
	}

	seen := make(map[string]struct{})
	for _, records := range recordsByDate {
		for _, rec := range records {
			if _, ok := seen[rec.PrimaryKey]; ok {
				t.Fatalf("duplicate primary key found in output: %q", rec.PrimaryKey)
			}
			seen[rec.PrimaryKey] = struct{}{}
		}
	}

	checkRecord(t, recordsByDate["2026-02-11"], jsBundleRecord{
		PrimaryKey: "2026-02-11/chart_a/resources/js/12345.a1b2c3.js",
		FilePath:   "2026-02-11/chart_a/resources/js/12345.a1b2c3.js",
		SizeBytes:  int64(len("alpha")),
		SHA256:     sha256Hex("alpha"),
		ChunkName:  "12345",
	})
	checkRecord(t, recordsByDate["2026-02-11"], jsBundleRecord{
		PrimaryKey: "2026-02-11/chart_b/resources/js/main.chunk.8899aa.js",
		FilePath:   "2026-02-11/chart_b/resources/js/main.chunk.8899aa.js",
		SizeBytes:  int64(len("beta")),
		SHA256:     sha256Hex("beta"),
		ChunkName:  "main",
	})
	checkRecord(t, recordsByDate["2026-02-12"], jsBundleRecord{
		PrimaryKey: "2026-02-12/chart_c/resources/js/runtime.fedcba9876.js",
		FilePath:   "2026-02-12/chart_c/resources/js/runtime.fedcba9876.js",
		SizeBytes:  int64(len("gamma")),
		SHA256:     sha256Hex("gamma"),
		ChunkName:  "runtime",
	})
}

func TestValidateNoDuplicatePrimaryKeys(t *testing.T) {
	err := validateNoDuplicatePrimaryKeys([]jsBundleRecord{
		{PrimaryKey: "dup"},
		{PrimaryKey: "dup"},
	})
	if err == nil {
		t.Fatal("validateNoDuplicatePrimaryKeys() error = nil, want non-nil")
	}
}

func TestIndexJSBundlesWritesAnalysisAndErrorArtifacts(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "research_data")

	validRelPath := "2026-02-11/chart_a/resources/js/main.001122.js"
	validJS := strings.Join([]string{
		`import apiClient from "./api/client";`,
		`const lazy = import("./lazy/chunk.js");`,
		`const util = require("./util");`,
		`function bootstrap() { return "/api/v1/orders"; }`,
		`const emitAction = () => "chart:event:update";`,
		`class ChartEngine {}`,
		`export { bootstrap, ChartEngine };`,
		`exports.emitAction = emitAction;`,
		`const wsChannel = "wss://feed.example/ws/prices";`,
		`const flag = "FEATURE_TRADING_PANEL";`,
	}, "\n")

	invalidRelPath := "2026-02-11/chart_a/resources/js/broken.334455.js"
	invalidJS := `function broken( {`

	if err := writeFixtureJS(dataDir, validRelPath, validJS); err != nil {
		t.Fatalf("writeFixtureJS(valid) error = %v", err)
	}
	if err := writeFixtureJS(dataDir, invalidRelPath, invalidJS); err != nil {
		t.Fatalf("writeFixtureJS(invalid) error = %v", err)
	}

	if err := indexJSBundles(context.Background(), dataDir); err != nil {
		t.Fatalf("indexJSBundles() error = %v", err)
	}

	analysisPath := filepath.Join(dataDir, "2026-02-11", analysisRelativeOutput)
	errorPath := filepath.Join(dataDir, "2026-02-11", errorsRelativeOutput)

	analysisRecords := parseJSONLAnalysisRecords(t, mustReadFile(t, analysisPath))
	if got := len(analysisRecords); got != 1 {
		t.Fatalf("analysis record count = %d, want 1", got)
	}

	rec := analysisRecords[0]
	if rec.PrimaryKey != validRelPath {
		t.Fatalf("analysis primary key = %q, want %q", rec.PrimaryKey, validRelPath)
	}
	assertContains(t, rec.Functions, "bootstrap")
	assertContains(t, rec.Functions, "emitAction")
	assertContains(t, rec.Classes, "ChartEngine")
	assertContains(t, rec.Exports, "ChartEngine")
	assertContains(t, rec.Exports, "bootstrap")
	assertContains(t, rec.Exports, "emitAction")
	assertContains(t, rec.ImportEdges, "./api/client")
	assertContains(t, rec.ImportEdges, "./lazy/chunk.js")
	assertContains(t, rec.RequireEdges, "./util")
	assertContainsAnchor(t, rec.SignalAnchors, jsSignalAnchor{Type: "api_route", Value: "/api/v1/orders"})
	assertContainsAnchor(t, rec.SignalAnchors, jsSignalAnchor{Type: "websocket_channel", Value: "wss://feed.example/ws/prices"})
	assertContainsAnchor(t, rec.SignalAnchors, jsSignalAnchor{Type: "action_event", Value: "chart:event:update"})
	assertContainsAnchor(t, rec.SignalAnchors, jsSignalAnchor{Type: "feature_flag", Value: "FEATURE_TRADING_PANEL"})

	errorRecords := parseJSONLErrorRecords(t, mustReadFile(t, errorPath))
	if got := len(errorRecords); got != 1 {
		t.Fatalf("analysis error record count = %d, want 1", got)
	}
	if errorRecords[0].PrimaryKey != invalidRelPath {
		t.Fatalf("error primary key = %q, want %q", errorRecords[0].PrimaryKey, invalidRelPath)
	}
	if errorRecords[0].Error == "" {
		t.Fatal("expected non-empty parse error message")
	}
}

func parseJSONLRecords(t *testing.T, data []byte) []jsBundleRecord {
	t.Helper()
	lines := slices.DeleteFunc(splitTrimLines(string(data)), func(line string) bool { return line == "" })
	records := make([]jsBundleRecord, 0, len(lines))
	for _, line := range lines {
		var rec jsBundleRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		records = append(records, rec)
	}
	return records
}

func parseJSONLAnalysisRecords(t *testing.T, data []byte) []jsBundleAnalysisRecord {
	t.Helper()
	lines := slices.DeleteFunc(splitTrimLines(string(data)), func(line string) bool { return line == "" })
	records := make([]jsBundleAnalysisRecord, 0, len(lines))
	for _, line := range lines {
		var rec jsBundleAnalysisRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("json.Unmarshal() analysis error = %v", err)
		}
		records = append(records, rec)
	}
	return records
}

func parseJSONLErrorRecords(t *testing.T, data []byte) []jsBundleAnalysisErrorRecord {
	t.Helper()
	lines := slices.DeleteFunc(splitTrimLines(string(data)), func(line string) bool { return line == "" })
	records := make([]jsBundleAnalysisErrorRecord, 0, len(lines))
	for _, line := range lines {
		var rec jsBundleAnalysisErrorRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("json.Unmarshal() parse error record error = %v", err)
		}
		records = append(records, rec)
	}
	return records
}

func splitTrimLines(data string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

func checkRecord(t *testing.T, records []jsBundleRecord, want jsBundleRecord) {
	t.Helper()
	for _, rec := range records {
		if rec.PrimaryKey == want.PrimaryKey {
			if rec != want {
				t.Fatalf("record mismatch for %q: got %+v want %+v", want.PrimaryKey, rec, want)
			}
			return
		}
	}
	t.Fatalf("missing record %q", want.PrimaryKey)
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return data
}

func sha256Hex(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func writeFixtureJS(dataDir, relPath, content string) error {
	fullPath := filepath.Join(dataDir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0o644)
}

func assertContains(t *testing.T, values []string, want string) {
	t.Helper()
	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("missing value %q in %v", want, values)
}

func assertContainsAnchor(t *testing.T, anchors []jsSignalAnchor, want jsSignalAnchor) {
	t.Helper()
	for _, anchor := range anchors {
		if anchor == want {
			return
		}
	}
	t.Fatalf("missing anchor %+v in %+v", want, anchors)
}
