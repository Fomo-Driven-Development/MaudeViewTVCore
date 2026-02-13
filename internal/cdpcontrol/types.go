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

// ChartTypeMap maps human-readable chart type names to TradingView ChartStyle enum values.
var ChartTypeMap = map[string]int{
	"bars":              0,
	"candles":           1,
	"line":              2,
	"area":              3,
	"renko":             4,
	"kagi":              5,
	"point_and_figure":  6,
	"line_break":        7,
	"heikin_ashi":       8,
	"hollow_candles":    9,
	"baseline":          10,
	"high_low":          12,
	"columns":           13,
	"line_with_markers": 14,
	"step_line":         15,
	"hlc_area":          16,
	"volume_candles":    19,
	"hlc_bars":          21,
}

// ChartTypeReverseMap maps TradingView ChartStyle enum values back to human-readable names.
var ChartTypeReverseMap = func() map[int]string {
	m := make(map[int]string, len(ChartTypeMap))
	for name, id := range ChartTypeMap {
		m[id] = name
	}
	return m
}()

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

// ShapeInfo describes a known drawing shape and its required point count.
type ShapeInfo struct {
	Name   string `json:"name"`   // internal shape name (e.g. "trend_line")
	Label  string `json:"label"`  // display name (e.g. "Trend Line")
	Points int    `json:"points"` // required point count (1 = single-point endpoint, 2+ = multipoint)
}

// ShapeGroup groups related shapes under a UI category.
type ShapeGroup struct {
	Category string      `json:"category"` // e.g. "lines"
	Label    string      `json:"label"`    // e.g. "Lines"
	Shapes   []ShapeInfo `json:"shapes"`
}

