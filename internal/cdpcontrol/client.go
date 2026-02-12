package cdpcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/target"
)

var chartURLPattern = regexp.MustCompile(`/chart/([^/?#]+)/?`)

// transientHints are substrings in error causes that indicate a transient
// failure worth retrying (e.g. broken connection, closed session).
var transientHints = []string{
	"context canceled",
	"target closed",
	"session closed",
	"websocket",
	"connection reset",
	"broken pipe",
	"eof",
	"connection refused",
	"connection closed",
}

type tabSession struct {
	info      ChartInfo
	mu        sync.Mutex
	sessionID string // CDP session ID from Target.attachToTarget
}

type Client struct {
	cdpURL      string
	tabFilter   string
	evalTimeout time.Duration

	mu            sync.Mutex
	cdp           *rawCDP
	tabs          map[target.ID]*tabSession
	chartToTarget map[string]target.ID

	chartLocksMu sync.Mutex
	chartLocks   map[string]*sync.Mutex
}

type evalEnvelope struct {
	OK           bool            `json:"ok"`
	Data         json.RawMessage `json:"data,omitempty"`
	ErrorCode    string          `json:"error_code,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

func NewClient(cdpURL, tabFilter string, evalTimeout time.Duration) *Client {
	return &Client{
		cdpURL:        cdpURL,
		tabFilter:     strings.ToLower(strings.TrimSpace(tabFilter)),
		evalTimeout:   evalTimeout,
		tabs:          make(map[target.ID]*tabSession),
		chartToTarget: make(map[string]target.ID),
		chartLocks:    make(map[string]*sync.Mutex),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connectLocked(ctx)
}

func (c *Client) connectLocked(ctx context.Context) error {
	if c.cdpURL == "" {
		return newError(CodeCDPUnavailable, "missing CDP URL", nil)
	}

	slog.Info("cdpcontrol connect start", "cdp_url", c.cdpURL)
	c.cleanupLocked()

	c.cdp = newRawCDP(c.cdpURL)
	if err := c.cdp.connect(ctx); err != nil {
		c.cdp = nil
		return newError(CodeCDPUnavailable, "connect to CDP failed", err)
	}

	if err := c.syncTabsLocked(ctx); err != nil {
		slog.Error("cdpcontrol initial tab sync failed", "error", err)
		c.cleanupLocked()
		return newError(CodeCDPUnavailable, "connect to CDP failed", err)
	}

	slog.Info("cdpcontrol connect ok", "cdp_url", c.cdpURL, "tabs", len(c.tabs))
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleanupLocked()
	return nil
}

func (c *Client) cleanupLocked() {
	// Detach from any active sessions without closing targets.
	if c.cdp != nil {
		for _, session := range c.tabs {
			if session == nil {
				continue
			}
			session.mu.Lock()
			if session.sessionID != "" {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_ = c.cdp.detachFromTarget(ctx, session.sessionID)
				cancel()
				session.sessionID = ""
			}
			session.mu.Unlock()
		}
		c.cdp.close()
		c.cdp = nil
	}
	c.tabs = make(map[target.ID]*tabSession)
	c.chartToTarget = make(map[string]target.ID)
}

func (c *Client) ListCharts(ctx context.Context) ([]ChartInfo, error) {
	if err := c.refreshTabs(ctx); err != nil {
		slog.Warn("cdpcontrol list charts failed", "error", err)
		return nil, err
	}

	c.mu.Lock()
	charts := make([]ChartInfo, 0, len(c.tabs))
	for _, s := range c.tabs {
		if s != nil {
			charts = append(charts, s.info)
		}
	}
	c.mu.Unlock()

	sort.Slice(charts, func(i, j int) bool {
		return charts[i].ChartID < charts[j].ChartID
	})
	slog.Debug("cdpcontrol list charts", "count", len(charts))
	return charts, nil
}

func (c *Client) GetSymbolInfo(ctx context.Context, chartID string) (SymbolInfo, error) {
	var out SymbolInfo
	err := c.evalOnChart(ctx, chartID, jsGetSymbolInfo(), &out)
	if err != nil {
		return SymbolInfo{}, err
	}
	return out, nil
}

func (c *Client) GetActiveChart(ctx context.Context) (ActiveChartInfo, error) {
	charts, err := c.ListCharts(ctx)
	if err != nil {
		return ActiveChartInfo{}, err
	}
	if len(charts) == 0 {
		return ActiveChartInfo{}, newError(CodeChartNotFound, "no chart tabs found", nil)
	}

	for _, ch := range charts {
		var out struct {
			ChartIndex int `json:"chart_index"`
			ChartCount int `json:"chart_count"`
		}
		if evalErr := c.evalOnChart(ctx, ch.ChartID, jsGetActiveChart(), &out); evalErr != nil {
			continue
		}
		return ActiveChartInfo{
			ChartID:    ch.ChartID,
			TargetID:   ch.TargetID,
			URL:        ch.URL,
			Title:      ch.Title,
			ChartIndex: out.ChartIndex,
			ChartCount: out.ChartCount,
		}, nil
	}

	// Fallback: return the first chart if JS eval fails on all.
	return ActiveChartInfo{
		ChartID:    charts[0].ChartID,
		TargetID:   charts[0].TargetID,
		URL:        charts[0].URL,
		Title:      charts[0].Title,
		ChartIndex: 0,
		ChartCount: len(charts),
	}, nil
}

func (c *Client) GetSymbol(ctx context.Context, chartID string) (string, error) {
	var out struct {
		Symbol string `json:"symbol"`
	}
	err := c.evalOnChart(ctx, chartID, jsGetSymbol(), &out)
	if err != nil {
		return "", err
	}
	return out.Symbol, nil
}

func (c *Client) SetSymbol(ctx context.Context, chartID, symbol string) (string, error) {
	var out struct {
		CurrentSymbol string `json:"current_symbol"`
	}
	err := c.evalOnChart(ctx, chartID, jsSetSymbol(symbol), &out)
	if err != nil {
		return "", err
	}
	return out.CurrentSymbol, nil
}

func (c *Client) GetResolution(ctx context.Context, chartID string) (string, error) {
	var out struct {
		Resolution string `json:"resolution"`
	}
	err := c.evalOnChart(ctx, chartID, jsGetResolution(), &out)
	if err != nil {
		return "", err
	}
	return out.Resolution, nil
}

func (c *Client) SetResolution(ctx context.Context, chartID, resolution string) (string, error) {
	// Fire the setResolution call; the JS returns immediately without reading
	// back the value because TradingView reloads chart data asynchronously.
	if err := c.evalOnChart(ctx, chartID, jsSetResolution(resolution), nil); err != nil {
		return "", err
	}

	// Give TradingView time to process the resolution change before verifying.
	select {
	case <-time.After(500 * time.Millisecond):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	return c.GetResolution(ctx, chartID)
}

func (c *Client) ExecuteAction(ctx context.Context, chartID, actionID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsExecuteAction(actionID), &out); err != nil {
		return err
	}
	if out.Status == "" {
		return newError(CodeEvalFailure, "empty action status", nil)
	}
	return nil
}

func (c *Client) ListStudies(ctx context.Context, chartID string) ([]Study, error) {
	var out struct {
		Studies []Study `json:"studies"`
	}
	err := c.evalOnChart(ctx, chartID, jsListStudies(), &out)
	if err != nil {
		return nil, err
	}
	if out.Studies == nil {
		return []Study{}, nil
	}
	return out.Studies, nil
}

func (c *Client) AddStudy(ctx context.Context, chartID, name string, inputs map[string]any, forceOverlay bool) (Study, error) {
	var out struct {
		Study Study `json:"study"`
	}
	err := c.evalOnChart(ctx, chartID, jsAddStudy(name, inputs, forceOverlay), &out)
	if err != nil {
		return Study{}, err
	}
	return out.Study, nil
}

func (c *Client) GetStudyInputs(ctx context.Context, chartID, studyID string) (StudyDetail, error) {
	var out StudyDetail
	err := c.evalOnChart(ctx, chartID, jsGetStudyInputs(studyID), &out)
	if err != nil {
		return StudyDetail{}, err
	}
	return out, nil
}

func (c *Client) ModifyStudyInputs(ctx context.Context, chartID, studyID string, inputs map[string]any) (StudyDetail, error) {
	var out StudyDetail
	err := c.evalOnChart(ctx, chartID, jsModifyStudyInputs(studyID, inputs), &out)
	if err != nil {
		return StudyDetail{}, err
	}
	return out, nil
}

func (c *Client) RemoveStudy(ctx context.Context, chartID, studyID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsRemoveStudy(studyID), &out); err != nil {
		return err
	}
	if out.Status == "" {
		return newError(CodeEvalFailure, "empty remove-study status", nil)
	}
	return nil
}

func (c *Client) Zoom(ctx context.Context, chartID, direction string) error {
	var out struct {
		Status    string `json:"status"`
		Direction string `json:"direction"`
	}
	if err := c.evalOnChart(ctx, chartID, jsZoom(direction), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) Scroll(ctx context.Context, chartID string, bars int) error {
	var out struct {
		Status string `json:"status"`
		Bars   int    `json:"bars"`
	}
	if err := c.evalOnChart(ctx, chartID, jsScroll(bars), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) ScrollToRealtime(ctx context.Context, chartID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsScrollToRealtime(), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) GoToDate(ctx context.Context, chartID string, timestamp int64) error {
	var out struct {
		Status    string `json:"status"`
		Timestamp int64  `json:"timestamp"`
	}
	if err := c.evalOnChart(ctx, chartID, jsGoToDate(timestamp), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetVisibleRange(ctx context.Context, chartID string) (VisibleRange, error) {
	var out VisibleRange
	if err := c.evalOnChart(ctx, chartID, jsGetVisibleRange(), &out); err != nil {
		return VisibleRange{}, err
	}
	return out, nil
}

func (c *Client) SetVisibleRange(ctx context.Context, chartID string, from, to float64) (VisibleRange, error) {
	var out VisibleRange
	if err := c.evalOnChart(ctx, chartID, jsSetVisibleRange(from, to), &out); err != nil {
		return VisibleRange{}, err
	}
	return out, nil
}

func (c *Client) ResetScales(ctx context.Context, chartID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsResetScales(), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) ListWatchlists(ctx context.Context) ([]WatchlistInfo, error) {
	var out struct {
		Watchlists []WatchlistInfo `json:"watchlists"`
	}
	if err := c.evalOnAnyChart(ctx, jsListWatchlists(), &out); err != nil {
		return nil, err
	}
	if out.Watchlists == nil {
		return []WatchlistInfo{}, nil
	}
	return out.Watchlists, nil
}

func (c *Client) GetActiveWatchlist(ctx context.Context) (WatchlistDetail, error) {
	var out WatchlistDetail
	if err := c.evalOnAnyChart(ctx, jsGetActiveWatchlist(), &out); err != nil {
		return WatchlistDetail{}, err
	}
	return out, nil
}

func (c *Client) SetActiveWatchlist(ctx context.Context, id string) (WatchlistInfo, error) {
	var out WatchlistInfo
	if err := c.evalOnAnyChart(ctx, jsSetActiveWatchlist(id), &out); err != nil {
		return WatchlistInfo{}, err
	}
	return out, nil
}

func (c *Client) GetWatchlist(ctx context.Context, id string) (WatchlistDetail, error) {
	var out WatchlistDetail
	if err := c.evalOnAnyChart(ctx, jsGetWatchlist(id), &out); err != nil {
		return WatchlistDetail{}, err
	}
	return out, nil
}

func (c *Client) CreateWatchlist(ctx context.Context, name string) (WatchlistInfo, error) {
	var out WatchlistInfo
	if err := c.evalOnAnyChart(ctx, jsCreateWatchlist(name), &out); err != nil {
		return WatchlistInfo{}, err
	}
	return out, nil
}

func (c *Client) RenameWatchlist(ctx context.Context, id, name string) (WatchlistInfo, error) {
	var out WatchlistInfo
	if err := c.evalOnAnyChart(ctx, jsRenameWatchlist(id, name), &out); err != nil {
		return WatchlistInfo{}, err
	}
	return out, nil
}

func (c *Client) AddWatchlistSymbols(ctx context.Context, id string, symbols []string) (WatchlistDetail, error) {
	var out WatchlistDetail
	if err := c.evalOnAnyChart(ctx, jsAddWatchlistSymbols(id, symbols), &out); err != nil {
		return WatchlistDetail{}, err
	}
	return out, nil
}

func (c *Client) RemoveWatchlistSymbols(ctx context.Context, id string, symbols []string) (WatchlistDetail, error) {
	var out WatchlistDetail
	if err := c.evalOnAnyChart(ctx, jsRemoveWatchlistSymbols(id, symbols), &out); err != nil {
		return WatchlistDetail{}, err
	}
	return out, nil
}

func (c *Client) FlagSymbol(ctx context.Context, id, symbol string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnAnyChart(ctx, jsFlagSymbol(id, symbol), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteWatchlist(ctx context.Context, id string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnAnyChart(ctx, jsDeleteWatchlist(id), &out); err != nil {
		return err
	}
	return nil
}

// --- ChartAPI methods ---

func (c *Client) ProbeChartApiDeep(ctx context.Context, chartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsProbeChartApiDeep(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ProbeChartApi(ctx context.Context, chartID string) (ChartApiProbe, error) {
	var out ChartApiProbe
	if err := c.evalOnChart(ctx, chartID, jsProbeChartApi(), &out); err != nil {
		return ChartApiProbe{}, err
	}
	if out.AccessPaths == nil {
		out.AccessPaths = []string{}
	}
	if out.Methods == nil {
		out.Methods = []string{}
	}
	return out, nil
}

func (c *Client) ResolveSymbol(ctx context.Context, chartID, symbol string) (ResolvedSymbolInfo, error) {
	var out ResolvedSymbolInfo
	if err := c.evalOnChart(ctx, chartID, jsResolveSymbol(symbol), &out); err != nil {
		return ResolvedSymbolInfo{}, err
	}
	return out, nil
}

func (c *Client) SwitchTimezone(ctx context.Context, chartID, tz string) error {
	var out struct {
		Timezone string `json:"timezone"`
	}
	if err := c.evalOnChart(ctx, chartID, jsSwitchTimezone(tz), &out); err != nil {
		return err
	}
	return nil
}

// --- Replay Manager methods ---

func (c *Client) ProbeReplayManager(ctx context.Context, chartID string) (ReplayManagerProbe, error) {
	var out ReplayManagerProbe
	if err := c.evalOnChart(ctx, chartID, jsProbeReplayManager(), &out); err != nil {
		return ReplayManagerProbe{}, err
	}
	if out.AccessPaths == nil {
		out.AccessPaths = []string{}
	}
	if out.Methods == nil {
		out.Methods = []string{}
	}
	if out.State == nil {
		out.State = map[string]any{}
	}
	return out, nil
}

func (c *Client) ScanReplayActivation(ctx context.Context, chartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsScanReplayActivation(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ProbeReplayManagerDeep(ctx context.Context, chartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsProbeReplayManagerDeep(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ActivateReplay(ctx context.Context, chartID string, date float64) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsActivateReplay(date), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ActivateReplayAuto(ctx context.Context, chartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsActivateReplayAuto(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) DeactivateReplay(ctx context.Context, chartID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsDeactivateReplay(), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetReplayStatus(ctx context.Context, chartID string) (ReplayStatus, error) {
	var out ReplayStatus
	if err := c.evalOnChart(ctx, chartID, jsGetReplayStatus(), &out); err != nil {
		return ReplayStatus{}, err
	}
	return out, nil
}

func (c *Client) StartReplay(ctx context.Context, chartID string, point float64) error {
	var out struct {
		Status string  `json:"status"`
		Point  float64 `json:"point"`
	}
	if err := c.evalOnChart(ctx, chartID, jsStartReplay(point), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) StopReplay(ctx context.Context, chartID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsStopReplay(), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) ReplayStep(ctx context.Context, chartID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsReplayStep(), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) StartAutoplay(ctx context.Context, chartID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsStartAutoplay(), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) StopAutoplay(ctx context.Context, chartID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsStopAutoplay(), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) ResetReplay(ctx context.Context, chartID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsResetReplay(), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) ChangeAutoplayDelay(ctx context.Context, chartID string, delay float64) (float64, error) {
	var out struct {
		Status string  `json:"status"`
		Delay  float64 `json:"delay"`
	}
	if err := c.evalOnChart(ctx, chartID, jsChangeAutoplayDelay(delay), &out); err != nil {
		return 0, err
	}
	return out.Delay, nil
}

// --- Backtesting Strategy API methods ---

func (c *Client) ProbeBacktestingApi(ctx context.Context, chartID string) (StrategyApiProbe, error) {
	var out StrategyApiProbe
	if err := c.evalOnChart(ctx, chartID, jsProbeBacktestingApi(), &out); err != nil {
		return StrategyApiProbe{}, err
	}
	if out.AccessPaths == nil {
		out.AccessPaths = []string{}
	}
	if out.Methods == nil {
		out.Methods = []string{}
	}
	if out.State == nil {
		out.State = map[string]any{}
	}
	return out, nil
}

func (c *Client) ListStrategies(ctx context.Context, chartID string) (any, error) {
	var out struct {
		Strategies any `json:"strategies"`
	}
	if err := c.evalOnChart(ctx, chartID, jsListStrategies(), &out); err != nil {
		return nil, err
	}
	return out.Strategies, nil
}

func (c *Client) GetActiveStrategy(ctx context.Context, chartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsGetActiveStrategy(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) SetActiveStrategy(ctx context.Context, chartID, strategyID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsSetActiveStrategy(strategyID), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) SetStrategyInput(ctx context.Context, chartID, name string, value any) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsSetStrategyInput(name, value), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetStrategyReport(ctx context.Context, chartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsGetStrategyReport(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetStrategyDateRange(ctx context.Context, chartID string) (any, error) {
	var out struct {
		DateRange any `json:"date_range"`
	}
	if err := c.evalOnChart(ctx, chartID, jsGetStrategyDateRange(), &out); err != nil {
		return nil, err
	}
	return out.DateRange, nil
}

func (c *Client) StrategyGotoDate(ctx context.Context, chartID string, timestamp float64, belowBar bool) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsStrategyGotoDate(timestamp, belowBar), &out); err != nil {
		return err
	}
	return nil
}

// --- Alerts REST API methods ---

func (c *Client) ScanAlertsAccess(ctx context.Context, chartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsScanAlertsAccess(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ProbeAlertsRestApi(ctx context.Context, chartID string) (AlertsApiProbe, error) {
	var out AlertsApiProbe
	if err := c.evalOnChart(ctx, chartID, jsProbeAlertsRestApi(), &out); err != nil {
		return AlertsApiProbe{}, err
	}
	if out.AccessPaths == nil {
		out.AccessPaths = []string{}
	}
	if out.Methods == nil {
		out.Methods = []string{}
	}
	if out.State == nil {
		out.State = map[string]any{}
	}
	return out, nil
}

func (c *Client) ProbeAlertsRestApiDeep(ctx context.Context, chartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsProbeAlertsRestApiDeep(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListAlerts(ctx context.Context) (any, error) {
	var out struct {
		Alerts any `json:"alerts"`
	}
	if err := c.evalOnAnyChart(ctx, jsListAlerts(), &out); err != nil {
		return nil, err
	}
	return out.Alerts, nil
}

func (c *Client) GetAlerts(ctx context.Context, ids []string) (any, error) {
	var out struct {
		Alerts any `json:"alerts"`
	}
	if err := c.evalOnAnyChart(ctx, jsGetAlerts(ids), &out); err != nil {
		return nil, err
	}
	return out.Alerts, nil
}

func (c *Client) CreateAlert(ctx context.Context, params map[string]any) (any, error) {
	var out struct {
		Alert any `json:"alert"`
	}
	if err := c.evalOnAnyChart(ctx, jsCreateAlert(params), &out); err != nil {
		return nil, err
	}
	return out.Alert, nil
}

func (c *Client) ModifyAlert(ctx context.Context, params map[string]any) (any, error) {
	var out struct {
		Alert any `json:"alert"`
	}
	if err := c.evalOnAnyChart(ctx, jsModifyAlert(params), &out); err != nil {
		return nil, err
	}
	return out.Alert, nil
}

func (c *Client) DeleteAlerts(ctx context.Context, ids []string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnAnyChart(ctx, jsDeleteAlerts(ids), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) StopAlerts(ctx context.Context, ids []string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnAnyChart(ctx, jsStopAlerts(ids), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) RestartAlerts(ctx context.Context, ids []string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnAnyChart(ctx, jsRestartAlerts(ids), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) CloneAlerts(ctx context.Context, ids []string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnAnyChart(ctx, jsCloneAlerts(ids), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) ListFires(ctx context.Context) (any, error) {
	var out struct {
		Fires any `json:"fires"`
	}
	if err := c.evalOnAnyChart(ctx, jsListFires(), &out); err != nil {
		return nil, err
	}
	return out.Fires, nil
}

func (c *Client) DeleteFires(ctx context.Context, ids []string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnAnyChart(ctx, jsDeleteFires(ids), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteAllFires(ctx context.Context) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnAnyChart(ctx, jsDeleteAllFires(), &out); err != nil {
		return err
	}
	return nil
}

// --- Drawing/Shape methods ---

func (c *Client) ListDrawings(ctx context.Context, chartID string) ([]Shape, error) {
	var out struct {
		Shapes []Shape `json:"shapes"`
	}
	err := c.evalOnChart(ctx, chartID, jsListDrawings(), &out)
	if err != nil {
		return nil, err
	}
	if out.Shapes == nil {
		return []Shape{}, nil
	}
	return out.Shapes, nil
}

func (c *Client) GetDrawing(ctx context.Context, chartID, shapeID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsGetDrawing(shapeID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateDrawing(ctx context.Context, chartID string, point ShapePoint, options map[string]any) (string, error) {
	var out struct {
		ID string `json:"id"`
	}
	if err := c.evalOnChart(ctx, chartID, jsCreateDrawing(jsJSON(point), jsJSON(options)), &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (c *Client) CreateMultipointDrawing(ctx context.Context, chartID string, points []ShapePoint, options map[string]any) (string, error) {
	var out struct {
		ID string `json:"id"`
	}
	if err := c.evalOnChart(ctx, chartID, jsCreateMultipointDrawing(jsJSON(points), jsJSON(options)), &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (c *Client) CloneDrawing(ctx context.Context, chartID, shapeID string) (string, error) {
	var out struct {
		ID string `json:"id"`
	}
	if err := c.evalOnChart(ctx, chartID, jsCloneDrawing(shapeID), &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (c *Client) RemoveDrawing(ctx context.Context, chartID, shapeID string, disableUndo bool) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsRemoveDrawing(shapeID, disableUndo), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveAllDrawings(ctx context.Context, chartID string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsRemoveAllDrawings(), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetDrawingToggles(ctx context.Context, chartID string) (DrawingToggles, error) {
	var out DrawingToggles
	if err := c.evalOnChart(ctx, chartID, jsGetDrawingToggles(), &out); err != nil {
		return DrawingToggles{}, err
	}
	return out, nil
}

func (c *Client) SetHideDrawings(ctx context.Context, chartID string, val bool) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsSetHideDrawings(val), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) SetLockDrawings(ctx context.Context, chartID string, val bool) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsSetLockDrawings(val), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) SetMagnet(ctx context.Context, chartID string, enabled bool, mode int) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsSetMagnet(enabled, mode), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) SetDrawingVisibility(ctx context.Context, chartID, shapeID string, visible bool) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsSetDrawingVisibility(shapeID, visible), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetDrawingTool(ctx context.Context, chartID string) (string, error) {
	var out struct {
		Tool string `json:"tool"`
	}
	if err := c.evalOnChart(ctx, chartID, jsGetDrawingTool(), &out); err != nil {
		return "", err
	}
	return out.Tool, nil
}

func (c *Client) SetDrawingTool(ctx context.Context, chartID, tool string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsSetDrawingTool(tool), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) SetDrawingZOrder(ctx context.Context, chartID, shapeID, action string) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsSetDrawingZOrder(shapeID, action), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) ExportDrawingsState(ctx context.Context, chartID string) (any, error) {
	var out struct {
		State any `json:"state"`
	}
	if err := c.evalOnChart(ctx, chartID, jsExportDrawingsState(), &out); err != nil {
		return nil, err
	}
	return out.State, nil
}

func (c *Client) ImportDrawingsState(ctx context.Context, chartID string, state any) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.evalOnChart(ctx, chartID, jsImportDrawingsState(jsJSON(state)), &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) TakeSnapshot(ctx context.Context, chartID, format, quality string, hideRes bool) (SnapshotResult, error) {
	var out SnapshotResult
	if err := c.evalOnChart(ctx, chartID, jsTakeSnapshot(format, quality, hideRes), &out); err != nil {
		return SnapshotResult{}, err
	}
	return out, nil
}

// --- Pine Editor methods (DOM-based) ---

func (c *Client) ProbePineDOM(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnAnyChart(ctx, jsProbePineDOM(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) TogglePineEditor(ctx context.Context) (PineState, error) {
	// Step 1: Locate the button to click and get its coordinates.
	var loc struct {
		Action string  `json:"action"` // "open" or "close"
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
	}
	if err := c.evalOnAnyChart(ctx, jsPineLocateToggleBtn(), &loc); err != nil {
		return PineState{}, err
	}

	// Step 2: Dispatch a trusted CDP mouse click at the button coordinates.
	if err := c.clickOnAnyChart(ctx, loc.X, loc.Y); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch trusted click", err)
	}

	// Step 3: Poll for the expected state change.
	var out PineState
	if loc.Action == "close" {
		if err := c.evalOnAnyChart(ctx, jsPineWaitForClose(), &out); err != nil {
			return PineState{}, err
		}
	} else {
		if err := c.evalOnAnyChart(ctx, jsPineWaitForOpen(), &out); err != nil {
			return PineState{}, err
		}
	}
	return out, nil
}

func (c *Client) GetPineStatus(ctx context.Context) (PineState, error) {
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineStatus(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) GetPineSource(ctx context.Context) (PineState, error) {
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineGetSource(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) SetPineSource(ctx context.Context, source string) (PineState, error) {
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineSetSource(source), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) SavePineScript(ctx context.Context) (PineState, error) {
	// Step 1: Focus the Monaco editor and check visibility.
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}

	// Step 2: Send Ctrl+S via trusted CDP key dispatch.
	// modifiers=2 means Ctrl.
	if err := c.sendKeysOnAnyChart(ctx, "s", "KeyS", 83, 2); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch Ctrl+S", err)
	}

	// Step 3: Wait for save to complete.
	if err := c.evalOnAnyChart(ctx, jsPineWaitForSave(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) AddPineToChart(ctx context.Context) (PineState, error) {
	// Step 1: Focus the Monaco editor and check visibility.
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}

	// Step 2: Send Ctrl+Enter via trusted CDP key dispatch.
	// modifiers=2 means Ctrl.
	if err := c.sendKeysOnAnyChart(ctx, "Enter", "Enter", 13, 2); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch Ctrl+Enter", err)
	}

	// Step 3: Wait for the script addition to process.
	if err := c.evalOnAnyChart(ctx, jsPineWaitForAddToChart(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

// ReloadPage reloads the browser tab for the given chart (or the first chart
// if chartID is empty). When hard is true, the browser cache is bypassed
// (equivalent to Shift+F5). After reload the session state is invalidated and
// the tab registry is refreshed so that subsequent calls use the new chart IDs.
func (c *Client) ReloadPage(ctx context.Context, chartID string, hard bool) error {
	chartID = strings.TrimSpace(chartID)

	// Resolve the target session.
	var session *tabSession
	var info ChartInfo
	var err error
	if chartID == "" {
		charts, lerr := c.ListCharts(ctx)
		if lerr != nil {
			return lerr
		}
		if len(charts) == 0 {
			return newError(CodeChartNotFound, "no chart tabs found", nil)
		}
		session, info, err = c.resolveChartSession(ctx, charts[0].ChartID)
	} else {
		session, info, err = c.resolveChartSession(ctx, chartID)
	}
	if err != nil {
		return err
	}

	// Ensure we have a CDP session on the target.
	c.mu.Lock()
	cdp := c.cdp
	c.mu.Unlock()
	if cdp == nil {
		return newError(CodeCDPUnavailable, "CDP client not connected", nil)
	}
	sessionID, err := c.ensureSession(ctx, cdp, session, info.TargetID)
	if err != nil {
		return err
	}

	// Send Page.reload.
	type reloadParams struct {
		IgnoreCache bool `json:"ignoreCache,omitempty"`
	}
	_, err = cdp.sendFlat(ctx, sessionID, "Page.reload", reloadParams{IgnoreCache: hard})
	if err != nil {
		return newError(CodeEvalFailure, "Page.reload failed", err)
	}

	// Invalidate the session — the JS context is destroyed on reload.
	session.mu.Lock()
	session.sessionID = ""
	session.mu.Unlock()

	// Wait for the page to finish loading by polling document.readyState.
	pollCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Brief pause to let the reload start.
	time.Sleep(500 * time.Millisecond)

	for {
		select {
		case <-pollCtx.Done():
			return newError(CodeEvalTimeout, "timed out waiting for page reload", pollCtx.Err())
		default:
		}

		// Re-attach if needed, then evaluate.
		sid, attachErr := c.ensureSession(pollCtx, cdp, session, info.TargetID)
		if attachErr != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		evalCtx, evalCancel := context.WithTimeout(pollCtx, 3*time.Second)
		raw, evalErr := cdp.evaluate(evalCtx, sid, `document.readyState`)
		evalCancel()
		if evalErr != nil {
			// Session was destroyed mid-load — reset and retry.
			session.mu.Lock()
			session.sessionID = ""
			session.mu.Unlock()
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if raw == "complete" {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Refresh the tab registry so new chart IDs are available.
	_ = c.refreshTabs(ctx)
	return nil
}

func (c *Client) GetPineConsole(ctx context.Context) ([]PineConsoleMessage, error) {
	var out struct {
		Messages []PineConsoleMessage `json:"messages"`
	}
	if err := c.evalOnAnyChart(ctx, jsPineGetConsole(), &out); err != nil {
		return nil, err
	}
	if out.Messages == nil {
		return []PineConsoleMessage{}, nil
	}
	return out.Messages, nil
}

// --- Pine Editor keyboard shortcut methods ---

func (c *Client) PineUndo(ctx context.Context) (PineState, error) {
	return c.pineKeyAction(ctx, "z", "KeyZ", 90, 2, 1, jsPineBriefWait(300))
}

func (c *Client) PineRedo(ctx context.Context) (PineState, error) {
	return c.pineKeyAction(ctx, "Z", "KeyZ", 90, 10, 1, jsPineBriefWait(300)) // Ctrl+Shift
}

func (c *Client) PineDeleteLine(ctx context.Context, count int) (PineState, error) {
	if count < 1 {
		count = 1
	}
	return c.pineKeyAction(ctx, "K", "KeyK", 75, 10, count, jsPineBriefWait(300)) // Ctrl+Shift+K
}

func (c *Client) PineMoveLine(ctx context.Context, direction string, count int) (PineState, error) {
	if count < 1 {
		count = 1
	}
	var key, code string
	var keyCode int
	if direction == "up" {
		key, code, keyCode = "ArrowUp", "ArrowUp", 38
	} else {
		key, code, keyCode = "ArrowDown", "ArrowDown", 40
	}
	return c.pineKeyAction(ctx, key, code, keyCode, 1, count, jsPineBriefWait(300)) // Alt+Arrow
}

func (c *Client) PineToggleComment(ctx context.Context) (PineState, error) {
	return c.pineKeyAction(ctx, "/", "Slash", 191, 2, 1, jsPineBriefWait(300)) // Ctrl+/
}

func (c *Client) PineToggleConsole(ctx context.Context) (PineState, error) {
	return c.pineKeyAction(ctx, "`", "Backquote", 192, 2, 1, jsPineBriefWait(500)) // Ctrl+`
}

