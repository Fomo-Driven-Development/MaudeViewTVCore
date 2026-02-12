package cdpcontrol

import "fmt"

const (
	CodeValidation     = "VALIDATION"
	CodeChartNotFound  = "CHART_NOT_FOUND"
	CodeAPIUnavailable = "API_UNAVAILABLE"
	CodeEvalFailure    = "EVAL_FAILURE"
	CodeEvalTimeout    = "EVAL_TIMEOUT"
	CodeCDPUnavailable    = "CDP_UNAVAILABLE"
	CodeSnapshotNotFound  = "SNAPSHOT_NOT_FOUND"
)

// CodedError is a typed error used for stable API mapping.
type CodedError struct {
	Code    string
	Message string
	Cause   error
}

func (e *CodedError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
}

func (e *CodedError) Unwrap() error { return e.Cause }

func newError(code, msg string, cause error) error {
	return &CodedError{Code: code, Message: msg, Cause: cause}
}

// ChartInfo describes a chart tab mapped from a browser target.
type ChartInfo struct {
	ChartID  string `json:"chart_id"`
	TargetID string `json:"target_id"`
	URL      string `json:"url"`
	Title    string `json:"title,omitempty"`
}

// Study describes a study entity from TradingView.
type Study struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ActiveChartInfo describes the currently active chart.
type ActiveChartInfo struct {
	ChartID    string `json:"chart_id"`
	TargetID   string `json:"target_id"`
	URL        string `json:"url"`
	Title      string `json:"title,omitempty"`
	ChartIndex int    `json:"chart_index"`
	ChartCount int    `json:"chart_count"`
}

// SymbolInfo describes extended metadata for a symbol.
type SymbolInfo struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Exchange    string `json:"exchange,omitempty"`
	Type        string `json:"type,omitempty"`
	Currency    string `json:"currency,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	PriceScale  int    `json:"pricescale,omitempty"`
	MinMov      int    `json:"minmov,omitempty"`
	HasIntraday bool   `json:"has_intraday,omitempty"`
	HasDaily    bool   `json:"has_daily,omitempty"`
}

// StudyDetail describes a study with its input parameters.
type StudyDetail struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Inputs map[string]any `json:"inputs"`
}

// WatchlistInfo describes a watchlist summary.
type WatchlistInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type,omitempty"`
	Active bool   `json:"active,omitempty"`
	Count  int    `json:"count"`
}

// WatchlistDetail describes a watchlist with its symbols.
type WatchlistDetail struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type,omitempty"`
	Symbols []string `json:"symbols"`
}

// VisibleRange describes the visible bar range on a chart.
type VisibleRange struct {
	From float64 `json:"from"`
	To   float64 `json:"to"`
}

// ChartApiProbe describes the result of probing for the chartApi() singleton.
type ChartApiProbe struct {
	Found       bool     `json:"found"`
	AccessPaths []string `json:"access_paths"`
	Methods     []string `json:"methods"`
}

// ReplayManagerProbe describes the result of probing for the _replayManager singleton.
type ReplayManagerProbe struct {
	Found       bool           `json:"found"`
	AccessPaths []string       `json:"access_paths"`
	Methods     []string       `json:"methods"`
	State       map[string]any `json:"state"`
}

// ReplayStatus describes the current state of the replay manager.
type ReplayStatus struct {
	IsReplayStarted   bool    `json:"is_replay_started"`
	IsReplayFinished  bool    `json:"is_replay_finished"`
	IsReplayConnected bool    `json:"is_replay_connected"`
	IsAutoplayStarted bool    `json:"is_autoplay_started"`
	ReplayPoint       any     `json:"replay_point"`
	ServerTime        any     `json:"server_time"`
	AutoplayDelay     float64 `json:"autoplay_delay"`
	Depth             any     `json:"depth"`
}

// AlertsApiProbe describes the result of probing for the getAlertsRestApi() singleton.
type AlertsApiProbe struct {
	Found       bool           `json:"found"`
	AccessPaths []string       `json:"access_paths"`
	Methods     []string       `json:"methods"`
	State       map[string]any `json:"state"`
}

// StrategyApiProbe describes the result of probing for the _backtestingStrategyApi singleton.
type StrategyApiProbe struct {
	Found       bool           `json:"found"`
	AccessPaths []string       `json:"access_paths"`
	Methods     []string       `json:"methods"`
	State       map[string]any `json:"state"`
}

// Shape describes a drawing entity from TradingView.
type Shape struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ShapePoint describes a point for drawing creation.
type ShapePoint struct {
	Time  float64 `json:"time"`
	Price float64 `json:"price"`
}

// DrawingToggles describes the toggle states for drawing tools.
type DrawingToggles struct {
	HideAll       *bool `json:"hide_all,omitempty"`
	LockAll       *bool `json:"lock_all,omitempty"`
	MagnetEnabled *bool `json:"magnet_enabled,omitempty"`
	MagnetMode    *int  `json:"magnet_mode,omitempty"`
}