// ShapeGroups defines all known drawing shapes, grouped by category.
// Order matches the TradingView left toolbar.
var ShapeGroups = []ShapeGroup{
	{Category: "lines", Label: "Lines", Shapes: []ShapeInfo{
		{Name: "trend_line", Label: "Trend Line", Points: 2},
		{Name: "ray", Label: "Ray", Points: 2},
		{Name: "info_line", Label: "Info Line", Points: 2},
		{Name: "extended", Label: "Extended Line", Points: 2},
		{Name: "trend_angle", Label: "Trend Angle", Points: 2},
		{Name: "horizontal_line", Label: "Horizontal Line", Points: 1},
		{Name: "horizontal_ray", Label: "Horizontal Ray", Points: 1},
		{Name: "vertical_line", Label: "Vertical Line", Points: 1},
		{Name: "cross_line", Label: "Cross Line", Points: 1},
	}},
	{Category: "channels", Label: "Channels", Shapes: []ShapeInfo{
		{Name: "parallel_channel", Label: "Parallel Channel", Points: 3},
		{Name: "regression_trend", Label: "Regression Trend", Points: 2},
		{Name: "flat_bottom", Label: "Flat Top/Bottom", Points: 3},
		{Name: "disjoint_angle", Label: "Disjoint Channel", Points: 3},
	}},
	{Category: "pitchforks", Label: "Pitchforks", Shapes: []ShapeInfo{
		{Name: "pitchfork", Label: "Pitchfork", Points: 3},
		{Name: "schiff_pitchfork", Label: "Schiff Pitchfork", Points: 3},
		{Name: "schiff_pitchfork_modified", Label: "Modified Schiff Pitchfork", Points: 3},
		{Name: "inside_pitchfork", Label: "Inside Pitchfork", Points: 3},
	}},
	{Category: "fibonacci", Label: "Fibonacci", Shapes: []ShapeInfo{
		{Name: "fib_retracement", Label: "Fib Retracement", Points: 2},
		{Name: "fib_trend_ext", Label: "Trend-Based Fib Extension", Points: 3},
		{Name: "fib_channel", Label: "Fib Channel", Points: 3},
		{Name: "fib_timezone", Label: "Fib Time Zone", Points: 2},
		{Name: "fib_speed_resist_fan", Label: "Fib Speed Resistance Fan", Points: 2},
		{Name: "fib_trend_time", Label: "Trend-Based Fib Time", Points: 3},
		{Name: "fib_circles", Label: "Fib Circles", Points: 2},
		{Name: "fib_spiral", Label: "Fib Spiral", Points: 2},
		{Name: "fib_speed_resist_arcs", Label: "Fib Speed Resistance Arcs", Points: 2},
		{Name: "fib_wedge", Label: "Fib Wedge", Points: 3},
		{Name: "pitchfan", Label: "Pitchfan", Points: 3},
	}},
	{Category: "gann", Label: "Gann", Shapes: []ShapeInfo{
		{Name: "gannbox_square", Label: "Gann Box", Points: 2},
		{Name: "gannbox_fixed", Label: "Gann Square Fixed", Points: 2},
		{Name: "gannbox", Label: "Gann Square", Points: 2},
		{Name: "gannbox_fan", Label: "Gann Fan", Points: 2},
	}},
	{Category: "patterns", Label: "Patterns", Shapes: []ShapeInfo{
		{Name: "xabcd_pattern", Label: "XABCD Pattern", Points: 5},
		{Name: "cypher_pattern", Label: "Cypher Pattern", Points: 5},
		{Name: "head_and_shoulders", Label: "Head and Shoulders", Points: 7},
		{Name: "abcd_pattern", Label: "ABCD Pattern", Points: 4},
		{Name: "triangle_pattern", Label: "Triangle Pattern", Points: 4},
		{Name: "3divers_pattern", Label: "Three Drives Pattern", Points: 7},
	}},
	{Category: "elliott_waves", Label: "Elliott Waves", Shapes: []ShapeInfo{
		{Name: "elliott_impulse_wave", Label: "Elliott Impulse Wave (12345)", Points: 6},
		{Name: "elliott_correction", Label: "Elliott Correction Wave (ABC)", Points: 4},
		{Name: "elliott_triangle_wave", Label: "Elliott Triangle Wave (ABCDE)", Points: 6},
		{Name: "elliott_double_combo", Label: "Elliott Double Combo Wave (WXY)", Points: 4},
		{Name: "elliott_triple_combo", Label: "Elliott Triple Combo Wave (WXYXZ)", Points: 6},
	}},
	{Category: "cycles", Label: "Cycles", Shapes: []ShapeInfo{
		{Name: "cyclic_lines", Label: "Cyclic Lines", Points: 2},
		{Name: "time_cycles", Label: "Time Cycles", Points: 2},
		{Name: "sine_line", Label: "Sine Line", Points: 2},
	}},
	{Category: "projection", Label: "Projection", Shapes: []ShapeInfo{
		{Name: "long_position", Label: "Long Position", Points: 2},
		{Name: "short_position", Label: "Short Position", Points: 2},
		{Name: "forecast", Label: "Forecast", Points: 2},
		{Name: "bars_pattern", Label: "Bars Pattern", Points: 2},
		{Name: "ghost_feed", Label: "Ghost Feed", Points: 5},
		{Name: "projection", Label: "Projection", Points: 3},
	}},
	{Category: "volume_based", Label: "Volume Based", Shapes: []ShapeInfo{
		{Name: "anchored_vwap", Label: "Anchored VWAP", Points: 1},
		{Name: "fixed_range_volume_profile", Label: "Fixed Range Volume Profile", Points: 2},
		{Name: "anchored_volume_profile", Label: "Anchored Volume Profile", Points: 1},
	}},
	{Category: "measurer", Label: "Measurer", Shapes: []ShapeInfo{
		{Name: "price_range", Label: "Price Range", Points: 2},
		{Name: "date_range", Label: "Date Range", Points: 2},
		{Name: "date_and_price_range", Label: "Date and Price Range", Points: 2},
	}},
	{Category: "brushes", Label: "Brushes", Shapes: []ShapeInfo{
		{Name: "brush", Label: "Brush", Points: -1},
		{Name: "highlighter", Label: "Highlighter", Points: -1},
	}},
	{Category: "arrows", Label: "Arrows", Shapes: []ShapeInfo{
		{Name: "arrow_marker", Label: "Arrow Marker", Points: 2},
		{Name: "arrow", Label: "Arrow", Points: 2},
		{Name: "arrow_up", Label: "Arrow Mark Up", Points: 1},
		{Name: "arrow_down", Label: "Arrow Mark Down", Points: 1},
	}},
	{Category: "shapes", Label: "Shapes", Shapes: []ShapeInfo{
		{Name: "rectangle", Label: "Rectangle", Points: 2},
		{Name: "rotated_rectangle", Label: "Rotated Rectangle", Points: 3},
		{Name: "path", Label: "Path", Points: -1},
		{Name: "circle", Label: "Circle", Points: 2},
		{Name: "ellipse", Label: "Ellipse", Points: 3},
		{Name: "polyline", Label: "Polyline", Points: -1},
		{Name: "triangle", Label: "Triangle", Points: 3},
		{Name: "arc", Label: "Arc", Points: 3},
		{Name: "curve", Label: "Curve", Points: 2},
		{Name: "double_curve", Label: "Double Curve", Points: 2},
	}},
	{Category: "text_and_notes", Label: "Text & Notes", Shapes: []ShapeInfo{
		{Name: "text", Label: "Text", Points: 1},
		{Name: "anchored_text", Label: "Anchored Text", Points: 1},
		{Name: "note", Label: "Note", Points: 1},
		{Name: "text_note", Label: "Text Note", Points: 2},
		{Name: "anchored_note", Label: "Anchored Note", Points: 1},
		{Name: "price_note", Label: "Price Note", Points: 2},
		{Name: "balloon", Label: "Pin", Points: 1},
		{Name: "table", Label: "Table", Points: 1},
		{Name: "callout", Label: "Callout", Points: 2},
		{Name: "comment", Label: "Comment", Points: 1},
		{Name: "price_label", Label: "Price Label", Points: 1},
		{Name: "signpost", Label: "Signpost", Points: 1},
		{Name: "flag", Label: "Flag Mark", Points: 1},
	}},
	{Category: "content", Label: "Content", Shapes: []ShapeInfo{
		{Name: "image", Label: "Image", Points: 1},
		{Name: "tweet", Label: "Tweet", Points: 1},
		{Name: "idea", Label: "Idea", Points: 1},
	}},
	{Category: "icons", Label: "Icons", Shapes: []ShapeInfo{
		{Name: "icon", Label: "Icon", Points: 1},
		{Name: "emoji", Label: "Emoji", Points: 1},
		{Name: "sticker", Label: "Sticker", Points: 1},
	}},
}

