package cdpcontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/chromedp/cdproto/target"
)

// --- Chart-targeted CDP input helpers ---

// resolveSession resolves a CDP session for a specific chart tab,
// mirroring resolveAnySession but for an explicit chartID.
func (c *Client) resolveSession(ctx context.Context, chartID string) (*rawCDP, string, error) {
	session, info, err := c.resolveChartSession(ctx, chartID)
	if err != nil {
		return nil, "", err
	}
	c.mu.Lock()
	cdp := c.cdp
	c.mu.Unlock()
	if cdp == nil {
		return nil, "", newError(CodeCDPUnavailable, "CDP client not connected", nil)
	}
	sessionID, err := c.ensureSession(ctx, cdp, session, info.TargetID)
	if err != nil {
		return nil, "", err
	}
	return cdp, sessionID, nil
}

// clickOnChart dispatches a trusted CDP mouse click on a specific chart tab.
func (c *Client) clickOnChart(ctx context.Context, chartID string, x, y float64) error {
	cdp, sessionID, err := c.resolveSession(ctx, chartID)
	if err != nil {
		return err
	}
	if err := cdp.dispatchMouseClick(ctx, sessionID, x, y); err != nil {
		return newError(CodeEvalFailure, "failed to dispatch trusted mouse click", err)
	}
	return nil
}

// sendKeysOnChart dispatches a trusted CDP key event on a specific chart tab.
// modifiers is a bitmask: 1=Alt, 2=Ctrl, 4=Meta, 8=Shift.
func (c *Client) sendKeysOnChart(ctx context.Context, chartID, key, code string, keyCode, modifiers int) error {
	cdp, sessionID, err := c.resolveSession(ctx, chartID)
	if err != nil {
		return err
	}
	if err := cdp.dispatchKeyEvent(ctx, sessionID, key, code, keyCode, modifiers); err != nil {
		return newError(CodeEvalFailure, "failed to dispatch trusted key event", err)
	}
	return nil
}

// insertTextOnChart types text into the focused element on a specific chart tab.
func (c *Client) insertTextOnChart(ctx context.Context, chartID, text string) error {
	cdp, sessionID, err := c.resolveSession(ctx, chartID)
	if err != nil {
		return err
	}
	if err := cdp.insertText(ctx, sessionID, text); err != nil {
		return newError(CodeEvalFailure, "failed to dispatch trusted text insertion", err)
	}
	return nil
}

// sendShortcutOnChart dispatches a key event + settle wait on a specific chart tab.
func (c *Client) sendShortcutOnChart(ctx context.Context, chartID, key, code string, keyCode, modifiers int, settle time.Duration, desc string) error {
	if err := c.sendKeysOnChart(ctx, chartID, key, code, keyCode, modifiers); err != nil {
		return newError(CodeEvalFailure, desc, err)
	}
	return sendShortcutWait(ctx, settle)
}

