package staticanalysis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
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