func (c *Client) PineInsertLineAbove(ctx context.Context) (PineState, error) {
	return c.pineKeyAction(ctx, "Enter", "Enter", 13, 10, 1, jsPineBriefWait(300)) // Ctrl+Shift+Enter
}

func (c *Client) PineNewTab(ctx context.Context) (PineState, error) {
	return c.pineKeyAction(ctx, "T", "KeyT", 84, 9, 1, jsPineBriefWait(500)) // Shift+Alt+T
}

func (c *Client) PineCommandPalette(ctx context.Context) (PineState, error) {
	return c.pineKeyAction(ctx, "F1", "F1", 112, 0, 1, jsPineBriefWait(500))
}

func (c *Client) PineNewIndicator(ctx context.Context) (PineState, error) {
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	// Chord: Ctrl+K then Ctrl+I
	if err := c.sendKeysOnAnyChart(ctx, "k", "KeyK", 75, 2); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch Ctrl+K", err)
	}
	time.Sleep(50 * time.Millisecond)
	if err := c.sendKeysOnAnyChart(ctx, "i", "KeyI", 73, 2); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch Ctrl+I", err)
	}
	if err := c.evalOnAnyChart(ctx, jsPineWaitForNewScript(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) PineNewStrategy(ctx context.Context) (PineState, error) {
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	// Chord: Ctrl+K then Ctrl+S
	if err := c.sendKeysOnAnyChart(ctx, "k", "KeyK", 75, 2); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch Ctrl+K", err)
	}
	time.Sleep(50 * time.Millisecond)
	if err := c.sendKeysOnAnyChart(ctx, "s", "KeyS", 83, 2); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch Ctrl+S", err)
	}
	if err := c.evalOnAnyChart(ctx, jsPineWaitForNewScript(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) PineGoToLine(ctx context.Context, line int) (PineState, error) {
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	// Ctrl+G opens Go to Line dialog
	if err := c.sendKeysOnAnyChart(ctx, "g", "KeyG", 71, 2); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch Ctrl+G", err)
	}
	time.Sleep(200 * time.Millisecond)
	// Type the line number
	if err := c.insertTextOnAnyChart(ctx, fmt.Sprintf("%d", line)); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to type line number", err)
	}
	time.Sleep(100 * time.Millisecond)
	// Enter to confirm
	if err := c.sendKeysOnAnyChart(ctx, "Enter", "Enter", 13, 0); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to confirm go-to-line", err)
	}
	if err := c.evalOnAnyChart(ctx, jsPineBriefWait(300), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) PineOpenScript(ctx context.Context, name string) (PineState, error) {
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	// Ctrl+O opens the script picker dialog
	if err := c.sendKeysOnAnyChart(ctx, "o", "KeyO", 79, 2); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch Ctrl+O", err)
	}
	time.Sleep(500 * time.Millisecond)
	// Type the script name to filter
	if err := c.insertTextOnAnyChart(ctx, name); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to type script name", err)
	}
	time.Sleep(500 * time.Millisecond)
	// Click the first matching result and wait for load; response includes close button coords.
	var clickResult struct {
		PineState
		CloseX float64 `json:"close_x"`
		CloseY float64 `json:"close_y"`
	}
	if err := c.evalOnAnyChart(ctx, jsPineClickFirstScriptResult(), &clickResult); err != nil {
		return PineState{}, err
	}
	out = clickResult.PineState
	// Dismiss the dialog with a trusted CDP click on the close button
	if clickResult.CloseX > 0 && clickResult.CloseY > 0 {
		_ = c.clickOnAnyChart(ctx, clickResult.CloseX, clickResult.CloseY)
		time.Sleep(200 * time.Millisecond)
	}
	return out, nil
}