// pineKeyActionOnChart focuses the Monaco editor, dispatches a key combo on a specific chart tab.
func (c *Client) pineKeyActionOnChart(ctx context.Context, chartID, key, code string, keyCode, modifiers, repeat int, waitJS string) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	for i := 0; i < repeat; i++ {
		if err := c.sendKeysOnChart(ctx, chartID, key, code, keyCode, modifiers); err != nil {
			return PineState{}, newError(CodeEvalFailure, "failed to dispatch key", err)
		}
	}
	if err := c.evalOnChart(ctx, chartID, waitJS, &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

// invalidateChartSession zeroes the session ID for a specific chartID so the
// next eval will re-attach. Used after navigation events.
func (c *Client) invalidateChartSession(chartID string) {
	c.mu.Lock()
	targetID, ok := c.chartToTarget[chartID]
	c.mu.Unlock()
	if !ok {
		return
	}
	c.mu.Lock()
	ts, ok := c.tabs[target.ID(targetID)]
	c.mu.Unlock()
	if ok {
		ts.mu.Lock()
		ts.sessionID = ""
		ts.mu.Unlock()
	}
}

// --- Watchlist *OnChart methods ---

func (c *Client) ListWatchlistsOnChart(ctx context.Context, chartID string) ([]WatchlistInfo, error) {
	var out struct {
		Watchlists []WatchlistInfo `json:"watchlists"`
	}
	if err := c.evalOnChart(ctx, chartID, jsListWatchlists(), &out); err != nil {
		return nil, err
	}
	if out.Watchlists == nil {
		return []WatchlistInfo{}, nil
	}
	return out.Watchlists, nil
}

func (c *Client) GetActiveWatchlistOnChart(ctx context.Context, chartID string) (WatchlistDetail, error) {
	var out WatchlistDetail
	if err := c.evalOnChart(ctx, chartID, jsGetActiveWatchlist(), &out); err != nil {
		return WatchlistDetail{}, err
	}
	return out, nil
}

func (c *Client) SetActiveWatchlistOnChart(ctx context.Context, chartID, id string) (WatchlistInfo, error) {
	var out WatchlistInfo
	if err := c.evalOnChart(ctx, chartID, jsSetActiveWatchlist(id), &out); err != nil {
		return WatchlistInfo{}, err
	}
	return out, nil
}

func (c *Client) GetWatchlistOnChart(ctx context.Context, chartID, id string) (WatchlistDetail, error) {
	var out WatchlistDetail
	if err := c.evalOnChart(ctx, chartID, jsGetWatchlist(id), &out); err != nil {
		return WatchlistDetail{}, err
	}
	return out, nil
}

func (c *Client) CreateWatchlistOnChart(ctx context.Context, chartID, name string) (WatchlistInfo, error) {
	var out WatchlistInfo
	if err := c.evalOnChart(ctx, chartID, jsCreateWatchlist(name), &out); err != nil {
		return WatchlistInfo{}, err
	}
	return out, nil
}

func (c *Client) RenameWatchlistOnChart(ctx context.Context, chartID, id, name string) (WatchlistInfo, error) {
	var out WatchlistInfo
	if err := c.evalOnChart(ctx, chartID, jsRenameWatchlist(id, name), &out); err != nil {
		return WatchlistInfo{}, err
	}
	return out, nil
}

func (c *Client) AddWatchlistSymbolsOnChart(ctx context.Context, chartID, id string, symbols []string) (WatchlistDetail, error) {
	var out WatchlistDetail
	if err := c.evalOnChart(ctx, chartID, jsAddWatchlistSymbols(id, symbols), &out); err != nil {
		return WatchlistDetail{}, err
	}
	return out, nil
}

func (c *Client) RemoveWatchlistSymbolsOnChart(ctx context.Context, chartID, id string, symbols []string) (WatchlistDetail, error) {
	var out WatchlistDetail
	if err := c.evalOnChart(ctx, chartID, jsRemoveWatchlistSymbols(id, symbols), &out); err != nil {
		return WatchlistDetail{}, err
	}
	return out, nil
}

func (c *Client) PinWatchlistOnChart(ctx context.Context, chartID, id string) (WatchlistInfo, error) {
	var out WatchlistInfo
	if err := c.evalOnChart(ctx, chartID, jsPinWatchlist(id), &out); err != nil {
		return WatchlistInfo{}, err
	}
	return out, nil
}

func (c *Client) FlagSymbolOnChart(ctx context.Context, chartID, id, symbol string) error {
	return c.doChartAction(ctx, chartID, jsFlagSymbol(id, symbol))
}

func (c *Client) DeleteWatchlistOnChart(ctx context.Context, chartID, id string) error {
	return c.doChartAction(ctx, chartID, jsDeleteWatchlist(id))
}

// --- Colored Watchlist *OnChart methods ---

func (c *Client) ListColoredWatchlistsOnChart(ctx context.Context, chartID string) ([]ColoredWatchlist, error) {
	var out struct {
		ColoredWatchlists []ColoredWatchlist `json:"colored_watchlists"`
	}
	if err := c.evalOnChart(ctx, chartID, jsListColoredWatchlists(), &out); err != nil {
		return nil, err
	}
	if out.ColoredWatchlists == nil {
		return []ColoredWatchlist{}, nil
	}
	return out.ColoredWatchlists, nil
}

func (c *Client) ReplaceColoredWatchlistOnChart(ctx context.Context, chartID, color string, symbols []string) (ColoredWatchlist, error) {
	var out ColoredWatchlist
	if err := c.evalOnChart(ctx, chartID, jsReplaceColoredWatchlist(color, symbols), &out); err != nil {
		return ColoredWatchlist{}, err
	}
	return out, nil
}

func (c *Client) AppendColoredWatchlistOnChart(ctx context.Context, chartID, color string, symbols []string) (ColoredWatchlist, error) {
	var out ColoredWatchlist
	if err := c.evalOnChart(ctx, chartID, jsAppendColoredWatchlist(color, symbols), &out); err != nil {
		return ColoredWatchlist{}, err
	}
	return out, nil
}

func (c *Client) RemoveColoredWatchlistOnChart(ctx context.Context, chartID, color string, symbols []string) (ColoredWatchlist, error) {
	var out ColoredWatchlist
	if err := c.evalOnChart(ctx, chartID, jsRemoveColoredWatchlist(color, symbols), &out); err != nil {
		return ColoredWatchlist{}, err
	}
	return out, nil
}

func (c *Client) BulkRemoveColoredWatchlistOnChart(ctx context.Context, chartID string, symbols []string) error {
	return c.doChartAction(ctx, chartID, jsBulkRemoveColoredWatchlist(symbols))
}

// --- Notes *OnChart methods ---

func (c *Client) ListNotesOnChart(ctx context.Context, chartID, symbol string) ([]Note, error) {
	var out struct {
		Notes []Note `json:"notes"`
	}
	if err := c.evalOnChart(ctx, chartID, jsListNotes(symbol), &out); err != nil {
		return nil, err
	}
	if out.Notes == nil {
		return []Note{}, nil
	}
	return out.Notes, nil
}

func (c *Client) CreateNoteOnChart(ctx context.Context, chartID, symbolFull, description, title, snapshotUID string) (Note, error) {
	var out Note
	if err := c.evalOnChart(ctx, chartID, jsCreateNote(symbolFull, description, title, snapshotUID), &out); err != nil {
		return Note{}, err
	}
	return out, nil
}

func (c *Client) EditNoteOnChart(ctx context.Context, chartID string, noteID int, description, title, snapshotUID string) (Note, error) {
	var out Note
	if err := c.evalOnChart(ctx, chartID, jsEditNote(noteID, description, title, snapshotUID), &out); err != nil {
		return Note{}, err
	}
	return out, nil
}

func (c *Client) DeleteNoteOnChart(ctx context.Context, chartID string, noteID int) error {
	return c.doChartAction(ctx, chartID, jsDeleteNote(noteID))
}

func (c *Client) TakeServerSnapshotOnChart(ctx context.Context, chartID string) (ServerSnapshotResult, error) {
	snapCtx, cancel := context.WithTimeout(ctx, 3*c.evalTimeout)
	defer cancel()

	var out ServerSnapshotResult
	if err := c.evalOnChart(snapCtx, chartID, jsTakeServerSnapshot(), &out); err != nil {
		return ServerSnapshotResult{}, err
	}
	return out, nil
}

// --- Alerts *OnChart methods ---

func (c *Client) ListAlertsOnChart(ctx context.Context, chartID string) (any, error) {
	var out struct {
		Alerts any `json:"alerts"`
	}
	if err := c.evalOnChart(ctx, chartID, jsListAlerts(), &out); err != nil {
		return nil, err
	}
	return out.Alerts, nil
}

func (c *Client) GetAlertsOnChart(ctx context.Context, chartID string, ids []string) (any, error) {
	var out struct {
		Alerts any `json:"alerts"`
	}
	if err := c.evalOnChart(ctx, chartID, jsGetAlerts(ids), &out); err != nil {
		return nil, err
	}
	return out.Alerts, nil
}

func (c *Client) CreateAlertOnChart(ctx context.Context, chartID string, params map[string]any) (any, error) {
	var out struct {
		Alert any `json:"alert"`
	}
	if err := c.evalOnChart(ctx, chartID, jsCreateAlert(params), &out); err != nil {
		return nil, err
	}
	return out.Alert, nil
}

func (c *Client) ModifyAlertOnChart(ctx context.Context, chartID string, params map[string]any) (any, error) {
	var out struct {
		Alert any `json:"alert"`
	}
	if err := c.evalOnChart(ctx, chartID, jsModifyAlert(params), &out); err != nil {
		return nil, err
	}
	return out.Alert, nil
}

func (c *Client) DeleteAlertsOnChart(ctx context.Context, chartID string, ids []string) error {
	return c.doChartAction(ctx, chartID, jsDeleteAlerts(ids))
}

func (c *Client) StopAlertsOnChart(ctx context.Context, chartID string, ids []string) error {
	return c.doChartAction(ctx, chartID, jsStopAlerts(ids))
}

func (c *Client) RestartAlertsOnChart(ctx context.Context, chartID string, ids []string) error {
	return c.doChartAction(ctx, chartID, jsRestartAlerts(ids))
}

func (c *Client) CloneAlertsOnChart(ctx context.Context, chartID string, ids []string) error {
	return c.doChartAction(ctx, chartID, jsCloneAlerts(ids))
}

func (c *Client) ListFiresOnChart(ctx context.Context, chartID string) (any, error) {
	var out struct {
		Fires any `json:"fires"`
	}
	if err := c.evalOnChart(ctx, chartID, jsListFires(), &out); err != nil {
		return nil, err
	}
	return out.Fires, nil
}

func (c *Client) DeleteFiresOnChart(ctx context.Context, chartID string, ids []string) error {
	return c.doChartAction(ctx, chartID, jsDeleteFires(ids))
}

func (c *Client) DeleteAllFiresOnChart(ctx context.Context, chartID string) error {
	return c.doChartAction(ctx, chartID, jsDeleteAllFires())
}

// --- Pine Editor *OnChart methods ---

func (c *Client) TogglePineEditorOnChart(ctx context.Context, chartID string) (PineState, error) {
	var loc struct {
		Action string  `json:"action"`
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
	}
	if err := c.evalOnChart(ctx, chartID, jsPineLocateToggleBtn(), &loc); err != nil {
		return PineState{}, err
	}

	if err := c.clickOnChart(ctx, chartID, loc.X, loc.Y); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch trusted click", err)
	}

	var out PineState
	if loc.Action == "close" {
		if err := c.evalOnChart(ctx, chartID, jsPineWaitForClose(), &out); err != nil {
			return PineState{}, err
		}
	} else {
		if err := c.evalOnChart(ctx, chartID, jsPineWaitForOpen(), &out); err != nil {
			return PineState{}, err
		}
	}
	return out, nil
}

