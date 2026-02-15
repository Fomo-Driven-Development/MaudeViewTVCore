package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerWatchlistHandlers(api huma.API, svc Service) {
	// --- Watchlist endpoints ---

	type listWatchlistsOutput struct {
		Body struct {
			Watchlists []cdpcontrol.WatchlistInfo `json:"watchlists"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "list-watchlists", Method: http.MethodGet, Path: "/api/v1/watchlists", Summary: "List all watchlists", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct{}) (*listWatchlistsOutput, error) {
			wls, err := svc.ListWatchlists(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &listWatchlistsOutput{}
			out.Body.Watchlists = wls
			return out, nil
		})

	type watchlistDetailOutput struct {
		Body cdpcontrol.WatchlistDetail
	}
	huma.Register(api, huma.Operation{OperationID: "get-active-watchlist", Method: http.MethodGet, Path: "/api/v1/watchlists/active", Summary: "Get active watchlist with symbols", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct{}) (*watchlistDetailOutput, error) {
			detail, err := svc.GetActiveWatchlist(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	type setActiveWatchlistInput struct {
		Body struct {
			ID string `json:"id" required:"true"`
		}
	}
	type watchlistInfoOutput struct {
		Body cdpcontrol.WatchlistInfo
	}
	huma.Register(api, huma.Operation{OperationID: "set-active-watchlist", Method: http.MethodPut, Path: "/api/v1/watchlists/active", Summary: "Set active watchlist by ID", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *setActiveWatchlistInput) (*watchlistInfoOutput, error) {
			info, err := svc.SetActiveWatchlist(ctx, input.Body.ID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	type watchlistIDInput struct {
		WatchlistID string `path:"watchlist_id"`
	}

	huma.Register(api, huma.Operation{OperationID: "create-watchlist", Method: http.MethodPost, Path: "/api/v1/watchlists", Summary: "Create new watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			Body struct {
				Name string `json:"name" required:"true"`
			}
		}) (*watchlistInfoOutput, error) {
			info, err := svc.CreateWatchlist(ctx, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "get-watchlist", Method: http.MethodGet, Path: "/api/v1/watchlist/{watchlist_id}", Summary: "Get watchlist detail with symbols", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *watchlistIDInput) (*watchlistDetailOutput, error) {
			detail, err := svc.GetWatchlist(ctx, input.WatchlistID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "rename-watchlist", Method: http.MethodPatch, Path: "/api/v1/watchlist/{watchlist_id}", Summary: "Rename watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			WatchlistID string `path:"watchlist_id"`
			Body        struct {
				Name string `json:"name" required:"true"`
			}
		}) (*watchlistInfoOutput, error) {
			info, err := svc.RenameWatchlist(ctx, input.WatchlistID, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "delete-watchlist", Method: http.MethodDelete, Path: "/api/v1/watchlist/{watchlist_id}", Summary: "Delete watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *watchlistIDInput) (*struct{}, error) {
			if err := svc.DeleteWatchlist(ctx, input.WatchlistID); err != nil {
				return nil, mapErr(err)
			}
			return &struct{}{}, nil
		})

	type symbolsBodyInput struct {
		WatchlistID string `path:"watchlist_id"`
		Body        struct {
			Symbols []string `json:"symbols" required:"true"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "add-symbols", Method: http.MethodPost, Path: "/api/v1/watchlist/{watchlist_id}/symbols", Summary: "Add symbol(s) to watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *symbolsBodyInput) (*watchlistDetailOutput, error) {
			detail, err := svc.AddWatchlistSymbols(ctx, input.WatchlistID, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "remove-symbols", Method: http.MethodDelete, Path: "/api/v1/watchlist/{watchlist_id}/symbols", Summary: "Remove symbol(s) from watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *symbolsBodyInput) (*watchlistDetailOutput, error) {
			detail, err := svc.RemoveWatchlistSymbols(ctx, input.WatchlistID, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "flag-symbol", Method: http.MethodPost, Path: "/api/v1/watchlist/{watchlist_id}/flag", Summary: "[EXPERIMENTAL] Flag/unflag a symbol", Description: "Uses React Fiber internals to find markSymbol(). Fragile — may break on React upgrades.", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			WatchlistID string `path:"watchlist_id"`
			Body        struct {
				Symbol string `json:"symbol" required:"true"`
			}
		}) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			if err := svc.FlagSymbol(ctx, input.WatchlistID, input.Body.Symbol); err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					Status string `json:"status"`
				}
			}{}
			out.Body.Status = "toggled"
			return out, nil
		})

	// --- Colored Watchlist endpoints ---

	type coloredWatchlistsOutput struct {
		Body struct {
			ColoredWatchlists []cdpcontrol.ColoredWatchlist `json:"colored_watchlists"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "list-colored-watchlists", Method: http.MethodGet, Path: "/api/v1/watchlists/colored", Summary: "List all colored watchlists", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct{}) (*coloredWatchlistsOutput, error) {
			wls, err := svc.ListColoredWatchlists(ctx)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &coloredWatchlistsOutput{}
			out.Body.ColoredWatchlists = wls
			return out, nil
		})

	type coloredWatchlistOutput struct {
		Body cdpcontrol.ColoredWatchlist
	}
	type coloredWatchlistColorInput struct {
		Color string `path:"color"`
		Body  struct {
			Symbols []string `json:"symbols" required:"true"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "replace-colored-watchlist", Method: http.MethodPut, Path: "/api/v1/watchlists/colored/{color}", Summary: "Replace entire colored watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *coloredWatchlistColorInput) (*coloredWatchlistOutput, error) {
			wl, err := svc.ReplaceColoredWatchlist(ctx, input.Color, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &coloredWatchlistOutput{}
			out.Body = wl
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "append-colored-watchlist", Method: http.MethodPost, Path: "/api/v1/watchlists/colored/{color}/append", Summary: "Add symbols to colored watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *coloredWatchlistColorInput) (*coloredWatchlistOutput, error) {
			wl, err := svc.AppendColoredWatchlist(ctx, input.Color, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &coloredWatchlistOutput{}
			out.Body = wl
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "remove-colored-watchlist", Method: http.MethodPost, Path: "/api/v1/watchlists/colored/{color}/remove", Summary: "Remove symbols from colored watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *coloredWatchlistColorInput) (*coloredWatchlistOutput, error) {
			wl, err := svc.RemoveColoredWatchlist(ctx, input.Color, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &coloredWatchlistOutput{}
			out.Body = wl
			return out, nil
		})

	type bulkRemoveColoredInput struct {
		Body struct {
			Symbols []string `json:"symbols" required:"true"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "bulk-remove-colored-watchlist", Method: http.MethodPost, Path: "/api/v1/watchlists/colored/bulk-remove", Summary: "Remove symbols from all colored watchlists", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *bulkRemoveColoredInput) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			if err := svc.BulkRemoveColoredWatchlist(ctx, input.Body.Symbols); err != nil {
				return nil, mapErr(err)
			}
			out := &struct {
				Body struct {
					Status string `json:"status"`
				}
			}{}
			out.Body.Status = "ok"
			return out, nil
		})
}
