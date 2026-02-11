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