func (c *Client) GetPineStatusOnChart(ctx context.Context, chartID string) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineStatus(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) GetPineSourceOnChart(ctx context.Context, chartID string) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineGetSource(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) SetPineSourceOnChart(ctx context.Context, chartID, source string) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineSetSource(source), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) SavePineScriptOnChart(ctx context.Context, chartID string) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	if err := c.sendKeysOnChart(ctx, chartID, "s", "KeyS", 83, 2); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch Ctrl+S", err)
	}
	if err := c.evalOnChart(ctx, chartID, jsPineWaitForSave(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) AddPineToChartOnChart(ctx context.Context, chartID string) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	if err := c.sendKeysOnChart(ctx, chartID, "Enter", "Enter", 13, 2); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to dispatch Ctrl+Enter", err)
	}
	if err := c.evalOnChart(ctx, chartID, jsPineWaitForAddToChart(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) GetPineConsoleOnChart(ctx context.Context, chartID string) ([]PineConsoleMessage, error) {
	var out struct {
		Messages []PineConsoleMessage `json:"messages"`
	}
	if err := c.evalOnChart(ctx, chartID, jsPineGetConsole(), &out); err != nil {
		return nil, err
	}
	if out.Messages == nil {
		return []PineConsoleMessage{}, nil
	}
	return out.Messages, nil
}

func (c *Client) PineFindReplaceOnChart(ctx context.Context, chartID, find, replace string) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineFindReplace(find, replace), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

// --- Pine keyboard shortcut *OnChart methods ---

func (c *Client) PineUndoOnChart(ctx context.Context, chartID string) (PineState, error) {
	return c.pineKeyActionOnChart(ctx, chartID, "z", "KeyZ", 90, 2, 1, jsPineBriefWait(300))
}

func (c *Client) PineRedoOnChart(ctx context.Context, chartID string) (PineState, error) {
	return c.pineKeyActionOnChart(ctx, chartID, "Z", "KeyZ", 90, 10, 1, jsPineBriefWait(300))
}

func (c *Client) PineDeleteLineOnChart(ctx context.Context, chartID string, count int) (PineState, error) {
	if count < 1 {
		count = 1
	}
	return c.pineKeyActionOnChart(ctx, chartID, "K", "KeyK", 75, 10, count, jsPineBriefWait(300))
}

func (c *Client) PineMoveLineOnChart(ctx context.Context, chartID, direction string, count int) (PineState, error) {
	if count < 1 {
		count = 1
	}
	var key, code string
	var keyCode int
	if direction == "up" {
		key, code, keyCode = "ArrowUp", "ArrowUp", 38
	} else {
		key, code, keyCode = "ArrowDown", "ArrowDown", 40
	}
	return c.pineKeyActionOnChart(ctx, chartID, key, code, keyCode, 1, count, jsPineBriefWait(300))
}

func (c *Client) PineToggleCommentOnChart(ctx context.Context, chartID string) (PineState, error) {
	return c.pineKeyActionOnChart(ctx, chartID, "/", "Slash", 191, 2, 1, jsPineBriefWait(300))
}

func (c *Client) PineToggleConsoleOnChart(ctx context.Context, chartID string) (PineState, error) {
	return c.pineKeyActionOnChart(ctx, chartID, "`", "Backquote", 192, 2, 1, jsPineBriefWait(500))
}

func (c *Client) PineInsertLineAboveOnChart(ctx context.Context, chartID string) (PineState, error) {
	return c.pineKeyActionOnChart(ctx, chartID, "Enter", "Enter", 13, 10, 1, jsPineBriefWait(300))
}

func (c *Client) PineNewTabOnChart(ctx context.Context, chartID string) (PineState, error) {
	return c.pineKeyActionOnChart(ctx, chartID, "T", "KeyT", 84, 9, 1, jsPineBriefWait(500))
}

func (c *Client) PineCommandPaletteOnChart(ctx context.Context, chartID string) (PineState, error) {
	return c.pineKeyActionOnChart(ctx, chartID, "F1", "F1", 112, 0, 1, jsPineBriefWait(500))
}

func (c *Client) PineNewIndicatorOnChart(ctx context.Context, chartID string) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	if err := c.sendShortcutOnChart(ctx, chartID, "k", "KeyK", 75, 2, uiSettleShort, "failed to dispatch Ctrl+K"); err != nil {
		return PineState{}, err
	}
	if err := c.sendShortcutOnChart(ctx, chartID, "i", "KeyI", 73, 2, uiSettleShort, "failed to dispatch Ctrl+I"); err != nil {
		return PineState{}, err
	}
	if err := c.evalOnChart(ctx, chartID, jsPineWaitForNewScript(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) PineNewStrategyOnChart(ctx context.Context, chartID string) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	if err := c.sendShortcutOnChart(ctx, chartID, "k", "KeyK", 75, 2, uiSettleShort, "failed to dispatch Ctrl+K"); err != nil {
		return PineState{}, err
	}
	if err := c.sendShortcutOnChart(ctx, chartID, "s", "KeyS", 83, 2, uiSettleShort, "failed to dispatch Ctrl+S"); err != nil {
		return PineState{}, err
	}
	if err := c.evalOnChart(ctx, chartID, jsPineWaitForNewScript(), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) PineGoToLineOnChart(ctx context.Context, chartID string, line int) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	if err := c.sendShortcutOnChart(ctx, chartID, "g", "KeyG", 71, 2, uiSettleMedium, "failed to dispatch Ctrl+G"); err != nil {
		return PineState{}, err
	}
	if err := c.insertTextOnChart(ctx, chartID, fmt.Sprintf("%d", line)); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to type line number", err)
	}
	time.Sleep(uiSettleShort)
	if err := c.sendShortcutOnChart(ctx, chartID, "Enter", "Enter", 13, 0, 0, "failed to confirm go-to-line"); err != nil {
		return PineState{}, err
	}
	if err := c.evalOnChart(ctx, chartID, jsPineBriefWait(300), &out); err != nil {
		return PineState{}, err
	}
	return out, nil
}

