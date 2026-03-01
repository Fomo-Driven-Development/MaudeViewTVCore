package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerScreencastHandlers(api huma.API, svc Service) {
	type startBody struct {
		ChartID string `json:"chart_id,omitempty" doc:"Chart tab ID (empty = first available)"`
		cdpcontrol.ScreencastOptions
	}
	type startInput struct {
		Body startBody
	}
	type screencastOutput struct {
		Body cdpcontrol.ScreencastInfo
	}

	huma.Register(api, huma.Operation{
		OperationID: "start-screencast",
		Method:      http.MethodPost,
		Path:        "/api/v1/screencast/start",
		Summary:     "Start a CDP screencast session",
		Tags:        []string{"Screencast"},
	}, func(ctx context.Context, input *startInput) (*screencastOutput, error) {
		info, err := svc.StartScreencast(ctx, input.Body.ChartID, input.Body.ScreencastOptions)
		if err != nil {
			return nil, mapErr(err)
		}
		out := &screencastOutput{}
		out.Body = info
		return out, nil
	})

	type stopInput struct {
		ID string `path:"id"`
	}
	huma.Register(api, huma.Operation{
		OperationID: "stop-screencast",
		Method:      http.MethodPost,
		Path:        "/api/v1/screencast/{id}/stop",
		Summary:     "Stop an active screencast session",
		Tags:        []string{"Screencast"},
	}, func(ctx context.Context, input *stopInput) (*screencastOutput, error) {
		info, err := svc.StopScreencast(ctx, input.ID)
		if err != nil {
			return nil, mapErr(err)
		}
		out := &screencastOutput{}
		out.Body = info
		return out, nil
	})

	type listOutput struct {
		Body []cdpcontrol.ScreencastInfo
	}
	huma.Register(api, huma.Operation{
		OperationID: "list-screencasts",
		Method:      http.MethodGet,
		Path:        "/api/v1/screencast",
		Summary:     "List all screencast sessions",
		Tags:        []string{"Screencast"},
	}, func(ctx context.Context, input *struct{}) (*listOutput, error) {
		sessions, err := svc.ListScreencasts(ctx)
		if err != nil {
			return nil, mapErr(err)
		}
		out := &listOutput{}
		out.Body = sessions
		return out, nil
	})

	type getInput struct {
		ID string `path:"id"`
	}
	huma.Register(api, huma.Operation{
		OperationID: "get-screencast",
		Method:      http.MethodGet,
		Path:        "/api/v1/screencast/{id}",
		Summary:     "Get screencast session info",
		Tags:        []string{"Screencast"},
	}, func(ctx context.Context, input *getInput) (*screencastOutput, error) {
		info, err := svc.GetScreencast(ctx, input.ID)
		if err != nil {
			return nil, mapErr(err)
		}
		out := &screencastOutput{}
		out.Body = info
		return out, nil
	})

	type deleteInput struct {
		ID string `path:"id"`
	}
	huma.Register(api, huma.Operation{
		OperationID: "delete-screencast",
		Method:      http.MethodDelete,
		Path:        "/api/v1/screencast/{id}",
		Summary:     "Delete a screencast session and its frame files",
		Tags:        []string{"Screencast"},
	}, func(ctx context.Context, input *deleteInput) (*struct{}, error) {
		if err := svc.DeleteScreencast(ctx, input.ID); err != nil {
			return nil, mapErr(err)
		}
		return nil, nil
	})
}
