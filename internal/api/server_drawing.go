package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerDrawingHandlers(api huma.API, svc Service) {
	// --- Drawing/Shape endpoints ---

	type shapeGroupsOutput struct {
		Body struct {
			Groups []cdpcontrol.ShapeGroup `json:"groups"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "list-drawing-shapes", Method: http.MethodGet, Path: "/api/v1/drawings/shapes", Summary: "List known drawing shapes grouped by category", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct{}) (*shapeGroupsOutput, error) {
			out := &shapeGroupsOutput{}
			out.Body.Groups = cdpcontrol.ShapeGroups
			return out, nil
		})

	type drawingListOutput struct {
		Body struct {
			ChartID string             `json:"chart_id"`
			Shapes  []cdpcontrol.Shape `json:"shapes"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "list-drawings", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/drawings", Summary: "List all drawings on chart", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *chartIDInput) (*drawingListOutput, error) {
			shapes, err := svc.ListDrawings(ctx, input.ChartID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &drawingListOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Shapes = shapes
			return out, nil
		})

	type shapeIDInput struct {
		ChartID string `path:"chart_id"`
		ShapeID string `path:"shape_id"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
	}
	type drawingDetailOutput struct {
		Body map[string]any
	}
	huma.Register(api, huma.Operation{OperationID: "get-drawing", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/drawings/{shape_id}", Summary: "Get drawing by ID", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *shapeIDInput) (*drawingDetailOutput, error) {
			detail, err := svc.GetDrawing(ctx, input.ChartID, input.ShapeID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &drawingDetailOutput{}
			out.Body = detail
			return out, nil
		})

	type createDrawingInput struct {
		ChartID string `path:"chart_id"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
		Body    struct {
			Point   cdpcontrol.ShapePoint `json:"point" required:"true"`
			Options map[string]any        `json:"options" required:"true"`
		}
	}
	type createDrawingOutput struct {
		Body struct {
			ChartID string `json:"chart_id"`
			ID      string `json:"id"`
			Status  string `json:"status"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "create-drawing", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/drawings", Summary: "Create a single-point drawing", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *createDrawingInput) (*createDrawingOutput, error) {
			id, err := svc.CreateDrawing(ctx, input.ChartID, input.Body.Point, input.Body.Options, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createDrawingOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.ID = id
			out.Body.Status = "created"
			return out, nil
		})

	type createMultipointInput struct {
		ChartID string `path:"chart_id"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
		Body    struct {
			Points  []cdpcontrol.ShapePoint `json:"points" required:"true"`
			Options map[string]any          `json:"options" required:"true"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "create-multipoint-drawing", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/drawings/multipoint", Summary: "Create a multi-point drawing", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *createMultipointInput) (*createDrawingOutput, error) {
			id, err := svc.CreateMultipointDrawing(ctx, input.ChartID, input.Body.Points, input.Body.Options, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createDrawingOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.ID = id
			out.Body.Status = "created"
			return out, nil
		})

	type createTweetInput struct {
		ChartID string `path:"chart_id"`
		Pane    int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
		Body    struct {
			TweetURL string `json:"tweet_url" required:"true" doc:"Full URL to the tweet (twitter.com or x.com)"`
		}
	}
	type createTweetOutput struct {
		Body struct {
			ChartID string `json:"chart_id"`
			ID      string `json:"id"`
			Status  string `json:"status"`
			URL     string `json:"url"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "create-tweet-drawing", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/drawings/tweet", Summary: "Create a tweet drawing from a URL", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *createTweetInput) (*createTweetOutput, error) {
			result, err := svc.CreateTweetDrawing(ctx, input.ChartID, input.Body.TweetURL, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createTweetOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.ID = result.ID
			out.Body.Status = result.Status
			out.Body.URL = result.URL
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "clone-drawing", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/drawings/{shape_id}/clone", Summary: "Clone a drawing", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *shapeIDInput) (*createDrawingOutput, error) {
			id, err := svc.CloneDrawing(ctx, input.ChartID, input.ShapeID, input.Pane)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &createDrawingOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.ID = id
			out.Body.Status = "cloned"
			return out, nil
		})

	type removeDrawingInput struct {
		ChartID     string `path:"chart_id"`
		ShapeID     string `path:"shape_id"`
		DisableUndo bool   `query:"disable_undo"`
		Pane        int    `query:"pane" default:"-1" doc:"Target pane index (0-based). Omit to use active pane."`
	}
	huma.Register(api, huma.Operation{OperationID: "remove-drawing", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/drawings/{shape_id}", Summary: "Remove a drawing", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *removeDrawingInput) (*struct{}, error) {
			if err := svc.RemoveDrawing(ctx, input.ChartID, input.ShapeID, input.DisableUndo, input.Pane); err != nil {
				return nil, mapErr(err)
			}
			return &struct{}{}, nil
		})

	huma.Register(api, huma.Operation{OperationID: "remove-all-drawings", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/drawings", Summary: "Remove all drawings", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *chartIDInput) (*struct{}, error) {
			if err := svc.RemoveAllDrawings(ctx, input.ChartID, input.Pane); err != nil {
				return nil, mapErr(err)
			}
			return &struct{}{}, nil
		})

	type drawingTogglesOutput struct {
		Body struct {
			ChartID string                    `json:"chart_id"`
			Toggles cdpcontrol.DrawingToggles `json:"toggles"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "get-drawing-toggles", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/drawings/toggles", Summary: "Get drawing toggle states", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *chartIDInput) (*drawingTogglesOutput, error) {
			toggles, err := svc.GetDrawingToggles(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &drawingTogglesOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Toggles = toggles
			return out, nil
		})

	type drawingStatusOutput struct {
		Body struct {
			ChartID string `json:"chart_id"`
			Status  string `json:"status"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "set-hide-drawings", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/toggles/hide", Summary: "Hide or show all drawings", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Value bool `json:"value"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.SetHideDrawings(ctx, input.ChartID, input.Body.Value); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "set"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-lock-drawings", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/toggles/lock", Summary: "Lock or unlock all drawings", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Value bool `json:"value"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.SetLockDrawings(ctx, input.ChartID, input.Body.Value); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "set"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-magnet", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/toggles/magnet", Summary: "Set magnet mode", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Enabled bool `json:"enabled"`
				Mode    int  `json:"mode,omitempty"`
			}
		}) (*drawingStatusOutput, error) {
			mode := input.Body.Mode
			if mode == 0 && !input.Body.Enabled {
				mode = -1
			}
			if err := svc.SetMagnet(ctx, input.ChartID, input.Body.Enabled, mode); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "set"
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-drawing-visibility", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/{shape_id}/visibility", Summary: "Set drawing visibility", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			ShapeID string `path:"shape_id"`
			Body    struct {
				Visible bool `json:"visible"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.SetDrawingVisibility(ctx, input.ChartID, input.ShapeID, input.Body.Visible); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "set"
			return out, nil
		})

	type drawingToolOutput struct {
		Body struct {
			ChartID string `json:"chart_id"`
			Tool    string `json:"tool"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "get-drawing-tool", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/drawings/tool", Summary: "Get selected drawing tool", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *chartIDInput) (*drawingToolOutput, error) {
			tool, err := svc.GetDrawingTool(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &drawingToolOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Tool = tool
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "set-drawing-tool", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/tool", Summary: "Set drawing tool", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Tool string `json:"tool" required:"true"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.SetDrawingTool(ctx, input.ChartID, input.Body.Tool); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "set"
			return out, nil
		})

	// --- Tool activation endpoints (thin wrappers around SetDrawingTool) ---

	for _, tool := range []struct {
		id, path, name, summary string
	}{
		{"activate-measure-tool", "measure", "measure", "Activate measure tool"},
		{"activate-zoom-tool", "zoom", "zoom", "Activate zoom-in selection tool"},
		{"activate-eraser-tool", "eraser", "eraser", "Activate eraser tool"},
		{"activate-cursor-tool", "cursor", "cursor", "Return to default cursor"},
	} {
		huma.Register(api, huma.Operation{
			OperationID: tool.id,
			Method:      http.MethodPost,
			Path:        "/api/v1/chart/{chart_id}/tools/" + tool.path,
			Summary:     tool.summary,
			Tags:        []string{"Tools"},
		}, func(ctx context.Context, input *chartIDInput) (*drawingStatusOutput, error) {
			if err := svc.SetDrawingTool(ctx, input.ChartID, tool.name); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "activated"
			return out, nil
		})
	}

	huma.Register(api, huma.Operation{OperationID: "set-drawing-z-order", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/drawings/{shape_id}/z-order", Summary: "Change drawing z-order", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			ShapeID string `path:"shape_id"`
			Body    struct {
				Action string `json:"action" required:"true" doc:"bring_forward, bring_to_front, send_backward, send_to_back"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.SetDrawingZOrder(ctx, input.ChartID, input.ShapeID, input.Body.Action); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "executed"
			return out, nil
		})

	type drawingsStateOutput struct {
		Body struct {
			ChartID string `json:"chart_id"`
			State   any    `json:"state"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "export-drawings-state", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/drawings/state", Summary: "Export all drawings state", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *chartIDInput) (*drawingsStateOutput, error) {
			state, err := svc.ExportDrawingsState(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &drawingsStateOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.State = state
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "import-drawings-state", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/drawings/state", Summary: "Import drawings state", Tags: []string{"Drawings"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				State any `json:"state" required:"true"`
			}
		}) (*drawingStatusOutput, error) {
			if err := svc.ImportDrawingsState(ctx, input.ChartID, input.Body.State); err != nil {
				return nil, mapErr(err)
			}
			out := &drawingStatusOutput{}
			out.Body.ChartID = input.ChartID
			out.Body.Status = "imported"
			return out, nil
		})

}
