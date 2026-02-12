package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/dgnsrekt/tv_agent/internal/cdpcontrol"
	"github.com/dgnsrekt/tv_agent/internal/snapshot"
	"github.com/google/uuid"
)

// Service wraps active TradingView control operations.
type Service struct {
	cdp   *cdpcontrol.Client
	snaps *snapshot.Store
}

func NewService(cdp *cdpcontrol.Client, snaps *snapshot.Store) *Service {
	return &Service{cdp: cdp, snaps: snaps}
}

func (s *Service) ListCharts(ctx context.Context) ([]cdpcontrol.ChartInfo, error) {
	return s.cdp.ListCharts(ctx)
}

func (s *Service) GetSymbolInfo(ctx context.Context, chartID string) (cdpcontrol.SymbolInfo, error) {
	return s.cdp.GetSymbolInfo(ctx, strings.TrimSpace(chartID))
}

func (s *Service) GetActiveChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error) {
	return s.cdp.GetActiveChart(ctx)
}

func (s *Service) GetSymbol(ctx context.Context, chartID string) (string, error) {
	return s.cdp.GetSymbol(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetSymbol(ctx context.Context, chartID, symbol string) (string, error) {
	if strings.TrimSpace(symbol) == "" {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbol is required"}
	}
	return s.cdp.SetSymbol(ctx, strings.TrimSpace(chartID), strings.TrimSpace(symbol))
}

func (s *Service) GetResolution(ctx context.Context, chartID string) (string, error) {
	return s.cdp.GetResolution(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetResolution(ctx context.Context, chartID, resolution string) (string, error) {
	if strings.TrimSpace(resolution) == "" {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "resolution is required"}
	}
	return s.cdp.SetResolution(ctx, strings.TrimSpace(chartID), strings.TrimSpace(resolution))
}

func (s *Service) ExecuteAction(ctx context.Context, chartID, actionID string) error {
	if strings.TrimSpace(actionID) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "action_id is required"}
	}
	return s.cdp.ExecuteAction(ctx, strings.TrimSpace(chartID), strings.TrimSpace(actionID))
}

func (s *Service) ListStudies(ctx context.Context, chartID string) ([]cdpcontrol.Study, error) {
	return s.cdp.ListStudies(ctx, strings.TrimSpace(chartID))
}

func (s *Service) AddStudy(ctx context.Context, chartID, name string, inputs map[string]any, forceOverlay bool) (cdpcontrol.Study, error) {
	if strings.TrimSpace(name) == "" {
		return cdpcontrol.Study{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "study name is required"}
	}
	return s.cdp.AddStudy(ctx, strings.TrimSpace(chartID), strings.TrimSpace(name), inputs, forceOverlay)
}

func (s *Service) GetStudyInputs(ctx context.Context, chartID, studyID string) (cdpcontrol.StudyDetail, error) {
	if strings.TrimSpace(studyID) == "" {
		return cdpcontrol.StudyDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "study_id is required"}
	}
	return s.cdp.GetStudyInputs(ctx, strings.TrimSpace(chartID), strings.TrimSpace(studyID))
}

func (s *Service) ModifyStudyInputs(ctx context.Context, chartID, studyID string, inputs map[string]any) (cdpcontrol.StudyDetail, error) {
	if strings.TrimSpace(studyID) == "" {
		return cdpcontrol.StudyDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "study_id is required"}
	}
	if len(inputs) == 0 {
		return cdpcontrol.StudyDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "inputs must not be empty"}
	}
	return s.cdp.ModifyStudyInputs(ctx, strings.TrimSpace(chartID), strings.TrimSpace(studyID), inputs)
}

func (s *Service) RemoveStudy(ctx context.Context, chartID, studyID string) error {
	if strings.TrimSpace(studyID) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "study_id is required"}
	}
	return s.cdp.RemoveStudy(ctx, strings.TrimSpace(chartID), strings.TrimSpace(studyID))
}

func (s *Service) Zoom(ctx context.Context, chartID, direction string) error {
	direction = strings.TrimSpace(strings.ToLower(direction))
	if direction != "in" && direction != "out" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "direction must be \"in\" or \"out\""}
	}
	return s.cdp.Zoom(ctx, strings.TrimSpace(chartID), direction)
}

func (s *Service) Scroll(ctx context.Context, chartID string, bars int) error {
	if bars == 0 {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "bars must be non-zero"}
	}
	return s.cdp.Scroll(ctx, strings.TrimSpace(chartID), bars)
}