func (c *Client) PineOpenScriptOnChart(ctx context.Context, chartID, name string) (PineState, error) {
	var out PineState
	if err := c.evalOnChart(ctx, chartID, jsPineFocusEditor(), &out); err != nil {
		return PineState{}, err
	}
	if err := c.sendShortcutOnChart(ctx, chartID, "o", "KeyO", 79, 2, uiSettleLong, "failed to dispatch Ctrl+O"); err != nil {
		return PineState{}, err
	}
	if err := c.insertTextOnChart(ctx, chartID, name); err != nil {
		return PineState{}, newError(CodeEvalFailure, "failed to type script name", err)
	}
	time.Sleep(uiSettleLong)
	var clickResult struct {
		PineState
		CloseX float64 `json:"close_x"`
		CloseY float64 `json:"close_y"`
	}
	if err := c.evalOnChart(ctx, chartID, jsPineClickFirstScriptResult(), &clickResult); err != nil {
		return PineState{}, err
	}
	out = clickResult.PineState
	if clickResult.CloseX > 0 && clickResult.CloseY > 0 {
		if err := c.clickOnChart(ctx, chartID, clickResult.CloseX, clickResult.CloseY); err != nil {
			slog.Debug("indicator dialog close click failed", "error", err)
		}
		time.Sleep(uiSettleMedium)
	}
	return out, nil
}

