package pipeline

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunModes(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{name: "static", mode: ModeStaticOnly, want: "static-analysis: complete\n"},
		{name: "runtime", mode: ModeRuntimeOnly, want: "runtime-probes: complete\n"},
		{name: "correlate", mode: ModeCorrelate, want: "correlation: complete\n"},
		{name: "report", mode: ModeReport, want: "reporting: complete\n"},
		{name: "validate", mode: ModeValidate, want: "validation: complete\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RESEARCHER_DATA_DIR", filepath.Join(t.TempDir(), "research_data"))
			var out bytes.Buffer
			if err := Run(context.Background(), &out, tc.mode); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if got := out.String(); got != tc.want {
				t.Fatalf("Run() output = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRunFullModeOrder(t *testing.T) {
	t.Setenv("RESEARCHER_DATA_DIR", filepath.Join(t.TempDir(), "research_data"))
	var out bytes.Buffer
	if err := Run(context.Background(), &out, ModeFull); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := out.String()
	wantOrder := []string{
		"static-analysis: complete\n",
		"runtime-probes: complete\n",
		"correlation: complete\n",
		"reporting: complete\n",
	}

	start := 0
	for _, fragment := range wantOrder {
		next := strings.Index(got[start:], fragment)
		if next < 0 {
			t.Fatalf("full mode output missing %q in %q", fragment, got)
		}
		start += next + len(fragment)
	}
}

func TestRunSmokeMode(t *testing.T) {
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

	t.Setenv("RESEARCHER_DATA_DIR", dataDir)
	t.Setenv("RESEARCHER_DOCS_DIR", docsDir)
	var out bytes.Buffer
	if err := Run(context.Background(), &out, ModeSmoke); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	got := out.String()
	for _, token := range []string{
		"static-analysis: complete\n",
		"runtime-probes: complete\n",
		"correlation: complete\n",
		"reporting: complete\n",
		"validation: complete\n",
		"smoke-test-report:",
	} {
		if !strings.Contains(got, token) {
			t.Fatalf("smoke mode output missing %q in %q", token, got)
		}
	}
}

func TestRunUnknownMode(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), &out, "nope")
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
}

func TestRunContextCancelled(t *testing.T) {
	t.Setenv("RESEARCHER_DATA_DIR", filepath.Join(t.TempDir(), "research_data"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out bytes.Buffer
	err := Run(ctx, &out, ModeStaticOnly)
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
}