func (s *Service) ScrollToRealtime(ctx context.Context, chartID string) error {
	return s.cdp.ScrollToRealtime(ctx, strings.TrimSpace(chartID))
}

func (s *Service) GoToDate(ctx context.Context, chartID string, timestamp int64) error {
	if timestamp <= 0 {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "timestamp must be positive"}
	}
	return s.cdp.GoToDate(ctx, strings.TrimSpace(chartID), timestamp)
}

func (s *Service) GetVisibleRange(ctx context.Context, chartID string) (cdpcontrol.VisibleRange, error) {
	return s.cdp.GetVisibleRange(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetVisibleRange(ctx context.Context, chartID string, from, to float64) (cdpcontrol.VisibleRange, error) {
	if from >= to {
		return cdpcontrol.VisibleRange{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "from must be less than to"}
	}
	return s.cdp.SetVisibleRange(ctx, strings.TrimSpace(chartID), from, to)
}

func (s *Service) ResetScales(ctx context.Context, chartID string) error {
	return s.cdp.ResetScales(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ListWatchlists(ctx context.Context) ([]cdpcontrol.WatchlistInfo, error) {
	return s.cdp.ListWatchlists(ctx)
}

func (s *Service) GetActiveWatchlist(ctx context.Context) (cdpcontrol.WatchlistDetail, error) {
	return s.cdp.GetActiveWatchlist(ctx)
}

func (s *Service) SetActiveWatchlist(ctx context.Context, id string) (cdpcontrol.WatchlistInfo, error) {
	if strings.TrimSpace(id) == "" {
		return cdpcontrol.WatchlistInfo{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "watchlist id is required"}
	}
	return s.cdp.SetActiveWatchlist(ctx, strings.TrimSpace(id))
}

func (s *Service) GetWatchlist(ctx context.Context, id string) (cdpcontrol.WatchlistDetail, error) {
	if strings.TrimSpace(id) == "" {
		return cdpcontrol.WatchlistDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "watchlist_id is required"}
	}
	return s.cdp.GetWatchlist(ctx, strings.TrimSpace(id))
}

func (s *Service) CreateWatchlist(ctx context.Context, name string) (cdpcontrol.WatchlistInfo, error) {
	if strings.TrimSpace(name) == "" {
		return cdpcontrol.WatchlistInfo{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "name is required"}
	}
	return s.cdp.CreateWatchlist(ctx, strings.TrimSpace(name))
}

func (s *Service) RenameWatchlist(ctx context.Context, id, name string) (cdpcontrol.WatchlistInfo, error) {
	if strings.TrimSpace(id) == "" {
		return cdpcontrol.WatchlistInfo{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "watchlist_id is required"}
	}
	if strings.TrimSpace(name) == "" {
		return cdpcontrol.WatchlistInfo{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "name is required"}
	}
	return s.cdp.RenameWatchlist(ctx, strings.TrimSpace(id), strings.TrimSpace(name))
}

func (s *Service) DeleteWatchlist(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "watchlist_id is required"}
	}
	return s.cdp.DeleteWatchlist(ctx, strings.TrimSpace(id))
}

func (s *Service) AddWatchlistSymbols(ctx context.Context, id string, symbols []string) (cdpcontrol.WatchlistDetail, error) {
	if strings.TrimSpace(id) == "" {
		return cdpcontrol.WatchlistDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "watchlist_id is required"}
	}
	if len(symbols) == 0 {
		return cdpcontrol.WatchlistDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbols must not be empty"}
	}
	return s.cdp.AddWatchlistSymbols(ctx, strings.TrimSpace(id), symbols)
}

func (s *Service) RemoveWatchlistSymbols(ctx context.Context, id string, symbols []string) (cdpcontrol.WatchlistDetail, error) {
	if strings.TrimSpace(id) == "" {
		return cdpcontrol.WatchlistDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "watchlist_id is required"}
	}
	if len(symbols) == 0 {
		return cdpcontrol.WatchlistDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbols must not be empty"}
	}
	return s.cdp.RemoveWatchlistSymbols(ctx, strings.TrimSpace(id), symbols)
}

func (s *Service) FlagSymbol(ctx context.Context, id, symbol string) error {
	if strings.TrimSpace(id) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "watchlist_id is required"}
	}
	if strings.TrimSpace(symbol) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbol is required"}
	}
	return s.cdp.FlagSymbol(ctx, strings.TrimSpace(id), strings.TrimSpace(symbol))
}

