package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dgnsrekt/tv_agent/internal/cdpcontrol"
	"github.com/dgnsrekt/tv_agent/internal/snapshot"
)

type stubService struct{}

func (s *stubService) ListCharts(ctx context.Context) ([]cdpcontrol.ChartInfo, error) {
	return []cdpcontrol.ChartInfo{}, nil
}
func (s *stubService) GetActiveChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error) {
	return cdpcontrol.ActiveChartInfo{}, nil
}
func (s *stubService) GetSymbolInfo(ctx context.Context, chartID string, pane int) (cdpcontrol.SymbolInfo, error) {
	return cdpcontrol.SymbolInfo{}, nil
}
func (s *stubService) GetSymbol(ctx context.Context, chartID string, pane int) (string, error) {
	return "", nil
}
func (s *stubService) SetSymbol(ctx context.Context, chartID, symbol string, pane int) (string, error) {
	return symbol, nil
}
func (s *stubService) GetResolution(ctx context.Context, chartID string, pane int) (string, error) {
	return "", nil
}
func (s *stubService) SetResolution(ctx context.Context, chartID, resolution string, pane int) (string, error) {
	return resolution, nil
}
func (s *stubService) ExecuteAction(ctx context.Context, chartID, actionID string) error { return nil }
func (s *stubService) ListStudies(ctx context.Context, chartID string, pane int) ([]cdpcontrol.Study, error) {
	return []cdpcontrol.Study{}, nil
}
func (s *stubService) AddStudy(ctx context.Context, chartID, name string, inputs map[string]any, forceOverlay bool, pane int) (cdpcontrol.Study, error) {
	return cdpcontrol.Study{}, nil
}
func (s *stubService) RemoveStudy(ctx context.Context, chartID, studyID string, pane int) error {
	return nil
}
func (s *stubService) GetStudyInputs(ctx context.Context, chartID, studyID string, pane int) (cdpcontrol.StudyDetail, error) {
	return cdpcontrol.StudyDetail{}, nil
}
func (s *stubService) ModifyStudyInputs(ctx context.Context, chartID, studyID string, inputs map[string]any, pane int) (cdpcontrol.StudyDetail, error) {
	return cdpcontrol.StudyDetail{}, nil
}
func (s *stubService) ListWatchlists(ctx context.Context) ([]cdpcontrol.WatchlistInfo, error) {
	return []cdpcontrol.WatchlistInfo{}, nil
}
func (s *stubService) GetActiveWatchlist(ctx context.Context) (cdpcontrol.WatchlistDetail, error) {
	return cdpcontrol.WatchlistDetail{}, nil
}
func (s *stubService) SetActiveWatchlist(ctx context.Context, id string) (cdpcontrol.WatchlistInfo, error) {
	return cdpcontrol.WatchlistInfo{}, nil
}
func (s *stubService) GetWatchlist(ctx context.Context, id string) (cdpcontrol.WatchlistDetail, error) {
	return cdpcontrol.WatchlistDetail{}, nil
}
func (s *stubService) CreateWatchlist(ctx context.Context, name string) (cdpcontrol.WatchlistInfo, error) {
	return cdpcontrol.WatchlistInfo{}, nil
}
func (s *stubService) RenameWatchlist(ctx context.Context, id, name string) (cdpcontrol.WatchlistInfo, error) {
	return cdpcontrol.WatchlistInfo{}, nil
}
func (s *stubService) DeleteWatchlist(ctx context.Context, id string) error { return nil }
func (s *stubService) AddWatchlistSymbols(ctx context.Context, id string, symbols []string) (cdpcontrol.WatchlistDetail, error) {
	return cdpcontrol.WatchlistDetail{}, nil
}
func (s *stubService) RemoveWatchlistSymbols(ctx context.Context, id string, symbols []string) (cdpcontrol.WatchlistDetail, error) {
	return cdpcontrol.WatchlistDetail{}, nil
}
func (s *stubService) FlagSymbol(ctx context.Context, id, symbol string) error { return nil }
func (s *stubService) Zoom(ctx context.Context, chartID, direction string) error { return nil }
func (s *stubService) Scroll(ctx context.Context, chartID string, bars int) error { return nil }
func (s *stubService) ScrollToRealtime(ctx context.Context, chartID string) error { return nil }
func (s *stubService) GoToDate(ctx context.Context, chartID string, timestamp int64) error {
	return nil
}
func (s *stubService) GetVisibleRange(ctx context.Context, chartID string) (cdpcontrol.VisibleRange, error) {
	return cdpcontrol.VisibleRange{}, nil
}
func (s *stubService) SetVisibleRange(ctx context.Context, chartID string, from, to float64) (cdpcontrol.VisibleRange, error) {
	return cdpcontrol.VisibleRange{}, nil
}
func (s *stubService) SetTimeFrame(ctx context.Context, chartID, preset, resolution string, pane int) (cdpcontrol.TimeFrameResult, error) {
	return cdpcontrol.TimeFrameResult{}, nil
}
func (s *stubService) ResetScales(ctx context.Context, chartID string) error { return nil }
func (s *stubService) ProbeChartApi(ctx context.Context, chartID string) (cdpcontrol.ChartApiProbe, error) {
	return cdpcontrol.ChartApiProbe{}, nil
}
func (s *stubService) ProbeChartApiDeep(ctx context.Context, chartID string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) ResolveSymbol(ctx context.Context, chartID, symbol string) (cdpcontrol.ResolvedSymbolInfo, error) {
	return cdpcontrol.ResolvedSymbolInfo{}, nil
}
func (s *stubService) SwitchTimezone(ctx context.Context, chartID, tz string) error { return nil }
func (s *stubService) ScanReplayActivation(ctx context.Context, chartID string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) ActivateReplay(ctx context.Context, chartID string, date float64) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) ActivateReplayAuto(ctx context.Context, chartID string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) DeactivateReplay(ctx context.Context, chartID string) error { return nil }
func (s *stubService) ProbeReplayManager(ctx context.Context, chartID string) (cdpcontrol.ReplayManagerProbe, error) {
	return cdpcontrol.ReplayManagerProbe{}, nil
}
func (s *stubService) ProbeReplayManagerDeep(ctx context.Context, chartID string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) GetReplayStatus(ctx context.Context, chartID string) (cdpcontrol.ReplayStatus, error) {
	return cdpcontrol.ReplayStatus{}, nil
}
func (s *stubService) StartReplay(ctx context.Context, chartID string, point float64) error {
	return nil
}
func (s *stubService) StopReplay(ctx context.Context, chartID string) error  { return nil }
func (s *stubService) ReplayStep(ctx context.Context, chartID string, count int) error { return nil }
func (s *stubService) StartAutoplay(ctx context.Context, chartID string) error { return nil }
func (s *stubService) StopAutoplay(ctx context.Context, chartID string) error  { return nil }
func (s *stubService) ResetReplay(ctx context.Context, chartID string) error   { return nil }
func (s *stubService) ChangeAutoplayDelay(ctx context.Context, chartID string, delay float64) (float64, error) {
	return delay, nil
}
func (s *stubService) ProbeBacktestingApi(ctx context.Context, chartID string) (cdpcontrol.StrategyApiProbe, error) {
	return cdpcontrol.StrategyApiProbe{}, nil
}
func (s *stubService) ListStrategies(ctx context.Context, chartID string) (any, error) {
	return []any{}, nil
}
func (s *stubService) GetActiveStrategy(ctx context.Context, chartID string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) SetActiveStrategy(ctx context.Context, chartID, strategyID string) error {
	return nil
}
func (s *stubService) SetStrategyInput(ctx context.Context, chartID, name string, value any) error {
	return nil
}
func (s *stubService) GetStrategyReport(ctx context.Context, chartID string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) GetStrategyDateRange(ctx context.Context, chartID string) (any, error) {
	return nil, nil
}
func (s *stubService) StrategyGotoDate(ctx context.Context, chartID string, timestamp float64, belowBar bool) error {
	return nil
}
func (s *stubService) ScanAlertsAccess(ctx context.Context, chartID string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) ProbeAlertsRestApi(ctx context.Context, chartID string) (cdpcontrol.AlertsApiProbe, error) {
	return cdpcontrol.AlertsApiProbe{}, nil
}
func (s *stubService) ProbeAlertsRestApiDeep(ctx context.Context, chartID string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) ListAlerts(ctx context.Context) (any, error)                { return nil, nil }
func (s *stubService) GetAlerts(ctx context.Context, ids []string) (any, error)   { return nil, nil }
func (s *stubService) CreateAlert(ctx context.Context, params map[string]any) (any, error) {
	return nil, nil
}
func (s *stubService) ModifyAlert(ctx context.Context, params map[string]any) (any, error) {
	return nil, nil
}
func (s *stubService) DeleteAlerts(ctx context.Context, ids []string) error  { return nil }
func (s *stubService) StopAlerts(ctx context.Context, ids []string) error    { return nil }
func (s *stubService) RestartAlerts(ctx context.Context, ids []string) error { return nil }
func (s *stubService) CloneAlerts(ctx context.Context, ids []string) error   { return nil }
func (s *stubService) ListFires(ctx context.Context) (any, error)            { return nil, nil }
func (s *stubService) DeleteFires(ctx context.Context, ids []string) error   { return nil }
func (s *stubService) DeleteAllFires(ctx context.Context) error { return nil }
func (s *stubService) ListDrawings(ctx context.Context, chartID string, pane int) ([]cdpcontrol.Shape, error) {
	return []cdpcontrol.Shape{}, nil
}
func (s *stubService) GetDrawing(ctx context.Context, chartID, shapeID string, pane int) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) CreateDrawing(ctx context.Context, chartID string, point cdpcontrol.ShapePoint, options map[string]any, pane int) (string, error) {
	return "", nil
}
func (s *stubService) CreateMultipointDrawing(ctx context.Context, chartID string, points []cdpcontrol.ShapePoint, options map[string]any, pane int) (string, error) {
	return "", nil
}
func (s *stubService) CloneDrawing(ctx context.Context, chartID, shapeID string, pane int) (string, error) {
	return "", nil
}
func (s *stubService) RemoveDrawing(ctx context.Context, chartID, shapeID string, disableUndo bool, pane int) error {
	return nil
}
func (s *stubService) RemoveAllDrawings(ctx context.Context, chartID string, pane int) error {
	return nil
}
func (s *stubService) GetDrawingToggles(ctx context.Context, chartID string) (cdpcontrol.DrawingToggles, error) {
	return cdpcontrol.DrawingToggles{}, nil
}
func (s *stubService) SetHideDrawings(ctx context.Context, chartID string, val bool) error {
	return nil
}
func (s *stubService) SetLockDrawings(ctx context.Context, chartID string, val bool) error {
	return nil
}
func (s *stubService) SetMagnet(ctx context.Context, chartID string, enabled bool, mode int) error {
	return nil
}
func (s *stubService) SetDrawingVisibility(ctx context.Context, chartID, shapeID string, visible bool) error {
	return nil
}
func (s *stubService) GetDrawingTool(ctx context.Context, chartID string) (string, error) {
	return "", nil
}
func (s *stubService) SetDrawingTool(ctx context.Context, chartID, tool string) error { return nil }
func (s *stubService) SetDrawingZOrder(ctx context.Context, chartID, shapeID, action string) error {
	return nil
}
func (s *stubService) ExportDrawingsState(ctx context.Context, chartID string) (any, error) {
	return nil, nil
}
func (s *stubService) ImportDrawingsState(ctx context.Context, chartID string, state any) error {
	return nil
}
func (s *stubService) BrowserScreenshot(ctx context.Context, format string, quality int, fullPage bool, notes string) (snapshot.SnapshotMeta, error) {
	return snapshot.SnapshotMeta{}, nil
}
func (s *stubService) TakeSnapshot(ctx context.Context, chartID, format, quality, notes string, pane int) (snapshot.SnapshotMeta, error) {
	return snapshot.SnapshotMeta{}, nil
}
func (s *stubService) GetPaneInfo(ctx context.Context) (cdpcontrol.PanesResult, error) {
	return cdpcontrol.PanesResult{}, nil
}
func (s *stubService) ListSnapshots(ctx context.Context) ([]snapshot.SnapshotMeta, error) {
	return []snapshot.SnapshotMeta{}, nil
}
func (s *stubService) GetSnapshot(ctx context.Context, id string) (snapshot.SnapshotMeta, error) {
	return snapshot.SnapshotMeta{}, nil
}
func (s *stubService) ReadSnapshotImage(ctx context.Context, id string) ([]byte, string, error) {
	return nil, "", nil
}
func (s *stubService) DeleteSnapshot(ctx context.Context, id string) error { return nil }
func (s *stubService) ReloadPage(ctx context.Context, mode string) (cdpcontrol.ReloadResult, error) {
	return cdpcontrol.ReloadResult{}, nil
}
func (s *stubService) TogglePineEditor(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) GetPineStatus(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) GetPineSource(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) SetPineSource(ctx context.Context, source string) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) SavePineScript(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) AddPineToChart(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) GetPineConsole(ctx context.Context) ([]cdpcontrol.PineConsoleMessage, error) {
	return []cdpcontrol.PineConsoleMessage{}, nil
}

