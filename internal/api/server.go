package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/snapshot"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Service interface {
	ListCharts(ctx context.Context) ([]cdpcontrol.ChartInfo, error)
	GetActiveChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error)
	GetSymbolInfo(ctx context.Context, chartID string, pane int) (cdpcontrol.SymbolInfo, error)
	GetSymbol(ctx context.Context, chartID string, pane int) (string, error)
	SetSymbol(ctx context.Context, chartID, symbol string, pane int) (string, error)
	GetResolution(ctx context.Context, chartID string, pane int) (string, error)
	SetResolution(ctx context.Context, chartID, resolution string, pane int) (string, error)
	GetChartType(ctx context.Context, chartID string, pane int) (int, error)
	SetChartType(ctx context.Context, chartID string, chartType int, pane int) (int, error)
	ExecuteAction(ctx context.Context, chartID, actionID string) error
	ListStudies(ctx context.Context, chartID string, pane int) ([]cdpcontrol.Study, error)
	AddStudy(ctx context.Context, chartID, name string, inputs map[string]any, forceOverlay bool, pane int) (cdpcontrol.Study, error)
	RemoveStudy(ctx context.Context, chartID, studyID string, pane int) error
	GetStudyInputs(ctx context.Context, chartID, studyID string, pane int) (cdpcontrol.StudyDetail, error)
	ModifyStudyInputs(ctx context.Context, chartID, studyID string, inputs map[string]any, pane int) (cdpcontrol.StudyDetail, error)
	AddCompare(ctx context.Context, chartID, symbol, mode, source string, pane int) (cdpcontrol.Study, error)
	ListCompares(ctx context.Context, chartID string, pane int) ([]cdpcontrol.Study, error)
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
	ResetView(ctx context.Context, chartID string) error
	UndoChart(ctx context.Context, chartID string) error
	RedoChart(ctx context.Context, chartID string) error
	GoToDate(ctx context.Context, chartID string, timestamp int64) error
	GetVisibleRange(ctx context.Context, chartID string) (cdpcontrol.VisibleRange, error)
	SetVisibleRange(ctx context.Context, chartID string, from, to float64) (cdpcontrol.VisibleRange, error)
	SetTimeFrame(ctx context.Context, chartID, preset, resolution string, pane int) (cdpcontrol.TimeFrameResult, error)
	ResetScales(ctx context.Context, chartID string) error
	GetChartToggles(ctx context.Context, chartID string, pane int) (cdpcontrol.ChartToggles, error)
	ToggleLogScale(ctx context.Context, chartID string) error
	ToggleAutoScale(ctx context.Context, chartID string) error
	ToggleExtendedHours(ctx context.Context, chartID string) error
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
	ReplayStep(ctx context.Context, chartID string, count int) error
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
	ListDrawings(ctx context.Context, chartID string, pane int) ([]cdpcontrol.Shape, error)
	GetDrawing(ctx context.Context, chartID, shapeID string, pane int) (map[string]any, error)
	CreateDrawing(ctx context.Context, chartID string, point cdpcontrol.ShapePoint, options map[string]any, pane int) (string, error)
	CreateMultipointDrawing(ctx context.Context, chartID string, points []cdpcontrol.ShapePoint, options map[string]any, pane int) (string, error)
	CreateTweetDrawing(ctx context.Context, chartID string, tweetURL string, pane int) (cdpcontrol.TweetDrawingResult, error)
	CloneDrawing(ctx context.Context, chartID, shapeID string, pane int) (string, error)
	RemoveDrawing(ctx context.Context, chartID, shapeID string, disableUndo bool, pane int) error
	RemoveAllDrawings(ctx context.Context, chartID string, pane int) error
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
	BrowserScreenshot(ctx context.Context, format string, quality int, fullPage bool, notes string) (snapshot.SnapshotMeta, error)
	TakeSnapshot(ctx context.Context, chartID, format, quality, notes string, pane int) (snapshot.SnapshotMeta, error)
	GetPaneInfo(ctx context.Context) (cdpcontrol.PanesResult, error)
	ListSnapshots(ctx context.Context) ([]snapshot.SnapshotMeta, error)
	GetSnapshot(ctx context.Context, id string) (snapshot.SnapshotMeta, error)
	ReadSnapshotImage(ctx context.Context, id string) ([]byte, string, error)
	DeleteSnapshot(ctx context.Context, id string) error
	ReloadPage(ctx context.Context, mode string) (cdpcontrol.ReloadResult, error)
	TogglePineEditor(ctx context.Context) (cdpcontrol.PineState, error)
	GetPineStatus(ctx context.Context) (cdpcontrol.PineState, error)
	GetPineSource(ctx context.Context) (cdpcontrol.PineState, error)
	SetPineSource(ctx context.Context, source string) (cdpcontrol.PineState, error)
	SavePineScript(ctx context.Context) (cdpcontrol.PineState, error)
	AddPineToChart(ctx context.Context) (cdpcontrol.PineState, error)
	GetPineConsole(ctx context.Context) ([]cdpcontrol.PineConsoleMessage, error)
	PineUndo(ctx context.Context) (cdpcontrol.PineState, error)
	PineRedo(ctx context.Context) (cdpcontrol.PineState, error)
	PineNewIndicator(ctx context.Context) (cdpcontrol.PineState, error)
	PineNewStrategy(ctx context.Context) (cdpcontrol.PineState, error)
	PineOpenScript(ctx context.Context, name string) (cdpcontrol.PineState, error)
	PineFindReplace(ctx context.Context, find, replace string) (cdpcontrol.PineState, error)
	PineGoToLine(ctx context.Context, line int) (cdpcontrol.PineState, error)
	PineDeleteLine(ctx context.Context, count int) (cdpcontrol.PineState, error)
	PineMoveLine(ctx context.Context, direction string, count int) (cdpcontrol.PineState, error)
	PineToggleComment(ctx context.Context) (cdpcontrol.PineState, error)
	PineToggleConsole(ctx context.Context) (cdpcontrol.PineState, error)
	PineInsertLineAbove(ctx context.Context) (cdpcontrol.PineState, error)
	PineNewTab(ctx context.Context) (cdpcontrol.PineState, error)
	PineCommandPalette(ctx context.Context) (cdpcontrol.PineState, error)
	ListLayouts(ctx context.Context) ([]cdpcontrol.LayoutInfo, error)
	GetLayoutFavorite(ctx context.Context) (cdpcontrol.LayoutFavoriteResult, error)
	ToggleLayoutFavorite(ctx context.Context) (cdpcontrol.LayoutFavoriteResult, error)
	GetLayoutStatus(ctx context.Context) (cdpcontrol.LayoutStatus, error)
	SwitchLayout(ctx context.Context, id int) (cdpcontrol.LayoutActionResult, error)
	SaveLayout(ctx context.Context) (cdpcontrol.LayoutActionResult, error)
	CloneLayout(ctx context.Context, name string) (cdpcontrol.LayoutActionResult, error)
	DeleteLayout(ctx context.Context, id int) (cdpcontrol.LayoutActionResult, error)
	RenameLayout(ctx context.Context, name string) (cdpcontrol.LayoutActionResult, error)
	SetLayoutGrid(ctx context.Context, template string) (cdpcontrol.LayoutStatus, error)
	NextChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error)
	PrevChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error)
	MaximizeChart(ctx context.Context) (cdpcontrol.LayoutStatus, error)
	ActivateChart(ctx context.Context, index int) (cdpcontrol.LayoutStatus, error)
	ToggleFullscreen(ctx context.Context) (cdpcontrol.LayoutStatus, error)
	DismissDialog(ctx context.Context) (cdpcontrol.LayoutActionResult, error)
	BatchDeleteLayouts(ctx context.Context, ids []int, skipActive bool) (cdpcontrol.BatchDeleteResult, error)
	PreviewLayout(ctx context.Context, id int, takeSnapshot bool) (cdpcontrol.LayoutDetail, error)
	DeepHealthCheck(ctx context.Context) (cdpcontrol.DeepHealthResult, error)
	SearchIndicators(ctx context.Context, chartID, query string) (cdpcontrol.IndicatorSearchResult, error)
	AddIndicatorBySearch(ctx context.Context, chartID, query string, index int) (cdpcontrol.IndicatorAddResult, error)
	ListFavoriteIndicators(ctx context.Context, chartID string) (cdpcontrol.IndicatorSearchResult, error)
	ToggleIndicatorFavorite(ctx context.Context, chartID, query string, index int) (cdpcontrol.IndicatorFavoriteResult, error)
	ProbeIndicatorDialogDOM(ctx context.Context) (map[string]any, error)
	GetCurrency(ctx context.Context, chartID string, pane int) (cdpcontrol.CurrencyInfo, error)
	SetCurrency(ctx context.Context, chartID, currency string, pane int) (cdpcontrol.CurrencyInfo, error)
	GetAvailableCurrencies(ctx context.Context, chartID string, pane int) ([]cdpcontrol.AvailableCurrency, error)
	GetUnit(ctx context.Context, chartID string, pane int) (cdpcontrol.UnitInfo, error)
	SetUnit(ctx context.Context, chartID, unit string, pane int) (cdpcontrol.UnitInfo, error)
	GetAvailableUnits(ctx context.Context, chartID string, pane int) ([]cdpcontrol.AvailableUnit, error)
	ListColoredWatchlists(ctx context.Context) ([]cdpcontrol.ColoredWatchlist, error)
	ReplaceColoredWatchlist(ctx context.Context, color string, symbols []string) (cdpcontrol.ColoredWatchlist, error)
	AppendColoredWatchlist(ctx context.Context, color string, symbols []string) (cdpcontrol.ColoredWatchlist, error)
	RemoveColoredWatchlist(ctx context.Context, color string, symbols []string) (cdpcontrol.ColoredWatchlist, error)
	BulkRemoveColoredWatchlist(ctx context.Context, symbols []string) error
	ListStudyTemplates(ctx context.Context) (cdpcontrol.StudyTemplateList, error)
	GetStudyTemplate(ctx context.Context, id int) (cdpcontrol.StudyTemplateEntry, error)
	ApplyStudyTemplate(ctx context.Context, chartID, name string) (cdpcontrol.StudyTemplateApplyResult, error)
	ProbeHotlistsManager(ctx context.Context) (cdpcontrol.HotlistsManagerProbe, error)
	ProbeHotlistsManagerDeep(ctx context.Context) (map[string]any, error)
	GetHotlistMarkets(ctx context.Context) (any, error)
	GetHotlistExchanges(ctx context.Context) ([]cdpcontrol.HotlistExchangeDetail, error)
	GetOneHotlist(ctx context.Context, exchange, group string) (cdpcontrol.HotlistResult, error)
	ProbeDataWindow(ctx context.Context, chartID string, pane int) (cdpcontrol.DataWindowProbe, error)
}

type chartIDInput struct {
	ChartID string `path:"chart_id"`
	Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
}

type activeChartOutput struct {
	Body cdpcontrol.ActiveChartInfo
}

type navStatusOutput struct {
	Body struct {
		ChartID string `json:"chart_id"`
		Status  string `json:"status"`
	}
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
		if _, err := w.Write([]byte(docsHTML)); err != nil {
			slog.Debug("docs response write failed", "error", err)
		}
	})

	registerChartHandlers(api, svc)
	registerWatchlistHandlers(api, svc)
	registerStudyHandlers(api, svc)
	registerDrawingHandlers(api, svc)
	registerReplayHandlers(api, svc)
	registerAlertHandlers(api, svc)
	registerPineHandlers(api, svc)
	registerLayoutHandlers(api, svc)
	registerMiscHandlers(api, svc)

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