// --- ChartAPI methods ---

func (s *Service) ProbeChartApiDeep(ctx context.Context, chartID string) (map[string]any, error) {
	return s.cdp.ProbeChartApiDeep(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ProbeChartApi(ctx context.Context, chartID string) (cdpcontrol.ChartApiProbe, error) {
	return s.cdp.ProbeChartApi(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ResolveSymbol(ctx context.Context, chartID, symbol string) (cdpcontrol.ResolvedSymbolInfo, error) {
	if strings.TrimSpace(symbol) == "" {
		return cdpcontrol.ResolvedSymbolInfo{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbol is required"}
	}
	return s.cdp.ResolveSymbol(ctx, strings.TrimSpace(chartID), strings.TrimSpace(symbol))
}

func (s *Service) SwitchTimezone(ctx context.Context, chartID, tz string) error {
	if strings.TrimSpace(tz) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "timezone is required"}
	}
	return s.cdp.SwitchTimezone(ctx, strings.TrimSpace(chartID), strings.TrimSpace(tz))
}

// --- Replay Manager methods ---

func (s *Service) ProbeReplayManager(ctx context.Context, chartID string) (cdpcontrol.ReplayManagerProbe, error) {
	return s.cdp.ProbeReplayManager(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ScanReplayActivation(ctx context.Context, chartID string) (map[string]any, error) {
	return s.cdp.ScanReplayActivation(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ProbeReplayManagerDeep(ctx context.Context, chartID string) (map[string]any, error) {
	return s.cdp.ProbeReplayManagerDeep(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ActivateReplay(ctx context.Context, chartID string, date float64) (map[string]any, error) {
	return s.cdp.ActivateReplay(ctx, strings.TrimSpace(chartID), date)
}

func (s *Service) ActivateReplayAuto(ctx context.Context, chartID string) (map[string]any, error) {
	return s.cdp.ActivateReplayAuto(ctx, strings.TrimSpace(chartID))
}

func (s *Service) DeactivateReplay(ctx context.Context, chartID string) error {
	return s.cdp.DeactivateReplay(ctx, strings.TrimSpace(chartID))
}

func (s *Service) GetReplayStatus(ctx context.Context, chartID string) (cdpcontrol.ReplayStatus, error) {
	return s.cdp.GetReplayStatus(ctx, strings.TrimSpace(chartID))
}

func (s *Service) StartReplay(ctx context.Context, chartID string, point float64) error {
	return s.cdp.StartReplay(ctx, strings.TrimSpace(chartID), point)
}

func (s *Service) StopReplay(ctx context.Context, chartID string) error {
	return s.cdp.StopReplay(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ReplayStep(ctx context.Context, chartID string) error {
	return s.cdp.ReplayStep(ctx, strings.TrimSpace(chartID))
}

func (s *Service) StartAutoplay(ctx context.Context, chartID string) error {
	return s.cdp.StartAutoplay(ctx, strings.TrimSpace(chartID))
}

func (s *Service) StopAutoplay(ctx context.Context, chartID string) error {
	return s.cdp.StopAutoplay(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ResetReplay(ctx context.Context, chartID string) error {
	return s.cdp.ResetReplay(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ChangeAutoplayDelay(ctx context.Context, chartID string, delay float64) (float64, error) {
	if delay <= 0 {
		return 0, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "delay must be positive"}
	}
	return s.cdp.ChangeAutoplayDelay(ctx, strings.TrimSpace(chartID), delay)
}

// --- Backtesting Strategy API methods ---

func (s *Service) ProbeBacktestingApi(ctx context.Context, chartID string) (cdpcontrol.StrategyApiProbe, error) {
	return s.cdp.ProbeBacktestingApi(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ListStrategies(ctx context.Context, chartID string) (any, error) {
	return s.cdp.ListStrategies(ctx, strings.TrimSpace(chartID))
}

func (s *Service) GetActiveStrategy(ctx context.Context, chartID string) (map[string]any, error) {
	return s.cdp.GetActiveStrategy(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetActiveStrategy(ctx context.Context, chartID, strategyID string) error {
	if strings.TrimSpace(strategyID) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "strategy_id is required"}
	}
	return s.cdp.SetActiveStrategy(ctx, strings.TrimSpace(chartID), strings.TrimSpace(strategyID))
}

func (s *Service) SetStrategyInput(ctx context.Context, chartID, name string, value any) error {
	if strings.TrimSpace(name) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "name is required"}
	}
	return s.cdp.SetStrategyInput(ctx, strings.TrimSpace(chartID), strings.TrimSpace(name), value)
}

func (s *Service) GetStrategyReport(ctx context.Context, chartID string) (map[string]any, error) {
	return s.cdp.GetStrategyReport(ctx, strings.TrimSpace(chartID))
}

func (s *Service) GetStrategyDateRange(ctx context.Context, chartID string) (any, error) {
	return s.cdp.GetStrategyDateRange(ctx, strings.TrimSpace(chartID))
}

func (s *Service) StrategyGotoDate(ctx context.Context, chartID string, timestamp float64, belowBar bool) error {
	return s.cdp.StrategyGotoDate(ctx, strings.TrimSpace(chartID), timestamp, belowBar)
}

// --- Alerts REST API methods ---

func (s *Service) ScanAlertsAccess(ctx context.Context, chartID string) (map[string]any, error) {
	return s.cdp.ScanAlertsAccess(ctx, strings.TrimSpace(chartID))
}



func (s *Service) ProbeAlertsRestApi(ctx context.Context, chartID string) (cdpcontrol.AlertsApiProbe, error) {
	return s.cdp.ProbeAlertsRestApi(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ProbeAlertsRestApiDeep(ctx context.Context, chartID string) (map[string]any, error) {
	return s.cdp.ProbeAlertsRestApiDeep(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ListAlerts(ctx context.Context) (any, error) {
	return s.cdp.ListAlerts(ctx)
}

func (s *Service) GetAlerts(ctx context.Context, ids []string) (any, error) {
	if len(ids) == 0 {
		return nil, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "alert_ids must not be empty"}
	}
	return s.cdp.GetAlerts(ctx, ids)
}

func (s *Service) CreateAlert(ctx context.Context, params map[string]any) (any, error) {
	if len(params) == 0 {
		return nil, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "params must not be empty"}
	}
	return s.cdp.CreateAlert(ctx, params)
}

func (s *Service) ModifyAlert(ctx context.Context, params map[string]any) (any, error) {
	if len(params) == 0 {
		return nil, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "params must not be empty"}
	}
	return s.cdp.ModifyAlert(ctx, params)
}

func (s *Service) DeleteAlerts(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "alert_ids must not be empty"}
	}
	return s.cdp.DeleteAlerts(ctx, ids)
}

func (s *Service) StopAlerts(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "alert_ids must not be empty"}
	}
	return s.cdp.StopAlerts(ctx, ids)
}

func (s *Service) RestartAlerts(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "alert_ids must not be empty"}
	}
	return s.cdp.RestartAlerts(ctx, ids)
}

func (s *Service) CloneAlerts(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "alert_ids must not be empty"}
	}
	return s.cdp.CloneAlerts(ctx, ids)
}

func (s *Service) ListFires(ctx context.Context) (any, error) {
	return s.cdp.ListFires(ctx)
}

func (s *Service) DeleteFires(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "fire_ids must not be empty"}
	}
	return s.cdp.DeleteFires(ctx, ids)
}

func (s *Service) DeleteAllFires(ctx context.Context) error {
	return s.cdp.DeleteAllFires(ctx)
}

// --- Drawing/Shape methods ---

func (s *Service) ListDrawings(ctx context.Context, chartID string) ([]cdpcontrol.Shape, error) {
	return s.cdp.ListDrawings(ctx, strings.TrimSpace(chartID))
}

func (s *Service) GetDrawing(ctx context.Context, chartID, shapeID string) (map[string]any, error) {
	if strings.TrimSpace(shapeID) == "" {
		return nil, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "shape_id is required"}
	}
	return s.cdp.GetDrawing(ctx, strings.TrimSpace(chartID), strings.TrimSpace(shapeID))
}

func (s *Service) CreateDrawing(ctx context.Context, chartID string, point cdpcontrol.ShapePoint, options map[string]any) (string, error) {
	if _, ok := options["shape"]; !ok {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "options must contain \"shape\" key"}
	}
	return s.cdp.CreateDrawing(ctx, strings.TrimSpace(chartID), point, options)
}

func (s *Service) CreateMultipointDrawing(ctx context.Context, chartID string, points []cdpcontrol.ShapePoint, options map[string]any) (string, error) {
	if len(points) < 2 {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "points must have at least 2 entries"}
	}
	if _, ok := options["shape"]; !ok {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "options must contain \"shape\" key"}
	}
	return s.cdp.CreateMultipointDrawing(ctx, strings.TrimSpace(chartID), points, options)
}

func (s *Service) CloneDrawing(ctx context.Context, chartID, shapeID string) (string, error) {
	if strings.TrimSpace(shapeID) == "" {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "shape_id is required"}
	}
	return s.cdp.CloneDrawing(ctx, strings.TrimSpace(chartID), strings.TrimSpace(shapeID))
}

