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
func (s *stubService) GetSymbolInfo(ctx context.Context, chartID string) (cdpcontrol.SymbolInfo, error) {
	return cdpcontrol.SymbolInfo{}, nil
}
func (s *stubService) GetSymbol(ctx context.Context, chartID string) (string, error) { return "", nil }
func (s *stubService) SetSymbol(ctx context.Context, chartID, symbol string) (string, error) {
	return symbol, nil
}
func (s *stubService) GetResolution(ctx context.Context, chartID string) (string, error) {
	return "", nil
}
func (s *stubService) SetResolution(ctx context.Context, chartID, resolution string) (string, error) {
	return resolution, nil
}
func (s *stubService) ExecuteAction(ctx context.Context, chartID, actionID string) error { return nil }
func (s *stubService) ListStudies(ctx context.Context, chartID string) ([]cdpcontrol.Study, error) {
	return []cdpcontrol.Study{}, nil
}
func (s *stubService) AddStudy(ctx context.Context, chartID, name string, inputs map[string]any, forceOverlay bool) (cdpcontrol.Study, error) {
	return cdpcontrol.Study{}, nil
}
func (s *stubService) RemoveStudy(ctx context.Context, chartID, studyID string) error { return nil }
func (s *stubService) GetStudyInputs(ctx context.Context, chartID, studyID string) (cdpcontrol.StudyDetail, error) {
	return cdpcontrol.StudyDetail{}, nil
}
func (s *stubService) ModifyStudyInputs(ctx context.Context, chartID, studyID string, inputs map[string]any) (cdpcontrol.StudyDetail, error) {
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
func (s *stubService) ReplayStep(ctx context.Context, chartID string) error  { return nil }
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
func (s *stubService) ListDrawings(ctx context.Context, chartID string) ([]cdpcontrol.Shape, error) {
	return []cdpcontrol.Shape{}, nil
}
func (s *stubService) GetDrawing(ctx context.Context, chartID, shapeID string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *stubService) CreateDrawing(ctx context.Context, chartID string, point cdpcontrol.ShapePoint, options map[string]any) (string, error) {
	return "", nil
}
func (s *stubService) CreateMultipointDrawing(ctx context.Context, chartID string, points []cdpcontrol.ShapePoint, options map[string]any) (string, error) {
	return "", nil
}
func (s *stubService) CloneDrawing(ctx context.Context, chartID, shapeID string) (string, error) {
	return "", nil
}
func (s *stubService) RemoveDrawing(ctx context.Context, chartID, shapeID string, disableUndo bool) error {
	return nil
}
func (s *stubService) RemoveAllDrawings(ctx context.Context, chartID string) error { return nil }
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
func (s *stubService) TakeSnapshot(ctx context.Context, chartID, format, quality string) (snapshot.SnapshotMeta, error) {
	return snapshot.SnapshotMeta{}, nil
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