func (c *Client) PineFindReplace(ctx context.Context, find, replace string) (PineState, error) {
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineFindReplace(find, replace), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) evalOnAnyChart(ctx context.Context, js string, out any) error {
	charts, err := c.ListCharts(ctx)
	if err != nil {
		return err
	}
	if len(charts) == 0 {
		return newError(CodeChartNotFound, "no chart tabs found", nil)
	}
	for _, ch := range charts {
		if evalErr := c.evalOnChart(ctx, ch.ChartID, js, out); evalErr == nil {
			return nil
		}
	}
	return c.evalOnChart(ctx, charts[0].ChartID, js, out)
}

// clickOnAnyChart dispatches a trusted CDP mouse click at the given coordinates
// on the first available chart tab session.
// resolveAnySession resolves a CDP session on the first available chart tab.
func (c *Client) resolveAnySession(ctx context.Context) (*rawCDP, string, error) {
	charts, err := c.ListCharts(ctx)
	if err != nil {
		return nil, "", err
	}
	if len(charts) == 0 {
		return nil, "", newError(CodeChartNotFound, "no chart tabs found", nil)
	}
	session, info, err := c.resolveChartSession(ctx, charts[0].ChartID)
	if err != nil {
		return nil, "", err
	}
	c.mu.Lock()
	cdp := c.cdp
	c.mu.Unlock()
	if cdp == nil {
		return nil, "", newError(CodeCDPUnavailable, "CDP client not connected", nil)
	}
	sessionID, err := c.ensureSession(ctx, cdp, session, info.TargetID)
	if err != nil {
		return nil, "", err
	}
	return cdp, sessionID, nil
}

