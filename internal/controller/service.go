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

func (s *Service) requireNonEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: fieldName + " is required"}
	}
	return nil
}

// ensurePane activates the given pane index before an operation.
// If pane < 0 it is a no-op (use current active pane).
func (s *Service) ensurePane(ctx context.Context, pane int) error {
	if pane < 0 {
		return nil
	}
	status, err := s.cdp.ActivateChart(ctx, pane)
	if err != nil {
		return err
	}
	if status.ActiveIndex != pane {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: fmt.Sprintf("pane %d out of range (chart_count=%d)", pane, status.ChartCount)}
	}
	return nil
}

func (s *Service) ListCharts(ctx context.Context) ([]cdpcontrol.ChartInfo, error) {
	return s.cdp.ListCharts(ctx)
}

func (s *Service) GetSymbolInfo(ctx context.Context, chartID string, pane int) (cdpcontrol.SymbolInfo, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.SymbolInfo{}, err
	}
	return s.cdp.GetSymbolInfo(ctx, strings.TrimSpace(chartID))
}

func (s *Service) GetActiveChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error) {
	return s.cdp.GetActiveChart(ctx)
}

func (s *Service) GetSymbol(ctx context.Context, chartID string, pane int) (string, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return "", err
	}
	return s.cdp.GetSymbol(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetSymbol(ctx context.Context, chartID, symbol string, pane int) (string, error) {
	if err := s.requireNonEmpty(symbol, "symbol"); err != nil {
		return "", err
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return "", err
	}
	return s.cdp.SetSymbol(ctx, strings.TrimSpace(chartID), strings.TrimSpace(symbol))
}

func (s *Service) GetResolution(ctx context.Context, chartID string, pane int) (string, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return "", err
	}
	return s.cdp.GetResolution(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetResolution(ctx context.Context, chartID, resolution string, pane int) (string, error) {
	if err := s.requireNonEmpty(resolution, "resolution"); err != nil {
		return "", err
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return "", err
	}
	return s.cdp.SetResolution(ctx, strings.TrimSpace(chartID), strings.TrimSpace(resolution))
}

func (s *Service) GetChartType(ctx context.Context, chartID string, pane int) (int, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return 0, err
	}
	return s.cdp.GetChartType(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetChartType(ctx context.Context, chartID string, chartType int, pane int) (int, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return 0, err
	}
	return s.cdp.SetChartType(ctx, strings.TrimSpace(chartID), chartType)
}

func (s *Service) ExecuteAction(ctx context.Context, chartID, actionID string) error {
	if err := s.requireNonEmpty(actionID, "action_id"); err != nil {
		return err
	}

	return s.cdp.ExecuteAction(ctx, strings.TrimSpace(chartID), strings.TrimSpace(actionID))
}

func (s *Service) ListStudies(ctx context.Context, chartID string, pane int) ([]cdpcontrol.Study, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return nil, err
	}
	return s.cdp.ListStudies(ctx, strings.TrimSpace(chartID))
}

func (s *Service) AddStudy(ctx context.Context, chartID, name string, inputs map[string]any, forceOverlay bool, pane int) (cdpcontrol.Study, error) {
	if err := s.requireNonEmpty(name, "study name"); err != nil {
		return cdpcontrol.Study{}, err
	}

	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.Study{}, err
	}
	return s.cdp.AddStudy(ctx, strings.TrimSpace(chartID), strings.TrimSpace(name), inputs, forceOverlay)
}

func (s *Service) GetStudyInputs(ctx context.Context, chartID, studyID string, pane int) (cdpcontrol.StudyDetail, error) {
	if err := s.requireNonEmpty(studyID, "study_id"); err != nil {
		return cdpcontrol.StudyDetail{}, err
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.StudyDetail{}, err
	}
	return s.cdp.GetStudyInputs(ctx, strings.TrimSpace(chartID), strings.TrimSpace(studyID))
}

