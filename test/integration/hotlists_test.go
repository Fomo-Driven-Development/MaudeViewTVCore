//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestHotlistsProbe(t *testing.T) {
	resp := env.GET(t, "/api/v1/hotlists/probe")
	requireStatus(t, resp, http.StatusOK)

	var probe struct {
		Found       bool     `json:"found"`
		AccessPaths []string `json:"access_paths"`
		Methods     []string `json:"methods"`
	}
	probe = decodeJSON[struct {
		Found       bool     `json:"found"`
		AccessPaths []string `json:"access_paths"`
		Methods     []string `json:"methods"`
	}](t, resp)
	t.Logf("hotlists probe: found=%v methods=%d access_paths=%v", probe.Found, len(probe.Methods), probe.AccessPaths)

	if !probe.Found {
		t.Skip("hotlistsManager not found — may not be loaded yet")
	}
}

func TestHotlistsProbeDeep(t *testing.T) {
	resp := env.GET(t, "/api/v1/hotlists/probe/deep")
	requireStatus(t, resp, http.StatusOK)

	var result map[string]any
	result = decodeJSON[map[string]any](t, resp)
	t.Logf("hotlists deep probe keys: %v", mapKeys(result))
}

func TestHotlistsMarkets(t *testing.T) {
	resp := env.GET(t, "/api/v1/hotlists/markets")
	requireStatus(t, resp, http.StatusOK)

	var result struct {
		Markets any `json:"markets"`
	}
	result = decodeJSON[struct {
		Markets any `json:"markets"`
	}](t, resp)
	t.Logf("hotlists markets: %v", result.Markets)
}

func TestHotlistsExchanges(t *testing.T) {
	resp := env.GET(t, "/api/v1/hotlists/exchanges")
	requireStatus(t, resp, http.StatusOK)

	var result struct {
		Exchanges []struct {
			Exchange string `json:"exchange"`
			Name     string `json:"name"`
			Groups   []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"groups"`
		} `json:"exchanges"`
	}
	result = decodeJSON[struct {
		Exchanges []struct {
			Exchange string `json:"exchange"`
			Name     string `json:"name"`
			Groups   []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"groups"`
		} `json:"exchanges"`
	}](t, resp)
	t.Logf("hotlists exchanges: count=%d", len(result.Exchanges))
	for i, ex := range result.Exchanges {
		if i < 5 {
			t.Logf("  exchange=%s name=%s groups=%d", ex.Exchange, ex.Name, len(ex.Groups))
		}
	}
}

func TestHotlistsGetOne(t *testing.T) {
	// First discover an exchange and group
	resp := env.GET(t, "/api/v1/hotlists/exchanges")
	requireStatus(t, resp, http.StatusOK)

	var exchanges struct {
		Exchanges []struct {
			Exchange string `json:"exchange"`
			Groups   []struct {
				ID string `json:"id"`
			} `json:"groups"`
		} `json:"exchanges"`
	}
	exchanges = decodeJSON[struct {
		Exchanges []struct {
			Exchange string `json:"exchange"`
			Groups   []struct {
				ID string `json:"id"`
			} `json:"groups"`
		} `json:"exchanges"`
	}](t, resp)

	if len(exchanges.Exchanges) == 0 {
		t.Skip("no exchanges available")
	}

	ex := exchanges.Exchanges[0]
	if len(ex.Groups) == 0 {
		t.Skip("no groups for first exchange")
	}

	group := ex.Groups[0].ID
	path := fmt.Sprintf("/api/v1/hotlists/%s/%s", ex.Exchange, group)
	resp2 := env.GET(t, path)
	requireStatus(t, resp2, http.StatusOK)

	var result struct {
		Exchange string `json:"exchange"`
		Group    string `json:"group"`
		Symbols  []struct {
			Symbol string         `json:"symbol"`
			Extra  map[string]any `json:"extra,omitempty"`
		} `json:"symbols"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp2.Body.Close()

	t.Logf("hotlist %s/%s: symbols=%d exchange_field=%s", ex.Exchange, group, len(result.Symbols), result.Exchange)
	if result.Exchange != ex.Exchange {
		t.Errorf("exchange = %q, want %q", result.Exchange, ex.Exchange)
	}
	for i, sym := range result.Symbols {
		if i < 5 {
			t.Logf("  symbol=%s extra=%v", sym.Symbol, sym.Extra)
		}
	}
}

// mapKeys returns the keys of a map for logging.
func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
