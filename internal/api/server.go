package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/dgnsrekt/tv_agent/internal/cdpcontrol"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Service interface {
	ListCharts(ctx context.Context) ([]cdpcontrol.ChartInfo, error)
	GetActiveChart(ctx context.Context) (cdpcontrol.ActiveChartInfo, error)
	GetSymbol(ctx context.Context, chartID string) (string, error)
	SetSymbol(ctx context.Context, chartID, symbol string) (string, error)
	GetResolution(ctx context.Context, chartID string) (string, error)
	SetResolution(ctx context.Context, chartID, resolution string) (string, error)
	ExecuteAction(ctx context.Context, chartID, actionID string) error
	ListStudies(ctx context.Context, chartID string) ([]cdpcontrol.Study, error)
	AddStudy(ctx context.Context, chartID, name string, inputs map[string]any, forceOverlay bool) (cdpcontrol.Study, error)
	RemoveStudy(ctx context.Context, chartID, studyID string) error
	GetStudyInputs(ctx context.Context, chartID, studyID string) (cdpcontrol.StudyDetail, error)
	ModifyStudyInputs(ctx context.Context, chartID, studyID string, inputs map[string]any) (cdpcontrol.StudyDetail, error)
}

func NewServer(svc Service) http.Handler {
	router := chi.NewMux()
	router.Use(middleware.RequestID)
	router.Use(requestLogger)
	router.Use(middleware.Recoverer)

	cfg := huma.DefaultConfig("TV Agent Controller API", "1.0.0")
	cfg.DocsPath = ""
	api := humachi.New(router, cfg)

	router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(docsHTML))
	})

	type healthOutput struct {
		Body struct {
			Status string `json:"status"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "health", Method: http.MethodGet, Path: "/health", Summary: "Health check", Tags: []string{"Health"}},
		func(ctx context.Context, input *struct{}) (*healthOutput, error) {
			out := &healthOutput{}
			out.Body.Status = "ok"
			return out, nil
		})

	type listChartsOutput struct {
		Body struct {
			Charts []cdpcontrol.ChartInfo `json:"charts"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "list-charts", Method: http.MethodGet, Path: "/api/v1/charts", Summary: "List chart tabs", Tags: []string{"Charts"}},
		func(ctx context.Context, input *struct{}) (*listChartsOutput, error) {
			charts, err := svc.ListCharts(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &listChartsOutput{}
			out.Body.Charts = charts
			return out, nil
		})

	type activeChartOutput struct {
		Body cdpcontrol.ActiveChartInfo
	}
	huma.Register(api, huma.Operation{OperationID: "get-active-chart", Method: http.MethodGet, Path: "/api/v1/charts/active", Summary: "Get active chart", Tags: []string{"Charts"}},
		func(ctx context.Context, input *struct{}) (*activeChartOutput, error) {
			info, err := svc.GetActiveChart(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &activeChartOutput{}
			out.Body = info
			return out, nil
		})

	type chartIDInput struct {
		ChartID string `path:"chart_id"`
	}
	type symbolInput struct {
		ChartID string `path:"chart_id"`
		Symbol  string `query:"symbol" required:"true"`
	}
	type symbolOutput struct {
		Body struct {
			ChartID         string `json:"chart_id"`
			RequestedSymbol string `json:"requested_symbol,omitempty"`
			CurrentSymbol   string `json:"current_symbol"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "get-symbol", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/symbol", Summary: "Get chart symbol", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *chartIDInput) (*symbolOutput, error) {
			symbol, err := svc.GetSymbol(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &symbolOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.CurrentSymbol = symbol
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-symbol", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/symbol", Summary: "Set chart symbol", Tags: []string{"Symbol"}},
		func(ctx context.Context, input *symbolInput) (*symbolOutput, error) {
			current, err := svc.SetSymbol(ctx, input.ChartID, input.Symbol)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &symbolOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.RequestedSymbol = input.Symbol
			out.Body.CurrentSymbol = current
			return out, nil
		})

	type resolutionInput struct {
		ChartID    string `path:"chart_id"`
		Resolution string `query:"resolution" required:"true"`
	}
	type resolutionOutput struct {
		Body struct {
			ChartID             string `json:"chart_id"`
			RequestedResolution string `json:"requested_resolution,omitempty"`
			CurrentResolution   string `json:"current_resolution"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "get-resolution", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/resolution", Summary: "Get chart resolution", Tags: []string{"Resolution"}},
		func(ctx context.Context, input *chartIDInput) (*resolutionOutput, error) {
			resolution, err := svc.GetResolution(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &resolutionOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.CurrentResolution = resolution
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-resolution", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/resolution", Summary: "Set chart resolution", Tags: []string{"Resolution"}},
		func(ctx context.Context, input *resolutionInput) (*resolutionOutput, error) {
			current, err := svc.SetResolution(ctx, input.ChartID, input.Resolution)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &resolutionOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.RequestedResolution = input.Resolution
			out.Body.CurrentResolution = current
			return out, nil
		})

	type actionInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			ActionID string `json:"action_id,omitempty"`
			Action   string `json:"action,omitempty"`
		}
	}
	type actionOutput struct {
		Body struct {
			ChartID  string `json:"chart_id"`
			ActionID string `json:"action_id"`
			Status   string `json:"status"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "execute-action", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/action", Summary: "Execute chart action", Tags: []string{"Action"}},
		func(ctx context.Context, input *actionInput) (*actionOutput, error) {
			actionID := strings.TrimSpace(input.Body.ActionID)
			if actionID == "" {
				actionID = strings.TrimSpace(input.Body.Action)
			}
			if actionID == "" {
				return nil, huma.Error400BadRequest("action_id is required")
			}

			if err := svc.ExecuteAction(ctx, input.ChartID, actionID); err != nil {
				return nil, mapErr(err)
			}
			out := &actionOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.ActionID = actionID
			out.Body.Status = "executed"
			return out, nil
		})

	type studyListOutput struct {
		Body struct {
			ChartID string             `json:"chart_id"`
			Studies []cdpcontrol.Study `json:"studies"`
		}
	}
	type addStudyInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			Name         string         `json:"name" required:"true"`
			Inputs       map[string]any `json:"inputs,omitempty"`
			ForceOverlay bool           `json:"force_overlay,omitempty"`
		}
	}
	type addStudyOutput struct {
		Body struct {
			ChartID string           `json:"chart_id"`
			Study   cdpcontrol.Study `json:"study"`
			Status  string           `json:"status"`
		}
	}
	type removeStudyInput struct {
		ChartID string `path:"chart_id"`
		StudyID string `path:"study_id"`
	}

	huma.Register(api, huma.Operation{OperationID: "list-studies", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/studies", Summary: "List studies", Tags: []string{"Studies"}},
		func(ctx context.Context, input *chartIDInput) (*studyListOutput, error) {
			studies, err := svc.ListStudies(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &studyListOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Studies = studies
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "add-study", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/studies", Summary: "Add study", Tags: []string{"Studies"}},
		func(ctx context.Context, input *addStudyInput) (*addStudyOutput, error) {
			study, err := svc.AddStudy(ctx, input.ChartID, input.Body.Name, input.Body.Inputs, input.Body.ForceOverlay)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &addStudyOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Study = study
			out.Body.Status = "added"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "remove-study", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/studies/{study_id}", Summary: "Remove study", Tags: []string{"Studies"}},
		func(ctx context.Context, input *removeStudyInput) (*struct{}, error) {
			if err := svc.RemoveStudy(ctx, input.ChartID, input.StudyID); err != nil {
				return nil, mapErr(err)
			}
			return &struct{}{}, nil
		})

	type getStudyOutput struct {
		Body struct {
			ChartID string                 `json:"chart_id"`
			Study   cdpcontrol.StudyDetail `json:"study"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "get-study", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/studies/{study_id}", Summary: "Get study detail with inputs", Tags: []string{"Studies"}},
		func(ctx context.Context, input *removeStudyInput) (*getStudyOutput, error) {
			detail, err := svc.GetStudyInputs(ctx, input.ChartID, input.StudyID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &getStudyOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Study = detail
			return out, nil
		})

	type modifyStudyInput struct {
		ChartID string `path:"chart_id"`
		StudyID string `path:"study_id"`
		Body    struct {
			Inputs map[string]any `json:"inputs" required:"true"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "modify-study", Method: http.MethodPatch, Path: "/api/v1/chart/{chart_id}/studies/{study_id}", Summary: "Modify study input parameters", Tags: []string{"Studies"}},
		func(ctx context.Context, input *modifyStudyInput) (*getStudyOutput, error) {
			detail, err := svc.ModifyStudyInputs(ctx, input.ChartID, input.StudyID, input.Body.Inputs)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &getStudyOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Study = detail
			return out, nil
		})

	return router
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	var coded *cdpcontrol.CodedError
	if errors.As(err, &coded) {
		switch coded.Code {
		case cdpcontrol.CodeValidation:
			return huma.Error400BadRequest(coded.Message)
		case cdpcontrol.CodeChartNotFound:
			return huma.Error404NotFound(coded.Message)
		case cdpcontrol.CodeEvalTimeout:
			return huma.Error504GatewayTimeout(coded.Message)
		case cdpcontrol.CodeAPIUnavailable, cdpcontrol.CodeCDPUnavailable:
			return huma.Error502BadGateway(coded.Message)
		default:
			return huma.Error500InternalServerError(fmt.Sprintf("%s: %s", coded.Code, coded.Message))
		}
	}
	return huma.Error500InternalServerError(err.Error())
}
