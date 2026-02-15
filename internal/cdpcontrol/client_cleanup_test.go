package cdpcontrol

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/target"
)

func TestCleanupLockedLogsDetachFailure(t *testing.T) {
	var buf bytes.Buffer
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() {
		slog.SetDefault(oldLogger)
	})

	client := &Client{
		cdp: &rawCDP{},
		tabs: map[target.ID]*tabSession{
			"target-1": {
				sessionID: "session-1",
			},
		},
	}
	client.cleanupLocked()

	if !strings.Contains(buf.String(), "detach cleanup failed") {
		t.Fatalf("expected detach cleanup debug log, got %q", buf.String())
	}
}
