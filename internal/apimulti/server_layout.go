package apimulti

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerLayoutHandlers(api huma.API, svc MultiService) {
	// All layout management endpoints moved under /chart/{chart_id}/ so agents route to a specific window.
	// Chart navigation endpoints (next/prev/maximize/activate) moved under /chart/{chart_id}/pane/.

	type listLayoutsOutput struct {
		Body struct {
			Layouts []cdpcontrol.LayoutInfo `json:"layouts"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-list-layouts", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/layouts", Summary: "List all saved layouts", Tags: []string{"Layout"}},
		func(ctx context.Context, input *chartIDInput) (*listLayoutsOutput, error) {
			layouts, err := svc.ListLayouts(ctx, input.ChartID)
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
	huma.Register(api, huma.Operation{OperationID: "multi-get-layout-status", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/layout/status", Summary: "Get current layout state", Tags: []string{"Layout"}},
		func(ctx context.Context, input *chartIDInput) (*layoutStatusOutput, error) {
			status, err := svc.GetLayoutStatus(ctx, input.ChartID)
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
	huma.Register(api, huma.Operation{OperationID: "multi-switch-layout", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/layout/switch", Summary: "Switch to a layout by numeric ID", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				ID int `json:"id" required:"true"`
			}
		}) (*layoutActionOutput, error) {
			result, err := svc.SwitchLayout(ctx, input.ChartID, input.Body.ID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutActionOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-save-layout", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/layout/save", Summary: "Save current layout", Tags: []string{"Layout"}},
		func(ctx context.Context, input *chartIDInput) (*layoutActionOutput, error) {
			result, err := svc.SaveLayout(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutActionOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-clone-layout", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/layout/clone", Summary: "Clone current layout with a new name", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Name string `json:"name" required:"true"`
			}
		}) (*layoutActionOutput, error) {
			result, err := svc.CloneLayout(ctx, input.ChartID, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutActionOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-delete-layout", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/layout/{layout_id}", Summary: "Delete a layout by ID", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			ChartID  string `path:"chart_id"`
			LayoutID int    `path:"layout_id"`
		}) (*struct{}, error) {
			_, err := svc.DeleteLayout(ctx, input.ChartID, input.LayoutID)
			if err != nil {
				return nil, mapErr(err)
			}
			return nil, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-rename-layout", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/layout/rename", Summary: "Rename current layout", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Name string `json:"name" required:"true"`
			}
		}) (*layoutActionOutput, error) {
			result, err := svc.RenameLayout(ctx, input.ChartID, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutActionOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-set-layout-grid", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/layout/grid", Summary: "Set grid template (e.g. s, 2h, 2v, 3h, 4)", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Template string `json:"template" required:"true"`
			}
		}) (*layoutStatusOutput, error) {
			status, err := svc.SetLayoutGrid(ctx, input.ChartID, input.Body.Template)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutStatusOutput{}
			out.Body = status
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-toggle-fullscreen", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/layout/fullscreen", Summary: "Toggle fullscreen mode", Tags: []string{"Layout"}},
		func(ctx context.Context, input *chartIDInput) (*layoutStatusOutput, error) {
			status, err := svc.ToggleFullscreen(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutStatusOutput{}
			out.Body = status
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-dismiss-dialog", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/layout/dismiss-dialog", Summary: "Dismiss any open modal/dialog (Escape key)", Tags: []string{"Layout"}},
		func(ctx context.Context, input *chartIDInput) (*layoutActionOutput, error) {
			result, err := svc.DismissDialog(ctx, input.ChartID)
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
	huma.Register(api, huma.Operation{OperationID: "multi-batch-delete-layouts", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/layouts/batch-delete", Summary: "Delete multiple layouts in one call", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				IDs        []int `json:"ids" required:"true"`
				SkipActive bool  `json:"skip_active,omitempty"`
			}
		}) (*batchDeleteOutput, error) {
			result, err := svc.BatchDeleteLayouts(ctx, input.ChartID, input.Body.IDs, input.Body.SkipActive)
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
	huma.Register(api, huma.Operation{OperationID: "multi-preview-layout", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/layout/preview", Summary: "Switch to a layout, gather details, and switch back", Tags: []string{"Layout"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				ID           int  `json:"id" required:"true"`
				TakeSnapshot bool `json:"take_snapshot,omitempty"`
			}
		}) (*layoutDetailOutput, error) {
			result, err := svc.PreviewLayout(ctx, input.ChartID, input.Body.ID, input.Body.TakeSnapshot)
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

	huma.Register(api, huma.Operation{OperationID: "multi-get-layout-favorite", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/layout/favorite", Summary: "Get current layout's favorite state", Tags: []string{"Layout"}},
		func(ctx context.Context, input *chartIDInput) (*layoutFavoriteOutput, error) {
			result, err := svc.GetLayoutFavorite(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutFavoriteOutput{}
			out.Body = result
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-toggle-layout-favorite", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/layout/favorite/toggle", Summary: "Toggle star/bookmark on current layout", Tags: []string{"Layout"}},
		func(ctx context.Context, input *chartIDInput) (*layoutFavoriteOutput, error) {
			result, err := svc.ToggleLayoutFavorite(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutFavoriteOutput{}
			out.Body = result
			return out, nil
		})

	// --- Chart pane navigation endpoints (moved under /chart/{chart_id}/pane/) ---

	huma.Register(api, huma.Operation{OperationID: "multi-next-chart", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pane/next", Summary: "Focus next chart pane (Tab)", Tags: []string{"Chart Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*activeChartOutput, error) {
			info, err := svc.NextChart(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &activeChartOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-prev-chart", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pane/prev", Summary: "Focus previous chart pane (Shift+Tab)", Tags: []string{"Chart Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*activeChartOutput, error) {
			info, err := svc.PrevChart(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &activeChartOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-maximize-chart", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pane/maximize", Summary: "Toggle maximize active pane (Alt+Enter)", Tags: []string{"Chart Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*layoutStatusOutput, error) {
			status, err := svc.MaximizeChart(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutStatusOutput{}
			out.Body = status
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-activate-chart", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pane/activate", Summary: "Set active chart by index (0-based)", Tags: []string{"Chart Navigation"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Index int `json:"index"`
			}
		}) (*layoutStatusOutput, error) {
			status, err := svc.ActivateChart(ctx, input.ChartID, input.Body.Index)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &layoutStatusOutput{}
			out.Body = status
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-get-panes", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/panes", Summary: "List all panes in the current grid layout", Tags: []string{"Chart Navigation"}},
		func(ctx context.Context, input *chartIDInput) (*struct{ Body cdpcontrol.PanesResult }, error) {
			result, err := svc.GetPaneInfo(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &struct{ Body cdpcontrol.PanesResult }{}
			out.Body = result
			return out, nil
		})
}
