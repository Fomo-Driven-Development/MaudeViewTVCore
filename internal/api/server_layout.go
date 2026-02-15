package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/tv_agent/internal/cdpcontrol"
)

func registerLayoutHandlers(api huma.API, svc Service) {
	// --- Layout management endpoints ---

	type listLayoutsOutput struct {
		Body struct {
			Layouts []cdpcontrol.LayoutInfo `json:"layouts"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "list-layouts", Method: http.MethodGet, Path: "/api/v1/layouts", Summary: "List all saved layouts", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct{}) (*listLayoutsOutput, error) {
			layouts, err := svc.ListLayouts(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &listLayoutsOutput{}
			out.Body.Layouts = layouts
			return out, nil
		})

	type layoutStatusOutput struct {
		Body cdpcontrol.LayoutStatus
	}
	huma.Register(api, huma.Operation{OperationID: "get-layout-status", Method: http.MethodGet, Path: "/api/v1/layout/status", Summary: "Get current layout state", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct{}) (*layoutStatusOutput, error) {
			status, err := svc.GetLayoutStatus(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutStatusOutput{}
			out.Body = status
			return out, nil
		})

	type layoutActionOutput struct {
		Body cdpcontrol.LayoutActionResult
	}
	huma.Register(api, huma.Operation{OperationID: "switch-layout", Method: http.MethodPost, Path: "/api/v1/layout/switch", Summary: "Switch to a layout by numeric ID", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			Body struct {
				ID int `json:"id" required:"true"`
			}
		}) (*layoutActionOutput, error) {
			result, err := svc.SwitchLayout(ctx, input.Body.ID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutActionOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "save-layout", Method: http.MethodPost, Path: "/api/v1/layout/save", Summary: "Save current layout", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct{}) (*layoutActionOutput, error) {
			result, err := svc.SaveLayout(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutActionOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "clone-layout", Method: http.MethodPost, Path: "/api/v1/layout/clone", Summary: "Clone current layout with a new name", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Name string `json:"name" required:"true"`
			}
		}) (*layoutActionOutput, error) {
			result, err := svc.CloneLayout(ctx, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutActionOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "delete-layout", Method: http.MethodDelete, Path: "/api/v1/layout/{layout_id}", Summary: "Delete a layout by ID", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			LayoutID int `path:"layout_id"`
		}) (*struct{}, error) {
			_, err := svc.DeleteLayout(ctx, input.LayoutID)
			if err != nil {
				return nil, mapErr(err)
			}
			return nil, nil
		})

	huma.Register(api, huma.Operation{OperationID: "rename-layout", Method: http.MethodPost, Path: "/api/v1/layout/rename", Summary: "Rename current layout", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Name string `json:"name" required:"true"`
			}
		}) (*layoutActionOutput, error) {
			result, err := svc.RenameLayout(ctx, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutActionOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-layout-grid", Method: http.MethodPost, Path: "/api/v1/layout/grid", Summary: "Set grid template (e.g. s, 2h, 2v, 3h, 4)", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Template string `json:"template" required:"true"`
			}
		}) (*layoutStatusOutput, error) {
			status, err := svc.SetLayoutGrid(ctx, input.Body.Template)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutStatusOutput{}
			out.Body = status
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "toggle-fullscreen", Method: http.MethodPost, Path: "/api/v1/layout/fullscreen", Summary: "Toggle fullscreen mode", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct{}) (*layoutStatusOutput, error) {
			status, err := svc.ToggleFullscreen(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutStatusOutput{}
			out.Body = status
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "dismiss-dialog", Method: http.MethodPost, Path: "/api/v1/layout/dismiss-dialog", Summary: "Dismiss any open modal/dialog (Escape key)", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct{}) (*layoutActionOutput, error) {
			result, err := svc.DismissDialog(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutActionOutput{}
			out.Body = result
			return out, nil
		})

	type batchDeleteOutput struct {
		Body cdpcontrol.BatchDeleteResult
	}
	huma.Register(api, huma.Operation{OperationID: "batch-delete-layouts", Method: http.MethodPost, Path: "/api/v1/layouts/batch-delete", Summary: "Delete multiple layouts in one call", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			Body struct {
				IDs        []int `json:"ids" required:"true"`
				SkipActive bool  `json:"skip_active,omitempty"`
			}
		}) (*batchDeleteOutput, error) {
			result, err := svc.BatchDeleteLayouts(ctx, input.Body.IDs, input.Body.SkipActive)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &batchDeleteOutput{}
			out.Body = result
			return out, nil
		})

	type layoutDetailOutput struct {
		Body cdpcontrol.LayoutDetail
	}
	huma.Register(api, huma.Operation{OperationID: "preview-layout", Method: http.MethodPost, Path: "/api/v1/layout/preview", Summary: "Switch to a layout, gather details, and switch back", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			Body struct {
				ID           int  `json:"id" required:"true"`
				TakeSnapshot bool `json:"take_snapshot,omitempty"`
			}
		}) (*layoutDetailOutput, error) {
			result, err := svc.PreviewLayout(ctx, input.Body.ID, input.Body.TakeSnapshot)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutDetailOutput{}
			out.Body = result
			return out, nil
		})

	type layoutFavoriteOutput struct {
		Body cdpcontrol.LayoutFavoriteResult
	}

	huma.Register(api, huma.Operation{OperationID: "get-layout-favorite", Method: http.MethodGet, Path: "/api/v1/layout/favorite", Summary: "Get current layout's favorite state", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct{}) (*layoutFavoriteOutput, error) {
			result, err := svc.GetLayoutFavorite(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutFavoriteOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "toggle-layout-favorite", Method: http.MethodPost, Path: "/api/v1/layout/favorite/toggle", Summary: "Toggle star/bookmark on current layout", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct{}) (*layoutFavoriteOutput, error) {
			result, err := svc.ToggleLayoutFavorite(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutFavoriteOutput{}
			out.Body = result
			return out, nil
		})

	// --- Chart navigation endpoints ---

	huma.Register(api, huma.Operation{OperationID: "next-chart", Method: http.MethodPost, Path: "/api/v1/chart/next", Summary: "Focus next chart pane (Tab)", Tags: []string{"Chart Navigation"}},
		func(ctx context.Context, input *struct{}) (*activeChartOutput, error) {
			info, err := svc.NextChart(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &activeChartOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "prev-chart", Method: http.MethodPost, Path: "/api/v1/chart/prev", Summary: "Focus previous chart pane (Shift+Tab)", Tags: []string{"Chart Navigation"}},
		func(ctx context.Context, input *struct{}) (*activeChartOutput, error) {
			info, err := svc.PrevChart(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &activeChartOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "maximize-chart", Method: http.MethodPost, Path: "/api/v1/chart/maximize", Summary: "Toggle maximize active pane (Alt+Enter)", Tags: []string{"Chart Navigation"}},
		func(ctx context.Context, input *struct{}) (*layoutStatusOutput, error) {
			status, err := svc.MaximizeChart(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutStatusOutput{}
			out.Body = status
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "activate-chart", Method: http.MethodPost, Path: "/api/v1/chart/activate", Summary: "Set active chart by index (0-based)", Tags: []string{"Chart Navigation"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Index int `json:"index"`
			}
		}) (*layoutStatusOutput, error) {
			status, err := svc.ActivateChart(ctx, input.Body.Index)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutStatusOutput{}
			out.Body = status
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "get-panes", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/panes", Summary: "List all panes in the current grid layout", Tags: []string{"Chart Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*struct{ Body cdpcontrol.PanesResult }, error) {
			result, err := svc.GetPaneInfo(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct{ Body cdpcontrol.PanesResult }{}
			out.Body = result
			return out, nil
		})

}
