package apimulti

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/snapshot"
)

func registerMiscHandlers(api huma.API, svc MultiService) {
	// --- Health (truly global — no chart_id needed) ---

	type healthOutput struct {
		Body struct {
			Status string `json:"status"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-health", Method: http.MethodGet, Path: "/health", Summary: "Health check", Tags: []string{"Health"}},
		func(ctx context.Context, input *struct{}) (*healthOutput, error) {
			out := &healthOutput{}
			out.Body.Status = "ok"
			return out, nil
		})

	type deepHealthOutput struct {
		Body cdpcontrol.DeepHealthResult
	}
	huma.Register(api, huma.Operation{OperationID: "multi-deep-health", Method: http.MethodGet, Path: "/api/v1/health/deep", Summary: "Deep health check", Tags: []string{"Health"}},
		func(ctx context.Context, input *struct{}) (*deepHealthOutput, error) {
			result, err := svc.DeepHealthCheck(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &deepHealthOutput{}
			out.Body = result
			return out, nil
		})

	// --- Strategy endpoints (chart-level, paths unchanged) ---

	type strategyProbeOutput struct {
		Body cdpcontrol.StrategyApiProbe
	}
	huma.Register(api, huma.Operation{OperationID: "multi-probe-strategy-api", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/strategy/probe", Summary: "Probe backtesting strategy API", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *chartIDInput) (*strategyProbeOutput, error) {
			probe, err := svc.ProbeBacktestingApi(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &strategyProbeOutput{}
			out.Body = probe
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-list-strategies", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/strategy/list", Summary: "List all loaded strategies", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *chartIDInput) (*struct {
			Body struct {
				Strategies any `json:"strategies"`
			}
		}, error) {
			strategies, err := svc.ListStrategies(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					Strategies any `json:"strategies"`
				}
			}{}
			out.Body.Strategies = strategies
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-get-active-strategy", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/strategy/active", Summary: "Get active strategy with inputs and metadata", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *chartIDInput) (*struct{ Body map[string]any }, error) {
			result, err := svc.GetActiveStrategy(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct{ Body map[string]any }{}
			out.Body = result
			return out, nil
		})

	type setStrategyInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			StrategyID string `json:"strategy_id" doc:"Entity ID of the strategy to activate"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-set-active-strategy", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/strategy/active", Summary: "Set active strategy by entity ID", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *setStrategyInput) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			if err := svc.SetActiveStrategy(ctx, input.ChartID, input.Body.StrategyID); err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					Status string `json:"status"`
				}
			}{}
			out.Body.Status = "set"
			return out, nil
		})

	type setStrategyInputInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			Name  string `json:"name" doc:"Input parameter name"`
			Value any    `json:"value" doc:"Input parameter value"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-set-strategy-input", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/strategy/input", Summary: "Set a strategy input parameter", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *setStrategyInputInput) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			if err := svc.SetStrategyInput(ctx, input.ChartID, input.Body.Name, input.Body.Value); err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					Status string `json:"status"`
				}
			}{}
			out.Body.Status = "set"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-get-strategy-report", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/strategy/report", Summary: "Get backtest report data", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *chartIDInput) (*struct{ Body map[string]any }, error) {
			result, err := svc.GetStrategyReport(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct{ Body map[string]any }{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-get-strategy-date-range", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/strategy/date-range", Summary: "Get backtest date range", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *chartIDInput) (*struct {
			Body struct {
				DateRange any `json:"date_range"`
			}
		}, error) {
			dateRange, err := svc.GetStrategyDateRange(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					DateRange any `json:"date_range"`
				}
			}{}
			out.Body.DateRange = dateRange
			return out, nil
		})

	type strategyGotoInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			Timestamp float64 `json:"timestamp" doc:"Bar timestamp to navigate to"`
			BelowBar  bool    `json:"below_bar,omitempty" doc:"Whether to position below the bar"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-strategy-goto-date", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/strategy/goto", Summary: "Navigate chart to a specific trade/bar timestamp", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *strategyGotoInput) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			if err := svc.StrategyGotoDate(ctx, input.ChartID, input.Body.Timestamp, input.Body.BelowBar); err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					Status string `json:"status"`
				}
			}{}
			out.Body.Status = "navigated"
			return out, nil
		})

	// --- Snapshot endpoints (global — local storage, no chart session needed) ---

	type takeSnapshotOutput struct {
		Body struct {
			Snapshot snapshot.SnapshotMeta `json:"snapshot"`
			URL      string                `json:"url"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-browser-screenshot", Method: http.MethodPost, Path: "/api/v1/browser_screenshot", Summary: "Take browser viewport screenshot", Tags: []string{"Snapshots"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Format   string `json:"format,omitempty" doc:"Image format: png (default) or jpeg"`
				Quality  int    `json:"quality,omitempty" doc:"JPEG quality 1-100 (ignored for PNG)"`
				FullPage bool   `json:"full_page,omitempty" doc:"Capture full scrollable page"`
				Notes    string `json:"notes,omitempty" doc:"Free-form annotation for the snapshot"`
			}
		}) (*takeSnapshotOutput, error) {
			meta, err := svc.BrowserScreenshot(ctx, input.Body.Format, input.Body.Quality, input.Body.FullPage, input.Body.Notes)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &takeSnapshotOutput{}
			out.Body.Snapshot = meta
			out.Body.URL = "/api/v1/snapshots/" + meta.ID + "/image"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-take-snapshot", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/snapshot", Summary: "Take chart snapshot", Tags: []string{"Snapshots"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
			Body    struct {
				Format  string `json:"format,omitempty"`
				Quality string `json:"quality,omitempty"`
				Notes   string `json:"notes,omitempty" doc:"Free-form annotation for the snapshot"`
			}
		}) (*takeSnapshotOutput, error) {
			meta, err := svc.TakeSnapshot(ctx, input.ChartID, input.Body.Format, input.Body.Quality, input.Body.Notes, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &takeSnapshotOutput{}
			out.Body.Snapshot = meta
			out.Body.URL = "/api/v1/snapshots/" + meta.ID + "/image"
			return out, nil
		})

	type listSnapshotsOutput struct {
		Body struct {
			Snapshots []snapshot.SnapshotMeta `json:"snapshots"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-list-snapshots", Method: http.MethodGet, Path: "/api/v1/snapshots", Summary: "List snapshots", Tags: []string{"Snapshots"}},
		func(ctx context.Context, input *struct{}) (*listSnapshotsOutput, error) {
			metas, err := svc.ListSnapshots(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &listSnapshotsOutput{}
			out.Body.Snapshots = metas
			if out.Body.Snapshots == nil {
				out.Body.Snapshots = []snapshot.SnapshotMeta{}
			}
			return out, nil
		})

	type snapshotIDInput struct {
		SnapshotID string `path:"snapshot_id"`
	}
	type getSnapshotOutput struct {
		Body snapshot.SnapshotMeta
	}
	huma.Register(api, huma.Operation{OperationID: "multi-get-snapshot-metadata", Method: http.MethodGet, Path: "/api/v1/snapshots/{snapshot_id}/metadata", Summary: "Get snapshot metadata", Tags: []string{"Snapshots"}},
		func(ctx context.Context, input *snapshotIDInput) (*getSnapshotOutput, error) {
			meta, err := svc.GetSnapshot(ctx, input.SnapshotID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &getSnapshotOutput{}
			out.Body = meta
			return out, nil
		})

	type deleteSnapshotOutput struct {
		Body struct {
			Status string `json:"status"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-delete-snapshot", Method: http.MethodDelete, Path: "/api/v1/snapshots/{snapshot_id}", Summary: "Delete snapshot", Tags: []string{"Snapshots"}},
		func(ctx context.Context, input *snapshotIDInput) (*deleteSnapshotOutput, error) {
			if err := svc.DeleteSnapshot(ctx, input.SnapshotID); err != nil {
				return nil, mapErr(err)
			}
			out := &deleteSnapshotOutput{}
			out.Body.Status = "deleted"
			return out, nil
		})

	type snapshotImageOutput struct {
		ContentType string `header:"Content-Type"`
		Body        []byte
	}
	huma.Register(api, huma.Operation{
		OperationID: "multi-get-snapshot-image",
		Method:      http.MethodGet,
		Path:        "/api/v1/snapshots/{snapshot_id}/image",
		Summary:     "Get snapshot image",
		Tags:        []string{"Snapshots"},
		Responses: map[string]*huma.Response{
			"200": {
				Description: "Snapshot image",
				Content: map[string]*huma.MediaType{
					"image/png": {
						Schema: &huma.Schema{Type: "string", Format: "binary"},
					},
				},
			},
		},
	}, func(ctx context.Context, input *snapshotIDInput) (*snapshotImageOutput, error) {
		data, format, err := svc.ReadSnapshotImage(ctx, input.SnapshotID)
		if err != nil {
			return nil, mapErr(err)
		}
		ct := "image/png"
		if format == "jpeg" {
			ct = "image/jpeg"
		}
		return &snapshotImageOutput{ContentType: ct, Body: data}, nil
	})

	// --- Reload page (session-level, moved under /chart/{chart_id}/) ---

	type reloadOutput struct {
		Body cdpcontrol.ReloadResult
	}
	huma.Register(api, huma.Operation{OperationID: "multi-reload-page", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/page/reload", Summary: "Reload the TradingView page", Description: "Reloads the browser tab. Mode: \"normal\" (default) or \"hard\" (bypass cache, like Shift+F5).", Tags: []string{"Page"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Mode string `json:"mode,omitempty" doc:"Reload mode: normal (default) or hard (bypass cache)" example:"normal" enum:"normal,hard"`
			}
		}) (*reloadOutput, error) {
			result, err := svc.ReloadPage(ctx, input.ChartID, input.Body.Mode)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &reloadOutput{}
			out.Body = result
			return out, nil
		})

	// --- Currency / Unit endpoints (chart-level, paths unchanged) ---

	type currencyInput struct {
		ChartID  string `path:"chart_id"`
		Currency string `query:"currency" required:"true"`
		Pane     int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
	}
	type currencyOutput struct {
		Body struct {
			ChartID  string                  `json:"chart_id"`
			Currency cdpcontrol.CurrencyInfo `json:"currency"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-get-currency", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/currency", Summary: "Get current currency", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *chartIDInput) (*currencyOutput, error) {
			info, err := svc.GetCurrency(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &currencyOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Currency = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-set-currency", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/currency", Summary: "Set currency (use 'null' to reset)", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *currencyInput) (*currencyOutput, error) {
			info, err := svc.SetCurrency(ctx, input.ChartID, input.Currency, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &currencyOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Currency = info
			return out, nil
		})

	type availableCurrenciesOutput struct {
		Body struct {
			ChartID    string                         `json:"chart_id"`
			Currencies []cdpcontrol.AvailableCurrency `json:"currencies"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-get-available-currencies", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/currency/available", Summary: "List available currencies", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *chartIDInput) (*availableCurrenciesOutput, error) {
			list, err := svc.GetAvailableCurrencies(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &availableCurrenciesOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Currencies = list
			return out, nil
		})

	type unitInput struct {
		ChartID string `path:"chart_id"`
		Unit    string `query:"unit" required:"true"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
	}
	type unitOutput struct {
		Body struct {
			ChartID string              `json:"chart_id"`
			Unit    cdpcontrol.UnitInfo `json:"unit"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-get-unit", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/unit", Summary: "Get current unit", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *chartIDInput) (*unitOutput, error) {
			info, err := svc.GetUnit(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &unitOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Unit = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-set-unit", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/unit", Summary: "Set unit (use 'null' to reset)", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *unitInput) (*unitOutput, error) {
			info, err := svc.SetUnit(ctx, input.ChartID, input.Unit, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &unitOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Unit = info
			return out, nil
		})

	type availableUnitsOutput struct {
		Body struct {
			ChartID string                     `json:"chart_id"`
			Units   []cdpcontrol.AvailableUnit `json:"units"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-get-available-units", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/unit/available", Summary: "List available units", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *chartIDInput) (*availableUnitsOutput, error) {
			list, err := svc.GetAvailableUnits(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &availableUnitsOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Units = list
			return out, nil
		})

	// --- Hotlists endpoints (session-level, moved under /chart/{chart_id}/hotlists/) ---

	type hotlistsProbeOutput struct {
		Body cdpcontrol.HotlistsManagerProbe
	}
	huma.Register(api, huma.Operation{OperationID: "multi-probe-hotlists", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/hotlists/probe", Summary: "Probe hotlistsManager() singleton", Tags: []string{"Hotlists"}},
		func(ctx context.Context, input *chartIDInput) (*hotlistsProbeOutput, error) {
			probe, err := svc.ProbeHotlistsManager(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &hotlistsProbeOutput{}
			out.Body = probe
			return out, nil
		})

	type hotlistsProbeDeepOutput struct {
		Body map[string]any
	}
	huma.Register(api, huma.Operation{OperationID: "multi-probe-hotlists-deep", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/hotlists/probe/deep", Summary: "Deep probe hotlistsManager() methods and properties", Tags: []string{"Hotlists"}},
		func(ctx context.Context, input *chartIDInput) (*hotlistsProbeDeepOutput, error) {
			probe, err := svc.ProbeHotlistsManagerDeep(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &hotlistsProbeDeepOutput{}
			out.Body = probe
			return out, nil
		})

	type hotlistMarketsOutput struct {
		Body struct {
			Markets any `json:"markets"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-get-hotlist-markets", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/hotlists/markets", Summary: "Get market organization", Tags: []string{"Hotlists"}},
		func(ctx context.Context, input *chartIDInput) (*hotlistMarketsOutput, error) {
			markets, err := svc.GetHotlistMarkets(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &hotlistMarketsOutput{}
			out.Body.Markets = markets
			return out, nil
		})

	type hotlistExchangesOutput struct {
		Body struct {
			Exchanges []cdpcontrol.HotlistExchangeDetail `json:"exchanges"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-get-hotlist-exchanges", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/hotlists/exchanges", Summary: "List available exchanges with names, flags, groups", Tags: []string{"Hotlists"}},
		func(ctx context.Context, input *chartIDInput) (*hotlistExchangesOutput, error) {
			exchanges, err := svc.GetHotlistExchanges(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &hotlistExchangesOutput{}
			out.Body.Exchanges = exchanges
			return out, nil
		})

	type hotlistResultOutput struct {
		Body cdpcontrol.HotlistResult
	}
	huma.Register(api, huma.Operation{OperationID: "multi-get-one-hotlist", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/hotlists/{exchange}/{group}", Summary: "Get symbols for a specific hotlist", Tags: []string{"Hotlists"}},
		func(ctx context.Context, input *struct {
			ChartID  string `path:"chart_id"`
			Exchange string `path:"exchange"`
			Group    string `path:"group"`
		}) (*hotlistResultOutput, error) {
			result, err := svc.GetOneHotlist(ctx, input.ChartID, input.Exchange, input.Group)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &hotlistResultOutput{}
			out.Body = result
			return out, nil
		})

	// --- Data Window endpoints (chart-level, paths unchanged) ---

	type dataWindowProbeOutput struct {
		Body cdpcontrol.DataWindowProbe
	}
	huma.Register(api, huma.Operation{OperationID: "multi-probe-data-window", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/data-window/probe", Summary: "Discover accessible data window state", Tags: []string{"Introspection"}},
		func(ctx context.Context, input *chartIDInput) (*dataWindowProbeOutput, error) {
			probe, err := svc.ProbeDataWindow(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &dataWindowProbeOutput{}
			out.Body = probe
			return out, nil
		})
}