// --- Layout *OnChart methods ---

func (c *Client) ListLayoutsOnChart(ctx context.Context, chartID string) ([]LayoutInfo, error) {
	var out struct {
		Layouts []LayoutInfo `json:"layouts"`
	}
	if err := c.evalOnChart(ctx, chartID, jsListLayouts(), &out); err != nil {
		return nil, err
	}
	if out.Layouts == nil {
		return []LayoutInfo{}, nil
	}
	return out.Layouts, nil
}

func (c *Client) GetLayoutFavoriteOnChart(ctx context.Context, chartID string) (LayoutFavoriteResult, error) {
	var out LayoutFavoriteResult
	if err := c.evalOnChart(ctx, chartID, jsGetLayoutFavorite(), &out); err != nil {
		return LayoutFavoriteResult{}, err
	}
	return out, nil
}

func (c *Client) ToggleLayoutFavoriteOnChart(ctx context.Context, chartID string) (LayoutFavoriteResult, error) {
	var out LayoutFavoriteResult
	if err := c.evalOnChart(ctx, chartID, jsToggleLayoutFavorite(), &out); err != nil {
		return LayoutFavoriteResult{}, err
	}
	return out, nil
}

func (c *Client) GetLayoutStatusOnChart(ctx context.Context, chartID string) (LayoutStatus, error) {
	var out LayoutStatus
	if err := c.evalOnChart(ctx, chartID, jsLayoutStatus(), &out); err != nil {
		return LayoutStatus{}, err
	}
	return out, nil
}

