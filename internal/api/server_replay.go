package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerReplayHandlers(api huma.API, svc Service) {
	// --- Replay endpoints ---

	huma.Register(api, huma.Operation{OperationID: "scan-replay-activation", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/replay/scan", Summary: "Scan for replay activation paths", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*struct{ Body map[string]any }, error) {
			result, err := svc.ScanReplayActivation(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct{ Body map[string]any }{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "activate-replay", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/replay/activate", Summary: "Activate replay at a date (unix timestamp)", Tags: []string{"Replay"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Date float64 `json:"date" required:"true"`
			}
		}) (*struct{ Body map[string]any }, error) {
			result, err := svc.ActivateReplay(ctx, input.ChartID, input.Body.Date)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct{ Body map[string]any }{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "activate-replay-auto", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/replay/activate/auto", Summary: "Activate replay at the first available date", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*struct{ Body map[string]any }, error) {
			result, err := svc.ActivateReplayAuto(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct{ Body map[string]any }{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "deactivate-replay", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/replay/deactivate", Summary: "Leave replay mode", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.DeactivateReplay(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "deactivated"
			return out, nil
		})

	type replayProbeOutput struct {
		Body cdpcontrol.ReplayManagerProbe
	}
	huma.Register(api, huma.Operation{OperationID: "probe-replay-manager", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/replay/probe", Summary: "Probe _replayManager singleton", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*replayProbeOutput, error) {
			probe, err := svc.ProbeReplayManager(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &replayProbeOutput{}
			out.Body = probe
			return out, nil
		})

	type replayProbeDeepOutput struct {
		Body map[string]any
	}
	huma.Register(api, huma.Operation{OperationID: "probe-replay-manager-deep", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/replay/probe/deep", Summary: "Deep probe _replayManager methods and state", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*replayProbeDeepOutput, error) {
			probe, err := svc.ProbeReplayManagerDeep(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &replayProbeDeepOutput{}
			out.Body = probe
			return out, nil
		})

	type replayStatusOutput struct {
		Body cdpcontrol.ReplayStatus
	}
	huma.Register(api, huma.Operation{OperationID: "get-replay-status", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/replay/status", Summary: "Get replay status", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*replayStatusOutput, error) {
			status, err := svc.GetReplayStatus(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &replayStatusOutput{}
			out.Body = status
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "start-replay", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/replay/start", Summary: "Start bar replay at a point", Tags: []string{"Replay"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Point float64 `json:"point" required:"true"`
			}
		}) (*navStatusOutput, error) {
			if err := svc.StartReplay(ctx, input.ChartID, input.Body.Point); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "started"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "stop-replay", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/replay/stop", Summary: "Stop bar replay", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.StopReplay(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "stopped"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "replay-step", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/replay/step", Summary: "Step forward N bars in replay (default 1)", Tags: []string{"Replay"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Count int `json:"count,omitempty" default:"1" minimum:"1" maximum:"500" doc:"Number of bars to step forward"`
			}
		}) (*navStatusOutput, error) {
			if err := svc.ReplayStep(ctx, input.ChartID, input.Body.Count); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = fmt.Sprintf("stepped %d", input.Body.Count)
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "start-autoplay", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/replay/autoplay/start", Summary: "Start autoplay", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.StartAutoplay(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "autoplay_started"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "stop-autoplay", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/replay/autoplay/stop", Summary: "Stop autoplay", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.StopAutoplay(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "autoplay_stopped"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "reset-replay", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/replay/reset", Summary: "Reset replay", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.ResetReplay(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "reset"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "change-autoplay-delay", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/replay/autoplay/delay", Summary: "Change autoplay delay", Tags: []string{"Replay"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Delay float64 `json:"delay" required:"true"`
			}
		}) (*struct {
			Body struct {
				ChartID string  `json:"chart_id"`
				Status  string  `json:"status"`
				Delay   float64 `json:"delay"`
			}
		}, error) {
			current, err := svc.ChangeAutoplayDelay(ctx, input.ChartID, input.Body.Delay)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					ChartID string  `json:"chart_id"`
					Status  string  `json:"status"`
					Delay   float64 `json:"delay"`
				}
			}{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "changed"
			out.Body.Delay = current
			return out, nil
		})

}
