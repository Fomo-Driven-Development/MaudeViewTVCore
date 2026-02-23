package apimulti

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func registerWatchlistHandlers(api huma.API, svc MultiService) {
	// All watchlist endpoints moved under /chart/{chart_id}/ so agents route to a specific window.

	type listWatchlistsOutput struct {
		Body struct {
			Watchlists []cdpcontrol.WatchlistInfo `json:"watchlists"`
		}
	}
	huma.Register(api, huma.Operation{OperationID: "multi-list-watchlists", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/watchlists", Summary: "List all watchlists", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *chartIDInput) (*listWatchlistsOutput, error) {
			wls, err := svc.ListWatchlists(ctx, input.ChartID)
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
	huma.Register(api, huma.Operation{OperationID: "multi-get-active-watchlist", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/watchlists/active", Summary: "Get active watchlist with symbols", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *chartIDInput) (*watchlistDetailOutput, error) {
			detail, err := svc.GetActiveWatchlist(ctx, input.ChartID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	type watchlistInfoOutput struct {
		Body cdpcontrol.WatchlistInfo
	}
	huma.Register(api, huma.Operation{OperationID: "multi-set-active-watchlist", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/watchlists/active", Summary: "Set active watchlist by ID", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				ID string `json:"id" required:"true"`
			}
		}) (*watchlistInfoOutput, error) {
			info, err := svc.SetActiveWatchlist(ctx, input.ChartID, input.Body.ID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-create-watchlist", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/watchlists", Summary: "Create new watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Name string `json:"name" required:"true"`
			}
		}) (*watchlistInfoOutput, error) {
			info, err := svc.CreateWatchlist(ctx, input.ChartID, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	type multiWatchlistIDInput struct {
		ChartID     string `path:"chart_id"`
		WatchlistID string `path:"watchlist_id"`
	}

	huma.Register(api, huma.Operation{OperationID: "multi-get-watchlist", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/watchlist/{watchlist_id}", Summary: "Get watchlist detail with symbols", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *multiWatchlistIDInput) (*watchlistDetailOutput, error) {
			detail, err := svc.GetWatchlist(ctx, input.ChartID, input.WatchlistID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-rename-watchlist", Method: http.MethodPatch, Path: "/api/v1/chart/{chart_id}/watchlist/{watchlist_id}", Summary: "Rename watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			ChartID     string `path:"chart_id"`
			WatchlistID string `path:"watchlist_id"`
			Body        struct {
				Name string `json:"name" required:"true"`
			}
		}) (*watchlistInfoOutput, error) {
			info, err := svc.RenameWatchlist(ctx, input.ChartID, input.WatchlistID, input.Body.Name)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-delete-watchlist", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/watchlist/{watchlist_id}", Summary: "Delete watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *multiWatchlistIDInput) (*struct{}, error) {
			if err := svc.DeleteWatchlist(ctx, input.ChartID, input.WatchlistID); err != nil {
				return nil, mapErr(err)
			}
			return &struct{}{}, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-activate-watchlist", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/watchlist/{watchlist_id}/active", Summary: "Set watchlist as active by path ID", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *multiWatchlistIDInput) (*watchlistInfoOutput, error) {
			info, err := svc.SetActiveWatchlist(ctx, input.ChartID, input.WatchlistID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-pin-watchlist", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/watchlist/{watchlist_id}/pin", Summary: "[EXPERIMENTAL] Toggle watchlist pin/star", Description: "Toggles the watchlist's pinned/favorite state via the favoritesService singleton. Pin state is stored in TradingView's settings system (not a dedicated REST endpoint). Fragile — may break if TradingView refactors the settings module.", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *multiWatchlistIDInput) (*watchlistInfoOutput, error) {
			info, err := svc.PinWatchlist(ctx, input.ChartID, input.WatchlistID)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistInfoOutput{}
			out.Body = info
			return out, nil
		})

	type multiSymbolsBodyInput struct {
		ChartID     string `path:"chart_id"`
		WatchlistID string `path:"watchlist_id"`
		Body        struct {
			Symbols []string `json:"symbols" required:"true"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-add-symbols", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/watchlist/{watchlist_id}/symbols", Summary: "Add symbol(s) to watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *multiSymbolsBodyInput) (*watchlistDetailOutput, error) {
			detail, err := svc.AddWatchlistSymbols(ctx, input.ChartID, input.WatchlistID, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-remove-symbols", Method: http.MethodDelete, Path: "/api/v1/chart/{chart_id}/watchlist/{watchlist_id}/symbols", Summary: "Remove symbol(s) from watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *multiSymbolsBodyInput) (*watchlistDetailOutput, error) {
			detail, err := svc.RemoveWatchlistSymbols(ctx, input.ChartID, input.WatchlistID, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &watchlistDetailOutput{}
			out.Body = detail
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-flag-symbol", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/watchlist/{watchlist_id}/flag", Summary: "[EXPERIMENTAL] Flag/unflag a symbol", Description: "Uses React Fiber internals to find markSymbol(). Fragile — may break on React upgrades.", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			ChartID     string `path:"chart_id"`
			WatchlistID string `path:"watchlist_id"`
			Body        struct {
				Symbol string `json:"symbol" required:"true"`
			}
		}) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			if err := svc.FlagSymbol(ctx, input.ChartID, input.WatchlistID, input.Body.Symbol); err != nil {
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
	huma.Register(api, huma.Operation{OperationID: "multi-list-colored-watchlists", Method: http.MethodGet, Path: "/api/v1/chart/{chart_id}/watchlists/colored", Summary: "List all colored watchlists", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *chartIDInput) (*coloredWatchlistsOutput, error) {
			wls, err := svc.ListColoredWatchlists(ctx, input.ChartID)
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
	type multiColoredInput struct {
		ChartID string `path:"chart_id"`
		Color   string `path:"color"`
		Body    struct {
			Symbols []string `json:"symbols" required:"true"`
		}
	}

	huma.Register(api, huma.Operation{OperationID: "multi-replace-colored-watchlist", Method: http.MethodPut, Path: "/api/v1/chart/{chart_id}/watchlists/colored/{color}", Summary: "Replace entire colored watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *multiColoredInput) (*coloredWatchlistOutput, error) {
			wl, err := svc.ReplaceColoredWatchlist(ctx, input.ChartID, input.Color, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &coloredWatchlistOutput{}
			out.Body = wl
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-append-colored-watchlist", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/watchlists/colored/{color}/append", Summary: "Add symbols to colored watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *multiColoredInput) (*coloredWatchlistOutput, error) {
			wl, err := svc.AppendColoredWatchlist(ctx, input.ChartID, input.Color, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &coloredWatchlistOutput{}
			out.Body = wl
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-remove-colored-watchlist", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/watchlists/colored/{color}/remove", Summary: "Remove symbols from colored watchlist", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *multiColoredInput) (*coloredWatchlistOutput, error) {
			wl, err := svc.RemoveColoredWatchlist(ctx, input.ChartID, input.Color, input.Body.Symbols)
			if err != nil {
				return nil, mapErr(err)
			}
			out := &coloredWatchlistOutput{}
			out.Body = wl
			return out, nil
		})

	huma.Register(api, huma.Operation{OperationID: "multi-bulk-remove-colored-watchlist", Method: http.MethodPost, Path: "/api/v1/chart/{chart_id}/watchlists/colored/bulk-remove", Summary: "Remove symbols from all colored watchlists", Tags: []string{"Watchlists"}},
		func(ctx context.Context, input *struct {
			ChartID string `path:"chart_id"`
			Body    struct {
				Symbols []string `json:"symbols" required:"true"`
			}
		}) (*struct {
			Body struct {
				Status string `json:"status"`
			}
		}, error) {
			if err := svc.BulkRemoveColoredWatchlist(ctx, input.ChartID, input.Body.Symbols); err != nil {
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