func (c *Client) SwitchLayoutOnChart(ctx context.Context, chartID string, id int) (LayoutActionResult, error) {
	// Step 1: Resolve the short URL for the target layout ID.
	var resolved struct {
		ShortURL string `json:"short_url"`
		Name     string `json:"name"`
	}
	if err := c.evalOnChart(ctx, chartID, jsSwitchLayoutResolveURL(id), &resolved); err != nil {
		return LayoutActionResult{}, err
	}
	if resolved.ShortURL == "" {
		return LayoutActionResult{}, newError(CodeValidation, fmt.Sprintf("layout %d not found or has no URL", id), nil)
	}

	// Step 1b: Suppress beforeunload handlers.
	if err := c.evalOnChart(ctx, chartID, jsSuppressBeforeunload(), &struct{}{}); err != nil {
		slog.Debug("beforeunload suppression eval failed", "error", err)
	}

	// Step 1c: Enable Page domain and register auto-accept handler.
	cdpConn, sessionID, resolveErr := c.resolveSession(ctx, chartID)
	var unregister func()
	if resolveErr == nil {
		if err := cdpConn.enablePageDomain(ctx, sessionID); err != nil {
			slog.Debug("enable page domain failed", "error", err)
		}
		sid := sessionID
		unregister = cdpConn.registerEventHandler("Page.javascriptDialogOpening", func(evtSessionID string, params json.RawMessage) {
			acceptCtx, acceptCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer acceptCancel()
			if err := cdpConn.handleJavaScriptDialog(acceptCtx, sid, true); err != nil {
				slog.Debug("auto-accept beforeunload dialog failed", "error", err)
			}
		})
	}

	// Step 2: Navigate via window.location.
	navJS := wrapJSEval(fmt.Sprintf(`window.location.href = "/chart/%s/"; return JSON.stringify({ok:true,data:{}});`, resolved.ShortURL))
	if err := c.evalOnChart(ctx, chartID, navJS, &struct{}{}); err != nil {
		slog.Debug("layout navigation eval expected failure", "error", err)
	}

	// Step 3: Invalidate only this chart's session.
	c.invalidateChartSession(chartID)

	pollCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() {
		if unregister != nil {
			unregister()
		}
	}()
	time.Sleep(2 * time.Second)

	for {
		select {
		case <-pollCtx.Done():
			return LayoutActionResult{}, newError(CodeEvalTimeout, "timed out waiting for layout switch", pollCtx.Err())
		default:
		}
		if err := c.refreshTabs(ctx); err != nil {
			slog.Debug("refresh tabs during layout switch failed", "error", err)
		}
		var readyOut struct{ Ready string }
		readyErr := c.evalOnChart(pollCtx, chartID, wrapJSEval(`return JSON.stringify({ok:true,data:{ready:document.readyState}});`), &readyOut)
		if readyErr == nil && readyOut.Ready == "complete" {
			break
		}
		time.Sleep(uiSettleLong)
	}

	status, statusErr := c.GetLayoutStatusOnChart(ctx, chartID)
	if statusErr != nil {
		return LayoutActionResult{Status: "switched", LayoutName: resolved.Name, LayoutID: resolved.ShortURL}, nil
	}
	return LayoutActionResult{Status: "switched", LayoutName: status.LayoutName, LayoutID: status.LayoutID}, nil
}

