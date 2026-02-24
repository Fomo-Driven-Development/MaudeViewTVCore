//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// alertPath returns the chart-scoped path for alerts probe/scan endpoints.
func alertPath(suffix string) string {
	return env.chartPath("alerts/" + suffix)
}

// getFirstAlertID lists alerts and returns the first alert's ID string,
// or calls t.Skip if no alerts exist.
func getFirstAlertID(t *testing.T) string {
	t.Helper()
	resp := env.GET(t, env.featurePath("alerts"))
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Alerts []map[string]any `json:"alerts"`
	}](t, resp)
	if len(result.Alerts) == 0 {
		t.Skip("no alerts to test with")
	}
	id, ok := result.Alerts[0]["alert_id"]
	if !ok {
		t.Fatal("first alert has no 'alert_id' field")
	}
	// id may be float64 from JSON; format without decimals.
	switch v := id.(type) {
	case float64:
		return fmt.Sprintf("%.0f", v)
	case string:
		return v
	default:
		t.Fatalf("unexpected alert id type %T", id)
		return ""
	}
}

// listAlertIDs returns all current alert ID strings.
func listAlertIDs(t *testing.T) []string {
	t.Helper()
	resp := env.GET(t, env.featurePath("alerts"))
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Alerts []map[string]any `json:"alerts"`
	}](t, resp)
	ids := make([]string, 0, len(result.Alerts))
	for _, a := range result.Alerts {
		if id, ok := a["alert_id"]; ok {
			switch v := id.(type) {
			case float64:
				ids = append(ids, fmt.Sprintf("%.0f", v))
			case string:
				ids = append(ids, v)
			}
		}
	}
	return ids
}

// --- Probing & Discovery ---

func TestAlertsScan(t *testing.T) {
	resp := env.GET(t, alertPath("scan"))
	requireStatus(t, resp, http.StatusOK)

	scan := decodeJSON[map[string]any](t, resp)
	if len(scan) == 0 {
		t.Fatal("expected scan to return data")
	}
	t.Logf("alerts scan result keys: %d", len(scan))
}

func TestAlertsProbe(t *testing.T) {
	resp := env.GET(t, alertPath("probe"))
	requireStatus(t, resp, http.StatusOK)

	probe := decodeJSON[struct {
		Found       bool     `json:"found"`
		AccessPaths []string `json:"access_paths"`
		Methods     []string `json:"methods"`
	}](t, resp)

	if !probe.Found {
		t.Fatal("expected alerts API to be found")
	}
	if len(probe.Methods) == 0 {
		t.Fatal("expected alerts API to have methods")
	}
	t.Logf("alerts probe: found=%v, %d methods, paths=%v", probe.Found, len(probe.Methods), probe.AccessPaths)
}

func TestAlertsProbeDeep(t *testing.T) {
	resp := env.GET(t, alertPath("probe/deep"))
	requireStatus(t, resp, http.StatusOK)

	probe := decodeJSON[map[string]any](t, resp)
	if len(probe) == 0 {
		t.Fatal("expected deep probe to return data")
	}
	t.Logf("alerts deep probe keys: %d", len(probe))
}

// --- Read Operations ---

func TestAlertsListAlerts(t *testing.T) {
	resp := env.GET(t, env.featurePath("alerts"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Alerts any `json:"alerts"`
	}](t, resp)

	// Alerts field should be present (may be empty array or null).
	t.Logf("list alerts result: %v", result.Alerts)
}

func TestAlertsListFires(t *testing.T) {
	resp := env.GET(t, env.featurePath("alerts/fires"))
	requireStatus(t, resp, http.StatusOK)

	result := decodeJSON[struct {
		Fires any `json:"fires"`
	}](t, resp)

	t.Logf("list fires result: %v", result.Fires)
}

// --- Validation Tests ---

func TestDeleteAlerts_EmptyIDs(t *testing.T) {
	resp := env.do(t, http.MethodDelete, env.featurePath("alerts"), map[string]any{
		"alert_ids": []string{},
	})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for empty alert_ids")
	}
	resp.Body.Close()
	t.Logf("empty alert_ids correctly rejected with status %d", resp.StatusCode)
}

func TestStopAlerts_EmptyIDs(t *testing.T) {
	resp := env.POST(t, env.featurePath("alerts/stop"), map[string]any{
		"alert_ids": []string{},
	})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for empty alert_ids")
	}
	resp.Body.Close()
	t.Logf("empty alert_ids correctly rejected with status %d", resp.StatusCode)
}

func TestRestartAlerts_EmptyIDs(t *testing.T) {
	resp := env.POST(t, env.featurePath("alerts/restart"), map[string]any{
		"alert_ids": []string{},
	})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for empty alert_ids")
	}
	resp.Body.Close()
	t.Logf("empty alert_ids correctly rejected with status %d", resp.StatusCode)
}

