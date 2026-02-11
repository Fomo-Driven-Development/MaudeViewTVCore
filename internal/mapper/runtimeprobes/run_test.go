package runtimeprobes

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

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
				},
			}, nil
		case target.ID("tab-2"):
			return probeLifecycle{
				Attached: true,
				Result: probeBootstrapResult{
					AlreadyInjected: true,
					URL:             tab.URL,
				},
			}, nil
		default:
			return probeLifecycle{}, errors.New("unexpected tab")
		}
	}

	attached, injected, err := bootstrapTargets(context.Background(), logger, targets, runner)
	if err != nil {
		t.Fatalf("bootstrapTargets() error = %v", err)
	}
	if attached != 2 {
		t.Fatalf("bootstrapTargets() attached = %d, want 2", attached)
	}
	if injected != 2 {
		t.Fatalf("bootstrapTargets() injected = %d, want 2", injected)
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
