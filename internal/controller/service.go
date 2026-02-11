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