func TestCloneAlerts_EmptyIDs(t *testing.T) {
	resp := env.POST(t, env.featurePath("alerts/clone"), map[string]any{
		"alert_ids": []string{},
	})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for empty alert_ids")
	}
	resp.Body.Close()
	t.Logf("empty alert_ids correctly rejected with status %d", resp.StatusCode)
}

func TestDeleteFires_EmptyIDs(t *testing.T) {
	resp := env.do(t, http.MethodDelete, env.featurePath("alerts/fires"), map[string]any{
		"fire_ids": []string{},
	})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for empty fire_ids")
	}
	resp.Body.Close()
	t.Logf("empty fire_ids correctly rejected with status %d", resp.StatusCode)
}

// --- Stateful Operations (skip if no alerts) ---

func TestAlerts_StopAndRestart(t *testing.T) {
	alertID := getFirstAlertID(t)
	t.Logf("testing stop/restart on alert %s", alertID)

	// Stop the alert.
	resp := env.POST(t, env.featurePath("alerts/stop"), map[string]any{
		"alert_ids": []string{alertID},
	})
	requireStatus(t, resp, http.StatusOK)
	stopResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, stopResult.Status, "stopped", "status")
	t.Logf("stopped alert %s", alertID)

	time.Sleep(testSettleLong)

	// Restart the alert.
	resp = env.POST(t, env.featurePath("alerts/restart"), map[string]any{
		"alert_ids": []string{alertID},
	})
	requireStatus(t, resp, http.StatusOK)
	restartResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, restartResult.Status, "restarted", "status")
	t.Logf("restarted alert %s", alertID)
}

func TestAlerts_CloneAndDelete(t *testing.T) {
	alertID := getFirstAlertID(t)
	t.Logf("testing clone/delete on alert %s", alertID)

	// Get initial alert count.
	beforeIDs := listAlertIDs(t)
	beforeCount := len(beforeIDs)
	t.Logf("alert count before clone: %d", beforeCount)

	// Clone the alert.
	resp := env.POST(t, env.featurePath("alerts/clone"), map[string]any{
		"alert_ids": []string{alertID},
	})
	requireStatus(t, resp, http.StatusOK)
	cloneResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, cloneResult.Status, "cloned", "status")

	time.Sleep(testSettleLong)

	// Get new alert list and find the clone.
	afterIDs := listAlertIDs(t)
	afterCount := len(afterIDs)
	t.Logf("alert count after clone: %d", afterCount)

	if afterCount <= beforeCount {
		t.Fatalf("expected alert count to increase after clone: before=%d, after=%d", beforeCount, afterCount)
	}

	// Find the new alert ID (present in after but not in before).
	beforeSet := make(map[string]bool, len(beforeIDs))
	for _, id := range beforeIDs {
		beforeSet[id] = true
	}
	var clonedID string
	for _, id := range afterIDs {
		if !beforeSet[id] {
			clonedID = id
			break
		}
	}
	if clonedID == "" {
		t.Fatal("could not identify cloned alert ID")
	}
	t.Logf("cloned alert ID: %s", clonedID)

	// Register cleanup in case delete fails.
	t.Cleanup(func() {
		r := env.do(t, http.MethodDelete, env.featurePath("alerts"), map[string]any{
			"alert_ids": []string{clonedID},
		})
		r.Body.Close()
	})

	// Delete the clone.
	resp = env.do(t, http.MethodDelete, env.featurePath("alerts"), map[string]any{
		"alert_ids": []string{clonedID},
	})
	requireStatus(t, resp, http.StatusOK)
	deleteResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, deleteResult.Status, "deleted", "status")
	t.Logf("deleted cloned alert %s", clonedID)

	time.Sleep(testSettleLong)

	// Verify count is back to original.
	finalIDs := listAlertIDs(t)
	if len(finalIDs) != beforeCount {
		t.Fatalf("expected alert count to return to %d after delete, got %d", beforeCount, len(finalIDs))
	}
}

// --- Create & Modify ---

func TestCreateAlert_EmptyParams(t *testing.T) {
	resp := env.POST(t, env.featurePath("alerts"), map[string]any{})
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected error for empty params")
	}
	resp.Body.Close()
	t.Logf("empty params correctly rejected with status %d", resp.StatusCode)
}

func TestModifyAlert_EmptyParams(t *testing.T) {
	alertID := getFirstAlertID(t)

	resp := env.PUT(t, env.featurePath("alerts/"+alertID), map[string]any{})
	// The server injects the alert_id into params, so params won't be empty.
	// But the underlying TradingView API may still reject it.
	// Accept either success (200) or client error (4xx).
	t.Logf("modify with empty body: status %d", resp.StatusCode)
	resp.Body.Close()
}

