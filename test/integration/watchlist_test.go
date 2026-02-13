//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestWatchlistCRUD(t *testing.T) {
	const wlName = "integration-test-crud"

	// --- Create ---
	resp := env.POST(t, "/api/v1/watchlists", map[string]any{"name": wlName})
	requireStatus(t, resp, http.StatusOK)
	created := decodeJSON[struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Count int    `json:"count"`
	}](t, resp)
	requireField(t, created.Name, wlName, "name")
	if created.ID == "" {
		t.Fatal("created watchlist has empty ID")
	}
	wlID := created.ID
	t.Logf("created watchlist %s (id=%s)", created.Name, wlID)

	// Ensure cleanup even if test fails partway through.
	t.Cleanup(func() {
		r := env.DELETE(t, "/api/v1/watchlist/"+wlID)
		r.Body.Close()
	})

	// --- Get (empty) ---
	resp = env.GET(t, "/api/v1/watchlist/"+wlID)
	requireStatus(t, resp, http.StatusOK)
	detail := decodeJSON[struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Symbols []string `json:"symbols"`
	}](t, resp)
	requireField(t, detail.ID, wlID, "id")
	if len(detail.Symbols) != 0 {
		t.Fatalf("new watchlist should be empty, got %d symbols", len(detail.Symbols))
	}

	// --- Add symbols ---
	resp = env.POST(t, "/api/v1/watchlist/"+wlID+"/symbols", map[string]any{
		"symbols": []string{"AAPL", "MSFT", "TSLA"},
	})
	requireStatus(t, resp, http.StatusOK)
	added := decodeJSON[struct {
		Symbols []string `json:"symbols"`
	}](t, resp)
	if len(added.Symbols) != 3 {
		t.Fatalf("expected 3 symbols after add, got %d: %v", len(added.Symbols), added.Symbols)
	}
	t.Logf("added symbols: %v", added.Symbols)

	// --- Remove one symbol (DELETE with body) ---
	resp = env.do(t, http.MethodDelete, "/api/v1/watchlist/"+wlID+"/symbols", map[string]any{
		"symbols": []string{"MSFT"},
	})
	requireStatus(t, resp, http.StatusOK)
	removed := decodeJSON[struct {
		Symbols []string `json:"symbols"`
	}](t, resp)
	if len(removed.Symbols) != 2 {
		t.Fatalf("expected 2 symbols after remove, got %d: %v", len(removed.Symbols), removed.Symbols)
	}
	// Verify MSFT is gone.
	for _, s := range removed.Symbols {
		if s == "MSFT" {
			t.Fatal("MSFT should have been removed")
		}
	}
	t.Logf("after remove: %v", removed.Symbols)

	// --- Rename ---
	const newName = "integration-test-renamed"
	resp = env.do(t, http.MethodPatch, "/api/v1/watchlist/"+wlID, map[string]any{
		"name": newName,
	})
	requireStatus(t, resp, http.StatusOK)
	renamed := decodeJSON[struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	}](t, resp)
	requireField(t, renamed.Name, newName, "name")
	requireField(t, renamed.Count, 2, "count")

	// --- Verify via Get ---
	resp = env.GET(t, "/api/v1/watchlist/"+wlID)
	requireStatus(t, resp, http.StatusOK)
	final := decodeJSON[struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Symbols []string `json:"symbols"`
	}](t, resp)
	requireField(t, final.Name, newName, "name")
	if len(final.Symbols) != 2 {
		t.Fatalf("expected 2 symbols in final get, got %d", len(final.Symbols))
	}

	// --- Delete ---
	resp = env.DELETE(t, "/api/v1/watchlist/"+wlID)
	requireStatus(t, resp, http.StatusNoContent)
	resp.Body.Close()

	// --- Verify deleted (should 500 with "watchlist not found") ---
	resp = env.GET(t, "/api/v1/watchlist/"+wlID)
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("watchlist should not exist after delete")
	}
	resp.Body.Close()
	t.Logf("watchlist %s deleted successfully", wlID)
}

func TestWatchlistListContainsCreated(t *testing.T) {
	// Create a watchlist, verify it appears in the listing, then clean up.
	resp := env.POST(t, "/api/v1/watchlists", map[string]any{"name": "integration-test-list"})
	requireStatus(t, resp, http.StatusOK)
	created := decodeJSON[struct {
		ID string `json:"id"`
	}](t, resp)
	t.Cleanup(func() {
		r := env.DELETE(t, "/api/v1/watchlist/"+created.ID)
		r.Body.Close()
	})

	resp = env.GET(t, "/api/v1/watchlists")
	requireStatus(t, resp, http.StatusOK)
	listing := decodeJSON[struct {
		Watchlists []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"watchlists"`
	}](t, resp)

	found := false
	for _, wl := range listing.Watchlists {
		if wl.ID == created.ID {
			found = true
			requireField(t, wl.Name, "integration-test-list", "name")
			break
		}
	}
	if !found {
		t.Fatalf("created watchlist %s not found in listing", created.ID)
	}
}