func (s *Service) RemoveDrawing(ctx context.Context, chartID, shapeID string, disableUndo bool) error {
	if strings.TrimSpace(shapeID) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "shape_id is required"}
	}
	return s.cdp.RemoveDrawing(ctx, strings.TrimSpace(chartID), strings.TrimSpace(shapeID), disableUndo)
}

func (s *Service) RemoveAllDrawings(ctx context.Context, chartID string) error {
	return s.cdp.RemoveAllDrawings(ctx, strings.TrimSpace(chartID))
}

func (s *Service) GetDrawingToggles(ctx context.Context, chartID string) (cdpcontrol.DrawingToggles, error) {
	return s.cdp.GetDrawingToggles(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetHideDrawings(ctx context.Context, chartID string, val bool) error {
	return s.cdp.SetHideDrawings(ctx, strings.TrimSpace(chartID), val)
}

func (s *Service) SetLockDrawings(ctx context.Context, chartID string, val bool) error {
	return s.cdp.SetLockDrawings(ctx, strings.TrimSpace(chartID), val)
}

func (s *Service) SetMagnet(ctx context.Context, chartID string, enabled bool, mode int) error {
	if mode < -1 || mode > 1 {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "mode must be 0, 1, or -1 (skip)"}
	}
	return s.cdp.SetMagnet(ctx, strings.TrimSpace(chartID), enabled, mode)
}

func (s *Service) SetDrawingVisibility(ctx context.Context, chartID, shapeID string, visible bool) error {
	if strings.TrimSpace(shapeID) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "shape_id is required"}
	}
	return s.cdp.SetDrawingVisibility(ctx, strings.TrimSpace(chartID), strings.TrimSpace(shapeID), visible)
}

