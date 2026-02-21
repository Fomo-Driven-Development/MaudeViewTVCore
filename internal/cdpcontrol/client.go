package cdpcontrol

import (
	"context"
	"encoding/base64"
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

const (
	uiSettleShort  = 50 * time.Millisecond
	uiSettleMedium = 300 * time.Millisecond
	uiSettleLong   = 500 * time.Millisecond
)

var chartURLPattern = regexp.MustCompile(`/chart/([^/?#]+)/?`)

// transientHints are substrings in error causes that indicate a transient
// failure worth retrying (e.g. broken connection, closed session).
var sendShortcutDispatch = func(c *Client, ctx context.Context, key, code string, keyCode, modifiers int) error {
	return c.sendKeysOnAnyChart(ctx, key, code, keyCode, modifiers)
}

var sendShortcutWait = func(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

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
				if err := c.cdp.detachFromTarget(ctx, session.sessionID); err != nil {
					slog.Debug("detach cleanup failed", "error", err)
				}
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
	case <-time.After(uiSettleLong):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	return c.GetResolution(ctx, chartID)
}

func (c *Client) GetChartType(ctx context.Context, chartID string) (int, error) {
	var out struct {
		ChartType int `json:"chart_type"`
	}
	if err := c.evalOnChart(ctx, chartID, jsGetChartType(), &out); err != nil {
		return 0, err
	}
	return out.ChartType, nil
}

func (c *Client) SetChartType(ctx context.Context, chartID string, chartType int) (int, error) {
	if err := c.evalOnChart(ctx, chartID, jsSetChartType(chartType), nil); err != nil {
		return 0, err
	}

	// Give TradingView time to process the chart type change before verifying.
	select {
	case <-time.After(uiSettleLong):
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	return c.GetChartType(ctx, chartID)
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

func (c *Client) ResetView(ctx context.Context, chartID string) error {
	// Alt+R via CDP — trusted "Reset chart view" keyboard shortcut.
	// modifiers: 1=Alt
	return c.sendShortcut(ctx, "r", "KeyR", 82, 1, uiSettleLong, "failed to send Alt+R")
}

func (c *Client) UndoChart(ctx context.Context, chartID string) error {
	// Ctrl+Z — chart-level undo for drawings, studies, etc.
	// modifiers: 2=Ctrl
	return c.sendShortcut(ctx, "z", "KeyZ", 90, 2, uiSettleMedium, "failed to send Ctrl+Z")
}

func (c *Client) RedoChart(ctx context.Context, chartID string) error {
	// Ctrl+Y — chart-level redo for drawings, studies, etc.
	// modifiers: 2=Ctrl
	return c.sendShortcut(ctx, "y", "KeyY", 89, 2, uiSettleMedium, "failed to send Ctrl+Y")
}

func (c *Client) GoToDate(ctx context.Context, chartID string, timestamp int64) error {
	// Convert Unix timestamp to YYYY-MM-DD string for the dialog textbox
	t := time.Unix(timestamp, 0).UTC()
	dateStr := t.Format("2006-01-02")

	// Step 1: Alt+G to open the "Go to" dialog (trusted CDP key event)
	// modifiers: 1=Alt
	if err := c.sendKeysOnAnyChart(ctx, "g", "KeyG", 71, 1); err != nil {
		return newError(CodeEvalFailure, "failed to send Alt+G", err)
	}

	// Step 2: Wait for dialog, fill date, focus textbox
	var fill struct {
		Status string `json:"status"`
		Date   string `json:"date"`
	}
	if err := c.evalOnAnyChart(ctx, jsGoToFillDate(dateStr), &fill); err != nil {
		return err
	}

	// Step 3: Enter to submit the form (trusted CDP key event)
	if err := c.sendKeysOnAnyChart(ctx, "Enter", "Enter", 13, 0); err != nil {
		return newError(CodeEvalFailure, "failed to send Enter", err)
	}

	// Step 4: Wait for dialog to close and data to load
	var result struct {
		Status string `json:"status"`
	}
	if err := c.evalOnAnyChart(ctx, jsGoToWaitClose(), &result); err != nil {
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

func (c *Client) SetTimeFrame(ctx context.Context, chartID, preset, resolution string) (TimeFrameResult, error) {
	var out TimeFrameResult
	if err := c.evalOnChart(ctx, chartID, jsSetTimeFrame(preset, resolution), &out); err != nil {
		return TimeFrameResult{}, err
	}
	return out, nil
}

func (c *Client) ResetScales(ctx context.Context, chartID string) error {
	return c.doChartAction(ctx, chartID, jsResetScales())
}

// --- Chart Toggles methods ---

func (c *Client) GetChartToggles(ctx context.Context, chartID string) (ChartToggles, error) {
	var out ChartToggles
	if err := c.evalOnChart(ctx, chartID, jsGetChartToggles(), &out); err != nil {
		return ChartToggles{}, err
	}
	return out, nil
}

func (c *Client) ToggleLogScale(ctx context.Context, chartID string) error {
	// Alt+L via CDP — trusted keyboard shortcut. modifiers: 1=Alt
	return c.sendShortcut(ctx, "l", "KeyL", 76, 1, uiSettleMedium, "failed to send Alt+L")
}

func (c *Client) ToggleAutoScale(ctx context.Context, chartID string) error {
	// Alt+A via CDP — trusted keyboard shortcut. modifiers: 1=Alt
	return c.sendShortcut(ctx, "a", "KeyA", 65, 1, uiSettleMedium, "failed to send Alt+A")
}

func (c *Client) ToggleExtendedHours(ctx context.Context, chartID string) error {
	// Alt+E via CDP — trusted keyboard shortcut. modifiers: 1=Alt
	return c.sendShortcut(ctx, "e", "KeyE", 69, 1, uiSettleMedium, "failed to send Alt+E")
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
	return c.doSessionAction(ctx, jsFlagSymbol(id, symbol))
}

func (c *Client) DeleteWatchlist(ctx context.Context, id string) error {
	return c.doSessionAction(ctx, jsDeleteWatchlist(id))
}

func (c *Client) DeepHealthCheck(ctx context.Context) (DeepHealthResult, error) {
	var out DeepHealthResult
	if err := c.evalOnAnyChart(ctx, jsDeepHealthCheck(), &out); err != nil {
		return DeepHealthResult{}, err
	}
	return out, nil
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
	initProbeDefaults(&out.AccessPaths, &out.Methods, nil)
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
	return c.doChartAction(ctx, chartID, jsSwitchTimezone(tz))
}

// --- Replay Manager methods ---

func (c *Client) ProbeReplayManager(ctx context.Context, chartID string) (ReplayManagerProbe, error) {
	var out ReplayManagerProbe
	if err := c.evalOnChart(ctx, chartID, jsProbeReplayManager(), &out); err != nil {
		return ReplayManagerProbe{}, err
	}
	initProbeDefaults(&out.AccessPaths, &out.Methods, &out.State)
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
	return c.doChartAction(ctx, chartID, jsDeactivateReplay())
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
	return c.doChartAction(ctx, chartID, jsStopReplay())
}

func (c *Client) ReplayStep(ctx context.Context, chartID string, count int) error {
	if count < 1 {
		count = 1
	}
	return c.doChartAction(ctx, chartID, jsReplayStep(count))
}

func (c *Client) StartAutoplay(ctx context.Context, chartID string) error {
	return c.doChartAction(ctx, chartID, jsStartAutoplay())
}

func (c *Client) StopAutoplay(ctx context.Context, chartID string) error {
	return c.doChartAction(ctx, chartID, jsStopAutoplay())
}

func (c *Client) ResetReplay(ctx context.Context, chartID string) error {
	return c.doChartAction(ctx, chartID, jsResetReplay())
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
	initProbeDefaults(&out.AccessPaths, &out.Methods, &out.State)
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
	return c.doChartAction(ctx, chartID, jsSetActiveStrategy(strategyID))
}

func (c *Client) SetStrategyInput(ctx context.Context, chartID, name string, value any) error {
	return c.doChartAction(ctx, chartID, jsSetStrategyInput(name, value))
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
	return c.doChartAction(ctx, chartID, jsStrategyGotoDate(timestamp, belowBar))
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
	initProbeDefaults(&out.AccessPaths, &out.Methods, &out.State)
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
	return c.doSessionAction(ctx, jsDeleteAlerts(ids))
}

func (c *Client) StopAlerts(ctx context.Context, ids []string) error {
	return c.doSessionAction(ctx, jsStopAlerts(ids))
}

func (c *Client) RestartAlerts(ctx context.Context, ids []string) error {
	return c.doSessionAction(ctx, jsRestartAlerts(ids))
}

func (c *Client) CloneAlerts(ctx context.Context, ids []string) error {
	return c.doSessionAction(ctx, jsCloneAlerts(ids))
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
	return c.doSessionAction(ctx, jsDeleteFires(ids))
}

func (c *Client) DeleteAllFires(ctx context.Context) error {
	return c.doSessionAction(ctx, jsDeleteAllFires())
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

func (c *Client) CreateTweetDrawing(ctx context.Context, chartID string, tweetURL string) (TweetDrawingResult, error) {
	var out TweetDrawingResult
	if err := c.evalOnChart(ctx, chartID, jsCreateTweetDrawing(tweetURL), &out); err != nil {
		return TweetDrawingResult{}, err
	}
	return out, nil
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
	return c.doChartAction(ctx, chartID, jsRemoveDrawing(shapeID, disableUndo))
}

func (c *Client) RemoveAllDrawings(ctx context.Context, chartID string) error {
	return c.doChartAction(ctx, chartID, jsRemoveAllDrawings())
}

func (c *Client) GetDrawingToggles(ctx context.Context, chartID string) (DrawingToggles, error) {
	var out DrawingToggles
	if err := c.evalOnChart(ctx, chartID, jsGetDrawingToggles(), &out); err != nil {
		return DrawingToggles{}, err
	}
	return out, nil
}

func (c *Client) SetHideDrawings(ctx context.Context, chartID string, val bool) error {
	return c.doChartAction(ctx, chartID, jsSetHideDrawings(val))
}

func (c *Client) SetLockDrawings(ctx context.Context, chartID string, val bool) error {
	return c.doChartAction(ctx, chartID, jsSetLockDrawings(val))
}

func (c *Client) SetMagnet(ctx context.Context, chartID string, enabled bool, mode int) error {
	return c.doChartAction(ctx, chartID, jsSetMagnet(enabled, mode))
}

func (c *Client) SetDrawingVisibility(ctx context.Context, chartID, shapeID string, visible bool) error {
	return c.doChartAction(ctx, chartID, jsSetDrawingVisibility(shapeID, visible))
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
	return c.doChartAction(ctx, chartID, jsSetDrawingTool(tool))
}

func (c *Client) SetDrawingZOrder(ctx context.Context, chartID, shapeID, action string) error {
	return c.doChartAction(ctx, chartID, jsSetDrawingZOrder(shapeID, action))
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
	return c.doChartAction(ctx, chartID, jsImportDrawingsState(jsJSON(state)))
}

// BrowserScreenshot captures a viewport screenshot via CDP Page.captureScreenshot.
// No TradingView JS is involved — this captures whatever is visible in the browser tab.
func (c *Client) BrowserScreenshot(ctx context.Context, format string, quality int, fullPage bool) ([]byte, error) {
	cdp, sessionID, err := c.resolveAnySession(ctx)
	if err != nil {
		return nil, err
	}
	b64, err := cdp.captureScreenshot(ctx, sessionID, format, quality, fullPage)
	if err != nil {
		return nil, newError(CodeEvalFailure, "browser screenshot failed", err)
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, newError(CodeEvalFailure, "decode screenshot base64", err)
	}
	return data, nil
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
	time.Sleep(uiSettleLong)

	for {
		select {
		case <-pollCtx.Done():
			return newError(CodeEvalTimeout, "timed out waiting for page reload", pollCtx.Err())
		default:
		}

		// Re-attach if needed, then evaluate.
		sid, attachErr := c.ensureSession(pollCtx, cdp, session, info.TargetID)
		if attachErr != nil {
			time.Sleep(uiSettleLong)
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
			time.Sleep(uiSettleLong)
			continue
		}
		if raw == "complete" {
			break
		}
		time.Sleep(uiSettleLong)
	}

	// Refresh the tab registry so new chart IDs are available.
	if err := c.refreshTabs(ctx); err != nil {
		slog.Debug("refresh tabs after chart reload failed", "error", err)
	}
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
	if err := c.sendShortcut(ctx, "k", "KeyK", 75, 2, uiSettleShort, "failed to dispatch Ctrl+K"); err != nil {
		return PineState{}, err
	}
	if err := c.sendShortcut(ctx, "i", "KeyI", 73, 2, uiSettleShort, "failed to dispatch Ctrl+I"); err != nil {
		return PineState{}, err
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
	if err := c.sendShortcut(ctx, "k", "KeyK", 75, 2, uiSettleShort, "failed to dispatch Ctrl+K"); err != nil {
		return PineState{}, err
	}
	if err := c.sendShortcut(ctx, "s", "KeyS", 83, 2, uiSettleShort, "failed to dispatch Ctrl+S"); err != nil {
		return PineState{}, err
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
	if err := c.sendShortcut(ctx, "g", "KeyG", 71, 2, uiSettleMedium, "failed to dispatch Ctrl+G"); err != nil {
		return PineState{}, err
	}
	// Type the line number
	if err := c.insertTextOnAnyChart(ctx, fmt.Sprintf("%d", line)); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to type line number", err)
	}
	time.Sleep(uiSettleShort)
	// Enter to confirm
	if err := c.sendShortcut(ctx, "Enter", "Enter", 13, 0, 0, "failed to confirm go-to-line"); err != nil {
		return PineState{}, err
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
	if err := c.sendShortcut(ctx, "o", "KeyO", 79, 2, uiSettleLong, "failed to dispatch Ctrl+O"); err != nil {
		return PineState{}, err
	}
	// Type the script name to filter
	if err := c.insertTextOnAnyChart(ctx, name); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to type script name", err)
	}
	time.Sleep(uiSettleLong)
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
		if err := c.clickOnAnyChart(ctx, clickResult.CloseX, clickResult.CloseY); err != nil {
			slog.Debug("indicator dialog close click failed", "error", err)
		}
		time.Sleep(uiSettleMedium)
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

// doChartAction evaluates JS on a chart tab, ignoring the response data.
// Used by methods that only need to confirm the eval succeeded.
func (c *Client) doChartAction(ctx context.Context, chartID, js string) error {
	return c.evalOnChart(ctx, chartID, js, nil)
}

// doSessionAction evaluates JS on any chart tab, ignoring the response data.
func (c *Client) doSessionAction(ctx context.Context, js string) error {
	return c.evalOnAnyChart(ctx, js, nil)
}

// initProbeDefaults ensures probe result slices/maps are non-nil for JSON.
func initProbeDefaults(paths *[]string, methods *[]string, state *map[string]any) {
	if *paths == nil {
		*paths = []string{}
	}
	if *methods == nil {
		*methods = []string{}
	}
	if state != nil && *state == nil {
		*state = map[string]any{}
	}
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
	if err := cdp.dispatchMouseClick(ctx, sessionID, x, y); err != nil {
		return newError(CodeEvalFailure, "failed to dispatch trusted mouse click", err)
	}
	return nil
}

func (c *Client) sendShortcut(ctx context.Context, key, code string, keyCode, modifiers int, settle time.Duration, desc string) error {
	if err := sendShortcutDispatch(c, ctx, key, code, keyCode, modifiers); err != nil {
		return newError(CodeEvalFailure, desc, err)
	}
	return sendShortcutWait(ctx, settle)
}

// sendKeysOnAnyChart dispatches a trusted CDP key event on the first chart's session.
// modifiers is a bitmask: 1=Alt, 2=Ctrl, 4=Meta, 8=Shift.
func (c *Client) sendKeysOnAnyChart(ctx context.Context, key, code string, keyCode, modifiers int) error {
	cdp, sessionID, err := c.resolveAnySession(ctx)
	if err != nil {
		return err
	}
	if err := cdp.dispatchKeyEvent(ctx, sessionID, key, code, keyCode, modifiers); err != nil {
		return newError(CodeEvalFailure, "failed to dispatch trusted key event", err)
	}
	return nil
}

// insertTextOnAnyChart types text into the currently focused element via CDP.
func (c *Client) insertTextOnAnyChart(ctx context.Context, text string) error {
	cdp, sessionID, err := c.resolveAnySession(ctx)
	if err != nil {
		return err
	}
	if err := cdp.insertText(ctx, sessionID, text); err != nil {
		return newError(CodeEvalFailure, "failed to dispatch trusted text insertion", err)
	}
	return nil
}

// typeTextOnAnyChart types text character-by-character using CDP key events.
// Unlike insertText, this triggers React/framework input handlers.
func (c *Client) typeTextOnAnyChart(ctx context.Context, text string) error {
	cdp, sessionID, err := c.resolveAnySession(ctx)
	if err != nil {
		return err
	}
	for _, ch := range text {
		if err := cdp.dispatchCharInput(ctx, sessionID, string(ch)); err != nil {
			return newError(CodeEvalFailure, "failed to dispatch trusted character input", err)
		}
	}
	return nil
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

// --- Indicator Dialog methods (DOM-based) ---

func (c *Client) ProbeIndicatorDialogDOM(ctx context.Context) (map[string]any, error) {
	if err := c.openAndSearchIndicators(ctx, "RSI"); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.evalOnAnyChart(ctx, jsProbeIndicatorDialogDOM(), &out); err != nil {
		c.dismissIndicatorDialog(ctx)
		return nil, err
	}
	c.dismissIndicatorDialog(ctx)
	return out, nil
}

// dismissIndicatorDialog sends Escape to close the indicator dialog and waits.
// Sends Escape twice to handle cases where the first press clears the search
// text without closing the dialog.
func (c *Client) dismissIndicatorDialog(ctx context.Context) {
	for range 2 {
		if err := c.sendKeysOnAnyChart(ctx, "Escape", "Escape", 27, 0); err != nil {
			slog.Debug("sendKeys while dismissing indicator dialog failed", "error", err)
		}
		if err := c.evalOnAnyChart(ctx, jsWaitForIndicatorDialogClosed(), nil); err != nil {
			slog.Debug("eval for indicator dialog close state failed", "error", err)
		}
	}
}

// openAndSearchIndicators opens the indicator dialog and types a search query.
func (c *Client) openAndSearchIndicators(ctx context.Context, query string) error {
	// Click the chart canvas to ensure it has focus before sending "/" shortcut.
	// Without focus, the "/" key may be ignored.
	if err := c.clickOnAnyChart(ctx, 400, 400); err != nil {
		slog.Debug("indicator search focus click failed", "error", err)
	}

	if err := c.sendKeysOnAnyChart(ctx, "/", "Slash", 191, 0); err != nil {
		return newError(CodeEvalFailure, "failed to send / key", err)
	}
	var dialogCheck struct {
		DialogFound bool    `json:"dialog_found"`
		InputX      float64 `json:"input_x"`
		InputY      float64 `json:"input_y"`
	}
	if err := c.evalOnAnyChart(ctx, jsWaitForIndicatorDialog(), &dialogCheck); err != nil {
		return err
	}
	if !dialogCheck.DialogFound {
		return newError(CodeAPIUnavailable, "indicator dialog did not open", nil)
	}
	// Type the search query using document.execCommand('insertText') which
	// fires all native events that React's controlled components respond to.
	if err := c.evalOnAnyChart(ctx, jsSetIndicatorSearchValue(query), nil); err != nil {
		c.dismissIndicatorDialog(ctx)
		return err
	}
	return nil
}

func (c *Client) SearchIndicators(ctx context.Context, chartID, query string) (IndicatorSearchResult, error) {
	if err := c.openAndSearchIndicators(ctx, query); err != nil {
		return IndicatorSearchResult{}, err
	}

	// Step 4: Scrape results
	var scraped struct {
		Results    []IndicatorResult `json:"results"`
		TotalCount int               `json:"total_count"`
	}
	if err := c.evalOnAnyChart(ctx, jsScrapeIndicatorResults(), &scraped); err != nil {
		c.dismissIndicatorDialog(ctx)
		return IndicatorSearchResult{}, err
	}

	// Step 5: Dismiss dialog
	c.dismissIndicatorDialog(ctx)

	if scraped.Results == nil {
		scraped.Results = []IndicatorResult{}
	}
	return IndicatorSearchResult{
		Status:     "ok",
		Query:      query,
		Results:    scraped.Results,
		TotalCount: scraped.TotalCount,
	}, nil
}

func (c *Client) AddIndicatorBySearch(ctx context.Context, chartID, query string, index int) (IndicatorAddResult, error) {
	if err := c.openAndSearchIndicators(ctx, query); err != nil {
		return IndicatorAddResult{}, err
	}

	// Step 4: Click result at index
	var clicked struct {
		Status string `json:"status"`
		Index  int    `json:"index"`
		Name   string `json:"name"`
	}
	if err := c.evalOnAnyChart(ctx, jsClickIndicatorResult(index), &clicked); err != nil {
		c.dismissIndicatorDialog(ctx)
		return IndicatorAddResult{}, err
	}

	// Step 5: Dismiss dialog
	c.dismissIndicatorDialog(ctx)

	return IndicatorAddResult{
		Status: "added",
		Query:  query,
		Index:  clicked.Index,
		Name:   clicked.Name,
	}, nil
}

func (c *Client) ListFavoriteIndicators(ctx context.Context, chartID string) (IndicatorSearchResult, error) {
	// Click chart to ensure focus, then open indicator dialog
	if err := c.clickOnAnyChart(ctx, 400, 400); err != nil {
		slog.Debug("favorites indicator focus click failed", "error", err)
	}
	if err := c.sendKeysOnAnyChart(ctx, "/", "Slash", 191, 0); err != nil {
		return IndicatorSearchResult{}, newError(CodeEvalFailure, "failed to send / key", err)
	}
	var dialogCheck struct {
		DialogFound bool    `json:"dialog_found"`
		InputX      float64 `json:"input_x"`
		InputY      float64 `json:"input_y"`
	}
	if err := c.evalOnAnyChart(ctx, jsWaitForIndicatorDialog(), &dialogCheck); err != nil {
		return IndicatorSearchResult{}, err
	}
	if !dialogCheck.DialogFound {
		return IndicatorSearchResult{}, newError(CodeAPIUnavailable, "indicator dialog did not open", nil)
	}

	// Click "Favorites" category in sidebar
	if err := c.evalOnAnyChart(ctx, jsClickIndicatorCategory("Favorites"), nil); err != nil {
		c.dismissIndicatorDialog(ctx)
		return IndicatorSearchResult{}, err
	}

	// Step 4: Scrape results
	var scraped struct {
		Results    []IndicatorResult `json:"results"`
		TotalCount int               `json:"total_count"`
	}
	if err := c.evalOnAnyChart(ctx, jsScrapeIndicatorResults(), &scraped); err != nil {
		c.dismissIndicatorDialog(ctx)
		return IndicatorSearchResult{}, err
	}

	// Step 5: Dismiss dialog
	c.dismissIndicatorDialog(ctx)

	if scraped.Results == nil {
		scraped.Results = []IndicatorResult{}
	}
	return IndicatorSearchResult{
		Status:     "ok",
		Category:   "Favorites",
		Results:    scraped.Results,
		TotalCount: scraped.TotalCount,
	}, nil
}

func (c *Client) ToggleIndicatorFavorite(ctx context.Context, chartID, query string, index int) (IndicatorFavoriteResult, error) {
	if err := c.openAndSearchIndicators(ctx, query); err != nil {
		return IndicatorFavoriteResult{}, err
	}

	// Step 4: Locate the star button coordinates
	var starLoc struct {
		Name       string  `json:"name"`
		Index      int     `json:"index"`
		IsFavorite bool    `json:"is_favorite"`
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
	}
	if err := c.evalOnAnyChart(ctx, jsLocateIndicatorFavoriteStar(index), &starLoc); err != nil {
		c.dismissIndicatorDialog(ctx)
		return IndicatorFavoriteResult{}, err
	}

	// Step 5: CDP trusted click on the star
	if err := c.clickOnAnyChart(ctx, starLoc.X, starLoc.Y); err != nil {
		c.dismissIndicatorDialog(ctx)
		return IndicatorFavoriteResult{}, newError(CodeEvalFailure, "failed to click favorite star", err)
	}

	// Step 6: Re-check the favorite state
	var newState struct {
		Name       string `json:"name"`
		IsFavorite bool   `json:"is_favorite"`
	}
	if err := c.evalOnAnyChart(ctx, jsCheckIndicatorFavoriteState(index), &newState); err != nil {
		c.dismissIndicatorDialog(ctx)
		return IndicatorFavoriteResult{}, err
	}

	// Step 7: Dismiss dialog
	c.dismissIndicatorDialog(ctx)

	return IndicatorFavoriteResult{
		Status:     "toggled",
		Query:      query,
		Name:       newState.Name,
		IsFavorite: newState.IsFavorite,
	}, nil
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
		return newError(CodeCDPUnavailable, "failed to list targets", err)
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

func (c *Client) GetLayoutFavorite(ctx context.Context) (LayoutFavoriteResult, error) {
	var out LayoutFavoriteResult
	if err := c.evalOnAnyChart(ctx, jsGetLayoutFavorite(), &out); err != nil {
		return LayoutFavoriteResult{}, err
	}
	return out, nil
}

func (c *Client) ToggleLayoutFavorite(ctx context.Context) (LayoutFavoriteResult, error) {
	var out LayoutFavoriteResult
	if err := c.evalOnAnyChart(ctx, jsToggleLayoutFavorite(), &out); err != nil {
		return LayoutFavoriteResult{}, err
	}
	return out, nil
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

	// Step 1b: Suppress beforeunload handlers to avoid blocking dialog.
	if err := c.evalOnAnyChart(ctx, jsSuppressBeforeunload(), &struct{}{}); err != nil {
		slog.Debug("beforeunload suppression eval failed", "error", err)
	}

	// Step 1c: Enable Page domain and register auto-accept handler for any
	// remaining beforeunload dialog the JS suppression didn't catch.
	cdpConn, sessionID, resolveErr := c.resolveAnySession(ctx)
	var unregister func()
	if resolveErr == nil {
		if err := cdpConn.enablePageDomain(ctx, sessionID); err != nil {
			slog.Debug("enable page domain failed", "error", err)
		}
		sid := sessionID
		unregister = cdpConn.registerEventHandler("Page.javascriptDialogOpening", func(evtSessionID string, params json.RawMessage) {
			acceptCtx, acceptCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer acceptCancel()
			if err := cdpConn.handleJavaScriptDialog(acceptCtx, sid, true); err != nil {
				slog.Debug("auto-accept beforeunload dialog failed", "error", err)
			}
		})
	}

	// Step 2: Navigate via window.location (triggers full page reload).
	navJS := wrapJSEval(fmt.Sprintf(`window.location.href = "/chart/%s/"; return JSON.stringify({ok:true,data:{}});`, resolved.ShortURL))
	if err := c.evalOnAnyChart(ctx, navJS, &struct{}{}); err != nil {
		slog.Debug("layout navigation eval expected failure", "error", err)
	}

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
	defer func() {
		if unregister != nil {
			unregister()
		}
	}()
	time.Sleep(2 * time.Second)

	for {
		select {
		case <-pollCtx.Done():
			return LayoutActionResult{}, newError(CodeEvalTimeout, "timed out waiting for layout switch", pollCtx.Err())
		default:
		}
		if err := c.refreshTabs(ctx); err != nil {
			slog.Debug("refresh tabs during layout switch failed", "error", err)
		}
		var readyOut struct{ Ready string }
		readyErr := c.evalOnAnyChart(pollCtx, wrapJSEval(`return JSON.stringify({ok:true,data:{ready:document.readyState}});`), &readyOut)
		if readyErr == nil && readyOut.Ready == "complete" {
			break
		}
		time.Sleep(uiSettleLong)
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

func (c *Client) DeleteLayout(ctx context.Context, id int) (LayoutActionResult, error) {
	var out LayoutActionResult
	if err := c.evalOnAnyChart(ctx, jsDeleteLayout(id), &out); err != nil {
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
	if err := c.sendShortcut(ctx, "Tab", "Tab", 9, 0, uiSettleMedium, "failed to dispatch Tab key"); err != nil {
		return ActiveChartInfo{}, err
	}
	return c.GetActiveChart(ctx)
}

func (c *Client) PrevChart(ctx context.Context) (ActiveChartInfo, error) {
	// Shift+Tab key (keyCode 9, modifiers 8=Shift)
	if err := c.sendShortcut(ctx, "Tab", "Tab", 9, 8, uiSettleMedium, "failed to dispatch Shift+Tab key"); err != nil {
		return ActiveChartInfo{}, err
	}
	return c.GetActiveChart(ctx)
}

func (c *Client) MaximizeChart(ctx context.Context) (LayoutStatus, error) {
	// Alt+Enter (keyCode 13, modifiers 1=Alt)
	if err := c.sendShortcut(ctx, "Enter", "Enter", 13, 1, uiSettleMedium, "failed to dispatch Alt+Enter key"); err != nil {
		return LayoutStatus{}, err
	}
	return c.GetLayoutStatus(ctx)
}

func (c *Client) ActivateChart(ctx context.Context, index int) (LayoutStatus, error) {
	var out LayoutStatus
	if err := c.evalOnAnyChart(ctx, jsActivateChart(index), &out); err != nil {
		return LayoutStatus{}, err
	}
	return out, nil
}

func (c *Client) GetPaneInfo(ctx context.Context) (PanesResult, error) {
	var out PanesResult
	if err := c.evalOnAnyChart(ctx, jsGetPaneInfo(), &out); err != nil {
		return PanesResult{}, err
	}
	if out.Panes == nil {
		out.Panes = []PaneInfo{}
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

// --- Currency / Unit client methods ---

func (c *Client) GetCurrency(ctx context.Context, chartID string) (CurrencyInfo, error) {
	var out CurrencyInfo
	if err := c.evalOnChart(ctx, chartID, jsGetCurrency(), &out); err != nil {
		return CurrencyInfo{}, err
	}
	return out, nil
}

func (c *Client) SetCurrency(ctx context.Context, chartID, currency string) (CurrencyInfo, error) {
	if err := c.evalOnChart(ctx, chartID, jsSetCurrency(currency), nil); err != nil {
		return CurrencyInfo{}, err
	}
	select {
	case <-time.After(1500 * time.Millisecond):
	case <-ctx.Done():
		return CurrencyInfo{}, ctx.Err()
	}
	return c.GetCurrency(ctx, chartID)
}

func (c *Client) GetAvailableCurrencies(ctx context.Context, chartID string) ([]AvailableCurrency, error) {
	var out struct {
		Currencies []AvailableCurrency `json:"currencies"`
	}
	if err := c.evalOnChart(ctx, chartID, jsGetAvailableCurrencies(), &out); err != nil {
		return nil, err
	}
	if out.Currencies == nil {
		return []AvailableCurrency{}, nil
	}
	return out.Currencies, nil
}

func (c *Client) GetUnit(ctx context.Context, chartID string) (UnitInfo, error) {
	var out UnitInfo
	if err := c.evalOnChart(ctx, chartID, jsGetUnit(), &out); err != nil {
		return UnitInfo{}, err
	}
	return out, nil
}

func (c *Client) SetUnit(ctx context.Context, chartID, unit string) (UnitInfo, error) {
	if err := c.evalOnChart(ctx, chartID, jsSetUnit(unit), nil); err != nil {
		return UnitInfo{}, err
	}
	select {
	case <-time.After(1500 * time.Millisecond):
	case <-ctx.Done():
		return UnitInfo{}, ctx.Err()
	}
	return c.GetUnit(ctx, chartID)
}

func (c *Client) GetAvailableUnits(ctx context.Context, chartID string) ([]AvailableUnit, error) {
	var out struct {
		Units []AvailableUnit `json:"units"`
	}
	if err := c.evalOnChart(ctx, chartID, jsGetAvailableUnits(), &out); err != nil {
		return nil, err
	}
	if out.Units == nil {
		return []AvailableUnit{}, nil
	}
	return out.Units, nil
}

// --- Colored Watchlist methods ---

func (c *Client) ListColoredWatchlists(ctx context.Context) ([]ColoredWatchlist, error) {
	var out struct {
		ColoredWatchlists []ColoredWatchlist `json:"colored_watchlists"`
	}
	if err := c.evalOnAnyChart(ctx, jsListColoredWatchlists(), &out); err != nil {
		return nil, err
	}
	if out.ColoredWatchlists == nil {
		return []ColoredWatchlist{}, nil
	}
	return out.ColoredWatchlists, nil
}

func (c *Client) ReplaceColoredWatchlist(ctx context.Context, color string, symbols []string) (ColoredWatchlist, error) {
	var out ColoredWatchlist
	if err := c.evalOnAnyChart(ctx, jsReplaceColoredWatchlist(color, symbols), &out); err != nil {
		return ColoredWatchlist{}, err
	}
	return out, nil
}

func (c *Client) AppendColoredWatchlist(ctx context.Context, color string, symbols []string) (ColoredWatchlist, error) {
	var out ColoredWatchlist
	if err := c.evalOnAnyChart(ctx, jsAppendColoredWatchlist(color, symbols), &out); err != nil {
		return ColoredWatchlist{}, err
	}
	return out, nil
}

func (c *Client) RemoveColoredWatchlist(ctx context.Context, color string, symbols []string) (ColoredWatchlist, error) {
	var out ColoredWatchlist
	if err := c.evalOnAnyChart(ctx, jsRemoveColoredWatchlist(color, symbols), &out); err != nil {
		return ColoredWatchlist{}, err
	}
	return out, nil
}

func (c *Client) BulkRemoveColoredWatchlist(ctx context.Context, symbols []string) error {
	return c.doSessionAction(ctx, jsBulkRemoveColoredWatchlist(symbols))
}

// --- Study Template methods ---

func (c *Client) ListStudyTemplates(ctx context.Context) (StudyTemplateList, error) {
	var out StudyTemplateList
	if err := c.evalOnAnyChart(ctx, jsListStudyTemplates(), &out); err != nil {
		return StudyTemplateList{}, err
	}
	return out, nil
}

func (c *Client) GetStudyTemplate(ctx context.Context, id int) (StudyTemplateEntry, error) {
	var out StudyTemplateEntry
	if err := c.evalOnAnyChart(ctx, jsGetStudyTemplate(id), &out); err != nil {
		return StudyTemplateEntry{}, err
	}
	return out, nil
}

func (c *Client) ApplyStudyTemplate(ctx context.Context, chartID, name string) (StudyTemplateApplyResult, error) {
	var out StudyTemplateApplyResult
	if err := c.evalOnChart(ctx, chartID, jsApplyStudyTemplateByName(name), &out); err != nil {
		return StudyTemplateApplyResult{}, err
	}
	return out, nil
}

// --- Hotlists Manager methods ---

func (c *Client) ProbeHotlistsManager(ctx context.Context) (HotlistsManagerProbe, error) {
	var out HotlistsManagerProbe
	if err := c.evalOnAnyChart(ctx, jsProbeHotlistsManager(), &out); err != nil {
		return HotlistsManagerProbe{}, err
	}
	initProbeDefaults(&out.AccessPaths, &out.Methods, &out.State)
	return out, nil
}

func (c *Client) ProbeHotlistsManagerDeep(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnAnyChart(ctx, jsProbeHotlistsManagerDeep(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetHotlistMarkets(ctx context.Context) (any, error) {
	var out struct {
		Markets any `json:"markets"`
	}
	if err := c.evalOnAnyChart(ctx, jsGetHotlistMarkets(), &out); err != nil {
		return nil, err
	}
	return out.Markets, nil
}

func (c *Client) GetHotlistExchanges(ctx context.Context) ([]HotlistExchangeDetail, error) {
	var out struct {
		Exchanges []HotlistExchangeDetail `json:"exchanges"`
	}
	if err := c.evalOnAnyChart(ctx, jsGetHotlistExchanges(), &out); err != nil {
		return nil, err
	}
	if out.Exchanges == nil {
		return []HotlistExchangeDetail{}, nil
	}
	return out.Exchanges, nil
}

func (c *Client) GetOneHotlist(ctx context.Context, exchange, group string) (HotlistResult, error) {
	var out HotlistResult
	if err := c.evalOnAnyChart(ctx, jsGetOneHotlist(exchange, group), &out); err != nil {
		return HotlistResult{}, err
	}
	if out.Symbols == nil {
		out.Symbols = []HotlistSymbol{}
	}
	return out, nil
}

// --- Data Window Probe methods ---

func (c *Client) ProbeDataWindow(ctx context.Context, chartID string) (DataWindowProbe, error) {
	var out DataWindowProbe
	if err := c.evalOnChart(ctx, chartID, jsProbeDataWindow(), &out); err != nil {
		return DataWindowProbe{}, err
	}
	if out.DOMElements == nil {
		out.DOMElements = []string{}
	}
	if out.CrosshairMethods == nil {
		out.CrosshairMethods = []string{}
	}
	if out.LegendElements == nil {
		out.LegendElements = []string{}
	}
	if out.ChartWidgetProps == nil {
		out.ChartWidgetProps = []string{}
	}
	if out.ModelProps == nil {
		out.ModelProps = []string{}
	}
	if out.DataWindowState == nil {
		out.DataWindowState = map[string]any{}
	}
	return out, nil
}
