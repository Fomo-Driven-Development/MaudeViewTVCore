package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerStudyHandlers(api huma.API, svc Service) {
	type studyListOutput struct {
		Body struct {
			ChartID string             `json:"chart_id"`
			Studies []cdpcontrol.Study `json:"studies"`
		}
	}
	type addStudyInput struct {
		ChartID string `path:"chart_id"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
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
	type studyPathInput struct {
		ChartID string `path:"chart_id"`
		StudyID string `path:"study_id"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
	}

	huma.Register(api, huma.Operation{OperationID: "list-studies", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/studies", Summary: "List studies", Tags: []string{"Studies"}},
		func(ctx context.Context, input *chartIDInput) (*studyListOutput, error) {
			studies, err := svc.ListStudies(ctx, input.ChartID, input.Pane)
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
			study, err := svc.AddStudy(ctx, input.ChartID, input.Body.Name, input.Body.Inputs, input.Body.ForceOverlay, input.Pane)
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
		func(ctx context.Context, input *studyPathInput) (*struct{}, error) {
			if err := svc.RemoveStudy(ctx, input.ChartID, input.StudyID, input.Pane); err != nil {
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
		func(ctx context.Context, input *studyPathInput) (*getStudyOutput, error) {
			detail, err := svc.GetStudyInputs(ctx, input.ChartID, input.StudyID, input.Pane)
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
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
		Body    struct {
			Inputs map[string]any `json:"inputs" required:"true"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "modify-study", Method: http.MethodPatch, Path: "/api/v1/chart/{chart_id}/studies/{study_id}", Summary: "Modify study input parameters", Tags: []string{"Studies"}},
		func(ctx context.Context, input *modifyStudyInput) (*getStudyOutput, error) {
			detail, err := svc.ModifyStudyInputs(ctx, input.ChartID, input.StudyID, input.Body.Inputs, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &getStudyOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Study = detail
			return out, nil
		})

	// --- Compare/Overlay convenience endpoints ---

	type addCompareInput struct {
		ChartID string `path:"chart_id"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
		Body    struct {
			Symbol string `json:"symbol" required:"true"`
			Mode   string `json:"mode,omitempty" doc:"overlay (default) or compare"`
			Source string `json:"source,omitempty" doc:"Price source for compare mode (default: close)"`
		}
	}
	type addCompareOutput struct {
		Body struct {
			ChartID string           `json:"chart_id"`
			Study   cdpcontrol.Study `json:"study"`
			Status  string           `json:"status"`
		}
	}
	type compareListOutput struct {
		Body struct {
			ChartID string             `json:"chart_id"`
			Studies []cdpcontrol.Study `json:"studies"`
		}
	}
	type comparePathInput = studyPathInput

	huma.Register(api, huma.Operation{OperationID: "add-compare", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/compare", Summary: "Add compare/overlay symbol", Tags: []string{"Compare"}},
		func(ctx context.Context, input *addCompareInput) (*addCompareOutput, error) {
			study, err := svc.AddCompare(ctx, input.ChartID, input.Body.Symbol, input.Body.Mode, input.Body.Source, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &addCompareOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Study = study
			out.Body.Status = "added"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "list-compares", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/compare", Summary: "List compare/overlay studies", Tags: []string{"Compare"}},
		func(ctx context.Context, input *chartIDInput) (*compareListOutput, error) {
			studies, err := svc.ListCompares(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &compareListOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Studies = studies
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "remove-compare", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/compare/{study_id}", Summary: "Remove compare/overlay study", Tags: []string{"Compare"}},
		func(ctx context.Context, input *comparePathInput) (*struct{}, error) {
			if err := svc.RemoveStudy(ctx, input.ChartID, input.StudyID, input.Pane); err != nil {
				return nil, mapErr(err)
			}
			return &struct{}{}, nil
		})

	// --- Study Template endpoints ---

	type studyTemplateListOutput struct {
		Body cdpcontrol.StudyTemplateList
	}
	huma.Register(api, huma.Operation{OperationID: "list-study-templates", Method: http.MethodGet, Path: "/api/v1/study-templates", Summary: "List all study templates", Tags: []string{"Studies"}},
		func(ctx context.Context, input *struct{}) (*studyTemplateListOutput, error) {
			list, err := svc.ListStudyTemplates(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &studyTemplateListOutput{}
			out.Body = list
			return out, nil
		})

	type studyTemplateOutput struct {
		Body cdpcontrol.StudyTemplateEntry
	}
	huma.Register(api, huma.Operation{OperationID: "get-study-template", Method: http.MethodGet, Path: "/api/v1/study-templates/{template_id}", Summary: "Get study template detail", Tags: []string{"Studies"}},
		func(ctx context.Context, input *struct {
			TemplateID int `path:"template_id"`
		}) (*studyTemplateOutput, error) {
			entry, err := svc.GetStudyTemplate(ctx, input.TemplateID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &studyTemplateOutput{}
			out.Body = entry
			return out, nil
		})
	type applyStudyTemplateInput struct {
		ChartID string `path:"chart_id"`
		Name    string `query:"name" required:"true" doc:"Template name to apply (case-insensitive match)"`
	}
	type applyStudyTemplateOutput struct {
		Body cdpcontrol.StudyTemplateApplyResult
	}
	huma.Register(api, huma.Operation{OperationID: "apply-study-template", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/study-templates/apply", Summary: "Apply a study template by name", Tags: []string{"Studies"}},
		func(ctx context.Context, input *applyStudyTemplateInput) (*applyStudyTemplateOutput, error) {
			result, err := svc.ApplyStudyTemplate(ctx, input.ChartID, input.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &applyStudyTemplateOutput{}
			out.Body = result
			return out, nil
		})

	type searchIndicatorsInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			Query string `json:"query" required:"true"`
		}
	}
	type indicatorSearchOutput struct {
		Body cdpcontrol.IndicatorSearchResult
	}
	huma.Register(api, huma.Operation{OperationID: "search-indicators", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/indicators/search", Summary: "Search indicators by query", Tags: []string{"Indicators"}},
		func(ctx context.Context, input *searchIndicatorsInput) (*indicatorSearchOutput, error) {
			result, err := svc.SearchIndicators(ctx, input.ChartID, input.Body.Query)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &indicatorSearchOutput{}
			out.Body = result
			return out, nil
		})

	type addIndicatorInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			Query string `json:"query" required:"true"`
			Index int    `json:"index"`
		}
	}
	type indicatorAddOutput struct {
		Body cdpcontrol.IndicatorAddResult
	}
	huma.Register(api, huma.Operation{OperationID: "add-indicator", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/indicators/add", Summary: "Search and add indicator by clicking Nth result", Tags: []string{"Indicators"}},
		func(ctx context.Context, input *addIndicatorInput) (*indicatorAddOutput, error) {
			result, err := svc.AddIndicatorBySearch(ctx, input.ChartID, input.Body.Query, input.Body.Index)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &indicatorAddOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "list-favorite-indicators", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/indicators/favorites", Summary: "List favorite indicators", Tags: []string{"Indicators"}},
		func(ctx context.Context, input *chartIDInput) (*indicatorSearchOutput, error) {
			result, err := svc.ListFavoriteIndicators(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &indicatorSearchOutput{}
			out.Body = result
			return out, nil
		})

	type toggleFavoriteInput struct {
		ChartID string `path:"chart_id"`
		Body    struct {
			Query string `json:"query" required:"true"`
			Index int    `json:"index"`
		}
	}
	type indicatorFavoriteOutput struct {
		Body cdpcontrol.IndicatorFavoriteResult
	}
	huma.Register(api, huma.Operation{OperationID: "toggle-indicator-favorite", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/indicators/favorite", Summary: "Toggle favorite on a search result", Tags: []string{"Indicators"}},
		func(ctx context.Context, input *toggleFavoriteInput) (*indicatorFavoriteOutput, error) {
			result, err := svc.ToggleIndicatorFavorite(ctx, input.ChartID, input.Body.Query, input.Body.Index)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &indicatorFavoriteOutput{}
			out.Body = result
			return out, nil
		})

	type probeIndicatorDOMOutput struct {
		Body map[string]any
	}
	huma.Register(api, huma.Operation{OperationID: "probe-indicator-dom", Method: http.MethodGet, Path: "/api/v1/indicators/probe-dom", Summary: "Probe indicator dialog DOM structure (debug)", Tags: []string{"Indicators"}},
		func(ctx context.Context, input *struct{}) (*probeIndicatorDOMOutput, error) {
			result, err := svc.ProbeIndicatorDialogDOM(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &probeIndicatorDOMOutput{}
			out.Body = result
			return out, nil
		})

}
