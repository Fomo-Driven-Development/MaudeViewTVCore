package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerNoteHandlers(api huma.API, svc Service) {
	type noteOutput struct {
		Body cdpcontrol.Note
	}

	type listNotesOutput struct {
		Body struct {
			Notes []cdpcontrol.Note `json:"notes"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "list-notes", Method: http.MethodGet, Path: "/api/v1/notes", Summary: "List all notes", Tags: []string{"Notes"}},
		func(ctx context.Context, input *struct {
			Symbol string `query:"symbol" doc:"Optional symbol filter (e.g. BTCUSD or COINBASE:BTCUSD)"`
		}) (*listNotesOutput, error) {
			notes, err := svc.ListNotes(ctx, input.Symbol)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &listNotesOutput{}
			out.Body.Notes = notes
			return out, nil
		})

	type noteIDInput struct {
		NoteID int `path:"note_id"`
	}

	huma.Register(api, huma.Operation{OperationID: "get-note", Method: http.MethodGet, Path: "/api/v1/notes/{note_id}", Summary: "Get single note by ID", Tags: []string{"Notes"}},
		func(ctx context.Context, input *noteIDInput) (*noteOutput, error) {
			note, err := svc.GetNote(ctx, input.NoteID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &noteOutput{}
			out.Body = note
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "create-note", Method: http.MethodPost, Path: "/api/v1/notes", Summary: "Create a new note", Tags: []string{"Notes"}},
		func(ctx context.Context, input *struct {
			Body struct {
				SymbolFull  string `json:"symbol_full" required:"true" doc:"Full symbol (e.g. COINBASE:BTCUSD)"`
				Description string `json:"description" required:"true" doc:"Note text content"`
				Title       string `json:"title,omitempty" doc:"Optional note title"`
				SnapshotUID string `json:"snapshot_uid,omitempty" doc:"Optional TradingView snapshot UID"`
			}
		}) (*noteOutput, error) {
			note, err := svc.CreateNote(ctx, input.Body.SymbolFull, input.Body.Description, input.Body.Title, input.Body.SnapshotUID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &noteOutput{}
			out.Body = note
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "edit-note", Method: http.MethodPut, Path: "/api/v1/notes/{note_id}", Summary: "Edit an existing note", Tags: []string{"Notes"}},
		func(ctx context.Context, input *struct {
			NoteID int `path:"note_id"`
			Body   struct {
				Description string `json:"description" required:"true" doc:"Updated note text"`
				Title       string `json:"title,omitempty" doc:"Optional updated title"`
				SnapshotUID string `json:"snapshot_uid,omitempty" doc:"Optional updated snapshot UID"`
			}
		}) (*noteOutput, error) {
			note, err := svc.EditNote(ctx, input.NoteID, input.Body.Description, input.Body.Title, input.Body.SnapshotUID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &noteOutput{}
			out.Body = note
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "delete-note", Method: http.MethodDelete, Path: "/api/v1/notes/{note_id}", Summary: "Delete a note", Tags: []string{"Notes"}},
		func(ctx context.Context, input *noteIDInput) (*struct{}, error) {
			if err := svc.DeleteNote(ctx, input.NoteID); err != nil {
				return nil, mapErr(err)
			}
			return nil, nil
		})

	type serverSnapshotOutput struct {
		Body cdpcontrol.ServerSnapshotResult
	}

	huma.Register(api, huma.Operation{OperationID: "take-server-snapshot", Method: http.MethodPost, Path: "/api/v1/notes/snapshot", Summary: "Upload chart screenshot to TradingView servers", Tags: []string{"Notes"}},
		func(ctx context.Context, input *struct{}) (*serverSnapshotOutput, error) {
			result, err := svc.TakeServerSnapshot(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &serverSnapshotOutput{}
			out.Body = result
			return out, nil
		})
}
