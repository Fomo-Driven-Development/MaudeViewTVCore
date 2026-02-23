package apimulti

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerAlertHandlers(api huma.API, svc MultiService) {
	// --- Alerts probe/scan endpoints (chart-level, paths unchanged) ---

	huma.Register(api, huma.Operation{OperationID: "multi-scan-alerts-access", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/alerts/scan", Summary: "Scan for alerts API access paths", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *chartIDInput) (*struct{ Body map[string]any }, error) {
			result, err := svc.ScanAlertsAccess(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct{ Body map[string]any }{}
			out.Body = result
			return out, nil
		})

	type alertsProbeOutput struct {
		Body cdpcontrol.AlertsApiProbe
	}
	huma.Register(api, huma.Operation{OperationID: "multi-probe-alerts-api", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/alerts/probe", Summary: "Probe getAlertsRestApi() singleton", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *chartIDInput) (*alertsProbeOutput, error) {
			probe, err := svc.ProbeAlertsRestApi(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &alertsProbeOutput{}
			out.Body = probe
			return out, nil
		})

	type alertsProbeDeepOutput struct {
		Body map[string]any
	}
	huma.Register(api, huma.Operation{OperationID: "multi-probe-alerts-api-deep", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/alerts/probe/deep", Summary: "Deep probe getAlertsRestApi() methods and state", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *chartIDInput) (*alertsProbeDeepOutput, error) {
			probe, err := svc.ProbeAlertsRestApiDeep(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &alertsProbeDeepOutput{}
			out.Body = probe
			return out, nil
		})

	// --- Alerts CRUD endpoints (session-level, moved under /chart/{chart_id}/) ---

	type alertsListOutput struct {
		Body struct {
			Alerts any `json:"alerts"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-list-alerts", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/alerts", Summary: "List all alerts", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *chartIDInput) (*alertsListOutput, error) {
			alerts, err := svc.ListAlerts(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &alertsListOutput{}
			out.Body.Alerts = alerts
			return out, nil
		})

	type alertOutput struct {
		Body struct {
			Alerts any `json:"alerts"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-get-alert", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/alerts/{alert_id}", Summary: "Get alert by ID", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			AlertID string `path:"alert_id"`
		}) (*alertOutput, error) {
			alerts, err := svc.GetAlerts(ctx, input.ChartID, []string{input.AlertID})
			if err != nil {
				return nil, mapErr(err)
			}
			out := &alertOutput{}
			out.Body.Alerts = alerts
			return out, nil
		})

	type createAlertOutput struct {
		Body struct {
			Alert any `json:"alert"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-create-alert", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/alerts", Summary: "Create a new alert", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *struct {
			ChartID string         `path:"chart_id"`
			Body    map[string]any
		}) (*createAlertOutput, error) {
			alert, err := svc.CreateAlert(ctx, input.ChartID, input.Body)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createAlertOutput{}
			out.Body.Alert = alert
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-modify-alert", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/alerts/{alert_id}", Summary: "Modify and restart an alert", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *struct {
			ChartID string         `path:"chart_id"`
			AlertID string         `path:"alert_id"`
			Body    map[string]any
		}) (*createAlertOutput, error) {
			params := input.Body
			if params == nil {
				params = map[string]any{}
			}
			params["id"] = input.AlertID
			alert, err := svc.ModifyAlert(ctx, input.ChartID, params)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createAlertOutput{}
			out.Body.Alert = alert
			return out, nil
		})

	type multiAlertIDsBodyInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			AlertIDs []string `json:"alert_ids" required:"true"`
		}
	}
	type alertStatusOutput struct {
		Body struct {
			Status string `json:"status"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-delete-alerts", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/alerts", Summary: "Delete alerts by IDs", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *multiAlertIDsBodyInput) (*alertStatusOutput, error) {
			if err := svc.DeleteAlerts(ctx, input.ChartID, input.Body.AlertIDs); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "deleted"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-stop-alerts", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/alerts/stop", Summary: "Stop alerts by IDs", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *multiAlertIDsBodyInput) (*alertStatusOutput, error) {
			if err := svc.StopAlerts(ctx, input.ChartID, input.Body.AlertIDs); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "stopped"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-restart-alerts", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/alerts/restart", Summary: "Restart alerts by IDs", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *multiAlertIDsBodyInput) (*alertStatusOutput, error) {
			if err := svc.RestartAlerts(ctx, input.ChartID, input.Body.AlertIDs); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "restarted"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-clone-alerts", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/alerts/clone", Summary: "Clone alerts by IDs", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *multiAlertIDsBodyInput) (*alertStatusOutput, error) {
			if err := svc.CloneAlerts(ctx, input.ChartID, input.Body.AlertIDs); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "cloned"
			return out, nil
		})

	type firesListOutput struct {
		Body struct {
			Fires any `json:"fires"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-list-fires", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/alerts/fires", Summary: "List all fired alerts", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *chartIDInput) (*firesListOutput, error) {
			fires, err := svc.ListFires(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &firesListOutput{}
			out.Body.Fires = fires
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-delete-fires", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/alerts/fires", Summary: "Delete fires by IDs", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				FireIDs []string `json:"fire_ids" required:"true"`
			}
		}) (*alertStatusOutput, error) {
			if err := svc.DeleteFires(ctx, input.ChartID, input.Body.FireIDs); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "deleted"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-delete-all-fires", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/alerts/fires/all", Summary: "Delete all fires", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *chartIDInput) (*alertStatusOutput, error) {
			if err := svc.DeleteAllFires(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "deleted"
			return out, nil
		})
}
