package apimulti

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerPineHandlers(api huma.API, svc MultiService) {
	// All Pine Editor endpoints moved under /chart/{chart_id}/pine/ so agents route to a specific window.

	type pineStateOutput struct {
		Body cdpcontrol.PineState
	}
	huma.Register(api, huma.Operation{OperationID: "multi-toggle-pine-editor", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/toggle", Summary: "Toggle Pine editor open/close via DOM click", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.TogglePineEditor(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-get-pine-status", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/pine/status", Summary: "Check if Pine editor is visible and Monaco is ready", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.GetPineStatus(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-get-pine-source", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/pine/source", Summary: "Read source from Monaco editor", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.GetPineSource(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-set-pine-source", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/pine/source", Summary: "Write source to Monaco editor", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Source string `json:"source" required:"true"`
			}
		}) (*pineStateOutput, error) {
			state, err := svc.SetPineSource(ctx, input.ChartID, input.Body.Source)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-save-pine-script", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/save", Summary: "Click save button in Pine editor toolbar", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.SavePineScript(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-add-pine-to-chart", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/add-to-chart", Summary: "Click 'Add to chart' button in Pine editor", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.AddPineToChart(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	type pineConsoleOutput struct {
		Body struct {
			Messages []cdpcontrol.PineConsoleMessage `json:"messages"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-get-pine-console", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/pine/console", Summary: "Read Pine console messages from DOM", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineConsoleOutput, error) {
			msgs, err := svc.GetPineConsole(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineConsoleOutput{}
			out.Body.Messages = msgs
			return out, nil
		})

	// --- Pine Editor keyboard shortcut endpoints ---

	huma.Register(api, huma.Operation{OperationID: "multi-pine-undo", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/undo", Summary: "Undo last edit (Ctrl+Z)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.PineUndo(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-redo", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/redo", Summary: "Redo last edit (Ctrl+Shift+Z)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.PineRedo(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-new-indicator", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/new-indicator", Summary: "Open new indicator script (Ctrl+K Ctrl+I)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.PineNewIndicator(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-new-strategy", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/new-strategy", Summary: "Open new strategy script (Ctrl+K Ctrl+S)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.PineNewStrategy(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-open-script", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/open-script", Summary: "Open script by name (Ctrl+O then type name)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Name string `json:"name" required:"true"`
			}
		}) (*pineStateOutput, error) {
			state, err := svc.PineOpenScript(ctx, input.ChartID, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-find-replace", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/find-replace", Summary: "Find and replace text in editor via Monaco API", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Find    string `json:"find" required:"true"`
				Replace string `json:"replace" required:"true"`
			}
		}) (*pineStateOutput, error) {
			state, err := svc.PineFindReplace(ctx, input.ChartID, input.Body.Find, input.Body.Replace)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-go-to-line", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/go-to-line", Summary: "Go to a specific line number (Ctrl+G)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Line int `json:"line" required:"true" minimum:"1"`
			}
		}) (*pineStateOutput, error) {
			state, err := svc.PineGoToLine(ctx, input.ChartID, input.Body.Line)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-delete-line", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/delete-line", Summary: "Delete current line (Ctrl+Shift+K)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Count int `json:"count,omitempty" minimum:"1"`
			}
		}) (*pineStateOutput, error) {
			count := input.Body.Count
			if count == 0 {
				count = 1
			}
			state, err := svc.PineDeleteLine(ctx, input.ChartID, count)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-move-line", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/move-line", Summary: "Move current line up or down (Alt+Arrow)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Direction string `json:"direction" required:"true" enum:"up,down"`
				Count     int    `json:"count,omitempty" minimum:"1"`
			}
		}) (*pineStateOutput, error) {
			count := input.Body.Count
			if count == 0 {
				count = 1
			}
			state, err := svc.PineMoveLine(ctx, input.ChartID, input.Body.Direction, count)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-toggle-comment", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/toggle-comment", Summary: "Toggle line comment (Ctrl+/)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.PineToggleComment(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-toggle-console", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/toggle-console", Summary: "Toggle Pine console panel (Ctrl+`)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.PineToggleConsole(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-insert-line", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/insert-line", Summary: "Insert blank line above (Ctrl+Shift+Enter)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.PineInsertLineAbove(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-new-tab", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/new-tab", Summary: "Open new editor tab (Shift+Alt+T)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.PineNewTab(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pine-command-palette", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/pine/command-palette", Summary: "Open command palette (F1)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *chartIDInput) (*pineStateOutput, error) {
			state, err := svc.PineCommandPalette(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})
}