func (s *stubService) PineUndo(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineRedo(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineNewIndicator(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineNewStrategy(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineOpenScript(ctx context.Context, name string) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineFindReplace(ctx context.Context, find, replace string) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineGoToLine(ctx context.Context, line int) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineDeleteLine(ctx context.Context, count int) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineMoveLine(ctx context.Context, direction string, count int) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineToggleComment(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineToggleConsole(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineInsertLineAbove(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineNewTab(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) PineCommandPalette(ctx context.Context) (cdpcontrol.PineState, error) {
	return cdpcontrol.PineState{}, nil
}
func (s *stubService) ListLayouts(ctx context.Context) ([]cdpcontrol.LayoutInfo, error) {
	return []cdpcontrol.LayoutInfo{}, nil
}
func (s *stubService) GetLayoutStatus(ctx context.Context) (cdpcontrol.LayoutStatus, error) {
	return cdpcontrol.LayoutStatus{}, nil
}
func (s *stubService) SwitchLayout(ctx context.Context, id int) (cdpcontrol.LayoutActionResult, error) {
	return cdpcontrol.LayoutActionResult{}, nil
}
func (s *stubService) SaveLayout(ctx context.Context) (cdpcontrol.LayoutActionResult, error) {
	return cdpcontrol.LayoutActionResult{}, nil
}
func (s *stubService) CloneLayout(ctx context.Context, name string) (cdpcontrol.LayoutActionResult, error) {
	return cdpcontrol.LayoutActionResult{}, nil
}
func (s *stubService) DeleteLayout(ctx context.Context, id int) (cdpcontrol.LayoutActionResult, error) {
	return cdpcontrol.LayoutActionResult{}, nil
}
func (s *stubService) RenameLayout(ctx context.Context, name string) (cdpcontrol.LayoutActionResult, error) {
	return cdpcontrol.LayoutActionResult{}, nil
}
func (s *stubService) SetLayoutGrid(ctx context.Context, template string) (cdpcontrol.LayoutStatus, error) {
	return cdpcontrol.LayoutStatus{}, nil
}
func (s *stubService) NextChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error) {
	return cdpcontrol.ActiveChartInfo{}, nil
}
func (s *stubService) PrevChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error) {
	return cdpcontrol.ActiveChartInfo{}, nil
}
func (s *stubService) MaximizeChart(ctx context.Context) (cdpcontrol.LayoutStatus, error) {
	return cdpcontrol.LayoutStatus{}, nil
}
func (s *stubService) ActivateChart(ctx context.Context, index int) (cdpcontrol.LayoutStatus, error) {
	return cdpcontrol.LayoutStatus{}, nil
}
func (s *stubService) ToggleFullscreen(ctx context.Context) (cdpcontrol.LayoutStatus, error) {
	return cdpcontrol.LayoutStatus{}, nil
}
func (s *stubService) DismissDialog(ctx context.Context) (cdpcontrol.LayoutActionResult, error) {
	return cdpcontrol.LayoutActionResult{}, nil
}
func (s *stubService) BatchDeleteLayouts(ctx context.Context, ids []int, skipActive bool) (cdpcontrol.BatchDeleteResult, error) {
	return cdpcontrol.BatchDeleteResult{}, nil
}
func (s *stubService) PreviewLayout(ctx context.Context, id int, takeSnapshot bool) (cdpcontrol.LayoutDetail, error) {
	return cdpcontrol.LayoutDetail{}, nil
}
func (s *stubService) DeepHealthCheck(ctx context.Context) (cdpcontrol.DeepHealthResult, error) {
	return cdpcontrol.DeepHealthResult{}, nil
}

func TestDocsDarkMode(t *testing.T) {
	h := NewServer(&stubService{})
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `data-theme="dark"`) {
		t.Fatalf("docs missing dark theme marker")
	}
}