func (s *Service) ModifyStudyInputs(ctx context.Context, chartID, studyID string, inputs map[string]any, pane int) (cdpcontrol.StudyDetail, error) {
	if err := s.requireNonEmpty(studyID, "study_id"); err != nil {
		return cdpcontrol.StudyDetail{}, err
	}
	if len(inputs) == 0 {
		return cdpcontrol.StudyDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "inputs must not be empty"}
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.StudyDetail{}, err
	}
	return s.cdp.ModifyStudyInputs(ctx, strings.TrimSpace(chartID), strings.TrimSpace(studyID), inputs)
}

func (s *Service) RemoveStudy(ctx context.Context, chartID, studyID string, pane int) error {
	if err := s.requireNonEmpty(studyID, "study_id"); err != nil {
		return err
	}

	if err := s.ensurePane(ctx, pane); err != nil {
		return err
	}
	return s.cdp.RemoveStudy(ctx, strings.TrimSpace(chartID), strings.TrimSpace(studyID))
}

// --- Compare/Overlay convenience methods ---

func (s *Service) AddCompare(ctx context.Context, chartID, symbol, mode, source string, pane int) (cdpcontrol.Study, error) {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return cdpcontrol.Study{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbol is required"}
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "overlay"
	}
	if mode != "overlay" && mode != "compare" {
		return cdpcontrol.Study{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "mode must be \"overlay\" or \"compare\""}
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.Study{}, err
	}

	var name string
	var forceOverlay bool
	inputs := map[string]any{"symbol": symbol}
	if mode == "overlay" {
		name = "Overlay"
		forceOverlay = true
	} else {
		name = "Compare"
		source = strings.TrimSpace(source)
		if source == "" {
			source = "close"
		}
		inputs["source"] = source
	}
	return s.cdp.AddStudy(ctx, strings.TrimSpace(chartID), name, inputs, forceOverlay)
}

func (s *Service) ListCompares(ctx context.Context, chartID string, pane int) ([]cdpcontrol.Study, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return nil, err
	}
	studies, err := s.cdp.ListStudies(ctx, strings.TrimSpace(chartID))
	if err != nil {
		return nil, err
	}
	var result []cdpcontrol.Study
	for _, st := range studies {
		if st.Name == "Overlay" || st.Name == "Compare" {
			result = append(result, st)
		}
	}
	if result == nil {
		result = []cdpcontrol.Study{}
	}
	return result, nil
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

func (s *Service) ResetView(ctx context.Context, chartID string) error {
	return s.cdp.ResetView(ctx, strings.TrimSpace(chartID))
}

func (s *Service) UndoChart(ctx context.Context, chartID string) error {
	return s.cdp.UndoChart(ctx, strings.TrimSpace(chartID))
}

func (s *Service) RedoChart(ctx context.Context, chartID string) error {
	return s.cdp.RedoChart(ctx, strings.TrimSpace(chartID))
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

var validPresets = map[string]bool{
	"1D": true, "5D": true, "1M": true, "3M": true, "6M": true,
	"YTD": true, "1Y": true, "5Y": true, "All": true,
}

func (s *Service) SetTimeFrame(ctx context.Context, chartID, preset, resolution string, pane int) (cdpcontrol.TimeFrameResult, error) {
	preset = strings.TrimSpace(preset)
	if preset == "" {
		return cdpcontrol.TimeFrameResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "preset is required"}
	}
	if !validPresets[preset] {
		return cdpcontrol.TimeFrameResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: fmt.Sprintf("invalid preset %q; valid values: 1D, 5D, 1M, 3M, 6M, YTD, 1Y, 5Y, All", preset)}
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.TimeFrameResult{}, err
	}
	return s.cdp.SetTimeFrame(ctx, strings.TrimSpace(chartID), preset, strings.TrimSpace(resolution))
}

func (s *Service) ResetScales(ctx context.Context, chartID string) error {
	return s.cdp.ResetScales(ctx, strings.TrimSpace(chartID))
}

// --- Chart Toggles methods ---