func (c *Client) SaveLayoutOnChart(ctx context.Context, chartID string) (LayoutActionResult, error) {
	var out LayoutActionResult
	if err := c.evalOnChart(ctx, chartID, jsSaveLayout(), &out); err != nil {
		return LayoutActionResult{}, err
	}
	return out, nil
}

func (c *Client) CloneLayoutOnChart(ctx context.Context, chartID, name string) (LayoutActionResult, error) {
	var out LayoutActionResult
	if err := c.evalOnChart(ctx, chartID, jsCloneLayout(name), &out); err != nil {
		return LayoutActionResult{}, err
	}
	return out, nil
}

func (c *Client) DeleteLayoutOnChart(ctx context.Context, chartID string, id int) (LayoutActionResult, error) {
	var out LayoutActionResult
	if err := c.evalOnChart(ctx, chartID, jsDeleteLayout(id), &out); err != nil {
		return LayoutActionResult{}, err
	}
	return out, nil
}

func (c *Client) RenameLayoutOnChart(ctx context.Context, chartID, name string) (LayoutActionResult, error) {
	var out LayoutActionResult
	if err := c.evalOnChart(ctx, chartID, jsRenameLayout(name), &out); err != nil {
		return LayoutActionResult{}, err
	}
	return out, nil
}

func (c *Client) SetLayoutGridOnChart(ctx context.Context, chartID, template string) (LayoutStatus, error) {
	var out LayoutStatus
	if err := c.evalOnChart(ctx, chartID, jsSetGrid(template), &out); err != nil {
		return LayoutStatus{}, err
	}
	return out, nil
}

func (c *Client) GetActiveChartOnChart(ctx context.Context, chartID string) (ActiveChartInfo, error) {
	charts, err := c.ListCharts(ctx)
	if err != nil {
		return ActiveChartInfo{}, err
	}
	var out struct {
		ChartIndex int `json:"chart_index"`
		ChartCount int `json:"chart_count"`
	}
	if err := c.evalOnChart(ctx, chartID, jsGetActiveChart(), &out); err != nil {
		// Fallback: find ChartInfo by chartID
		for _, ch := range charts {
			if ch.ChartID == chartID {
				return ActiveChartInfo{ChartID: ch.ChartID, TargetID: ch.TargetID, URL: ch.URL, Title: ch.Title}, nil
			}
		}
		return ActiveChartInfo{}, err
	}
	for _, ch := range charts {
		if ch.ChartID == chartID {
			return ActiveChartInfo{
				ChartID:    ch.ChartID,
				TargetID:   ch.TargetID,
				URL:        ch.URL,
				Title:      ch.Title,
				ChartIndex: out.ChartIndex,
				ChartCount: out.ChartCount,
			}, nil
		}
	}
	return ActiveChartInfo{ChartIndex: out.ChartIndex, ChartCount: out.ChartCount}, nil
}

func (c *Client) NextChartOnChart(ctx context.Context, chartID string) (ActiveChartInfo, error) {
	if err := c.sendShortcutOnChart(ctx, chartID, "Tab", "Tab", 9, 0, uiSettleMedium, "failed to dispatch Tab key"); err != nil {
		return ActiveChartInfo{}, err
	}
	return c.GetActiveChartOnChart(ctx, chartID)
}

func (c *Client) PrevChartOnChart(ctx context.Context, chartID string) (ActiveChartInfo, error) {
	if err := c.sendShortcutOnChart(ctx, chartID, "Tab", "Tab", 9, 8, uiSettleMedium, "failed to dispatch Shift+Tab key"); err != nil {
		return ActiveChartInfo{}, err
	}
	return c.GetActiveChartOnChart(ctx, chartID)
}

func (c *Client) MaximizeChartOnChart(ctx context.Context, chartID string) (LayoutStatus, error) {
	if err := c.sendShortcutOnChart(ctx, chartID, "Enter", "Enter", 13, 1, uiSettleMedium, "failed to dispatch Alt+Enter key"); err != nil {
		return LayoutStatus{}, err
	}
	return c.GetLayoutStatusOnChart(ctx, chartID)
}

