package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerAlertHandlers(api huma.API, svc Service) {
	// --- Alerts endpoints ---

	huma.Register(api, huma.Operation{OperationID: "scan-alerts-access", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/alerts/scan", Summary: "Scan for alerts API access paths", Tags: []string{"Alerts"}},
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
	huma.Register(api, huma.Operation{OperationID: "probe-alerts-api", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/alerts/probe", Summary: "Probe getAlertsRestApi() singleton", Tags: []string{"Alerts"}},
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
	huma.Register(api, huma.Operation{OperationID: "probe-alerts-api-deep", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/alerts/probe/deep", Summary: "Deep probe getAlertsRestApi() methods and state", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *chartIDInput) (*alertsProbeDeepOutput, error) {
			probe, err := svc.ProbeAlertsRestApiDeep(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &alertsProbeDeepOutput{}
			out.Body = probe
			return out, nil
		})

	type alertsListOutput struct {
		Body struct {
			Alerts any `json:"alerts"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "list-alerts", Method: http.MethodGet, Path: "/api/v1/alerts", Summary: "List all alerts", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *struct{}) (*alertsListOutput, error) {
			alerts, err := svc.ListAlerts(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &alertsListOutput{}
			out.Body.Alerts = alerts
			return out, nil
		})

	type alertIDInput struct {
		AlertID string `path:"alert_id"`
	}
	type alertOutput struct {
		Body struct {
			Alerts any `json:"alerts"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "get-alert", Method: http.MethodGet, Path: "/api/v1/alerts/{alert_id}", Summary: "Get alert by ID", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *alertIDInput) (*alertOutput, error) {
			alerts, err := svc.GetAlerts(ctx, []string{input.AlertID})
			if err != nil {
				return nil, mapErr(err)
			}
			out := &alertOutput{}
			out.Body.Alerts = alerts
			return out, nil
		})

	type createAlertInput struct {
		Body map[string]any
	}
	type createAlertOutput struct {
		Body struct {
			Alert any `json:"alert"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "create-alert", Method: http.MethodPost, Path: "/api/v1/alerts", Summary: "Create a new alert", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *createAlertInput) (*createAlertOutput, error) {
			alert, err := svc.CreateAlert(ctx, input.Body)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createAlertOutput{}
			out.Body.Alert = alert
			return out, nil
		})

	type modifyAlertInput struct {
		AlertID string `path:"alert_id"`
		Body    map[string]any
	}
	huma.Register(api, huma.Operation{OperationID: "modify-alert", Method: http.MethodPut, Path: "/api/v1/alerts/{alert_id}", Summary: "Modify and restart an alert", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *modifyAlertInput) (*createAlertOutput, error) {
			params := input.Body
			if params == nil {
				params = map[string]any{}
			}
			params["id"] = input.AlertID
			alert, err := svc.ModifyAlert(ctx, params)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createAlertOutput{}
			out.Body.Alert = alert
			return out, nil
		})

	type alertIDsBodyInput struct {
		Body struct {
			AlertIDs []string `json:"alert_ids" required:"true"`
		}
	}
	type alertStatusOutput struct {
		Body struct {
			Status string `json:"status"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "delete-alerts", Method: http.MethodDelete, Path: "/api/v1/alerts", Summary: "Delete alerts by IDs", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *alertIDsBodyInput) (*alertStatusOutput, error) {
			if err := svc.DeleteAlerts(ctx, input.Body.AlertIDs); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "deleted"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "stop-alerts", Method: http.MethodPost, Path: "/api/v1/alerts/stop", Summary: "Stop alerts by IDs", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *alertIDsBodyInput) (*alertStatusOutput, error) {
			if err := svc.StopAlerts(ctx, input.Body.AlertIDs); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "stopped"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "restart-alerts", Method: http.MethodPost, Path: "/api/v1/alerts/restart", Summary: "Restart alerts by IDs", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *alertIDsBodyInput) (*alertStatusOutput, error) {
			if err := svc.RestartAlerts(ctx, input.Body.AlertIDs); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "restarted"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "clone-alerts", Method: http.MethodPost, Path: "/api/v1/alerts/clone", Summary: "Clone alerts by IDs", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *alertIDsBodyInput) (*alertStatusOutput, error) {
			if err := svc.CloneAlerts(ctx, input.Body.AlertIDs); err != nil {
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
	huma.Register(api, huma.Operation{OperationID: "list-fires", Method: http.MethodGet, Path: "/api/v1/alerts/fires", Summary: "List all fired alerts", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *struct{}) (*firesListOutput, error) {
			fires, err := svc.ListFires(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &firesListOutput{}
			out.Body.Fires = fires
			return out, nil
		})

	type fireIDsBodyInput struct {
		Body struct {
			FireIDs []string `json:"fire_ids" required:"true"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "delete-fires", Method: http.MethodDelete, Path: "/api/v1/alerts/fires", Summary: "Delete fires by IDs", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *fireIDsBodyInput) (*alertStatusOutput, error) {
			if err := svc.DeleteFires(ctx, input.Body.FireIDs); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "deleted"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "delete-all-fires", Method: http.MethodDelete, Path: "/api/v1/alerts/fires/all", Summary: "Delete all fires", Tags: []string{"Alerts"}},
		func(ctx context.Context, input *struct{}) (*alertStatusOutput, error) {
			if err := svc.DeleteAllFires(ctx); err != nil {
				return nil, mapErr(err)
			}
			out := &alertStatusOutput{}
			out.Body.Status = "deleted"
			return out, nil
		})

}
