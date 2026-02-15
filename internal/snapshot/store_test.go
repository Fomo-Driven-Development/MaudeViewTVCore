package snapshot

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeleteLogsImageCleanupFailureWhenImageMissing(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	store := &Store{dir: dir}
	id := "123e4567-e89b-12d3-a456-426614174000"
	jsonPath := filepath.Join(dir, id+".json")

	meta := SnapshotMeta{
		ID:     id,
		Format: "png",
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}
	if err := os.WriteFile(jsonPath, metaBytes, 0o644); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	var buf bytes.Buffer
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() {
		slog.SetDefault(oldLogger)
	})

	if err := store.Delete(id); err != nil {
		t.Fatalf("Delete() = %v; want nil", err)
	}

	if !strings.Contains(buf.String(), "snapshot image cleanup failed") {
		t.Fatalf("expected image cleanup debug log, got %q", buf.String())
	}
}