func (c *Client) clickOnAnyChart(ctx context.Context, x, y float64) error {
	cdp, sessionID, err := c.resolveAnySession(ctx)
	if err != nil {
		return err
	}
	return cdp.dispatchMouseClick(ctx, sessionID, x, y)
}

// sendKeysOnAnyChart dispatches a trusted CDP key event on the first chart's session.
// modifiers is a bitmask: 1=Alt, 2=Ctrl, 4=Meta, 8=Shift.
func (c *Client) sendKeysOnAnyChart(ctx context.Context, key, code string, keyCode, modifiers int) error {
	cdp, sessionID, err := c.resolveAnySession(ctx)
	if err != nil {
		return err
	}
	return cdp.dispatchKeyEvent(ctx, sessionID, key, code, keyCode, modifiers)
}

// insertTextOnAnyChart types text into the currently focused element via CDP.
func (c *Client) insertTextOnAnyChart(ctx context.Context, text string) error {
	cdp, sessionID, err := c.resolveAnySession(ctx)
	if err != nil {
		return err
	}
	return cdp.insertText(ctx, sessionID, text)
}

// pineKeyAction is a helper that focuses the Monaco editor, dispatches a key combo
// (optionally repeated), then evaluates a wait/poll JS function.
func (c *Client) pineKeyAction(ctx context.Context, key, code string, keyCode, modifiers, repeat int, waitJS string) (PineState, error) {
	var out PineState
	if err := c.evalOnAnyChart(ctx, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	for i := 0; i < repeat; i++ {
		if err := c.sendKeysOnAnyChart(ctx, key, code, keyCode, modifiers); err != nil {
			return PineState{}, newError(CodeEvalFailure, "failed to dispatch key", err)
		}
	}
	if err := c.evalOnAnyChart(ctx, waitJS, &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) evalOnChart(ctx context.Context, chartID, js string, out any) error {
	chartID = strings.TrimSpace(chartID)
	if chartID == "" {
		return newError(CodeChartNotFound, "chart id is required", nil)
	}

	lock := c.chartLock(chartID)
	lock.Lock()
	defer lock.Unlock()

	// First attempt.
	slog.Debug("cdpcontrol eval on chart", "chart_id", chartID)
	session, info, err := c.resolveChartSession(ctx, chartID)
	if err != nil {
		slog.Warn("cdpcontrol chart resolve failed", "chart_id", chartID, "error", err)
	} else {
		slog.Debug("cdpcontrol chart resolved", "chart_id", chartID, "target_id", info.TargetID)
		err = c.evalOnSession(ctx, session, info.TargetID, js, out)
	}
	if err == nil {
		return nil
	}
	if !c.shouldRetry(err) {
		return err
	}

	// Retry after recovery.
	slog.Warn("cdpcontrol eval retry after transient failure", "chart_id", chartID, "error", err)
	if c.asCode(err, CodeCDPUnavailable) {
		if recErr := c.reconnect(ctx); recErr != nil {
			slog.Error("cdpcontrol reconnect failed during retry", "chart_id", chartID, "error", recErr)
			return recErr
		}
	} else {
		if syncErr := c.refreshTabs(ctx); syncErr != nil {
			slog.Warn("cdpcontrol tab refresh failed during retry", "chart_id", chartID, "error", syncErr)
		}
	}

	slog.Debug("cdpcontrol eval on chart (retry)", "chart_id", chartID)
	session, info, err = c.resolveChartSession(ctx, chartID)
	if err != nil {
		slog.Warn("cdpcontrol chart resolve failed (retry)", "chart_id", chartID, "error", err)
		return err
	}
	slog.Debug("cdpcontrol chart resolved (retry)", "chart_id", chartID, "target_id", info.TargetID)
	return c.evalOnSession(ctx, session, info.TargetID, js, out)
}

func (c *Client) evalOnSession(ctx context.Context, session *tabSession, targetID, js string, out any) error {
	c.mu.Lock()
	cdp := c.cdp
	c.mu.Unlock()
	if cdp == nil {
		return newError(CodeCDPUnavailable, "CDP client not connected", nil)
	}

	// Ensure we have a session attached to this target.
	sessionID, err := c.ensureSession(ctx, cdp, session, targetID)
	if err != nil {
		return err
	}

	evalCtx, evalCancel := context.WithTimeout(ctx, c.evalTimeout)
	defer evalCancel()

	raw, err := cdp.evaluate(evalCtx, sessionID, js)
	if err != nil {
		slog.Warn("cdpcontrol eval failed", "target_id", targetID, "error", err)
		// Reset session so a fresh attach happens on retry.
		session.mu.Lock()
		session.sessionID = ""
		session.mu.Unlock()

		if errors.Is(err, context.DeadlineExceeded) || errors.Is(evalCtx.Err(), context.DeadlineExceeded) {
			return newError(CodeEvalTimeout, "evaluation timed out", err)
		}
		return newError(CodeEvalFailure, "evaluation failed", err)
	}

	var env evalEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		return newError(CodeEvalFailure, "invalid evaluation envelope", err)
	}
	if !env.OK {
		code := env.ErrorCode
		if code == "" {
			code = CodeEvalFailure
		}
		if code == CodeChartNotFound {
			return newError(CodeChartNotFound, env.ErrorMessage, nil)
		}
		if code == CodeAPIUnavailable {
			return newError(CodeAPIUnavailable, env.ErrorMessage, nil)
		}
		return newError(code, env.ErrorMessage, nil)
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return newError(CodeEvalFailure, "invalid evaluation data", err)
	}
	return nil
}

// ensureSession returns a CDP session ID for the target, attaching if needed.
func (c *Client) ensureSession(ctx context.Context, cdp *rawCDP, session *tabSession, targetID string) (string, error) {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.sessionID != "" {
		return session.sessionID, nil
	}

	sid, err := cdp.attachToTarget(ctx, targetID)
	if err != nil {
		return "", newError(CodeCDPUnavailable, "attach to target failed", err)
	}
	session.sessionID = sid
	slog.Debug("cdpcontrol session attached", "target_id", targetID, "session_id", sid)
	return sid, nil
}

func (c *Client) resolveChartSession(ctx context.Context, chartID string) (*tabSession, ChartInfo, error) {
	session, info, found := c.lookupChartSession(chartID)
	if found {
		return session, info, nil
	}

	if err := c.refreshTabs(ctx); err != nil {
		return nil, ChartInfo{}, err
	}

	session, info, found = c.lookupChartSession(chartID)
	if found {
		return session, info, nil
	}

	return nil, ChartInfo{}, newError(CodeChartNotFound, "chart not found: "+chartID, nil)
}

func (c *Client) lookupChartSession(chartID string) (*tabSession, ChartInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	targetID, ok := c.chartToTarget[chartID]
	if !ok {
		return nil, ChartInfo{}, false
	}
	session := c.tabs[targetID]
	if session == nil {
		return nil, ChartInfo{}, false
	}
	return session, session.info, true
}

func (c *Client) refreshTabs(ctx context.Context) error {
	if err := c.ensureConnected(ctx); err != nil {
		return err
	}

	c.mu.Lock()
	err := c.syncTabsLocked(ctx)
	c.mu.Unlock()
	if err == nil {
		return nil
	}

	return newError(CodeCDPUnavailable, "failed to list targets", err)
}

func (c *Client) reconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connectLocked(ctx)
}

