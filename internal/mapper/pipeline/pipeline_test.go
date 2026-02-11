package pipeline

import (
	"bytes"
	"context"
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
