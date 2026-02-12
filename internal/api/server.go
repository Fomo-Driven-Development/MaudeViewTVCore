package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/dgnsrekt/tv_agent/internal/cdpcontrol"
	"github.com/dgnsrekt/tv_agent/internal/snapshot"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Service interface {
	ListCharts(ctx context.Context) ([]cdpcontrol.ChartInfo, error)
	GetActiveChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error)
	GetSymbolInfo(ctx context.Context, chartID string) (cdpcontrol.SymbolInfo, error)
	GetSymbol(ctx context.Context, chartID string) (string, error)
	SetSymbol(ctx context.Context, chartID, symbol string) (string, error)
	GetResolution(ctx context.Context, chartID string) (string, error)
	SetResolution(ctx context.Context, chartID, resolution string) (string, error)
	ExecuteAction(ctx context.Context, chartID, actionID string) error
	ListStudies(ctx context.Context, chartID string) ([]cdpcontrol.Study, error)
	AddStudy(ctx context.Context, chartID, name string, inputs map[string]any, forceOverlay bool) (cdpcontrol.Study, error)
	RemoveStudy(ctx context.Context, chartID, studyID string) error
	GetStudyInputs(ctx context.Context, chartID, studyID string) (cdpcontrol.StudyDetail, error)
	ModifyStudyInputs(ctx context.Context, chartID, studyID string, inputs map[string]any) (cdpcontrol.StudyDetail, error)
	ListWatchlists(ctx context.Context) ([]cdpcontrol.WatchlistInfo, error)
	GetActiveWatchlist(ctx context.Context) (cdpcontrol.WatchlistDetail, error)
	SetActiveWatchlist(ctx context.Context, id string) (cdpcontrol.WatchlistInfo, error)
	GetWatchlist(ctx context.Context, id string) (cdpcontrol.WatchlistDetail, error)
	CreateWatchlist(ctx context.Context, name string) (cdpcontrol.WatchlistInfo, error)
	RenameWatchlist(ctx context.Context, id, name string) (cdpcontrol.WatchlistInfo, error)
	DeleteWatchlist(ctx context.Context, id string) error
	AddWatchlistSymbols(ctx context.Context, id string, symbols []string) (cdpcontrol.WatchlistDetail, error)
	RemoveWatchlistSymbols(ctx context.Context, id string, symbols []string) (cdpcontrol.WatchlistDetail, error)
	FlagSymbol(ctx context.Context, id, symbol string) error
	Zoom(ctx context.Context, chartID, direction string) error
	Scroll(ctx context.Context, chartID string, bars int) error
	ScrollToRealtime(ctx context.Context, chartID string) error
	GoToDate(ctx context.Context, chartID string, timestamp int64) error
	GetVisibleRange(ctx context.Context, chartID string) (cdpcontrol.VisibleRange, error)
	SetVisibleRange(ctx context.Context, chartID string, from, to float64) (cdpcontrol.VisibleRange, error)
	ResetScales(ctx context.Context, chartID string) error
	ProbeChartApi(ctx context.Context, chartID string) (cdpcontrol.ChartApiProbe, error)
	ProbeChartApiDeep(ctx context.Context, chartID string) (map[string]any, error)
	ResolveSymbol(ctx context.Context, chartID, symbol string) (cdpcontrol.ResolvedSymbolInfo, error)
	SwitchTimezone(ctx context.Context, chartID, tz string) error
	ProbeReplayManager(ctx context.Context, chartID string) (cdpcontrol.ReplayManagerProbe, error)
	ProbeReplayManagerDeep(ctx context.Context, chartID string) (map[string]any, error)
	ScanReplayActivation(ctx context.Context, chartID string) (map[string]any, error)
	ActivateReplay(ctx context.Context, chartID string, date float64) (map[string]any, error)
	ActivateReplayAuto(ctx context.Context, chartID string) (map[string]any, error)
	DeactivateReplay(ctx context.Context, chartID string) error
	GetReplayStatus(ctx context.Context, chartID string) (cdpcontrol.ReplayStatus, error)
	StartReplay(ctx context.Context, chartID string, point float64) error
	StopReplay(ctx context.Context, chartID string) error
	ReplayStep(ctx context.Context, chartID string) error
	StartAutoplay(ctx context.Context, chartID string) error
	StopAutoplay(ctx context.Context, chartID string) error
	ResetReplay(ctx context.Context, chartID string) error
	ChangeAutoplayDelay(ctx context.Context, chartID string, delay float64) (float64, error)
	ProbeBacktestingApi(ctx context.Context, chartID string) (cdpcontrol.StrategyApiProbe, error)
	ListStrategies(ctx context.Context, chartID string) (any, error)
	GetActiveStrategy(ctx context.Context, chartID string) (map[string]any, error)
	SetActiveStrategy(ctx context.Context, chartID, strategyID string) error
	SetStrategyInput(ctx context.Context, chartID, name string, value any) error
	GetStrategyReport(ctx context.Context, chartID string) (map[string]any, error)
	GetStrategyDateRange(ctx context.Context, chartID string) (any, error)
	StrategyGotoDate(ctx context.Context, chartID string, timestamp float64, belowBar bool) error
	ScanAlertsAccess(ctx context.Context, chartID string) (map[string]any, error)
	ProbeAlertsRestApi(ctx context.Context, chartID string) (cdpcontrol.AlertsApiProbe, error)
	ProbeAlertsRestApiDeep(ctx context.Context, chartID string) (map[string]any, error)
	ListAlerts(ctx context.Context) (any, error)
	GetAlerts(ctx context.Context, ids []string) (any, error)
	CreateAlert(ctx context.Context, params map[string]any) (any, error)
	ModifyAlert(ctx context.Context, params map[string]any) (any, error)
	DeleteAlerts(ctx context.Context, ids []string) error
	StopAlerts(ctx context.Context, ids []string) error
	RestartAlerts(ctx context.Context, ids []string) error
	CloneAlerts(ctx context.Context, ids []string) error
	ListFires(ctx context.Context) (any, error)
	DeleteFires(ctx context.Context, ids []string) error
	DeleteAllFires(ctx context.Context) error
	ListDrawings(ctx context.Context, chartID string) ([]cdpcontrol.Shape, error)
	GetDrawing(ctx context.Context, chartID, shapeID string) (map[string]any, error)
	CreateDrawing(ctx context.Context, chartID string, point cdpcontrol.ShapePoint, options map[string]any) (string, error)
	CreateMultipointDrawing(ctx context.Context, chartID string, points []cdpcontrol.ShapePoint, options map[string]any) (string, error)
	CloneDrawing(ctx context.Context, chartID, shapeID string) (string, error)
	RemoveDrawing(ctx context.Context, chartID, shapeID string, disableUndo bool) error
	RemoveAllDrawings(ctx context.Context, chartID string) error
	GetDrawingToggles(ctx context.Context, chartID string) (cdpcontrol.DrawingToggles, error)
	SetHideDrawings(ctx context.Context, chartID string, val bool) error
	SetLockDrawings(ctx context.Context, chartID string, val bool) error
	SetMagnet(ctx context.Context, chartID string, enabled bool, mode int) error
	SetDrawingVisibility(ctx context.Context, chartID, shapeID string, visible bool) error
	GetDrawingTool(ctx context.Context, chartID string) (string, error)
	SetDrawingTool(ctx context.Context, chartID, tool string) error
	SetDrawingZOrder(ctx context.Context, chartID, shapeID, action string) error
	ExportDrawingsState(ctx context.Context, chartID string) (any, error)
	ImportDrawingsState(ctx context.Context, chartID string, state any) error
	TakeSnapshot(ctx context.Context, chartID, format, quality string) (snapshot.SnapshotMeta, error)
	ListSnapshots(ctx context.Context) ([]snapshot.SnapshotMeta, error)
	GetSnapshot(ctx context.Context, id string) (snapshot.SnapshotMeta, error)
	ReadSnapshotImage(ctx context.Context, id string) ([]byte, string, error)
	DeleteSnapshot(ctx context.Context, id string) error
}