func (c *Client) syncTabsLocked(ctx context.Context) error {
	if c.cdp == nil {
		return newError(CodeCDPUnavailable, "CDP client not connected", nil)
	}

	targets, err := c.cdp.listTargets(ctx)
	if err != nil {
		return err
	}

	expected := make(map[target.ID]ChartInfo)
	for _, t := range targets {
		if t.Type != "page" {
			continue
		}
		if c.tabFilter != "" && !strings.Contains(strings.ToLower(t.URL), c.tabFilter) {
			continue
		}
		chartID := chartIDFromURL(t.URL)
		if chartID == "" {
			continue
		}
		expected[t.TargetID] = ChartInfo{
			ChartID:  chartID,
			TargetID: string(t.TargetID),
			URL:      t.URL,
			Title:    t.Title,
		}
	}

	for targetID := range c.tabs {
		if _, ok := expected[targetID]; ok {
			continue
		}
		delete(c.tabs, targetID)
	}

	for targetID, info := range expected {
		session := c.tabs[targetID]
		if session != nil {
			session.info = info
			continue
		}
		c.tabs[targetID] = &tabSession{info: info}
	}

	c.chartToTarget = make(map[string]target.ID, len(c.tabs))
	for targetID, session := range c.tabs {
		if session == nil {
			continue
		}
		c.chartToTarget[session.info.ChartID] = targetID
	}

	// Prune chart locks for charts no longer present.
	c.chartLocksMu.Lock()
	for id := range c.chartLocks {
		if _, ok := c.chartToTarget[id]; !ok {
			delete(c.chartLocks, id)
		}
	}
	c.chartLocksMu.Unlock()

	slog.Debug("cdpcontrol tab sync", "targets", len(targets), "charts", len(c.chartToTarget))
	return nil
}

