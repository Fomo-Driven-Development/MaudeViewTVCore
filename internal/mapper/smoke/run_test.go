package smoke

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunSmokeSequenceOnCapturedData(t *testing.T) {
	baseDir := t.TempDir()
	dataDir := filepath.Join(baseDir, "research_data")
	docsDir := filepath.Join(baseDir, "docs")
	dateRoot := filepath.Join(dataDir, "2026-02-11")

	jsPath := filepath.Join(dateRoot, "chart_test/resources/js/main.001122.js")
	if err := os.MkdirAll(filepath.Dir(jsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll(jsPath) error = %v", err)
	}
	if err := os.WriteFile(jsPath, []byte("function tradingPanel(){return true;}"), 0o644); err != nil {
		t.Fatalf("WriteFile(jsPath) error = %v", err)
	}

	httpPath := filepath.Join(dateRoot, "chart_test/http/TAB12345.jsonl")
	httpLine := `{"timestamp":"2026-02-11T16:48:33Z","request_id":"http-1","tab_id":"TAB12345","url":"https://www.tradingview.com/chart/test","method":"GET"}` + "\n"
	if err := os.MkdirAll(filepath.Dir(httpPath), 0o755); err != nil {
		t.Fatalf("MkdirAll(httpPath) error = %v", err)
	}
	if err := os.WriteFile(httpPath, []byte(httpLine), 0o644); err != nil {
		t.Fatalf("WriteFile(httpPath) error = %v", err)
	}

	wsPath := filepath.Join(dateRoot, "chart_test/websocket/TAB12345.jsonl")
	wsLine := `{"timestamp":"2026-02-11T16:48:34Z","request_id":"ws-1","tab_id":"TAB12345","url":"wss://prodata.tradingview.com/socket.io/websocket","event_type":"frame_received","direction":"incoming"}` + "\n"
	if err := os.MkdirAll(filepath.Dir(wsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll(wsPath) error = %v", err)
	}
	if err := os.WriteFile(wsPath, []byte(wsLine), 0o644); err != nil {
		t.Fatalf("WriteFile(wsPath) error = %v", err)
	}

	now := time.Date(2026, 2, 11, 17, 0, 0, 0, time.UTC)
	var out bytes.Buffer
	runtimeStage := func(_ context.Context, w io.Writer) error {
		_, _ = w.Write([]byte("runtime-probes: complete\n"))
		return nil
	}
	if err := runSmoke(context.Background(), &out, dataDir, docsDir, now, runtimeStage); err != nil {
		t.Fatalf("runSmoke() error = %v", err)
	}

	gotOut := out.String()
	for _, token := range []string{
		"static-analysis: complete\n",
		"runtime-probes: complete\n",
		"correlation: complete\n",
		"reporting: complete\n",
		"validation: complete\n",
		"smoke-test-report:",
	} {
		if !strings.Contains(gotOut, token) {
			t.Fatalf("runSmoke output missing %q in %q", token, gotOut)
		}
	}

	reportPath := filepath.Join(docsDir, "mapper-smoke-test-20260211T170000Z.md")
	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile(report) error = %v", err)
	}
	report := string(reportData)
	if !strings.Contains(report, "Run timestamp (UTC): `2026-02-11T17:00:00Z`") {
		t.Fatalf("report missing run timestamp: %s", report)
	}
	if !strings.Contains(report, "Bundle count:") {
		t.Fatalf("report missing bundle count: %s", report)
	}
	if !strings.Contains(report, "Trace session count:") {
		t.Fatalf("report missing trace session count: %s", report)
	}
	if !strings.Contains(report, "Correlated capability count:") {
		t.Fatalf("report missing correlated capability count: %s", report)
	}
	if !strings.Contains(report, "/mapper/runtime-probes/runtime-trace.jsonl") {
		t.Fatalf("report missing artifact path: %s", report)
	}

	for _, artifact := range []string{
		"mapper/static-analysis/js-bundle-index.jsonl",
		"mapper/static-analysis/js-bundle-analysis.jsonl",
		"mapper/static-analysis/js-bundle-dependency-graph.jsonl",
		"mapper/runtime-probes/runtime-trace.jsonl",
		"mapper/runtime-probes/trace-sessions.jsonl",
		"mapper/correlation/capability-correlations.jsonl",
		"mapper/reporting/capability-matrix.jsonl",
	} {
		path := filepath.Join(dateRoot, artifact)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %q: %v", path, err)
		}
	}
}