func (s *Service) GetDrawingTool(ctx context.Context, chartID string) (string, error) {
	return s.cdp.GetDrawingTool(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetDrawingTool(ctx context.Context, chartID, tool string) error {
	if strings.TrimSpace(tool) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "tool is required"}
	}
	return s.cdp.SetDrawingTool(ctx, strings.TrimSpace(chartID), strings.TrimSpace(tool))
}

func (s *Service) SetDrawingZOrder(ctx context.Context, chartID, shapeID, action string) error {
	if strings.TrimSpace(shapeID) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "shape_id is required"}
	}
	valid := map[string]bool{"bring_forward": true, "bring_to_front": true, "send_backward": true, "send_to_back": true}
	a := strings.TrimSpace(action)
	if !valid[a] {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "action must be one of: bring_forward, bring_to_front, send_backward, send_to_back"}
	}
	return s.cdp.SetDrawingZOrder(ctx, strings.TrimSpace(chartID), strings.TrimSpace(shapeID), a)
}

func (s *Service) ExportDrawingsState(ctx context.Context, chartID string) (any, error) {
	return s.cdp.ExportDrawingsState(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ImportDrawingsState(ctx context.Context, chartID string, state any) error {
	if state == nil {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "state is required"}
	}
	return s.cdp.ImportDrawingsState(ctx, strings.TrimSpace(chartID), state)
}

// --- Snapshot methods ---

func (s *Service) TakeSnapshot(ctx context.Context, chartID, format, quality string) (snapshot.SnapshotMeta, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "png"
	}
	if format != "png" && format != "jpeg" {
		return snapshot.SnapshotMeta{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "format must be \"png\" or \"jpeg\""}
	}

	result, err := s.cdp.TakeSnapshot(ctx, strings.TrimSpace(chartID), format, quality, false)
	if err != nil {
		return snapshot.SnapshotMeta{}, err
	}

	imageData, err := decodeDataURL(result.DataURL)
	if err != nil {
		return snapshot.SnapshotMeta{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeEvalFailure, Message: fmt.Sprintf("decode data url: %v", err)}
	}

	meta := snapshot.SnapshotMeta{
		ID:        uuid.New().String(),
		ChartID:   strings.TrimSpace(chartID),
		Format:    format,
		Width:     result.Width,
		Height:    result.Height,
		SizeBytes: len(imageData),
		CreatedAt: time.Now().UTC(),
	}

	if len(result.Metadata.Charts) > 0 {
		c := result.Metadata.Charts[0]
		meta.Symbol = c.Meta.Symbol
		meta.Exchange = c.Meta.Exchange
		meta.Resolution = c.Meta.Resolution
		meta.Description = c.Meta.Description
	}
	meta.Theme = result.Metadata.Theme
	meta.Layout = result.Metadata.Layout

	if err := s.snaps.Save(meta, imageData); err != nil {
		return snapshot.SnapshotMeta{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeEvalFailure, Message: fmt.Sprintf("save snapshot: %v", err)}
	}

	return meta, nil
}