func (c *Client) ActivateChartOnChart(ctx context.Context, chartID string, index int) (LayoutStatus, error) {
	var out LayoutStatus
	if err := c.evalOnChart(ctx, chartID, jsActivateChart(index), &out); err != nil {
		return LayoutStatus{}, err
	}
	return out, nil
}

func (c *Client) GetPaneInfoOnChart(ctx context.Context, chartID string) (PanesResult, error) {
	var out PanesResult
	if err := c.evalOnChart(ctx, chartID, jsGetPaneInfo(), &out); err != nil {
		return PanesResult{}, err
	}
	if out.Panes == nil {
		out.Panes = []PaneInfo{}
	}
	return out, nil
}

func (c *Client) ToggleFullscreenOnChart(ctx context.Context, chartID string) (LayoutStatus, error) {
	var out LayoutStatus
	if err := c.evalOnChart(ctx, chartID, jsToggleFullscreen(), &out); err != nil {
		return LayoutStatus{}, err
	}
	return out, nil
}

func (c *Client) DismissDialogOnChart(ctx context.Context, chartID string) (LayoutActionResult, error) {
	if err := c.sendKeysOnChart(ctx, chartID, "Escape", "Escape", 27, 0); err != nil {
		return LayoutActionResult{}, newError(CodeEvalFailure, "failed to dispatch Escape key", err)
	}
	return LayoutActionResult{Status: "dismissed"}, nil
}

// --- Study Template *OnChart methods ---

func (c *Client) ListStudyTemplatesOnChart(ctx context.Context, chartID string) (StudyTemplateList, error) {
	var out StudyTemplateList
	if err := c.evalOnChart(ctx, chartID, jsListStudyTemplates(), &out); err != nil {
		return StudyTemplateList{}, err
	}
	return out, nil
}

func (c *Client) GetStudyTemplateOnChart(ctx context.Context, chartID string, id int) (StudyTemplateEntry, error) {
	var out StudyTemplateEntry
	if err := c.evalOnChart(ctx, chartID, jsGetStudyTemplate(id), &out); err != nil {
		return StudyTemplateEntry{}, err
	}
	return out, nil
}

// --- Hotlists *OnChart methods ---

func (c *Client) ProbeHotlistsManagerOnChart(ctx context.Context, chartID string) (HotlistsManagerProbe, error) {
	var out HotlistsManagerProbe
	if err := c.evalOnChart(ctx, chartID, jsProbeHotlistsManager(), &out); err != nil {
		return HotlistsManagerProbe{}, err
	}
	initProbeDefaults(&out.AccessPaths, &out.Methods, &out.State)
	return out, nil
}

func (c *Client) ProbeHotlistsManagerDeepOnChart(ctx context.Context, chartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsProbeHotlistsManagerDeep(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetHotlistMarketsOnChart(ctx context.Context, chartID string) (any, error) {
	var out struct {
		Markets any `json:"markets"`
	}
	if err := c.evalOnChart(ctx, chartID, jsGetHotlistMarkets(), &out); err != nil {
		return nil, err
	}
	return out.Markets, nil
}

func (c *Client) GetHotlistExchangesOnChart(ctx context.Context, chartID string) ([]HotlistExchangeDetail, error) {
	var out struct {
		Exchanges []HotlistExchangeDetail `json:"exchanges"`
	}
	if err := c.evalOnChart(ctx, chartID, jsGetHotlistExchanges(), &out); err != nil {
		return nil, err
	}
	if out.Exchanges == nil {
		return []HotlistExchangeDetail{}, nil
	}
	return out.Exchanges, nil
}

func (c *Client) GetOneHotlistOnChart(ctx context.Context, chartID, exchange, group string) (HotlistResult, error) {
	var out HotlistResult
	if err := c.evalOnChart(ctx, chartID, jsGetOneHotlist(exchange, group), &out); err != nil {
		return HotlistResult{}, err
	}
	if out.Symbols == nil {
		out.Symbols = []HotlistSymbol{}
	}
	return out, nil
}

// --- Indicator Dialog *OnChart method ---

func (c *Client) ProbeIndicatorDialogDOMOnChart(ctx context.Context, chartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.evalOnChart(ctx, chartID, jsProbeIndicatorDialogDOM(), &out); err != nil {
		return nil, err
	}
	return out, nil
}
