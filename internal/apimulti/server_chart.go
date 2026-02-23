package apimulti

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerChartHandlers(api huma.API, svc MultiService) {
	type listChartsOutput struct {
		Body struct {
			Charts []cdpcontrol.ChartInfo `json:"charts"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-list-charts", Method: http.MethodGet, Path: "/api/v1/charts", Summary: "List chart tabs", Tags: []string{"Charts"}},
		func(ctx context.Context, input *struct{}) (*listChartsOutput, error) {
			charts, err := svc.ListCharts(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &listChartsOutput{}
			out.Body.Charts = charts
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-get-active-chart", Method: http.MethodGet, Path: "/api/v1/charts/active", Summary: "Get active chart", Tags: []string{"Charts"}},
		func(ctx context.Context, input *struct{}) (*activeChartOutput, error) {
			info, err := svc.GetActiveChart(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &activeChartOutput{}
			out.Body = info
			return out, nil
		})

	type symbolInput struct {
		ChartID string `path:"chart_id"`
		Symbol  string `query:"symbol" required:"true"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
	}
	type symbolOutput struct {
		Body struct {
			ChartID         string `json:"chart_id"`
			RequestedSymbol string `json:"requested_symbol,omitempty"`
			CurrentSymbol   string `json:"current_symbol"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-get-symbol", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/symbol", Summary: "Get chart symbol", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *chartIDInput) (*symbolOutput, error) {
			symbol, err := svc.GetSymbol(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &symbolOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.CurrentSymbol = symbol
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-set-symbol", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/symbol", Summary: "Set chart symbol", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *symbolInput) (*symbolOutput, error) {
			current, err := svc.SetSymbol(ctx, input.ChartID, input.Symbol, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &symbolOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.RequestedSymbol = input.Symbol
			out.Body.CurrentSymbol = current
			return out, nil
		})

	type symbolInfoOutput struct {
		Body struct {
			ChartID    string                `json:"chart_id"`
			SymbolInfo cdpcontrol.SymbolInfo `json:"symbol_info"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-get-symbol-info", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/symbol/info", Summary: "Get extended symbol metadata", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *chartIDInput) (*symbolInfoOutput, error) {
			info, err := svc.GetSymbolInfo(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &symbolInfoOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.SymbolInfo = info
			return out, nil
		})

	type resolutionInput struct {
		ChartID    string `path:"chart_id"`
		Resolution string `query:"resolution" required:"true"`
		Pane       int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
	}
	type resolutionOutput struct {
		Body struct {
			ChartID             string `json:"chart_id"`
			RequestedResolution string `json:"requested_resolution,omitempty"`
			CurrentResolution   string `json:"current_resolution"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-get-resolution", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/resolution", Summary: "Get chart resolution", Tags: []string{"Resolution"}},
		func(ctx context.Context, input *chartIDInput) (*resolutionOutput, error) {
			resolution, err := svc.GetResolution(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &resolutionOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.CurrentResolution = resolution
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-set-resolution", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/resolution", Summary: "Set chart resolution", Tags: []string{"Resolution"}},
		func(ctx context.Context, input *resolutionInput) (*resolutionOutput, error) {
			current, err := svc.SetResolution(ctx, input.ChartID, input.Resolution, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &resolutionOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.RequestedResolution = input.Resolution
			out.Body.CurrentResolution = current
			return out, nil
		})

	type chartTypeInput struct {
		ChartID string `path:"chart_id"`
		Type    string `query:"type" required:"true"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
	}
	type chartTypeOutput struct {
		Body struct {
			ChartID       string `json:"chart_id"`
			ChartType     string `json:"chart_type"`
			ChartTypeID   int    `json:"chart_type_id"`
			RequestedType string `json:"requested_type,omitempty"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-get-chart-type", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/chart-type", Summary: "Get chart type (bar style)", Tags: []string{"ChartType"}},
		func(ctx context.Context, input *chartIDInput) (*chartTypeOutput, error) {
			typeID, err := svc.GetChartType(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			name := cdpcontrol.ChartTypeReverseMap[typeID]
			if name == "" {
				name = fmt.Sprintf("unknown_%d", typeID)
			}
			out := &chartTypeOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.ChartType = name
			out.Body.ChartTypeID = typeID
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-set-chart-type", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/chart-type", Summary: "Set chart type (bar style)", Tags: []string{"ChartType"}},
		func(ctx context.Context, input *chartTypeInput) (*chartTypeOutput, error) {
			typeID, ok := cdpcontrol.ChartTypeMap[strings.ToLower(strings.TrimSpace(input.Type))]
			if !ok {
				return nil, huma.Error400BadRequest(fmt.Sprintf("unknown chart type %q; valid types: bars, candles, line, area, renko, kagi, point_and_figure, line_break, heikin_ashi, hollow_candles, baseline, high_low, columns, line_with_markers, step_line, hlc_area, volume_candles, hlc_bars", input.Type))
			}
			current, err := svc.SetChartType(ctx, input.ChartID, typeID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			name := cdpcontrol.ChartTypeReverseMap[current]
			if name == "" {
				name = fmt.Sprintf("unknown_%d", current)
			}
			out := &chartTypeOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.RequestedType = input.Type
			out.Body.ChartType = name
			out.Body.ChartTypeID = current
			return out, nil
		})

	type actionInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			ActionID string `json:"action_id,omitempty"`
			Action   string `json:"action,omitempty"`
		}
	}
	type actionOutput struct {
		Body struct {
			ChartID  string `json:"chart_id"`
			ActionID string `json:"action_id"`
			Status   string `json:"status"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-execute-action", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/action", Summary: "Execute chart action", Tags: []string{"Action"}},
		func(ctx context.Context, input *actionInput) (*actionOutput, error) {
			actionID := strings.TrimSpace(input.Body.ActionID)
			if actionID == "" {
				actionID = strings.TrimSpace(input.Body.Action)
			}
			if actionID == "" {
				return nil, huma.Error400BadRequest("action_id is required")
			}
			if err := svc.ExecuteAction(ctx, input.ChartID, actionID); err != nil {
				return nil, mapErr(err)
			}
			out := &actionOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.ActionID = actionID
			out.Body.Status = "executed"
			return out, nil
		})

	type visibleRangeOutput struct {
		Body struct {
			ChartID string  `json:"chart_id"`
			From    float64 `json:"from"`
			To      float64 `json:"to"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-get-visible-range", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/visible-range", Summary: "Get visible bar range", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*visibleRangeOutput, error) {
			r, err := svc.GetVisibleRange(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &visibleRangeOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.From = r.From
			out.Body.To = r.To
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-set-visible-range", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/visible-range", Summary: "Set visible bar range", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				From float64 `json:"from" required:"true"`
				To   float64 `json:"to" required:"true"`
			}
		}) (*visibleRangeOutput, error) {
			r, err := svc.SetVisibleRange(ctx, input.ChartID, input.Body.From, input.Body.To)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &visibleRangeOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.From = r.From
			out.Body.To = r.To
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-set-timeframe", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/timeframe", Summary: "Set time frame preset (1D, 5D, 1M, 3M, 6M, YTD, 1Y, 5Y, All)", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *struct {
			ChartID    string `path:"chart_id"`
			Preset     string `query:"preset" required:"true" doc:"Time frame preset: 1D, 5D, 1M, 3M, 6M, YTD, 1Y, 5Y, All"`
			Resolution string `query:"resolution" doc:"Optional resolution override (e.g. 1, 5, 15, 60, 1D). Omit to let TradingView pick."`
			Pane       int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
		}) (*struct {
			Body struct {
				ChartID    string  `json:"chart_id"`
				Preset     string  `json:"preset"`
				Resolution string  `json:"resolution"`
				From       float64 `json:"from"`
				To         float64 `json:"to"`
			}
		}, error) {
			r, err := svc.SetTimeFrame(ctx, input.ChartID, input.Preset, input.Resolution, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					ChartID    string  `json:"chart_id"`
					Preset     string  `json:"preset"`
					Resolution string  `json:"resolution"`
					From       float64 `json:"from"`
					To         float64 `json:"to"`
				}
			}{}
			out.Body.ChartID = input.ChartID
			out.Body.Preset = r.Preset
			out.Body.Resolution = r.Resolution
			out.Body.From = r.From
			out.Body.To = r.To
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-zoom", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/zoom", Summary: "Zoom in or out", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Direction string `json:"direction" required:"true"`
			}
		}) (*struct {
			Body struct {
				ChartID   string `json:"chart_id"`
				Status    string `json:"status"`
				Direction string `json:"direction"`
			}
		}, error) {
			if err := svc.Zoom(ctx, input.ChartID, input.Body.Direction); err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					ChartID   string `json:"chart_id"`
					Status    string `json:"status"`
					Direction string `json:"direction"`
				}
			}{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			out.Body.Direction = input.Body.Direction
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-scroll", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/scroll", Summary: "Scroll chart by bars", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Bars int `json:"bars" required:"true"`
			}
		}) (*struct {
			Body struct {
				ChartID string `json:"chart_id"`
				Status  string `json:"status"`
				Bars    int    `json:"bars"`
			}
		}, error) {
			if err := svc.Scroll(ctx, input.ChartID, input.Body.Bars); err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					ChartID string `json:"chart_id"`
					Status  string `json:"status"`
					Bars    int    `json:"bars"`
				}
			}{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			out.Body.Bars = input.Body.Bars
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-reset-view", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/reset-view", Summary: "Reset chart view", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.ResetView(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-chart-undo", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/undo", Summary: "Chart Undo (Ctrl+Z)", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.UndoChart(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-chart-redo", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/redo", Summary: "Chart Redo (Ctrl+Y)", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.RedoChart(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-go-to-date", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/go-to-date", Summary: "Navigate chart to a specific date", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Timestamp int64 `json:"timestamp" required:"true"`
			}
		}) (*struct {
			Body struct {
				ChartID   string `json:"chart_id"`
				Status    string `json:"status"`
				Timestamp int64  `json:"timestamp"`
			}
		}, error) {
			if err := svc.GoToDate(ctx, input.ChartID, input.Body.Timestamp); err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					ChartID   string `json:"chart_id"`
					Status    string `json:"status"`
					Timestamp int64  `json:"timestamp"`
				}
			}{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			out.Body.Timestamp = input.Body.Timestamp
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-reset-scales", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/reset-scales", Summary: "Reset all chart scales", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.ResetScales(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			return out, nil
		})

	type chartTogglesOutput struct {
		Body struct {
			ChartID       string `json:"chart_id"`
			LogScale      *bool  `json:"log_scale,omitempty"`
			AutoScale     *bool  `json:"auto_scale,omitempty"`
			ExtendedHours *bool  `json:"extended_hours,omitempty"`
		}
	}
	type toggleResultOutput struct {
		Body struct {
			ChartID string                  `json:"chart_id"`
			Status  string                  `json:"status"`
			Before  cdpcontrol.ChartToggles `json:"before"`
			After   cdpcontrol.ChartToggles `json:"after"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-get-chart-toggles", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/toggles", Summary: "Get log scale, auto scale, and extended hours state", Tags: []string{"Chart Toggles"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
		}) (*chartTogglesOutput, error) {
			toggles, err := svc.GetChartToggles(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &chartTogglesOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.LogScale = toggles.LogScale
			out.Body.AutoScale = toggles.AutoScale
			out.Body.ExtendedHours = toggles.ExtendedHours
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-toggle-log-scale", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/toggles/log-scale", Summary: "Toggle logarithmic price scale (Alt+L)", Tags: []string{"Chart Toggles"}},
		func(ctx context.Context, input *chartIDInput) (*toggleResultOutput, error) {
			before, _ := svc.GetChartToggles(ctx, input.ChartID, -1)
			if err := svc.ToggleLogScale(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			after, _ := svc.GetChartToggles(ctx, input.ChartID, -1)
			out := &toggleResultOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "toggled"
			out.Body.Before = before
			out.Body.After = after
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-toggle-auto-scale", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/toggles/auto-scale", Summary: "Toggle auto-fitting price scale (Alt+A)", Tags: []string{"Chart Toggles"}},
		func(ctx context.Context, input *chartIDInput) (*toggleResultOutput, error) {
			before, _ := svc.GetChartToggles(ctx, input.ChartID, -1)
			if err := svc.ToggleAutoScale(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			after, _ := svc.GetChartToggles(ctx, input.ChartID, -1)
			out := &toggleResultOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "toggled"
			out.Body.Before = before
			out.Body.After = after
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-toggle-extended-hours", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/toggles/extended-hours", Summary: "Toggle extended trading hours (Alt+E)", Tags: []string{"Chart Toggles"}},
		func(ctx context.Context, input *chartIDInput) (*toggleResultOutput, error) {
			before, _ := svc.GetChartToggles(ctx, input.ChartID, -1)
			if err := svc.ToggleExtendedHours(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			after, _ := svc.GetChartToggles(ctx, input.ChartID, -1)
			out := &toggleResultOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "toggled"
			out.Body.Before = before
			out.Body.After = after
			return out, nil
		})

	type chartApiProbeOutput struct {
		Body cdpcontrol.ChartApiProbe
	}
	huma.Register(api, huma.Operation{OperationID: "multi-probe-chart-api", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/chart-api/probe", Summary: "Probe chartApi() singleton", Tags: []string{"ChartAPI"}},
		func(ctx context.Context, input *chartIDInput) (*chartApiProbeOutput, error) {
			probe, err := svc.ProbeChartApi(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &chartApiProbeOutput{}
			out.Body = probe
			return out, nil
		})

	type chartApiProbeDeepOutput struct {
		Body map[string]any
	}
	huma.Register(api, huma.Operation{OperationID: "multi-probe-chart-api-deep", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/chart-api/probe/deep", Summary: "Deep probe chartApi() methods and state", Tags: []string{"ChartAPI"}},
		func(ctx context.Context, input *chartIDInput) (*chartApiProbeDeepOutput, error) {
			probe, err := svc.ProbeChartApiDeep(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &chartApiProbeDeepOutput{}
			out.Body = probe
			return out, nil
		})

	type resolveSymbolInput struct {
		ChartID string `path:"chart_id"`
		Symbol  string `query:"symbol" required:"true"`
	}
	type resolveSymbolOutput struct {
		Body cdpcontrol.ResolvedSymbolInfo
	}
	huma.Register(api, huma.Operation{OperationID: "multi-resolve-symbol", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/chart-api/resolve-symbol", Summary: "Resolve metadata for any symbol", Tags: []string{"ChartAPI"}},
		func(ctx context.Context, input *resolveSymbolInput) (*resolveSymbolOutput, error) {
			info, err := svc.ResolveSymbol(ctx, input.ChartID, input.Symbol)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &resolveSymbolOutput{}
			out.Body = info
			return out, nil
		})

	type switchTimezoneInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			Timezone string `json:"timezone" required:"true"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-switch-timezone", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/chart-api/timezone", Summary: "Switch chart timezone", Tags: []string{"ChartAPI"}},
		func(ctx context.Context, input *switchTimezoneInput) (*navStatusOutput, error) {
			if err := svc.SwitchTimezone(ctx, input.ChartID, input.Body.Timezone); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			return out, nil
		})

	type exportInput struct {
		ChartID string `path:"chart_id"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
	}
	type exportOutput struct {
		Body cdpcontrol.ChartExportResult
	}
	huma.Register(api, huma.Operation{
		OperationID: "multi-export-chart-data",
		Method:      http.MethodGet,
		Path:        "/api/v1/chart/{chart_id}/export",
		Summary:     "Export chart data (OHLCV + all studies)",
		Description: "Returns all visible bars with OHLCV and every study plot column. " +
			"Equivalent to TradingView's native Download chart data dialog.",
		Tags: []string{"Data"},
	}, func(ctx context.Context, input *exportInput) (*exportOutput, error) {
		result, err := svc.ExportChartData(ctx, input.ChartID, input.Pane)
		if err != nil {
			return nil, mapErr(err)
		}
		return &exportOutput{Body: result}, nil
	})
}
