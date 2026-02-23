package apimulti

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerNoteHandlers(api huma.API, svc MultiService) {
	// All notes endpoints moved under /chart/{chart_id}/ so agents route to a specific window.

	type noteOutput struct {
		Body cdpcontrol.Note
	}

	type listNotesOutput struct {
		Body struct {
			Notes []cdpcontrol.Note `json:"notes"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-list-notes", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/notes", Summary: "List all notes", Tags: []string{"Notes"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Symbol  string `query:"symbol" doc:"Optional symbol filter (e.g. BTCUSD or COINBASE:BTCUSD)"`
		}) (*listNotesOutput, error) {
			notes, err := svc.ListNotes(ctx, input.ChartID, input.Symbol)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &listNotesOutput{}
			out.Body.Notes = notes
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-get-note", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/notes/{note_id}", Summary: "Get single note by ID", Tags: []string{"Notes"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			NoteID  int    `path:"note_id"`
		}) (*noteOutput, error) {
			note, err := svc.GetNote(ctx, input.ChartID, input.NoteID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &noteOutput{}
			out.Body = note
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-create-note", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/notes", Summary: "Create a new note", Tags: []string{"Notes"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				SymbolFull  string `json:"symbol_full" required:"true" doc:"Full symbol (e.g. COINBASE:BTCUSD)"`
				Description string `json:"description" required:"true" doc:"Note text content"`
				Title       string `json:"title,omitempty" doc:"Optional note title"`
				SnapshotUID string `json:"snapshot_uid,omitempty" doc:"Optional TradingView snapshot UID"`
			}
		}) (*noteOutput, error) {
			note, err := svc.CreateNote(ctx, input.ChartID, input.Body.SymbolFull, input.Body.Description, input.Body.Title, input.Body.SnapshotUID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &noteOutput{}
			out.Body = note
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-edit-note", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/notes/{note_id}", Summary: "Edit an existing note", Tags: []string{"Notes"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			NoteID  int    `path:"note_id"`
			Body    struct {
				Description string `json:"description" required:"true" doc:"Updated note text"`
				Title       string `json:"title,omitempty" doc:"Optional updated title"`
				SnapshotUID string `json:"snapshot_uid,omitempty" doc:"Optional updated snapshot UID"`
			}
		}) (*noteOutput, error) {
			note, err := svc.EditNote(ctx, input.ChartID, input.NoteID, input.Body.Description, input.Body.Title, input.Body.SnapshotUID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &noteOutput{}
			out.Body = note
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-delete-note", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/notes/{note_id}", Summary: "Delete a note", Tags: []string{"Notes"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			NoteID  int    `path:"note_id"`
		}) (*struct{}, error) {
			if err := svc.DeleteNote(ctx, input.ChartID, input.NoteID); err != nil {
				return nil, mapErr(err)
			}
			return nil, nil
		})

	type serverSnapshotOutput struct {
		Body cdpcontrol.ServerSnapshotResult
	}

	huma.Register(api, huma.Operation{OperationID: "multi-take-server-snapshot", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/notes/snapshot", Summary: "Upload chart screenshot to TradingView servers", Tags: []string{"Notes"}},
		func(ctx context.Context, input *chartIDInput) (*serverSnapshotOutput, error) {
			result, err := svc.TakeServerSnapshot(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &serverSnapshotOutput{}
			out.Body = result
			return out, nil
		})
}