func NewServer(svc Service) http.Handler {
	router := chi.NewMux()
	router.Use(middleware.RequestID)
	router.Use(requestLogger)
	router.Use(middleware.Recoverer)

	cfg := huma.DefaultConfig("TV Agent Controller API", "1.0.0")
	cfg.DocsPath = ""
	api := humachi.New(router, cfg)

	router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(docsHTML))
	})

	type healthOutput struct {
		Body struct {
			Status string `json:"status"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "health", Method: http.MethodGet, Path: "/health", Summary: "Health check", Tags: []string{"Health"}},
		func(ctx context.Context, input *struct{}) (*healthOutput, error) {
			out := &healthOutput{}
			out.Body.Status = "ok"
			return out, nil
		})

	type listChartsOutput struct {
		Body struct {
			Charts []cdpcontrol.ChartInfo `json:"charts"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "list-charts", Method: http.MethodGet, Path: "/api/v1/charts", Summary: "List chart tabs", Tags: []string{"Charts"}},
		func(ctx context.Context, input *struct{}) (*listChartsOutput, error) {
			charts, err := svc.ListCharts(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &listChartsOutput{}
			out.Body.Charts = charts
			return out, nil
		})

	type activeChartOutput struct {
		Body cdpcontrol.ActiveChartInfo
	}
	huma.Register(api, huma.Operation{OperationID: "get-active-chart", Method: http.MethodGet, Path: "/api/v1/charts/active", Summary: "Get active chart", Tags: []string{"Charts"}},
		func(ctx context.Context, input *struct{}) (*activeChartOutput, error) {
			info, err := svc.GetActiveChart(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &activeChartOutput{}
			out.Body = info
			return out, nil
		})

	type chartIDInput struct {
		ChartID string `path:"chart_id"`
	}
	type symbolInput struct {
		ChartID string `path:"chart_id"`
		Symbol  string `query:"symbol" required:"true"`
	}
	type symbolOutput struct {
		Body struct {
			ChartID         string `json:"chart_id"`
			RequestedSymbol string `json:"requested_symbol,omitempty"`
			CurrentSymbol   string `json:"current_symbol"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "get-symbol", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/symbol", Summary: "Get chart symbol", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *chartIDInput) (*symbolOutput, error) {
			symbol, err := svc.GetSymbol(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &symbolOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.CurrentSymbol = symbol
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-symbol", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/symbol", Summary: "Set chart symbol", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *symbolInput) (*symbolOutput, error) {
			current, err := svc.SetSymbol(ctx, input.ChartID, input.Symbol)
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
	huma.Register(api, huma.Operation{OperationID: "get-symbol-info", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/symbol/info", Summary: "Get extended symbol metadata", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *chartIDInput) (*symbolInfoOutput, error) {
			info, err := svc.GetSymbolInfo(ctx, input.ChartID)
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
	}
	type resolutionOutput struct {
		Body struct {
			ChartID             string `json:"chart_id"`
			RequestedResolution string `json:"requested_resolution,omitempty"`
			CurrentResolution   string `json:"current_resolution"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "get-resolution", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/resolution", Summary: "Get chart resolution", Tags: []string{"Resolution"}},
		func(ctx context.Context, input *chartIDInput) (*resolutionOutput, error) {
			resolution, err := svc.GetResolution(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &resolutionOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.CurrentResolution = resolution
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-resolution", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/resolution", Summary: "Set chart resolution", Tags: []string{"Resolution"}},
		func(ctx context.Context, input *resolutionInput) (*resolutionOutput, error) {
			current, err := svc.SetResolution(ctx, input.ChartID, input.Resolution)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &resolutionOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.RequestedResolution = input.Resolution
			out.Body.CurrentResolution = current
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

	huma.Register(api, huma.Operation{OperationID: "execute-action", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/action", Summary: "Execute chart action", Tags: []string{"Action"}},
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

	type studyListOutput struct {
		Body struct {
			ChartID string             `json:"chart_id"`
			Studies []cdpcontrol.Study `json:"studies"`
		}
	}
	type addStudyInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			Name         string         `json:"name" required:"true"`
			Inputs       map[string]any `json:"inputs,omitempty"`
			ForceOverlay bool           `json:"force_overlay,omitempty"`
		}
	}
	type addStudyOutput struct {
		Body struct {
			ChartID string           `json:"chart_id"`
			Study   cdpcontrol.Study `json:"study"`
			Status  string           `json:"status"`
		}
	}
	type removeStudyInput struct {
		ChartID string `path:"chart_id"`
		StudyID string `path:"study_id"`
	}

	huma.Register(api, huma.Operation{OperationID: "list-studies", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/studies", Summary: "List studies", Tags: []string{"Studies"}},
		func(ctx context.Context, input *chartIDInput) (*studyListOutput, error) {
			studies, err := svc.ListStudies(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &studyListOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Studies = studies
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "add-study", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/studies", Summary: "Add study", Tags: []string{"Studies"}},
		func(ctx context.Context, input *addStudyInput) (*addStudyOutput, error) {
			study, err := svc.AddStudy(ctx, input.ChartID, input.Body.Name, input.Body.Inputs, input.Body.ForceOverlay)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &addStudyOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Study = study
			out.Body.Status = "added"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "remove-study", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/studies/{study_id}", Summary: "Remove study", Tags: []string{"Studies"}},
		func(ctx context.Context, input *removeStudyInput) (*struct{}, error) {
			if err := svc.RemoveStudy(ctx, input.ChartID, input.StudyID); err != nil {
				return nil, mapErr(err)
			}
			return &struct{}{}, nil
		})

	type getStudyOutput struct {
		Body struct {
			ChartID string                 `json:"chart_id"`
			Study   cdpcontrol.StudyDetail `json:"study"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "get-study", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/studies/{study_id}", Summary: "Get study detail with inputs", Tags: []string{"Studies"}},
		func(ctx context.Context, input *removeStudyInput) (*getStudyOutput, error) {
			detail, err := svc.GetStudyInputs(ctx, input.ChartID, input.StudyID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &getStudyOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Study = detail
			return out, nil
		})

	type modifyStudyInput struct {
		ChartID string `path:"chart_id"`
		StudyID string `path:"study_id"`
		Body    struct {
			Inputs map[string]any `json:"inputs" required:"true"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "modify-study", Method: http.MethodPatch, Path: "/api/v1/chart/{chart_id}/studies/{study_id}", Summary: "Modify study input parameters", Tags: []string{"Studies"}},
		func(ctx context.Context, input *modifyStudyInput) (*getStudyOutput, error) {
			detail, err := svc.ModifyStudyInputs(ctx, input.ChartID, input.StudyID, input.Body.Inputs)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &getStudyOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Study = detail
			return out, nil
		})

	// --- Watchlist endpoints ---

	type listWatchlistsOutput struct {
		Body struct {
			Watchlists []cdpcontrol.WatchlistInfo `json:"watchlists"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "list-watchlists", Method: http.MethodGet, Path: "/api/v1/watchlists", Summary: "List all watchlists", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct{}) (*listWatchlistsOutput, error) {
			wls, err := svc.ListWatchlists(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &listWatchlistsOutput{}
			out.Body.Watchlists = wls
			return out, nil
		})

	type watchlistDetailOutput struct {
		Body cdpcontrol.WatchlistDetail
	}
	huma.Register(api, huma.Operation{OperationID: "get-active-watchlist", Method: http.MethodGet, Path: "/api/v1/watchlists/active", Summary: "Get active watchlist with symbols", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct{}) (*watchlistDetailOutput, error) {
			detail, err := svc.GetActiveWatchlist(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	type setActiveWatchlistInput struct {
		Body struct {
			ID string `json:"id" required:"true"`
		}
	}
	type watchlistInfoOutput struct {
		Body cdpcontrol.WatchlistInfo
	}
	huma.Register(api, huma.Operation{OperationID: "set-active-watchlist", Method: http.MethodPut, Path: "/api/v1/watchlists/active", Summary: "Set active watchlist by ID", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *setActiveWatchlistInput) (*watchlistInfoOutput, error) {
			info, err := svc.SetActiveWatchlist(ctx, input.Body.ID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	type watchlistIDInput struct {
		WatchlistID string `path:"watchlist_id"`
	}

	huma.Register(api, huma.Operation{OperationID: "create-watchlist", Method: http.MethodPost, Path: "/api/v1/watchlists", Summary: "Create new watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Name string `json:"name" required:"true"`
			}
		}) (*watchlistInfoOutput, error) {
			info, err := svc.CreateWatchlist(ctx, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "get-watchlist", Method: http.MethodGet, Path: "/api/v1/watchlist/{watchlist_id}", Summary: "Get watchlist detail with symbols", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *watchlistIDInput) (*watchlistDetailOutput, error) {
			detail, err := svc.GetWatchlist(ctx, input.WatchlistID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "rename-watchlist", Method: http.MethodPatch, Path: "/api/v1/watchlist/{watchlist_id}", Summary: "Rename watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			WatchlistID string `path:"watchlist_id"`
			Body        struct {
				Name string `json:"name" required:"true"`
			}
		}) (*watchlistInfoOutput, error) {
			info, err := svc.RenameWatchlist(ctx, input.WatchlistID, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "delete-watchlist", Method: http.MethodDelete, Path: "/api/v1/watchlist/{watchlist_id}", Summary: "Delete watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *watchlistIDInput) (*struct{}, error) {
			if err := svc.DeleteWatchlist(ctx, input.WatchlistID); err != nil {
				return nil, mapErr(err)
			}
			return &struct{}{}, nil
		})

	type symbolsBodyInput struct {
		WatchlistID string `path:"watchlist_id"`
		Body        struct {
			Symbols []string `json:"symbols" required:"true"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "add-symbols", Method: http.MethodPost, Path: "/api/v1/watchlist/{watchlist_id}/symbols", Summary: "Add symbol(s) to watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *symbolsBodyInput) (*watchlistDetailOutput, error) {
			detail, err := svc.AddWatchlistSymbols(ctx, input.WatchlistID, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "remove-symbols", Method: http.MethodDelete, Path: "/api/v1/watchlist/{watchlist_id}/symbols", Summary: "Remove symbol(s) from watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *symbolsBodyInput) (*watchlistDetailOutput, error) {
			detail, err := svc.RemoveWatchlistSymbols(ctx, input.WatchlistID, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "flag-symbol", Method: http.MethodPost, Path: "/api/v1/watchlist/{watchlist_id}/flag", Summary: "Flag/unflag a symbol", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			WatchlistID string `path:"watchlist_id"`
			Body        struct {
				Symbol string `json:"symbol" required:"true"`
			}
		}) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			if err := svc.FlagSymbol(ctx, input.WatchlistID, input.Body.Symbol); err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					Status string `json:"status"`
				}
			}{}
			out.Body.Status = "toggled"
			return out, nil
		})

	// --- Navigation endpoints ---

	type navStatusOutput struct {
		Body struct {
			ChartID string `json:"chart_id"`
			Status  string `json:"status"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "zoom", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/zoom", Summary: "Zoom in or out", Tags: []string{"Navigation"}},
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

	huma.Register(api, huma.Operation{OperationID: "scroll", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/scroll", Summary: "Scroll chart by bars", Tags: []string{"Navigation"}},
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

	huma.Register(api, huma.Operation{OperationID: "scroll-to-realtime", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/scroll/to-realtime", Summary: "Scroll to latest bar", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.ScrollToRealtime(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "go-to-date", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/go-to-date", Summary: "Navigate chart to a specific date", Tags: []string{"Navigation"}},
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

	type visibleRangeOutput struct {
		Body struct {
			ChartID string  `json:"chart_id"`
			From    float64 `json:"from"`
			To      float64 `json:"to"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "get-visible-range", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/visible-range", Summary: "Get visible bar range", Tags: []string{"Navigation"}},
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

	huma.Register(api, huma.Operation{OperationID: "set-visible-range", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/visible-range", Summary: "Set visible bar range", Tags: []string{"Navigation"}},
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

	huma.Register(api, huma.Operation{OperationID: "reset-scales", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/reset-scales", Summary: "Reset all chart scales", Tags: []string{"Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.ResetScales(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			return out, nil
		})

	// --- ChartAPI endpoints ---

	type chartApiProbeOutput struct {
		Body cdpcontrol.ChartApiProbe
	}
	huma.Register(api, huma.Operation{OperationID: "probe-chart-api", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/chart-api/probe", Summary: "Probe chartApi() singleton", Tags: []string{"ChartAPI"}},
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
	huma.Register(api, huma.Operation{OperationID: "probe-chart-api-deep", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/chart-api/probe/deep", Summary: "Deep probe chartApi() methods and state", Tags: []string{"ChartAPI"}},
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
	huma.Register(api, huma.Operation{OperationID: "resolve-symbol", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/chart-api/resolve-symbol", Summary: "Resolve metadata for any symbol", Tags: []string{"ChartAPI"}},
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
	huma.Register(api, huma.Operation{OperationID: "switch-timezone", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/chart-api/timezone", Summary: "Switch chart timezone", Tags: []string{"ChartAPI"}},
		func(ctx context.Context, input *switchTimezoneInput) (*navStatusOutput, error) {
			if err := svc.SwitchTimezone(ctx, input.ChartID, input.Body.Timezone); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			return out, nil
		})

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

	huma.Register(api, huma.Operation{OperationID: "replay-step", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/replay/step", Summary: "Step one bar forward in replay", Tags: []string{"Replay"}},
		func(ctx context.Context, input *chartIDInput) (*navStatusOutput, error) {
			if err := svc.ReplayStep(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			out := &navStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "stepped"
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

	// --- Strategy endpoints ---

	type strategyProbeOutput struct {
		Body cdpcontrol.StrategyApiProbe
	}
	huma.Register(api, huma.Operation{OperationID: "probe-strategy-api", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/strategy/probe", Summary: "Probe backtesting strategy API", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *chartIDInput) (*strategyProbeOutput, error) {
			probe, err := svc.ProbeBacktestingApi(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &strategyProbeOutput{}
			out.Body = probe
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "list-strategies", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/strategy/list", Summary: "List all loaded strategies", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *chartIDInput) (*struct{ Body struct{ Strategies any `json:"strategies"` } }, error) {
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

	huma.Register(api, huma.Operation{OperationID: "get-active-strategy", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/strategy/active", Summary: "Get active strategy with inputs and metadata", Tags: []string{"Strategy"}},
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
	huma.Register(api, huma.Operation{OperationID: "set-active-strategy", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/strategy/active", Summary: "Set active strategy by entity ID", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *setStrategyInput) (*struct{ Body struct{ Status string `json:"status"` } }, error) {
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
	huma.Register(api, huma.Operation{OperationID: "set-strategy-input", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/strategy/input", Summary: "Set a strategy input parameter", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *setStrategyInputInput) (*struct{ Body struct{ Status string `json:"status"` } }, error) {
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

	huma.Register(api, huma.Operation{OperationID: "get-strategy-report", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/strategy/report", Summary: "Get backtest report data", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *chartIDInput) (*struct{ Body map[string]any }, error) {
			result, err := svc.GetStrategyReport(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct{ Body map[string]any }{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "get-strategy-date-range", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/strategy/date-range", Summary: "Get backtest date range", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *chartIDInput) (*struct{ Body struct{ DateRange any `json:"date_range"` } }, error) {
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
	huma.Register(api, huma.Operation{OperationID: "strategy-goto-date", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/strategy/goto", Summary: "Navigate chart to a specific trade/bar timestamp", Tags: []string{"Strategy"}},
		func(ctx context.Context, input *strategyGotoInput) (*struct{ Body struct{ Status string `json:"status"` } }, error) {
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

	// --- Drawing/Shape endpoints ---

	type drawingListOutput struct {
		Body struct {
			ChartID string             `json:"chart_id"`
			Shapes  []cdpcontrol.Shape `json:"shapes"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "list-drawings", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/drawings", Summary: "List all drawings on chart", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *chartIDInput) (*drawingListOutput, error) {
			shapes, err := svc.ListDrawings(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &drawingListOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Shapes = shapes
			return out, nil
		})

	type shapeIDInput struct {
		ChartID string `path:"chart_id"`
		ShapeID string `path:"shape_id"`
	}
	type drawingDetailOutput struct {
		Body map[string]any
	}
	huma.Register(api, huma.Operation{OperationID: "get-drawing", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/drawings/{shape_id}", Summary: "Get drawing by ID", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *shapeIDInput) (*drawingDetailOutput, error) {
			detail, err := svc.GetDrawing(ctx, input.ChartID, input.ShapeID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &drawingDetailOutput{}
			out.Body = detail
			return out, nil
		})

	type createDrawingInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			Point   cdpcontrol.ShapePoint `json:"point" required:"true"`
			Options map[string]any        `json:"options" required:"true"`
		}
	}
	type createDrawingOutput struct {
		Body struct {
			ChartID string `json:"chart_id"`
			ID      string `json:"id"`
			Status  string `json:"status"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "create-drawing", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/drawings", Summary: "Create a single-point drawing", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *createDrawingInput) (*createDrawingOutput, error) {
			id, err := svc.CreateDrawing(ctx, input.ChartID, input.Body.Point, input.Body.Options)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createDrawingOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.ID = id
			out.Body.Status = "created"
			return out, nil
		})

	type createMultipointInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			Points  []cdpcontrol.ShapePoint `json:"points" required:"true"`
			Options map[string]any          `json:"options" required:"true"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "create-multipoint-drawing", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/drawings/multipoint", Summary: "Create a multi-point drawing", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *createMultipointInput) (*createDrawingOutput, error) {
			id, err := svc.CreateMultipointDrawing(ctx, input.ChartID, input.Body.Points, input.Body.Options)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createDrawingOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.ID = id
			out.Body.Status = "created"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "clone-drawing", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/drawings/{shape_id}/clone", Summary: "Clone a drawing", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *shapeIDInput) (*createDrawingOutput, error) {
			id, err := svc.CloneDrawing(ctx, input.ChartID, input.ShapeID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createDrawingOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.ID = id
			out.Body.Status = "cloned"
			return out, nil
		})

	type removeDrawingInput struct {
		ChartID    string `path:"chart_id"`
		ShapeID    string `path:"shape_id"`
		DisableUndo bool  `query:"disable_undo"`
	}
	huma.Register(api, huma.Operation{OperationID: "remove-drawing", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/drawings/{shape_id}", Summary: "Remove a drawing", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *removeDrawingInput) (*struct{}, error) {
			if err := svc.RemoveDrawing(ctx, input.ChartID, input.ShapeID, input.DisableUndo); err != nil {
				return nil, mapErr(err)
			}
			return &struct{}{}, nil
		})

	huma.Register(api, huma.Operation{OperationID: "remove-all-drawings", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/drawings", Summary: "Remove all drawings", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *chartIDInput) (*struct{}, error) {
			if err := svc.RemoveAllDrawings(ctx, input.ChartID); err != nil {
				return nil, mapErr(err)
			}
			return &struct{}{}, nil
		})

	type drawingTogglesOutput struct {
		Body struct {
			ChartID string                  `json:"chart_id"`
			Toggles cdpcontrol.DrawingToggles `json:"toggles"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "get-drawing-toggles", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/drawings/toggles", Summary: "Get drawing toggle states", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *chartIDInput) (*drawingTogglesOutput, error) {
			toggles, err := svc.GetDrawingToggles(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &drawingTogglesOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Toggles = toggles
			return out, nil
		})

	type drawingStatusOutput struct {
		Body struct {
			ChartID string `json:"chart_id"`
			Status  string `json:"status"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "set-hide-drawings", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/toggles/hide", Summary: "Hide or show all drawings", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Value bool `json:"value"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.SetHideDrawings(ctx, input.ChartID, input.Body.Value); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "set"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-lock-drawings", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/toggles/lock", Summary: "Lock or unlock all drawings", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Value bool `json:"value"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.SetLockDrawings(ctx, input.ChartID, input.Body.Value); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "set"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-magnet", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/toggles/magnet", Summary: "Set magnet mode", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Enabled bool `json:"enabled"`
				Mode    int  `json:"mode,omitempty"`
			}
		}) (*drawingStatusOutput, error) {
			mode := input.Body.Mode
			if mode == 0 && !input.Body.Enabled {
				mode = -1
			}
			if err := svc.SetMagnet(ctx, input.ChartID, input.Body.Enabled, mode); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "set"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-drawing-visibility", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/{shape_id}/visibility", Summary: "Set drawing visibility", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			ShapeID string `path:"shape_id"`
			Body    struct {
				Visible bool `json:"visible"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.SetDrawingVisibility(ctx, input.ChartID, input.ShapeID, input.Body.Visible); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "set"
			return out, nil
		})

	type drawingToolOutput struct {
		Body struct {
			ChartID string `json:"chart_id"`
			Tool    string `json:"tool"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "get-drawing-tool", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/drawings/tool", Summary: "Get selected drawing tool", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *chartIDInput) (*drawingToolOutput, error) {
			tool, err := svc.GetDrawingTool(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &drawingToolOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Tool = tool
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-drawing-tool", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/tool", Summary: "Set drawing tool", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Tool string `json:"tool" required:"true"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.SetDrawingTool(ctx, input.ChartID, input.Body.Tool); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "set"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-drawing-z-order", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/drawings/{shape_id}/z-order", Summary: "Change drawing z-order", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			ShapeID string `path:"shape_id"`
			Body    struct {
				Action string `json:"action" required:"true" doc:"bring_forward, bring_to_front, send_backward, send_to_back"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.SetDrawingZOrder(ctx, input.ChartID, input.ShapeID, input.Body.Action); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			return out, nil
		})

	type drawingsStateOutput struct {
		Body struct {
			ChartID string `json:"chart_id"`
			State   any    `json:"state"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "export-drawings-state", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/drawings/state", Summary: "Export all drawings state", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *chartIDInput) (*drawingsStateOutput, error) {
			state, err := svc.ExportDrawingsState(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &drawingsStateOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.State = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "import-drawings-state", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/state", Summary: "Import drawings state", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				State any `json:"state" required:"true"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.ImportDrawingsState(ctx, input.ChartID, input.Body.State); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "imported"
			return out, nil
		})

	// --- Snapshot endpoints ---

	type takeSnapshotOutput struct {
		Body struct {
			Snapshot snapshot.SnapshotMeta `json:"snapshot"`
			URL      string               `json:"url"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "take-snapshot", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/snapshot", Summary: "Take chart snapshot", Tags: []string{"Snapshots"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Format  string `json:"format,omitempty"`
				Quality string `json:"quality,omitempty"`
			}
		}) (*takeSnapshotOutput, error) {
			meta, err := svc.TakeSnapshot(ctx, input.ChartID, input.Body.Format, input.Body.Quality)
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
	huma.Register(api, huma.Operation{OperationID: "list-snapshots", Method: http.MethodGet, Path: "/api/v1/snapshots", Summary: "List snapshots", Tags: []string{"Snapshots"}},
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
	huma.Register(api, huma.Operation{OperationID: "get-snapshot", Method: http.MethodGet, Path: "/api/v1/snapshots/{snapshot_id}", Summary: "Get snapshot metadata", Tags: []string{"Snapshots"}},
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
	huma.Register(api, huma.Operation{OperationID: "delete-snapshot", Method: http.MethodDelete, Path: "/api/v1/snapshots/{snapshot_id}", Summary: "Delete snapshot", Tags: []string{"Snapshots"}},
		func(ctx context.Context, input *snapshotIDInput) (*deleteSnapshotOutput, error) {
			if err := svc.DeleteSnapshot(ctx, input.SnapshotID); err != nil {
				return nil, mapErr(err)
			}
			out := &deleteSnapshotOutput{}
			out.Body.Status = "deleted"
			return out, nil
		})

	// Raw chi handler for serving snapshot image bytes (Huma doesn't support binary responses).
	router.Get("/api/v1/snapshots/{snapshot_id}/image", func(w http.ResponseWriter, r *http.Request) {
		snapshotID := chi.URLParam(r, "snapshot_id")
		data, format, err := svc.ReadSnapshotImage(r.Context(), snapshotID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		ct := "image/png"
		if format == "jpeg" {
			ct = "image/jpeg"
		}
		w.Header().Set("Content-Type", ct)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		_, _ = w.Write(data)
	})

	return router
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	var coded *cdpcontrol.CodedError
	if errors.As(err, &coded) {
		switch coded.Code {
		case cdpcontrol.CodeValidation:
			return huma.Error400BadRequest(coded.Message)
		case cdpcontrol.CodeChartNotFound, cdpcontrol.CodeSnapshotNotFound:
			return huma.Error404NotFound(coded.Message)
		case cdpcontrol.CodeEvalTimeout:
			return huma.Error504GatewayTimeout(coded.Message)
		case cdpcontrol.CodeAPIUnavailable, cdpcontrol.CodeCDPUnavailable:
			return huma.Error502BadGateway(coded.Message)
		default:
			return huma.Error500InternalServerError(fmt.Sprintf("%s: %s", coded.Code, coded.Message))
		}
	}
	return huma.Error500InternalServerError(err.Error())
}
