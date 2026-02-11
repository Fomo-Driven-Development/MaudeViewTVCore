package staticanalysis

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update", false, "update golden fixtures")

type indexerGoldenOutputs struct {
	Index    []jsBundleRecord              `json:"index"`
	Analysis []jsBundleAnalysisRecord      `json:"analysis"`
	Errors   []jsBundleAnalysisErrorRecord `json:"errors"`
	Graph    []jsBundleGraphNode           `json:"graph"`
}

func TestParseIndexedBundleSourceGolden(t *testing.T) {
	sourcePath := filepath.Join("testdata", "bundles", "extractor_representative.min.js")
	sourceBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
	}

	got, err := parseIndexedBundleSource(string(sourceBytes))
	if err != nil {
		t.Fatalf("parseIndexedBundleSource() error = %v", err)
	}

	assertGoldenJSON(t, filepath.Join("testdata", "golden", "extractor_representative.json"), got)
}

func TestBuildBundleDependencyGraphGolden(t *testing.T) {
	mainPath := "2026-02-11/chart_a/resources/js/main.001122.js"
	mainSourcePath := filepath.Join("testdata", "indexer_input", filepath.FromSlash(mainPath))
	mainSource, err := os.ReadFile(mainSourcePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", mainSourcePath, err)
	}

	extracted, err := parseIndexedBundleSource(string(mainSource))
	if err != nil {
		t.Fatalf("parseIndexedBundleSource(main) error = %v", err)
	}

	indexRecords := []jsBundleRecord{
		{PrimaryKey: mainPath, FilePath: mainPath, ChunkName: "main"},
		{PrimaryKey: "2026-02-11/chart_a/resources/js/chartView.js", FilePath: "2026-02-11/chart_a/resources/js/chartView.js", ChunkName: "chartView"},
		{PrimaryKey: "2026-02-11/chart_a/resources/js/tradingPanel.js", FilePath: "2026-02-11/chart_a/resources/js/tradingPanel.js", ChunkName: "tradingPanel"},
	}
	analysisRecords := []jsBundleAnalysisRecord{{
		PrimaryKey:    mainPath,
		FilePath:      mainPath,
		ChunkName:     "main",
		Functions:     extracted.Functions,
		Classes:       extracted.Classes,
		Exports:       extracted.Exports,
		ImportEdges:   extracted.ImportEdges,
		RequireEdges:  extracted.RequireEdges,
		SignalAnchors: extracted.Anchors,
	}}

	got := buildBundleDependencyGraph(indexRecords, analysisRecords)
	assertGoldenJSON(t, filepath.Join("testdata", "golden", "graph_builder_representative.json"), got)
}

func TestIndexJSBundlesGolden(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "research_data")
	inputRoot := filepath.Join("testdata", "indexer_input")
	if err := copyTree(inputRoot, dataDir); err != nil {
		t.Fatalf("copyTree(%q, %q) error = %v", inputRoot, dataDir, err)
	}

	if err := indexJSBundles(context.Background(), dataDir); err != nil {
		t.Fatalf("indexJSBundles() error = %v", err)
	}

	outputRoot := filepath.Join(dataDir, "2026-02-11")
	got := indexerGoldenOutputs{
		Index:    decodeJSONLLines[jsBundleRecord](t, mustReadFile(t, filepath.Join(outputRoot, indexRelativeOutput))),
		Analysis: decodeJSONLLines[jsBundleAnalysisRecord](t, mustReadFile(t, filepath.Join(outputRoot, analysisRelativeOutput))),
		Errors:   decodeJSONLLines[jsBundleAnalysisErrorRecord](t, mustReadFile(t, filepath.Join(outputRoot, errorsRelativeOutput))),
		Graph:    decodeJSONLLines[jsBundleGraphNode](t, mustReadFile(t, filepath.Join(outputRoot, graphRelativeOutput))),
	}

	assertGoldenJSON(t, filepath.Join("testdata", "golden", "indexer_outputs_representative.json"), got)
}

func decodeJSONLLines[T any](t *testing.T, data []byte) []T {
	t.Helper()
	lines := slices.DeleteFunc(strings.Split(string(data), "\n"), func(line string) bool { return strings.TrimSpace(line) == "" })
	out := make([]T, 0, len(lines))
	for _, line := range lines {
		var item T
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			t.Fatalf("json.Unmarshal(%q) error = %v", line, err)
		}
		out = append(out, item)
	}
	return out
}

func assertGoldenJSON(t *testing.T, goldenPath string, got any) {
	t.Helper()
	gotBytes, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent() error = %v", err)
	}
	gotBytes = append(gotBytes, '\n')

	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(goldenPath), err)
		}
		if err := os.WriteFile(goldenPath, gotBytes, 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", goldenPath, err)
		}
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", goldenPath, err)
	}
	if string(wantBytes) != string(gotBytes) {
		t.Fatalf("golden mismatch for %s\nre-run with: go test ./internal/mapper/staticanalysis -run %s -update", goldenPath, t.Name())
	}
}

func copyTree(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstDir, relPath)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0o644)
	})
}