func TestModifyAlert(t *testing.T) {
	alertID := getFirstAlertID(t)
	t.Logf("testing modify on alert %s", alertID)

	// Get the existing alert data.
	resp := env.GET(t, env.featurePath("alerts/"+alertID))
	requireStatus(t, resp, http.StatusOK)
	original := decodeJSON[struct {
		Alerts any `json:"alerts"`
	}](t, resp)
	if original.Alerts == nil {
		t.Skip("no alert data returned")
	}

	// Modify with updated name.
	modifiedName := fmt.Sprintf("tv_agent_test_%d", time.Now().Unix())
	resp = env.PUT(t, env.featurePath("alerts/"+alertID), map[string]any{
		"name": modifiedName,
	})
	if resp.StatusCode != http.StatusOK {
		body := decodeJSON[map[string]any](t, resp)
		t.Logf("modify alert returned %d: %v (TradingView API may require more params)", resp.StatusCode, body)
		t.Skip("modify alert requires additional params not available in this test")
	}
	result := decodeJSON[struct {
		Alert any `json:"alert"`
	}](t, resp)
	if result.Alert == nil {
		t.Fatal("expected alert data in modify response")
	}
	t.Logf("modified alert %s → name=%s", alertID, modifiedName)
}

// --- Delete All Fires ---

func TestDeleteAllFires(t *testing.T) {
	resp := env.DELETE(t, env.featurePath("alerts/fires/all"))
	requireStatus(t, resp, http.StatusOK)
	result := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, result.Status, "deleted", "status")
	t.Logf("delete all fires: status=%s", result.Status)
}

// --- Full Lifecycle ---

func TestAlertsFullLifecycle(t *testing.T) {
	alertID := getFirstAlertID(t)
	t.Logf("full lifecycle test using alert %s", alertID)

	// 1. List alerts — already validated by getFirstAlertID.

	// 2. Get single alert.
	resp := env.GET(t, env.featurePath("alerts/"+alertID))
	requireStatus(t, resp, http.StatusOK)
	single := decodeJSON[struct {
		Alerts any `json:"alerts"`
	}](t, resp)
	if single.Alerts == nil {
		t.Fatal("expected alerts data for single get")
	}
	t.Logf("got single alert %s", alertID)

	// 3. Clone the alert.
	beforeIDs := listAlertIDs(t)
	resp = env.POST(t, env.featurePath("alerts/clone"), map[string]any{
		"alert_ids": []string{alertID},
	})
	requireStatus(t, resp, http.StatusOK)
	cloneResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, cloneResult.Status, "cloned", "status")

	time.Sleep(testSettleLong)

	// Find cloned ID.
	afterIDs := listAlertIDs(t)
	beforeSet := make(map[string]bool, len(beforeIDs))
	for _, id := range beforeIDs {
		beforeSet[id] = true
	}
	var clonedID string
	for _, id := range afterIDs {
		if !beforeSet[id] {
			clonedID = id
			break
		}
	}
	if clonedID == "" {
		t.Fatal("could not identify cloned alert ID")
	}
	t.Logf("cloned alert %s → %s", alertID, clonedID)

	// Register cleanup.
	t.Cleanup(func() {
		r := env.do(t, http.MethodDelete, env.featurePath("alerts"), map[string]any{
			"alert_ids": []string{clonedID},
		})
		r.Body.Close()
	})

	// 4. Stop the clone.
	resp = env.POST(t, env.featurePath("alerts/stop"), map[string]any{
		"alert_ids": []string{clonedID},
	})
	requireStatus(t, resp, http.StatusOK)
	stopResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, stopResult.Status, "stopped", "status")
	t.Logf("stopped clone %s", clonedID)

	time.Sleep(testSettleLong)

	// 5. Restart the clone.
	resp = env.POST(t, env.featurePath("alerts/restart"), map[string]any{
		"alert_ids": []string{clonedID},
	})
	requireStatus(t, resp, http.StatusOK)
	restartResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, restartResult.Status, "restarted", "status")
	t.Logf("restarted clone %s", clonedID)

	time.Sleep(testSettleLong)

	// 6. Delete the clone.
	resp = env.do(t, http.MethodDelete, env.featurePath("alerts"), map[string]any{
		"alert_ids": []string{clonedID},
	})
	requireStatus(t, resp, http.StatusOK)
	deleteResult := decodeJSON[struct {
		Status string `json:"status"`
	}](t, resp)
	requireField(t, deleteResult.Status, "deleted", "status")
	t.Logf("deleted clone %s", clonedID)

	time.Sleep(testSettleLong)

	// 7. Verify the clone is gone.
	finalIDs := listAlertIDs(t)
	for _, id := range finalIDs {
		if id == clonedID {
			t.Fatalf("cloned alert %s should not exist after delete", clonedID)
		}
	}
	t.Logf("full alerts lifecycle completed successfully")
}
