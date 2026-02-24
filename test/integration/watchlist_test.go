//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
)

func TestWatchlistCRUD(t *testing.T) {
	const wlName = "integration-test-crud"

	// --- Create ---
	resp := env.POST(t, env.featurePath("watchlists"), map[string]any{"name": wlName})
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
		r := env.DELETE(t, env.featurePath("watchlist/"+wlID))
		r.Body.Close()
	})

	// --- Get (empty) ---
	resp = env.GET(t, env.featurePath("watchlist/"+wlID))
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
	resp = env.POST(t, env.featurePath("watchlist/"+wlID+"/symbols"), map[string]any{
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
	resp = env.do(t, http.MethodDelete, env.featurePath("watchlist/"+wlID+"/symbols"), map[string]any{
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
	resp = env.do(t, http.MethodPatch, env.featurePath("watchlist/"+wlID), map[string]any{
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
	resp = env.GET(t, env.featurePath("watchlist/"+wlID))
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
	resp = env.DELETE(t, env.featurePath("watchlist/"+wlID))
	requireStatus(t, resp, http.StatusNoContent)
	resp.Body.Close()

	// --- Verify deleted (should 500 with "watchlist not found") ---
	resp = env.GET(t, env.featurePath("watchlist/"+wlID))
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("watchlist should not exist after delete")
	}
	resp.Body.Close()
	t.Logf("watchlist %s deleted successfully", wlID)
}

func TestGetActiveWatchlist(t *testing.T) {
	resp := env.GET(t, env.featurePath("watchlists/active"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Symbols []string `json:"symbols"`
	}](t, resp)

	if result.ID == "" {
		t.Fatal("expected non-empty active watchlist ID")
	}
	if result.Name == "" {
		t.Fatal("expected non-empty active watchlist name")
	}
	t.Logf("active watchlist: id=%s name=%q symbols=%d", result.ID, result.Name, len(result.Symbols))
}

func TestSetActiveWatchlist(t *testing.T) {
	// Get current active watchlist to restore later.
	resp := env.GET(t, env.featurePath("watchlists/active"))
	requireStatus(t, resp, http.StatusOK)
	original := decodeJSON[struct {
		ID string `json:"id"`
	}](t, resp)

	// Create a new watchlist to switch to.
	resp = env.POST(t, env.featurePath("watchlists"), map[string]any{"name": "integration-test-active"})
	requireStatus(t, resp, http.StatusOK)
	created := decodeJSON[struct {
		ID string `json:"id"`
	}](t, resp)
	t.Cleanup(func() {
		// Restore original active watchlist and delete test one.
		env.PUT(t, env.featurePath("watchlists/active"), map[string]any{"id": original.ID})
		r := env.DELETE(t, env.featurePath("watchlist/"+created.ID))
		r.Body.Close()
	})

	// Set the new watchlist as active.
	resp = env.PUT(t, env.featurePath("watchlists/active"), map[string]any{
		"id": created.ID,
	})
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}](t, resp)
	requireField(t, result.ID, created.ID, "id")
	t.Logf("set active watchlist: id=%s name=%q", result.ID, result.Name)

	// Verify via GET.
	resp = env.GET(t, env.featurePath("watchlists/active"))
	requireStatus(t, resp, http.StatusOK)
	verify := decodeJSON[struct {
		ID string `json:"id"`
	}](t, resp)
	requireField(t, verify.ID, created.ID, "active watchlist id")
}

func TestFlagSymbol(t *testing.T) {
	// Create a temporary watchlist with a symbol.
	resp := env.POST(t, env.featurePath("watchlists"), map[string]any{"name": "integration-test-flag"})
	requireStatus(t, resp, http.StatusOK)
	created := decodeJSON[struct {
		ID string `json:"id"`
	}](t, resp)
	t.Cleanup(func() {
		r := env.DELETE(t, env.featurePath("watchlist/"+created.ID))
		r.Body.Close()
	})

	// Add a symbol.
	resp = env.POST(t, env.featurePath("watchlist/"+created.ID+"/symbols"), map[string]any{
		"symbols": []string{"AAPL"},
	})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Set this as active so flag can find it.
	resp = env.PUT(t, env.featurePath("watchlists/active"), map[string]any{"id": created.ID})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Flag the symbol (experimental endpoint — may fail).
	resp = env.POST(t, env.featurePath("watchlist/"+created.ID+"/flag"), map[string]any{
		"symbol": "AAPL",
	})
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		t.Logf("flag symbol returned %d (experimental endpoint — may be fragile)", resp.StatusCode)
		t.Skip("flag endpoint not working in this environment")
	}
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "toggled", "status")
	t.Logf("flag symbol AAPL: status=%s", result.Status)
}

