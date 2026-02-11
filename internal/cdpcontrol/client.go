package cdpcontrol

import (
	"context"
	"encoding/json"
	"errors"
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
