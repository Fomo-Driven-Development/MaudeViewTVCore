//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
)

func TestListSnapshots(t *testing.T) {
	resp := env.GET(t, "/api/v1/snapshots")
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Snapshots []map[string]any `json:"snapshots"`
	}](t, resp)

	// Snapshots list may be empty — that's OK.
	t.Logf("snapshots count: %d", len(result.Snapshots))
}

func TestSnapshotLifecycle(t *testing.T) {
	// 1. Take a chart snapshot to create one.
	resp := env.POST(t, env.chartPath("snapshot"), map[string]any{
		"format": "png",
	})
	requireStatus(t, resp, http.StatusOK)
	created := decodeJSON[struct {
		Snapshot struct {
			ID string `json:"id"`
		} `json:"snapshot"`
		URL string `json:"url"`
	}](t, resp)

	if created.Snapshot.ID == "" {
		t.Fatal("expected snapshot ID after creation")
	}
	snapshotID := created.Snapshot.ID
	t.Logf("created snapshot: id=%s url=%s", snapshotID, created.URL)

	// Register cleanup.
	t.Cleanup(func() {
		r := env.DELETE(t, "/api/v1/snapshots/"+snapshotID)
		r.Body.Close()
	})

	// 2. List snapshots — should contain our new snapshot.
	resp = env.GET(t, "/api/v1/snapshots")
	requireStatus(t, resp, http.StatusOK)
	listing := decodeJSON[struct {
		Snapshots []struct {
			ID string `json:"id"`
		} `json:"snapshots"`
	}](t, resp)

	found := false
	for _, s := range listing.Snapshots {
		if s.ID == snapshotID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("snapshot %s not found in listing", snapshotID)
	}
	t.Logf("snapshot found in listing (%d total)", len(listing.Snapshots))

	// 3. Get snapshot metadata.
	resp = env.GET(t, fmt.Sprintf("/api/v1/snapshots/%s/metadata", snapshotID))
	requireStatus(t, resp, http.StatusOK)
	meta := decodeJSON[map[string]any](t, resp)
	if len(meta) == 0 {
		t.Fatal("expected snapshot metadata")
	}
	t.Logf("snapshot metadata keys: %d", len(meta))

	// 4. Delete the snapshot.
	resp = env.DELETE(t, "/api/v1/snapshots/"+snapshotID)
	requireStatus(t, resp, http.StatusOK)
	deleteResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, deleteResult.Status, "deleted", "status")
	t.Logf("deleted snapshot %s", snapshotID)

	// 5. Verify it's gone from the listing.
	resp = env.GET(t, "/api/v1/snapshots")
	requireStatus(t, resp, http.StatusOK)
	afterListing := decodeJSON[struct {
		Snapshots []struct {
			ID string `json:"id"`
		} `json:"snapshots"`
	}](t, resp)
	for _, s := range afterListing.Snapshots {
		if s.ID == snapshotID {
			t.Fatalf("snapshot %s should not exist after delete", snapshotID)
		}
	}
}
