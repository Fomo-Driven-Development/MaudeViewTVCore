package controller

import (
	"context"
	"strings"

	"github.com/dgnsrekt/tv_agent/internal/cdpcontrol"
)

// Service wraps active TradingView control operations.
type Service struct {
	cdp *cdpcontrol.Client
}

func NewService(cdp *cdpcontrol.Client) *Service {
	return &Service{cdp: cdp}
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