func (s *Service) ListSnapshots(ctx context.Context) ([]snapshot.SnapshotMeta, error) {
	return s.snaps.List()
}

func (s *Service) GetSnapshot(ctx context.Context, id string) (snapshot.SnapshotMeta, error) {
	if strings.TrimSpace(id) == "" {
		return snapshot.SnapshotMeta{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "snapshot_id is required"}
	}
	meta, err := s.snaps.Get(strings.TrimSpace(id))
	if err != nil {
		return snapshot.SnapshotMeta{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeSnapshotNotFound, Message: err.Error()}
	}
	return meta, nil
}

func (s *Service) ReadSnapshotImage(ctx context.Context, id string) ([]byte, string, error) {
	if strings.TrimSpace(id) == "" {
		return nil, "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "snapshot_id is required"}
	}
	data, format, err := s.snaps.ReadImage(strings.TrimSpace(id))
	if err != nil {
		return nil, "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeSnapshotNotFound, Message: err.Error()}
	}
	return data, format, nil
}

func (s *Service) DeleteSnapshot(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "snapshot_id is required"}
	}
	if err := s.snaps.Delete(strings.TrimSpace(id)); err != nil {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeSnapshotNotFound, Message: err.Error()}
	}
	return nil
}

// --- Health methods ---

func (s *Service) DeepHealthCheck(ctx context.Context) (cdpcontrol.DeepHealthResult, error) {
	return s.cdp.DeepHealthCheck(ctx)
}

// --- Page methods ---

func (s *Service) ReloadPage(ctx context.Context, mode string) (cdpcontrol.ReloadResult, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "normal"
	}
	if mode != "normal" && mode != "hard" {
		return cdpcontrol.ReloadResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "mode must be \"normal\" or \"hard\""}
	}
	hard := mode == "hard"
	if err := s.cdp.ReloadPage(ctx, "", hard); err != nil {
		return cdpcontrol.ReloadResult{}, err
	}
	return cdpcontrol.ReloadResult{Status: "reloaded", Mode: mode}, nil
}

// --- Pine Editor methods (DOM-based) ---

func (s *Service) TogglePineEditor(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.TogglePineEditor(ctx)
}

func (s *Service) GetPineStatus(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.GetPineStatus(ctx)
}

func (s *Service) GetPineSource(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.GetPineSource(ctx)
}

func (s *Service) SetPineSource(ctx context.Context, source string) (cdpcontrol.PineState, error) {
	if strings.TrimSpace(source) == "" {
		return cdpcontrol.PineState{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "source is required"}
	}
	return s.cdp.SetPineSource(ctx, source)
}

func (s *Service) SavePineScript(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.SavePineScript(ctx)
}

func (s *Service) AddPineToChart(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.AddPineToChart(ctx)
}

func (s *Service) GetPineConsole(ctx context.Context) ([]cdpcontrol.PineConsoleMessage, error) {
	return s.cdp.GetPineConsole(ctx)
}

func (s *Service) PineUndo(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.PineUndo(ctx)
}

func (s *Service) PineRedo(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.PineRedo(ctx)
}

func (s *Service) PineNewIndicator(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.PineNewIndicator(ctx)
}

func (s *Service) PineNewStrategy(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.PineNewStrategy(ctx)
}

func (s *Service) PineOpenScript(ctx context.Context, name string) (cdpcontrol.PineState, error) {
	if strings.TrimSpace(name) == "" {
		return cdpcontrol.PineState{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "name is required"}
	}
	return s.cdp.PineOpenScript(ctx, name)
}

func (s *Service) PineFindReplace(ctx context.Context, find, replace string) (cdpcontrol.PineState, error) {
	if find == "" {
		return cdpcontrol.PineState{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "find is required"}
	}
	return s.cdp.PineFindReplace(ctx, find, replace)
}