// KnownShapes is a flat lookup built from ShapeGroups for validation.
var KnownShapes = func() map[string]ShapeInfo {
	m := make(map[string]ShapeInfo)
	for _, g := range ShapeGroups {
		for _, s := range g.Shapes {
			m[s.Name] = s
		}
	}
	return m
}()

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

// PaneInfo describes one chart pane in a multi-pane grid layout.
type PaneInfo struct {
	Index      int    `json:"index"`
	Symbol     string `json:"symbol"`
	Exchange   string `json:"exchange,omitempty"`
	Resolution string `json:"resolution,omitempty"`
}

// PanesResult describes all panes in the current grid layout.
type PanesResult struct {
	GridTemplate string     `json:"grid_template"`
	ChartCount   int        `json:"chart_count"`
	ActiveIndex  int        `json:"active_index"`
	Panes        []PaneInfo `json:"panes"`
}

// TimeFrameResult describes the result of setting a time frame preset.
type TimeFrameResult struct {
	Preset     string  `json:"preset"`
	Resolution string  `json:"resolution"`
	From       float64 `json:"from"`
	To         float64 `json:"to"`
}

// ReloadResult describes the result of a page reload.
type ReloadResult struct {
	Status string `json:"status"`
	Mode   string `json:"mode"`
}

// IndicatorResult describes a single indicator entry from the Indicators dialog.
type IndicatorResult struct {
	Name       string `json:"name"`
	Author     string `json:"author,omitempty"`
	Boosts     int    `json:"boosts,omitempty"`
	IsFavorite bool   `json:"is_favorite"`
	Index      int    `json:"index"`
}

// IndicatorSearchResult describes the result of an indicator search or category browse.
type IndicatorSearchResult struct {
	Status     string            `json:"status"`
	Query      string            `json:"query,omitempty"`
	Category   string            `json:"category,omitempty"`
	Results    []IndicatorResult `json:"results"`
	TotalCount int               `json:"total_count"`
}

// IndicatorAddResult describes the result of adding an indicator via search.
type IndicatorAddResult struct {
	Status string `json:"status"`
	Query  string `json:"query"`
	Index  int    `json:"index"`
	Name   string `json:"name"`
}

// IndicatorFavoriteResult describes the result of toggling an indicator's favorite status.
type IndicatorFavoriteResult struct {
	Status     string `json:"status"`
	Query      string `json:"query"`
	Name       string `json:"name"`
	IsFavorite bool   `json:"is_favorite"`
}
