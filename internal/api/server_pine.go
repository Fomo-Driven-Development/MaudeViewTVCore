package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/tv_agent/internal/cdpcontrol"
)

func registerPineHandlers(api huma.API, svc Service) {
	// --- Pine Editor endpoints (DOM-based) ---

	type pineStateOutput struct {
		Body cdpcontrol.PineState
	}
	huma.Register(api, huma.Operation{OperationID: "toggle-pine-editor", Method: http.MethodPost, Path: "/api/v1/pine/toggle", Summary: "Toggle Pine editor open/close via DOM click", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.TogglePineEditor(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "get-pine-status", Method: http.MethodGet, Path: "/api/v1/pine/status", Summary: "Check if Pine editor is visible and Monaco is ready", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.GetPineStatus(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "get-pine-source", Method: http.MethodGet, Path: "/api/v1/pine/source", Summary: "Read source from Monaco editor", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.GetPineSource(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-pine-source", Method: http.MethodPut, Path: "/api/v1/pine/source", Summary: "Write source to Monaco editor", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Source string `json:"source" required:"true"`
			}
		}) (*pineStateOutput, error) {
			state, err := svc.SetPineSource(ctx, input.Body.Source)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "save-pine-script", Method: http.MethodPost, Path: "/api/v1/pine/save", Summary: "Click save button in Pine editor toolbar", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.SavePineScript(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "add-pine-to-chart", Method: http.MethodPost, Path: "/api/v1/pine/add-to-chart", Summary: "Click 'Add to chart' button in Pine editor", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.AddPineToChart(ctx)
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
	huma.Register(api, huma.Operation{OperationID: "get-pine-console", Method: http.MethodGet, Path: "/api/v1/pine/console", Summary: "Read Pine console messages from DOM", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineConsoleOutput, error) {
			msgs, err := svc.GetPineConsole(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineConsoleOutput{}
			out.Body.Messages = msgs
			return out, nil
		})

	// --- Pine Editor keyboard shortcut endpoints ---

	huma.Register(api, huma.Operation{OperationID: "pine-undo", Method: http.MethodPost, Path: "/api/v1/pine/undo", Summary: "Undo last edit (Ctrl+Z)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.PineUndo(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-redo", Method: http.MethodPost, Path: "/api/v1/pine/redo", Summary: "Redo last edit (Ctrl+Shift+Z)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.PineRedo(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-new-indicator", Method: http.MethodPost, Path: "/api/v1/pine/new-indicator", Summary: "Open new indicator script (Ctrl+K Ctrl+I)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.PineNewIndicator(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-new-strategy", Method: http.MethodPost, Path: "/api/v1/pine/new-strategy", Summary: "Open new strategy script (Ctrl+K Ctrl+S)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.PineNewStrategy(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-open-script", Method: http.MethodPost, Path: "/api/v1/pine/open-script", Summary: "Open script by name (Ctrl+O then type name)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Name string `json:"name" required:"true"`
			}
		}) (*pineStateOutput, error) {
			state, err := svc.PineOpenScript(ctx, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-find-replace", Method: http.MethodPost, Path: "/api/v1/pine/find-replace", Summary: "Find and replace text in editor via Monaco API", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Find    string `json:"find" required:"true"`
				Replace string `json:"replace" required:"true"`
			}
		}) (*pineStateOutput, error) {
			state, err := svc.PineFindReplace(ctx, input.Body.Find, input.Body.Replace)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-go-to-line", Method: http.MethodPost, Path: "/api/v1/pine/go-to-line", Summary: "Go to a specific line number (Ctrl+G)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Line int `json:"line" required:"true" minimum:"1"`
			}
		}) (*pineStateOutput, error) {
			state, err := svc.PineGoToLine(ctx, input.Body.Line)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-delete-line", Method: http.MethodPost, Path: "/api/v1/pine/delete-line", Summary: "Delete current line (Ctrl+Shift+K)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Count int `json:"count,omitempty" minimum:"1"`
			}
		}) (*pineStateOutput, error) {
			count := input.Body.Count
			if count == 0 {
				count = 1
			}
			state, err := svc.PineDeleteLine(ctx, count)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-move-line", Method: http.MethodPost, Path: "/api/v1/pine/move-line", Summary: "Move current line up or down (Alt+Arrow)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Direction string `json:"direction" required:"true" enum:"up,down"`
				Count     int    `json:"count,omitempty" minimum:"1"`
			}
		}) (*pineStateOutput, error) {
			count := input.Body.Count
			if count == 0 {
				count = 1
			}
			state, err := svc.PineMoveLine(ctx, input.Body.Direction, count)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-toggle-comment", Method: http.MethodPost, Path: "/api/v1/pine/toggle-comment", Summary: "Toggle line comment (Ctrl+/)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.PineToggleComment(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-toggle-console", Method: http.MethodPost, Path: "/api/v1/pine/toggle-console", Summary: "Toggle Pine console panel (Ctrl+`)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.PineToggleConsole(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-insert-line", Method: http.MethodPost, Path: "/api/v1/pine/insert-line", Summary: "Insert blank line above (Ctrl+Shift+Enter)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.PineInsertLineAbove(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-new-tab", Method: http.MethodPost, Path: "/api/v1/pine/new-tab", Summary: "Open new editor tab (Shift+Alt+T)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.PineNewTab(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "pine-command-palette", Method: http.MethodPost, Path: "/api/v1/pine/command-palette", Summary: "Open command palette (F1)", Tags: []string{"Pine Editor"}},
		func(ctx context.Context, input *struct{}) (*pineStateOutput, error) {
			state, err := svc.PineCommandPalette(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &pineStateOutput{}
			out.Body = state
			return out, nil
		})

}