// ResolvedSymbolInfo describes extended metadata for any symbol resolved via chartApi().
type ResolvedSymbolInfo struct {
	Symbol          string `json:"symbol"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	Exchange        string `json:"exchange,omitempty"`
	Type            string `json:"type,omitempty"`
	Currency        string `json:"currency,omitempty"`
	Timezone        string `json:"timezone,omitempty"`
	PriceScale      int    `json:"pricescale,omitempty"`
	MinMov          int    `json:"minmov,omitempty"`
	HasIntraday     bool   `json:"has_intraday,omitempty"`
	HasDaily        bool   `json:"has_daily,omitempty"`
	Session         string `json:"session,omitempty"`
	SessionHolidays string `json:"session_holidays,omitempty"`
}

// PineState describes the current state of the Pine editor panel.
type PineState struct {
	Status       string `json:"status"`
	IsVisible    bool   `json:"is_visible"`
	MonacoReady  bool   `json:"monaco_ready"`
	ScriptName   string `json:"script_name,omitempty"`
	ScriptSource string `json:"script_source,omitempty"`
	SourceLength int    `json:"source_length,omitempty"`
	LineCount    int    `json:"line_count,omitempty"`
	MatchCount   int    `json:"match_count,omitempty"`
}

// PineConsoleMessage describes a single Pine console output message.
type PineConsoleMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// SnapshotResult is the raw result from the in-page screenshot JS eval.
type SnapshotResult struct {
	DataURL  string          `json:"data_url"`
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Metadata SnapshotRawMeta `json:"metadata"`
}

// SnapshotRawMeta is the metadata envelope from api._chartWidgetCollection.images().
type SnapshotRawMeta struct {
	Layout string              `json:"layout,omitempty"`
	Theme  string              `json:"theme,omitempty"`
	Charts []SnapshotChartInfo `json:"charts,omitempty"`
}

// SnapshotChartInfo describes one chart pane inside the snapshot metadata.
type SnapshotChartInfo struct {
	Meta   SnapshotSymbolMeta `json:"meta"`
	OHLC   []string           `json:"ohlc,omitempty"`
	Quotes map[string]string  `json:"quotes,omitempty"`
}

// SnapshotSymbolMeta describes the symbol metadata for a chart pane.
type SnapshotSymbolMeta struct {
	Symbol      string `json:"symbol,omitempty"`
	Exchange    string `json:"exchange,omitempty"`
	Resolution  string `json:"resolution,omitempty"`
	Description string `json:"description,omitempty"`
}

// LayoutInfo describes a saved layout entry.
type LayoutInfo struct {
	ID       int    `json:"id"`
	URL      string `json:"url"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol,omitempty"`
	Interval string `json:"interval,omitempty"`
	Modified int64  `json:"modified,omitempty"`
	Favorite bool   `json:"favorite,omitempty"`
}

// LayoutStatus describes the current layout state.
type LayoutStatus struct {
	LayoutName   string `json:"layout_name"`
	LayoutID     string `json:"layout_id"`
	GridTemplate string `json:"grid_template"`
	ChartCount   int    `json:"chart_count"`
	ActiveIndex  int    `json:"active_index"`
	IsMaximized  bool   `json:"is_maximized"`
	IsFullscreen bool   `json:"is_fullscreen"`
	HasChanges   bool   `json:"has_changes"`
}

// LayoutActionResult describes the result of a layout action.
type LayoutActionResult struct {
	Status     string `json:"status"`
	LayoutName string `json:"layout_name,omitempty"`
	LayoutID   string `json:"layout_id,omitempty"`
}

// DeepHealthResult describes the availability of every implementation mechanism.
type DeepHealthResult struct {
	TradingViewAPI bool `json:"tradingview_api"`
	ChartWidget    bool `json:"chart_widget"`
	WebpackRequire bool `json:"webpack_require"`
	AlertsAPI      bool `json:"alerts_api"`
	WatchlistREST  bool `json:"watchlist_rest"`
	ReplayAPI      bool `json:"replay_api"`
	BacktestingAPI bool `json:"backtesting_api"`
	PineEditorDOM  bool `json:"pine_editor_dom"`
	MonacoWebpack  bool `json:"monaco_webpack"`
	LoadChart      bool `json:"load_chart"`
	SaveChart      bool `json:"save_chart"`
	ChartAPI       bool `json:"chart_api"`
}

// LayoutDetail describes a layout with its full content (studies, drawings, etc.).
type LayoutDetail struct {
	Info         LayoutInfo   `json:"layout_info"`
	Status       LayoutStatus `json:"status"`
	Studies      []Study      `json:"studies"`
	DrawingCount int          `json:"drawing_count"`
	SnapshotURL  string       `json:"snapshot_url,omitempty"`
	PreviousID   int          `json:"previous_layout_id,omitempty"`
}

// BatchDeleteResult describes the result of a batch layout delete operation.
type BatchDeleteResult struct {
	Deleted []int              `json:"deleted"`
	Skipped []int              `json:"skipped,omitempty"`
	Errors  []BatchDeleteError `json:"errors,omitempty"`
}

// BatchDeleteError describes a single layout deletion failure.
type BatchDeleteError struct {
	ID    int    `json:"id"`
	Error string `json:"error"`
}

// ReloadResult describes the result of a page reload.
type ReloadResult struct {
	Status string `json:"status"`
	Mode   string `json:"mode"`
}
