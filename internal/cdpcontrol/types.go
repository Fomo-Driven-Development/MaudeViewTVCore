package cdpcontrol

import "fmt"

const (
	CodeValidation     = "VALIDATION"
	CodeChartNotFound  = "CHART_NOT_FOUND"
	CodeAPIUnavailable = "API_UNAVAILABLE"
	CodeEvalFailure    = "EVAL_FAILURE"
	CodeEvalTimeout    = "EVAL_TIMEOUT"
	CodeCDPUnavailable = "CDP_UNAVAILABLE"
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