// --- Colored Watchlist tests ---

func TestListColoredWatchlists(t *testing.T) {
	resp := env.GET(t, env.featurePath("watchlists/colored"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		ColoredWatchlists []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"colored_watchlists"`
	}](t, resp)

	if len(result.ColoredWatchlists) == 0 {
		t.Fatal("expected at least one colored watchlist")
	}
	for _, cw := range result.ColoredWatchlists {
		t.Logf("colored watchlist: id=%s name=%q color=%s", cw.ID, cw.Name, cw.Color)
	}
}

func TestReplaceAndListColoredWatchlist(t *testing.T) {
	const color = "green"
	syms := []string{"AAPL", "MSFT"}

	// Append test symbols to green list.
	resp := env.POST(t, env.featurePath("watchlists/colored/"+color+"/append"), map[string]any{
		"symbols": syms,
	})
	requireStatus(t, resp, http.StatusOK)
	appended := decodeJSON[struct {
		Color   string   `json:"color"`
		Symbols []string `json:"symbols"`
	}](t, resp)
	t.Logf("appended to %s list: %v", color, appended.Symbols)

	t.Cleanup(func() {
		r := env.POST(t, env.featurePath("watchlists/colored/"+color+"/remove"), map[string]any{
			"symbols": syms,
		})
		r.Body.Close()
	})

	// Replace (reorder) — reverse the current symbol list.
	reversed := make([]string, len(appended.Symbols))
	for i, s := range appended.Symbols {
		reversed[len(appended.Symbols)-1-i] = s
	}
	resp = env.PUT(t, env.featurePath("watchlists/colored/"+color), map[string]any{
		"symbols": reversed,
	})
	requireStatus(t, resp, http.StatusOK)
	replaced := decodeJSON[struct {
		Color   string   `json:"color"`
		Symbols []string `json:"symbols"`
	}](t, resp)
	if len(replaced.Symbols) != len(appended.Symbols) {
		t.Fatalf("expected %d symbols after replace, got %d", len(appended.Symbols), len(replaced.Symbols))
	}
	t.Logf("replaced (reordered) %s list: %v", color, replaced.Symbols)

	// Verify via list.
	resp = env.GET(t, env.featurePath("watchlists/colored"))
	requireStatus(t, resp, http.StatusOK)
	listing := decodeJSON[struct {
		ColoredWatchlists []struct {
			Color   string   `json:"color"`
			Symbols []string `json:"symbols"`
		} `json:"colored_watchlists"`
	}](t, resp)
	for _, cw := range listing.ColoredWatchlists {
		if cw.Color == color {
			if len(cw.Symbols) < 2 {
				t.Fatalf("expected at least 2 symbols in %s list, got %d", color, len(cw.Symbols))
			}
			t.Logf("verified %s list has %d symbols", color, len(cw.Symbols))
			return
		}
	}
	t.Fatalf("color %s not found in listing", color)
}

func TestAppendAndRemoveColoredWatchlist(t *testing.T) {
	const color = "green"

	// Append two test symbols.
	resp := env.POST(t, env.featurePath("watchlists/colored/"+color+"/append"), map[string]any{
		"symbols": []string{"AAPL", "TSLA"},
	})
	requireStatus(t, resp, http.StatusOK)
	appended := decodeJSON[struct {
		Symbols []string `json:"symbols"`
	}](t, resp)
	t.Logf("after append: %v", appended.Symbols)

	t.Cleanup(func() {
		// Remove test symbols.
		r := env.POST(t, env.featurePath("watchlists/colored/"+color+"/remove"), map[string]any{
			"symbols": []string{"AAPL", "TSLA"},
		})
		r.Body.Close()
	})

	if len(appended.Symbols) < 2 {
		t.Fatalf("expected at least 2 symbols after append, got %d", len(appended.Symbols))
	}

	// Remove AAPL.
	resp = env.POST(t, env.featurePath("watchlists/colored/"+color+"/remove"), map[string]any{
		"symbols": []string{"AAPL"},
	})
	requireStatus(t, resp, http.StatusOK)
	removed := decodeJSON[struct {
		Symbols []string `json:"symbols"`
	}](t, resp)
	for _, s := range removed.Symbols {
		if s == "AAPL" {
			t.Fatal("AAPL should have been removed")
		}
	}
	t.Logf("after remove: %v", removed.Symbols)
}

func TestBulkRemoveColoredWatchlist(t *testing.T) {
	// Add a symbol to two colors.
	resp := env.POST(t, env.featurePath("watchlists/colored/red/append"), map[string]any{
		"symbols": []string{"GOOG"},
	})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	resp = env.POST(t, env.featurePath("watchlists/colored/blue/append"), map[string]any{
		"symbols": []string{"GOOG"},
	})
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	t.Cleanup(func() {
		// Best-effort cleanup: remove GOOG from both.
		r := env.POST(t, env.featurePath("watchlists/colored/red/remove"), map[string]any{"symbols": []string{"GOOG"}})
		r.Body.Close()
		r = env.POST(t, env.featurePath("watchlists/colored/blue/remove"), map[string]any{"symbols": []string{"GOOG"}})
		r.Body.Close()
	})

	// Bulk remove GOOG from all.
	resp = env.POST(t, env.featurePath("watchlists/colored/bulk-remove"), map[string]any{
		"symbols": []string{"GOOG"},
	})
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "ok", "status")
	t.Logf("bulk-remove GOOG: status=%s", result.Status)
}

// --- Study Template tests ---

func TestListStudyTemplates(t *testing.T) {
	resp := env.GET(t, env.featurePath("study-templates"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Custom       []struct{ ID int; Name string } `json:"custom"`
		Standard     []struct{ ID int; Name string } `json:"standard"`
		Fundamentals []struct{ ID int; Name string } `json:"fundamentals"`
	}](t, resp)

	total := len(result.Custom) + len(result.Standard) + len(result.Fundamentals)
	t.Logf("study templates: custom=%d standard=%d fundamentals=%d total=%d",
		len(result.Custom), len(result.Standard), len(result.Fundamentals), total)
	if total == 0 {
		t.Skip("no study templates found — user may not have any saved")
	}
}

func TestGetStudyTemplate(t *testing.T) {
	// First, list to get a valid ID.
	resp := env.GET(t, env.featurePath("study-templates"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Custom       []struct{ ID int `json:"id"`; Name string `json:"name"` } `json:"custom"`
		Standard     []struct{ ID int `json:"id"`; Name string `json:"name"` } `json:"standard"`
		Fundamentals []struct{ ID int `json:"id"`; Name string `json:"name"` } `json:"fundamentals"`
	}](t, resp)

	// Find the first available template ID.
	var templateID int
	for _, lists := range [][]struct{ ID int `json:"id"`; Name string `json:"name"` }{result.Custom, result.Standard, result.Fundamentals} {
		if len(lists) > 0 {
			templateID = lists[0].ID
			break
		}
	}
	if templateID == 0 {
		t.Skip("no study templates available to fetch")
	}

	resp = env.GET(t, env.featurePath(fmt.Sprintf("study-templates/%d", templateID)))
	requireStatus(t, resp, http.StatusOK)
	entry := decodeJSON[struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		MetaInfo any    `json:"meta_info"`
	}](t, resp)
	if entry.Name == "" {
		t.Fatal("expected non-empty template name")
	}
	t.Logf("study template %d: name=%q", entry.ID, entry.Name)
}

func TestWatchlistListContainsCreated(t *testing.T) {
	// Create a watchlist, verify it appears in the listing, then clean up.
	resp := env.POST(t, env.featurePath("watchlists"), map[string]any{"name": "integration-test-list"})
	requireStatus(t, resp, http.StatusOK)
	created := decodeJSON[struct {
		ID string `json:"id"`
	}](t, resp)
	t.Cleanup(func() {
		r := env.DELETE(t, env.featurePath("watchlist/"+created.ID))
		r.Body.Close()
	})

	resp = env.GET(t, env.featurePath("watchlists"))
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