func (c *Client) ensureConnected(ctx context.Context) error {
	c.mu.Lock()
	connected := c.cdp != nil
	c.mu.Unlock()
	if connected {
		return nil
	}
	return c.reconnect(ctx)
}

func (c *Client) chartLock(chartID string) *sync.Mutex {
	c.chartLocksMu.Lock()
	defer c.chartLocksMu.Unlock()
	m, ok := c.chartLocks[chartID]
	if !ok {
		m = &sync.Mutex{}
		c.chartLocks[chartID] = m
	}
	return m
}

func (c *Client) shouldRetry(err error) bool {
	var coded *CodedError
	if !errors.As(err, &coded) {
		return false
	}

	switch coded.Code {
	case CodeCDPUnavailable:
		return true
	case CodeChartNotFound:
		return false
	case CodeEvalFailure:
		if coded.Cause == nil {
			return false
		}
		cause := strings.ToLower(coded.Cause.Error())
		for _, hint := range transientHints {
			if strings.Contains(cause, hint) {
				return true
			}
		}
	}
	return false
}

func (c *Client) asCode(err error, code string) bool {
	var coded *CodedError
	if !errors.As(err, &coded) {
		return false
	}
	return coded.Code == code
}

func chartIDFromURL(url string) string {
	m := chartURLPattern.FindStringSubmatch(url)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// --- Layout management methods ---

func (c *Client) ListLayouts(ctx context.Context) ([]LayoutInfo, error) {
	var out struct {
		Layouts []LayoutInfo `json:"layouts"`
	}
	if err := c.evalOnAnyChart(ctx, jsListLayouts(), &out); err != nil {
		return nil, err
	}
	if out.Layouts == nil {
		return []LayoutInfo{}, nil
	}
	return out.Layouts, nil
}

func (c *Client) GetLayoutStatus(ctx context.Context) (LayoutStatus, error) {
	var out LayoutStatus
	if err := c.evalOnAnyChart(ctx, jsLayoutStatus(), &out); err != nil {
		return LayoutStatus{}, err
	}
	return out, nil
}

func (c *Client) SwitchLayout(ctx context.Context, id int) (LayoutActionResult, error) {
	// Step 1: Resolve the short URL for the target layout ID.
	var resolved struct {
		ShortURL string `json:"short_url"`
		Name     string `json:"name"`
	}
	if err := c.evalOnAnyChart(ctx, jsSwitchLayoutResolveURL(id), &resolved); err != nil {
		return LayoutActionResult{}, err
	}
	if resolved.ShortURL == "" {
		return LayoutActionResult{}, newError(CodeValidation, fmt.Sprintf("layout %d not found or has no URL", id), nil)
	}

	// Step 2: Navigate via window.location (triggers full page reload).
	navJS := wrapJSEval(fmt.Sprintf(`window.location.href = "/chart/%s/"; return JSON.stringify({ok:true,data:{}});`, resolved.ShortURL))
	_ = c.evalOnAnyChart(ctx, navJS, &struct{}{}) // will likely error due to navigation

	// Step 3: Invalidate sessions and poll until the new page is ready.
	c.mu.Lock()
	for _, ts := range c.tabs {
		ts.mu.Lock()
		ts.sessionID = ""
		ts.mu.Unlock()
	}
	c.mu.Unlock()

	pollCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	time.Sleep(2 * time.Second)

	for {
		select {
		case <-pollCtx.Done():
			return LayoutActionResult{}, newError(CodeEvalTimeout, "timed out waiting for layout switch", pollCtx.Err())
		default:
		}
		_ = c.refreshTabs(ctx)
		var readyOut struct{ Ready string }
		readyErr := c.evalOnAnyChart(pollCtx, wrapJSEval(`return JSON.stringify({ok:true,data:{ready:document.readyState}});`), &readyOut)
		if readyErr == nil && readyOut.Ready == "complete" {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Step 4: Read the new layout status.
	status, statusErr := c.GetLayoutStatus(ctx)
	if statusErr != nil {
		return LayoutActionResult{Status: "switched", LayoutName: resolved.Name, LayoutID: resolved.ShortURL}, nil
	}
	return LayoutActionResult{Status: "switched", LayoutName: status.LayoutName, LayoutID: status.LayoutID}, nil
}

func (c *Client) SaveLayout(ctx context.Context) (LayoutActionResult, error) {
	var out LayoutActionResult
	if err := c.evalOnAnyChart(ctx, jsSaveLayout(), &out); err != nil {
		return LayoutActionResult{}, err
	}
	return out, nil
}

func (c *Client) CloneLayout(ctx context.Context, name string) (LayoutActionResult, error) {
	var out LayoutActionResult
	if err := c.evalOnAnyChart(ctx, jsCloneLayout(name), &out); err != nil {
		return LayoutActionResult{}, err
	}
	return out, nil
}

func (c *Client) RenameLayout(ctx context.Context, name string) (LayoutActionResult, error) {
	var out LayoutActionResult
	if err := c.evalOnAnyChart(ctx, jsRenameLayout(name), &out); err != nil {
		return LayoutActionResult{}, err
	}
	return out, nil
}

func (c *Client) SetLayoutGrid(ctx context.Context, template string) (LayoutStatus, error) {
	var out LayoutStatus
	if err := c.evalOnAnyChart(ctx, jsSetGrid(template), &out); err != nil {
		return LayoutStatus{}, err
	}
	return out, nil
}

func (c *Client) NextChart(ctx context.Context) (ActiveChartInfo, error) {
	// Tab key (keyCode 9, no modifiers)
	if err := c.sendKeysOnAnyChart(ctx, "Tab", "Tab", 9, 0); err != nil {
		return ActiveChartInfo{}, newError(CodeEvalFailure, "failed to dispatch Tab key", err)
	}
	time.Sleep(200 * time.Millisecond)
	return c.GetActiveChart(ctx)
}

func (c *Client) PrevChart(ctx context.Context) (ActiveChartInfo, error) {
	// Shift+Tab key (keyCode 9, modifiers 8=Shift)
	if err := c.sendKeysOnAnyChart(ctx, "Tab", "Tab", 9, 8); err != nil {
		return ActiveChartInfo{}, newError(CodeEvalFailure, "failed to dispatch Shift+Tab key", err)
	}
	time.Sleep(200 * time.Millisecond)
	return c.GetActiveChart(ctx)
}

func (c *Client) MaximizeChart(ctx context.Context) (LayoutStatus, error) {
	// Alt+Enter (keyCode 13, modifiers 1=Alt)
	if err := c.sendKeysOnAnyChart(ctx, "Enter", "Enter", 13, 1); err != nil {
		return LayoutStatus{}, newError(CodeEvalFailure, "failed to dispatch Alt+Enter key", err)
	}
	time.Sleep(200 * time.Millisecond)
	return c.GetLayoutStatus(ctx)
}

func (c *Client) ActivateChart(ctx context.Context, index int) (LayoutStatus, error) {
	var out LayoutStatus
	if err := c.evalOnAnyChart(ctx, jsActivateChart(index), &out); err != nil {
		return LayoutStatus{}, err
	}
	return out, nil
}

func (c *Client) ToggleFullscreen(ctx context.Context) (LayoutStatus, error) {
	var out LayoutStatus
	if err := c.evalOnAnyChart(ctx, jsToggleFullscreen(), &out); err != nil {
		return LayoutStatus{}, err
	}
	return out, nil
}

func (c *Client) DismissDialog(ctx context.Context) (LayoutActionResult, error) {
	// Escape key (keyCode 27, no modifiers) — closes any open modal/dialog
	if err := c.sendKeysOnAnyChart(ctx, "Escape", "Escape", 27, 0); err != nil {
		return LayoutActionResult{}, newError(CodeEvalFailure, "failed to dispatch Escape key", err)
	}
	return LayoutActionResult{Status: "dismissed"}, nil
}

func jsString(v string) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func jsJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func buildIIFE(async bool, body string) string {
	prefix := "(function(){\n"
	if async {
		prefix = "(async function(){\n"
	}
	return prefix + `try {
` + body + `
} catch (err) {
return JSON.stringify({ok:false,error_code:"` + CodeEvalFailure + `",error_message:String(err && err.message || err)});
}
})()`
}

func wrapJSEval(body string) string      { return buildIIFE(false, body) }
func wrapJSEvalAsync(body string) string  { return buildIIFE(true, body) }