func (s *Service) PineGoToLine(ctx context.Context, line int) (cdpcontrol.PineState, error) {
	if line < 1 {
		return cdpcontrol.PineState{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "line must be >= 1"}
	}
	return s.cdp.PineGoToLine(ctx, line)
}

func (s *Service) PineDeleteLine(ctx context.Context, count int) (cdpcontrol.PineState, error) {
	return s.cdp.PineDeleteLine(ctx, count)
}

func (s *Service) PineMoveLine(ctx context.Context, direction string, count int) (cdpcontrol.PineState, error) {
	direction = strings.ToLower(strings.TrimSpace(direction))
	if direction != "up" && direction != "down" {
		return cdpcontrol.PineState{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "direction must be \"up\" or \"down\""}
	}
	return s.cdp.PineMoveLine(ctx, direction, count)
}

func (s *Service) PineToggleComment(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.PineToggleComment(ctx)
}

func (s *Service) PineToggleConsole(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.PineToggleConsole(ctx)
}

func (s *Service) PineInsertLineAbove(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.PineInsertLineAbove(ctx)
}

func (s *Service) PineNewTab(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.PineNewTab(ctx)
}

func (s *Service) PineCommandPalette(ctx context.Context) (cdpcontrol.PineState, error) {
	return s.cdp.PineCommandPalette(ctx)
}

// --- Layout management methods ---

func (s *Service) ListLayouts(ctx context.Context) ([]cdpcontrol.LayoutInfo, error) {
	return s.cdp.ListLayouts(ctx)
}

func (s *Service) GetLayoutStatus(ctx context.Context) (cdpcontrol.LayoutStatus, error) {
	return s.cdp.GetLayoutStatus(ctx)
}

func (s *Service) SwitchLayout(ctx context.Context, id int) (cdpcontrol.LayoutActionResult, error) {
	if id <= 0 {
		return cdpcontrol.LayoutActionResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "id must be > 0"}
	}
	return s.cdp.SwitchLayout(ctx, id)
}

func (s *Service) SaveLayout(ctx context.Context) (cdpcontrol.LayoutActionResult, error) {
	return s.cdp.SaveLayout(ctx)
}

func (s *Service) CloneLayout(ctx context.Context, name string) (cdpcontrol.LayoutActionResult, error) {
	if strings.TrimSpace(name) == "" {
		return cdpcontrol.LayoutActionResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "name is required"}
	}
	return s.cdp.CloneLayout(ctx, strings.TrimSpace(name))
}

func (s *Service) DeleteLayout(ctx context.Context, id int) (cdpcontrol.LayoutActionResult, error) {
	if id <= 0 {
		return cdpcontrol.LayoutActionResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "id must be positive"}
	}
	return s.cdp.DeleteLayout(ctx, id)
}

func (s *Service) RenameLayout(ctx context.Context, name string) (cdpcontrol.LayoutActionResult, error) {
	if strings.TrimSpace(name) == "" {
		return cdpcontrol.LayoutActionResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "name is required"}
	}
	return s.cdp.RenameLayout(ctx, strings.TrimSpace(name))
}

func (s *Service) SetLayoutGrid(ctx context.Context, template string) (cdpcontrol.LayoutStatus, error) {
	if strings.TrimSpace(template) == "" {
		return cdpcontrol.LayoutStatus{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "template is required"}
	}
	return s.cdp.SetLayoutGrid(ctx, strings.TrimSpace(template))
}

func (s *Service) NextChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error) {
	return s.cdp.NextChart(ctx)
}

func (s *Service) PrevChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error) {
	return s.cdp.PrevChart(ctx)
}

func (s *Service) MaximizeChart(ctx context.Context) (cdpcontrol.LayoutStatus, error) {
	return s.cdp.MaximizeChart(ctx)
}

func (s *Service) ActivateChart(ctx context.Context, index int) (cdpcontrol.LayoutStatus, error) {
	if index < 0 {
		return cdpcontrol.LayoutStatus{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "index must be >= 0"}
	}
	return s.cdp.ActivateChart(ctx, index)
}

func (s *Service) ToggleFullscreen(ctx context.Context) (cdpcontrol.LayoutStatus, error) {
	return s.cdp.ToggleFullscreen(ctx)
}

func (s *Service) DismissDialog(ctx context.Context) (cdpcontrol.LayoutActionResult, error) {
	return s.cdp.DismissDialog(ctx)
}