func (s *Service) GetChartToggles(ctx context.Context, chartID string, pane int) (cdpcontrol.ChartToggles, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.ChartToggles{}, err
	}
	return s.cdp.GetChartToggles(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ToggleLogScale(ctx context.Context, chartID string) error {
	return s.cdp.ToggleLogScale(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ToggleAutoScale(ctx context.Context, chartID string) error {
	return s.cdp.ToggleAutoScale(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ToggleExtendedHours(ctx context.Context, chartID string) error {
	return s.cdp.ToggleExtendedHours(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ListWatchlists(ctx context.Context) ([]cdpcontrol.WatchlistInfo, error) {
	return s.cdp.ListWatchlists(ctx)
}

func (s *Service) GetActiveWatchlist(ctx context.Context) (cdpcontrol.WatchlistDetail, error) {
	return s.cdp.GetActiveWatchlist(ctx)
}

func (s *Service) SetActiveWatchlist(ctx context.Context, id string) (cdpcontrol.WatchlistInfo, error) {
	if err := s.requireNonEmpty(id, "watchlist id"); err != nil {
		return cdpcontrol.WatchlistInfo{}, err
	}

	return s.cdp.SetActiveWatchlist(ctx, strings.TrimSpace(id))
}

func (s *Service) GetWatchlist(ctx context.Context, id string) (cdpcontrol.WatchlistDetail, error) {
	if err := s.requireNonEmpty(id, "watchlist_id"); err != nil {
		return cdpcontrol.WatchlistDetail{}, err
	}
	return s.cdp.GetWatchlist(ctx, strings.TrimSpace(id))
}

func (s *Service) CreateWatchlist(ctx context.Context, name string) (cdpcontrol.WatchlistInfo, error) {
	if err := s.requireNonEmpty(name, "name"); err != nil {
		return cdpcontrol.WatchlistInfo{}, err
	}
	return s.cdp.CreateWatchlist(ctx, strings.TrimSpace(name))
}

func (s *Service) RenameWatchlist(ctx context.Context, id, name string) (cdpcontrol.WatchlistInfo, error) {
	if err := s.requireNonEmpty(id, "watchlist_id"); err != nil {
		return cdpcontrol.WatchlistInfo{}, err
	}
	if err := s.requireNonEmpty(name, "name"); err != nil {
		return cdpcontrol.WatchlistInfo{}, err
	}
	return s.cdp.RenameWatchlist(ctx, strings.TrimSpace(id), strings.TrimSpace(name))
}

func (s *Service) DeleteWatchlist(ctx context.Context, id string) error {
	if err := s.requireNonEmpty(id, "watchlist_id"); err != nil {
		return err
	}

	return s.cdp.DeleteWatchlist(ctx, strings.TrimSpace(id))
}

func (s *Service) AddWatchlistSymbols(ctx context.Context, id string, symbols []string) (cdpcontrol.WatchlistDetail, error) {
	if err := s.requireNonEmpty(id, "watchlist_id"); err != nil {
		return cdpcontrol.WatchlistDetail{}, err
	}

	if len(symbols) == 0 {
		return cdpcontrol.WatchlistDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbols must not be empty"}
	}
	return s.cdp.AddWatchlistSymbols(ctx, strings.TrimSpace(id), symbols)
}

func (s *Service) RemoveWatchlistSymbols(ctx context.Context, id string, symbols []string) (cdpcontrol.WatchlistDetail, error) {
	if err := s.requireNonEmpty(id, "watchlist_id"); err != nil {
		return cdpcontrol.WatchlistDetail{}, err
	}
	if len(symbols) == 0 {
		return cdpcontrol.WatchlistDetail{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbols must not be empty"}
	}
	return s.cdp.RemoveWatchlistSymbols(ctx, strings.TrimSpace(id), symbols)
}

func (s *Service) FlagSymbol(ctx context.Context, id, symbol string) error {
	if err := s.requireNonEmpty(id, "watchlist_id"); err != nil {
		return err
	}

	if err := s.requireNonEmpty(symbol, "symbol"); err != nil {
		return err
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
	if err := s.requireNonEmpty(symbol, "symbol"); err != nil {
		return cdpcontrol.ResolvedSymbolInfo{}, err
	}

	return s.cdp.ResolveSymbol(ctx, strings.TrimSpace(chartID), strings.TrimSpace(symbol))
}

func (s *Service) SwitchTimezone(ctx context.Context, chartID, tz string) error {
	if err := s.requireNonEmpty(tz, "timezone"); err != nil {
		return err
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

func (s *Service) ReplayStep(ctx context.Context, chartID string, count int) error {
	return s.cdp.ReplayStep(ctx, strings.TrimSpace(chartID), count)
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
	if err := s.requireNonEmpty(strategyID, "strategy_id"); err != nil {
		return err
	}

	return s.cdp.SetActiveStrategy(ctx, strings.TrimSpace(chartID), strings.TrimSpace(strategyID))
}

func (s *Service) SetStrategyInput(ctx context.Context, chartID, name string, value any) error {
	if err := s.requireNonEmpty(name, "name"); err != nil {
		return err
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

func (s *Service) ListDrawings(ctx context.Context, chartID string, pane int) ([]cdpcontrol.Shape, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return nil, err
	}
	return s.cdp.ListDrawings(ctx, strings.TrimSpace(chartID))
}

func (s *Service) GetDrawing(ctx context.Context, chartID, shapeID string, pane int) (map[string]any, error) {
	if err := s.requireNonEmpty(shapeID, "shape_id"); err != nil {
		return nil, err
	}

	if err := s.ensurePane(ctx, pane); err != nil {
		return nil, err
	}
	return s.cdp.GetDrawing(ctx, strings.TrimSpace(chartID), strings.TrimSpace(shapeID))
}

func (s *Service) CreateDrawing(ctx context.Context, chartID string, point cdpcontrol.ShapePoint, options map[string]any, pane int) (string, error) {
	shapeName, ok := options["shape"].(string)
	if !ok || shapeName == "" {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "options must contain \"shape\" key with a string value"}
	}
	if info, known := cdpcontrol.KnownShapes[shapeName]; known && info.Points > 0 && info.Points != 1 {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: fmt.Sprintf("%q requires %d points; use the multipoint endpoint", shapeName, info.Points)}
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return "", err
	}
	return s.cdp.CreateDrawing(ctx, strings.TrimSpace(chartID), point, options)
}

func (s *Service) CreateMultipointDrawing(ctx context.Context, chartID string, points []cdpcontrol.ShapePoint, options map[string]any, pane int) (string, error) {
	if len(points) < 2 {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "points must have at least 2 entries"}
	}
	shapeName, ok := options["shape"].(string)
	if !ok || shapeName == "" {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "options must contain \"shape\" key with a string value"}
	}
	if info, known := cdpcontrol.KnownShapes[shapeName]; known && info.Points > 0 && info.Points != len(points) {
		return "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: fmt.Sprintf("%q requires exactly %d points, got %d", shapeName, info.Points, len(points))}
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return "", err
	}
	return s.cdp.CreateMultipointDrawing(ctx, strings.TrimSpace(chartID), points, options)
}

func (s *Service) CloneDrawing(ctx context.Context, chartID, shapeID string, pane int) (string, error) {
	if err := s.requireNonEmpty(shapeID, "shape_id"); err != nil {
		return "", err
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return "", err
	}
	return s.cdp.CloneDrawing(ctx, strings.TrimSpace(chartID), strings.TrimSpace(shapeID))
}

func (s *Service) RemoveDrawing(ctx context.Context, chartID, shapeID string, disableUndo bool, pane int) error {
	if err := s.requireNonEmpty(shapeID, "shape_id"); err != nil {
		return err
	}

	if err := s.ensurePane(ctx, pane); err != nil {
		return err
	}
	return s.cdp.RemoveDrawing(ctx, strings.TrimSpace(chartID), strings.TrimSpace(shapeID), disableUndo)
}

func (s *Service) RemoveAllDrawings(ctx context.Context, chartID string, pane int) error {
	if err := s.ensurePane(ctx, pane); err != nil {
		return err
	}
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
	if err := s.requireNonEmpty(shapeID, "shape_id"); err != nil {
		return err
	}

	return s.cdp.SetDrawingVisibility(ctx, strings.TrimSpace(chartID), strings.TrimSpace(shapeID), visible)
}

func (s *Service) GetDrawingTool(ctx context.Context, chartID string) (string, error) {
	return s.cdp.GetDrawingTool(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetDrawingTool(ctx context.Context, chartID, tool string) error {
	if err := s.requireNonEmpty(tool, "tool"); err != nil {
		return err
	}

	return s.cdp.SetDrawingTool(ctx, strings.TrimSpace(chartID), strings.TrimSpace(tool))
}

func (s *Service) SetDrawingZOrder(ctx context.Context, chartID, shapeID, action string) error {
	if err := s.requireNonEmpty(shapeID, "shape_id"); err != nil {
		return err
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

func (s *Service) BrowserScreenshot(ctx context.Context, format string, quality int, fullPage bool, notes string) (snapshot.SnapshotMeta, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "png"
	}
	if format != "png" && format != "jpeg" {
		return snapshot.SnapshotMeta{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "format must be \"png\" or \"jpeg\""}
	}

	imageData, err := s.cdp.BrowserScreenshot(ctx, format, quality, fullPage)
	if err != nil {
		return snapshot.SnapshotMeta{}, err
	}

	meta := snapshot.SnapshotMeta{
		ID:        uuid.New().String(),
		ChartID:   "browser",
		Format:    format,
		SizeBytes: len(imageData),
		CreatedAt: time.Now().UTC(),
		Notes:     strings.TrimSpace(notes),
	}

	if err := s.snaps.Save(meta, imageData); err != nil {
		return snapshot.SnapshotMeta{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeEvalFailure, Message: fmt.Sprintf("save snapshot: %v", err)}
	}

	return meta, nil
}

func (s *Service) GetPaneInfo(ctx context.Context) (cdpcontrol.PanesResult, error) {
	return s.cdp.GetPaneInfo(ctx)
}

func (s *Service) TakeSnapshot(ctx context.Context, chartID, format, quality, notes string, pane int) (snapshot.SnapshotMeta, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "png"
	}
	if format != "png" && format != "jpeg" {
		return snapshot.SnapshotMeta{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "format must be \"png\" or \"jpeg\""}
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return snapshot.SnapshotMeta{}, err
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
		Notes:     strings.TrimSpace(notes),
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
	if err := s.requireNonEmpty(id, "snapshot_id"); err != nil {
		return snapshot.SnapshotMeta{}, err
	}

	meta, err := s.snaps.Get(strings.TrimSpace(id))
	if err != nil {
		return snapshot.SnapshotMeta{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeSnapshotNotFound, Message: err.Error()}
	}
	return meta, nil
}

func (s *Service) ReadSnapshotImage(ctx context.Context, id string) ([]byte, string, error) {
	if err := s.requireNonEmpty(id, "snapshot_id"); err != nil {
		return nil, "", err
	}

	data, format, err := s.snaps.ReadImage(strings.TrimSpace(id))
	if err != nil {
		return nil, "", &cdpcontrol.CodedError{Code: cdpcontrol.CodeSnapshotNotFound, Message: err.Error()}
	}
	return data, format, nil
}

func (s *Service) DeleteSnapshot(ctx context.Context, id string) error {
	if err := s.requireNonEmpty(id, "snapshot_id"); err != nil {
		return err
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
	if err := s.requireNonEmpty(source, "source"); err != nil {
		return cdpcontrol.PineState{}, err
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
	if err := s.requireNonEmpty(name, "name"); err != nil {
		return cdpcontrol.PineState{}, err
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

func (s *Service) GetLayoutFavorite(ctx context.Context) (cdpcontrol.LayoutFavoriteResult, error) {
	return s.cdp.GetLayoutFavorite(ctx)
}

func (s *Service) ToggleLayoutFavorite(ctx context.Context) (cdpcontrol.LayoutFavoriteResult, error) {
	return s.cdp.ToggleLayoutFavorite(ctx)
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
	if err := s.requireNonEmpty(name, "name"); err != nil {
		return cdpcontrol.LayoutActionResult{}, err
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
	if err := s.requireNonEmpty(name, "name"); err != nil {
		return cdpcontrol.LayoutActionResult{}, err
	}
	return s.cdp.RenameLayout(ctx, strings.TrimSpace(name))
}

func (s *Service) SetLayoutGrid(ctx context.Context, template string) (cdpcontrol.LayoutStatus, error) {
	if err := s.requireNonEmpty(template, "template"); err != nil {
		return cdpcontrol.LayoutStatus{}, err
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
		snap, err := s.TakeSnapshot(ctx, chartID, "png", "", "", -1)
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

// --- Indicator Dialog methods ---

func (s *Service) SearchIndicators(ctx context.Context, chartID, query string) (cdpcontrol.IndicatorSearchResult, error) {
	if err := s.requireNonEmpty(query, "query"); err != nil {
		return cdpcontrol.IndicatorSearchResult{}, err
	}
	return s.cdp.SearchIndicators(ctx, strings.TrimSpace(chartID), strings.TrimSpace(query))
}

func (s *Service) AddIndicatorBySearch(ctx context.Context, chartID, query string, index int) (cdpcontrol.IndicatorAddResult, error) {
	if err := s.requireNonEmpty(query, "query"); err != nil {
		return cdpcontrol.IndicatorAddResult{}, err
	}
	if index < 0 {
		return cdpcontrol.IndicatorAddResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "index must be >= 0"}
	}
	return s.cdp.AddIndicatorBySearch(ctx, strings.TrimSpace(chartID), strings.TrimSpace(query), index)
}

func (s *Service) ListFavoriteIndicators(ctx context.Context, chartID string) (cdpcontrol.IndicatorSearchResult, error) {
	return s.cdp.ListFavoriteIndicators(ctx, strings.TrimSpace(chartID))
}

func (s *Service) ToggleIndicatorFavorite(ctx context.Context, chartID, query string, index int) (cdpcontrol.IndicatorFavoriteResult, error) {
	if err := s.requireNonEmpty(query, "query"); err != nil {
		return cdpcontrol.IndicatorFavoriteResult{}, err
	}
	if index < 0 {
		return cdpcontrol.IndicatorFavoriteResult{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "index must be >= 0"}
	}
	return s.cdp.ToggleIndicatorFavorite(ctx, strings.TrimSpace(chartID), strings.TrimSpace(query), index)
}

func (s *Service) ProbeIndicatorDialogDOM(ctx context.Context) (map[string]any, error) {
	return s.cdp.ProbeIndicatorDialogDOM(ctx)
}

// --- Currency / Unit methods ---

func (s *Service) GetCurrency(ctx context.Context, chartID string, pane int) (cdpcontrol.CurrencyInfo, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.CurrencyInfo{}, err
	}
	return s.cdp.GetCurrency(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetCurrency(ctx context.Context, chartID, currency string, pane int) (cdpcontrol.CurrencyInfo, error) {
	currency = strings.TrimSpace(currency)
	if currency == "" {
		return cdpcontrol.CurrencyInfo{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "currency is required (use \"null\" to reset)"}
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.CurrencyInfo{}, err
	}
	return s.cdp.SetCurrency(ctx, strings.TrimSpace(chartID), currency)
}

func (s *Service) GetAvailableCurrencies(ctx context.Context, chartID string, pane int) ([]cdpcontrol.AvailableCurrency, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return nil, err
	}
	return s.cdp.GetAvailableCurrencies(ctx, strings.TrimSpace(chartID))
}

func (s *Service) GetUnit(ctx context.Context, chartID string, pane int) (cdpcontrol.UnitInfo, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.UnitInfo{}, err
	}
	return s.cdp.GetUnit(ctx, strings.TrimSpace(chartID))
}

func (s *Service) SetUnit(ctx context.Context, chartID, unit string, pane int) (cdpcontrol.UnitInfo, error) {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return cdpcontrol.UnitInfo{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "unit is required (use \"null\" to reset)"}
	}
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.UnitInfo{}, err
	}
	return s.cdp.SetUnit(ctx, strings.TrimSpace(chartID), unit)
}

func (s *Service) GetAvailableUnits(ctx context.Context, chartID string, pane int) ([]cdpcontrol.AvailableUnit, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return nil, err
	}
	return s.cdp.GetAvailableUnits(ctx, strings.TrimSpace(chartID))
}

// --- Colored Watchlist methods ---

var validColors = map[string]bool{
	"red": true, "orange": true, "green": true, "purple": true, "blue": true,
}

func (s *Service) ListColoredWatchlists(ctx context.Context) ([]cdpcontrol.ColoredWatchlist, error) {
	return s.cdp.ListColoredWatchlists(ctx)
}

func (s *Service) ReplaceColoredWatchlist(ctx context.Context, color string, symbols []string) (cdpcontrol.ColoredWatchlist, error) {
	if !validColors[color] {
		return cdpcontrol.ColoredWatchlist{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "color must be one of: red, orange, green, purple, blue"}
	}
	if len(symbols) == 0 {
		return cdpcontrol.ColoredWatchlist{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbols must not be empty"}
	}
	return s.cdp.ReplaceColoredWatchlist(ctx, color, symbols)
}

func (s *Service) AppendColoredWatchlist(ctx context.Context, color string, symbols []string) (cdpcontrol.ColoredWatchlist, error) {
	if !validColors[color] {
		return cdpcontrol.ColoredWatchlist{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "color must be one of: red, orange, green, purple, blue"}
	}
	if len(symbols) == 0 {
		return cdpcontrol.ColoredWatchlist{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbols must not be empty"}
	}
	return s.cdp.AppendColoredWatchlist(ctx, color, symbols)
}

func (s *Service) RemoveColoredWatchlist(ctx context.Context, color string, symbols []string) (cdpcontrol.ColoredWatchlist, error) {
	if !validColors[color] {
		return cdpcontrol.ColoredWatchlist{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "color must be one of: red, orange, green, purple, blue"}
	}
	if len(symbols) == 0 {
		return cdpcontrol.ColoredWatchlist{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbols must not be empty"}
	}
	return s.cdp.RemoveColoredWatchlist(ctx, color, symbols)
}

func (s *Service) BulkRemoveColoredWatchlist(ctx context.Context, symbols []string) error {
	if len(symbols) == 0 {
		return &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "symbols must not be empty"}
	}
	return s.cdp.BulkRemoveColoredWatchlist(ctx, symbols)
}

// --- Study Template methods ---

func (s *Service) ListStudyTemplates(ctx context.Context) (cdpcontrol.StudyTemplateList, error) {
	return s.cdp.ListStudyTemplates(ctx)
}

func (s *Service) GetStudyTemplate(ctx context.Context, id int) (cdpcontrol.StudyTemplateEntry, error) {
	if id <= 0 {
		return cdpcontrol.StudyTemplateEntry{}, &cdpcontrol.CodedError{Code: cdpcontrol.CodeValidation, Message: "template_id must be > 0"}
	}
	return s.cdp.GetStudyTemplate(ctx, id)
}

// --- Hotlists Manager methods ---

func (s *Service) ProbeHotlistsManager(ctx context.Context) (cdpcontrol.HotlistsManagerProbe, error) {
	return s.cdp.ProbeHotlistsManager(ctx)
}

func (s *Service) ProbeHotlistsManagerDeep(ctx context.Context) (map[string]any, error) {
	return s.cdp.ProbeHotlistsManagerDeep(ctx)
}

func (s *Service) GetHotlistMarkets(ctx context.Context) (any, error) {
	return s.cdp.GetHotlistMarkets(ctx)
}

func (s *Service) GetHotlistExchanges(ctx context.Context) ([]cdpcontrol.HotlistExchangeDetail, error) {
	return s.cdp.GetHotlistExchanges(ctx)
}

func (s *Service) GetOneHotlist(ctx context.Context, exchange, group string) (cdpcontrol.HotlistResult, error) {
	if err := s.requireNonEmpty(exchange, "exchange"); err != nil {
		return cdpcontrol.HotlistResult{}, err
	}
	if err := s.requireNonEmpty(group, "group"); err != nil {
		return cdpcontrol.HotlistResult{}, err
	}
	return s.cdp.GetOneHotlist(ctx, strings.TrimSpace(exchange), strings.TrimSpace(group))
}

// --- Data Window Probe methods ---

func (s *Service) ProbeDataWindow(ctx context.Context, chartID string, pane int) (cdpcontrol.DataWindowProbe, error) {
	if err := s.ensurePane(ctx, pane); err != nil {
		return cdpcontrol.DataWindowProbe{}, err
	}
	return s.cdp.ProbeDataWindow(ctx, strings.TrimSpace(chartID))
}

func decodeDataURL(dataURL string) ([]byte, error) {
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data URL format")
	}
	return base64.StdEncoding.DecodeString(parts[1])
}