func (s *Service) BatchDeleteLayouts(ctx context.Context, ids []int, skipActive bool) (cdpcontrol.BatchDeleteResult, error) {
	if len(ids) == 0 {
		return cdpcontrol.BatchDeleteResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "ids must not be empty"}
	}
	if len(ids) > 50 {
		return cdpcontrol.BatchDeleteResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "max 50 layouts per batch"}
	}

	var activeID int
	if skipActive {
		status, err := s.cdp.GetLayoutStatus(ctx)
		if err == nil {
			// LayoutStatus.LayoutID is a short URL, not numeric. Look up
			// the numeric ID from the layouts list.
			layouts, lErr := s.cdp.ListLayouts(ctx)
			if lErr == nil {
				for _, l := range layouts {
					if l.URL == status.LayoutID {
						activeID = l.ID
						break
					}
				}
			}
		}
	}

	var result cdpcontrol.BatchDeleteResult
	for i, id := range ids {
		if skipActive && id == activeID {
			result.Skipped = append(result.Skipped, id)
			continue
		}
		_, err := s.cdp.DeleteLayout(ctx, id)
		if err != nil {
			result.Errors = append(result.Errors, cdpcontrol.BatchDeleteError{ID: id, Error: err.Error()})
		} else {
			result.Deleted = append(result.Deleted, id)
		}
		if i < len(ids)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	return result, nil
}

func (s *Service) PreviewLayout(ctx context.Context, id int, takeSnapshot bool) (cdpcontrol.LayoutDetail, error) {
	if id <= 0 {
		return cdpcontrol.LayoutDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "id must be > 0"}
	}

	// Get the layouts list for both current-ID lookup and target validation.
	layouts, err := s.cdp.ListLayouts(ctx)
	if err != nil {
		return cdpcontrol.LayoutDetail{}, err
	}

	// Resolve the current layout's numeric ID by matching the short URL
	// from LayoutStatus against the layout list (LayoutStatus.LayoutID is
	// a short URL like "HXdrcgc8", not a numeric ID).
	var previousID int
	currentStatus, statusErr := s.cdp.GetLayoutStatus(ctx)
	if statusErr == nil {
		for _, l := range layouts {
			if l.URL == currentStatus.LayoutID {
				previousID = l.ID
				break
			}
		}
	}

	// Find the target layout info.
	var layoutInfo cdpcontrol.LayoutInfo
	found := false
	for _, l := range layouts {
		if l.ID == id {
			layoutInfo = l
			found = true
			break
		}
	}
	if !found {
		return cdpcontrol.LayoutDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: fmt.Sprintf("layout %d not found", id)}
	}

	// If already on this layout, no need to switch.
	alreadyOnTarget := previousID == id

	if !alreadyOnTarget {
		if _, err := s.cdp.SwitchLayout(ctx, id); err != nil {
			return cdpcontrol.LayoutDetail{}, fmt.Errorf("switch to layout %d: %w", id, err)
		}
		// Allow TradingView API to fully initialize after page load.
		time.Sleep(2 * time.Second)
	}

	// Gather layout details.
	detail := cdpcontrol.LayoutDetail{
		Info:       layoutInfo,
		PreviousID: previousID,
	}

	if status, err := s.cdp.GetLayoutStatus(ctx); err == nil {
		detail.Status = status
	}

	// Use first chart for study/drawing queries.
	charts, _ := s.cdp.ListCharts(ctx)
	chartID := ""
	if len(charts) > 0 {
		chartID = charts[0].ChartID
	}

	if chartID != "" {
		if studies, err := s.cdp.ListStudies(ctx, chartID); err == nil {
			detail.Studies = studies
		}
		if drawings, err := s.cdp.ListDrawings(ctx, chartID); err == nil {
			detail.DrawingCount = len(drawings)
		}
	}

	if takeSnapshot && chartID != "" {
		snap, err := s.TakeSnapshot(ctx, chartID, "png", "")
		if err == nil {
			detail.SnapshotURL = fmt.Sprintf("/api/v1/snapshots/%s/image", snap.ID)
		}
	}

	// Switch back to the original layout.
	if !alreadyOnTarget && previousID > 0 {
		_, _ = s.cdp.SwitchLayout(ctx, previousID)
	}

	return detail, nil
}

func decodeDataURL(dataURL string) ([]byte, error) {
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data URL format")
	}
	return base64.StdEncoding.DecodeString(parts[1])
}

