package cdpcontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// rawCDP is a minimal CDP client that evaluates JS on browser targets without
// chromedp's heavy session initialisation (SetAutoAttach, SetDiscoverTargets,
// Page.Enable, DOM.Enable, etc.).  Those commands destabilise some browser
// builds and cause the browser process to exit when service workers are
// auto-attached.
type rawCDP struct {
	httpBase string // e.g. "http://127.0.0.1:9220"

	mu   sync.Mutex
	conn net.Conn
	seq  atomic.Int64

	pending   map[int64]chan json.RawMessage
	pendingMu sync.Mutex

	eventMu       sync.RWMutex
	eventHandlers map[string][]eventHandler
}

type eventHandler struct {
	id int64
	fn func(sessionID string, params json.RawMessage)
}

func newRawCDP(httpBase string) *rawCDP {
	return &rawCDP{
		httpBase:      strings.TrimRight(httpBase, "/"),
		pending:       make(map[int64]chan json.RawMessage),
		eventHandlers: make(map[string][]eventHandler),
	}
}

// connect dials the browser-level WebSocket endpoint.
func (r *rawCDP) connect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn != nil {
		return nil
	}

	wsURL, err := r.browserWSURL(ctx)
	if err != nil {
		return fmt.Errorf("rawcdp: browser ws url: %w", err)
	}

	slog.Debug("rawcdp connecting", "ws_url", wsURL)
	conn, _, _, err := ws.Dial(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("rawcdp: dial: %w", err)
	}

	r.conn = conn
	r.pending = make(map[int64]chan json.RawMessage)
	go r.readLoop()
	return nil
}

func (r *rawCDP) close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn != nil {
		r.conn.Close()
		r.conn = nil
	}
}

// readLoop processes incoming messages and dispatches responses to waiters.
func (r *rawCDP) readLoop() {
	for {
		r.mu.Lock()
		conn := r.conn
		r.mu.Unlock()
		if conn == nil {
			return
		}

		data, err := wsutil.ReadServerText(conn)
		if err != nil {
			slog.Debug("rawcdp read loop exit", "error", err)
			r.closeAllPending()
			return
		}

		var msg struct {
			ID        int64           `json:"id"`
			Method    string          `json:"method"`
			SessionID string          `json:"sessionId"`
			Params    json.RawMessage `json:"params"`
		}
		if json.Unmarshal(data, &msg) != nil {
			continue
		}
		if msg.ID > 0 {
			r.pendingMu.Lock()
			ch, ok := r.pending[msg.ID]
			if ok {
				delete(r.pending, msg.ID)
			}
			r.pendingMu.Unlock()
			if ok {
				ch <- json.RawMessage(data)
			}
		} else if msg.Method != "" {
			r.dispatchEvent(msg.Method, msg.SessionID, msg.Params)
		}
	}
}

func (r *rawCDP) closeAllPending() {
	r.pendingMu.Lock()
	defer r.pendingMu.Unlock()
	for id, ch := range r.pending {
		close(ch)
		delete(r.pending, id)
	}
}

// deletePending removes a pending response channel by ID.
func (r *rawCDP) deletePending(id int64) {
	r.pendingMu.Lock()
	delete(r.pending, id)
	r.pendingMu.Unlock()
}

// sendRaw marshals an envelope, sends it over the WebSocket, and waits for
// the response keyed by the given id.
func (r *rawCDP) sendRaw(ctx context.Context, id int64, envelope any) (json.RawMessage, error) {
	r.mu.Lock()
	conn := r.conn
	r.mu.Unlock()
	if conn == nil {
		return nil, fmt.Errorf("rawcdp: not connected")
	}

	ch := make(chan json.RawMessage, 1)
	r.pendingMu.Lock()
	r.pending[id] = ch
	r.pendingMu.Unlock()

	data, err := json.Marshal(envelope)
	if err != nil {
		r.deletePending(id)
		return nil, fmt.Errorf("rawcdp: marshal: %w", err)
	}

	r.mu.Lock()
	err = wsutil.WriteClientText(conn, data)
	r.mu.Unlock()
	if err != nil {
		r.deletePending(id)
		return nil, fmt.Errorf("rawcdp: send: %w", err)
	}

	select {
	case resp, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("rawcdp: connection closed")
		}
		return resp, nil
	case <-ctx.Done():
		r.deletePending(id)
		return nil, ctx.Err()
	}
}

// send sends a CDP command and waits for the matching response.
func (r *rawCDP) send(ctx context.Context, method string, params any) (json.RawMessage, error) {
	id := r.seq.Add(1)
	req := struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params any    `json:"params,omitempty"`
	}{ID: id, Method: method, Params: params}
	return r.sendRaw(ctx, id, req)
}

// attachToTarget attaches a flat session to the given target.
func (r *rawCDP) attachToTarget(ctx context.Context, targetID string) (string, error) {
	params := struct {
		TargetID string `json:"targetId"`
		Flatten  bool   `json:"flatten"`
	}{TargetID: targetID, Flatten: true}

	raw, err := r.send(ctx, "Target.attachToTarget", params)
	if err != nil {
		return "", err
	}

	var resp struct {
		Result struct {
			SessionID string `json:"sessionId"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", fmt.Errorf("rawcdp: unmarshal attach: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("rawcdp: attach: %s", resp.Error.Message)
	}
	return resp.Result.SessionID, nil
}

// evaluate runs JS on the given session and returns the string result.
func (r *rawCDP) evaluate(ctx context.Context, sessionID, js string) (string, error) {
	params := struct {
		Expression    string `json:"expression"`
		ReturnByValue bool   `json:"returnByValue"`
		AwaitPromise  bool   `json:"awaitPromise"`
	}{Expression: js, ReturnByValue: true, AwaitPromise: true}

	raw, err := r.sendFlat(ctx, sessionID, "Runtime.evaluate", params)
	if err != nil {
		return "", err
	}

	var resp struct {
		Result struct {
			Value json.RawMessage `json:"value"`
			Type  string          `json:"type"`
		} `json:"result"`
		ExceptionDetails *struct {
			Text string `json:"text"`
		} `json:"exceptionDetails"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", fmt.Errorf("rawcdp: unmarshal eval: %w", err)
	}
	if resp.ExceptionDetails != nil {
		return "", fmt.Errorf("rawcdp: eval exception: %s", resp.ExceptionDetails.Text)
	}

	// String results come back as JSON-encoded strings.
	var s string
	if err := json.Unmarshal(resp.Result.Value, &s); err != nil {
		return string(resp.Result.Value), nil
	}
	return s, nil
}

// sendFlat sends a command on a flattened session (sessionId in the outer envelope).
func (r *rawCDP) sendFlat(ctx context.Context, sessionID, method string, params any) (json.RawMessage, error) {
	id := r.seq.Add(1)
	req := struct {
		ID        int64  `json:"id"`
		Method    string `json:"method"`
		SessionID string `json:"sessionId,omitempty"`
		Params    any    `json:"params,omitempty"`
	}{ID: id, Method: method, SessionID: sessionID, Params: params}

	resp, err := r.sendRaw(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Extract the inner "result" field.
	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp, &envelope); err != nil {
		return resp, nil
	}
	if envelope.Error != nil {
		return nil, fmt.Errorf("rawcdp: %s: %s", method, envelope.Error.Message)
	}
	return envelope.Result, nil
}

// detachFromTarget detaches from a session without closing the target.
func (r *rawCDP) detachFromTarget(ctx context.Context, sessionID string) error {
	params := struct {
		SessionID string `json:"sessionId"`
	}{SessionID: sessionID}

	_, err := r.send(ctx, "Target.detachFromTarget", params)
	return err
}

// listTargets fetches open targets via the HTTP /json/list endpoint.
func (r *rawCDP) listTargets(ctx context.Context) ([]*target.Info, error) {
	listCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(listCtx, http.MethodGet, r.httpBase+"/json/list", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rawcdp: /json/list: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var entries []struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}

	out := make([]*target.Info, 0, len(entries))
	for _, e := range entries {
		out = append(out, &target.Info{
			TargetID: target.ID(e.ID),
			Type:     e.Type,
			Title:    e.Title,
			URL:      e.URL,
		})
	}
	return out, nil
}

// dispatchMouseClick sends trusted CDP Input.dispatchMouseEvent commands
// (mousePressed + mouseReleased) at the given coordinates on a session.
// This produces isTrusted=true events, equivalent to real user clicks.
func (r *rawCDP) dispatchMouseClick(ctx context.Context, sessionID string, x, y float64) error {
	pressed := struct {
		Type       string  `json:"type"`
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		Button     string  `json:"button"`
		ClickCount int     `json:"clickCount"`
	}{Type: "mousePressed", X: x, Y: y, Button: "left", ClickCount: 1}

	if _, err := r.sendFlat(ctx, sessionID, "Input.dispatchMouseEvent", pressed); err != nil {
		return fmt.Errorf("rawcdp: mousePressed: %w", err)
	}

	released := struct {
		Type       string  `json:"type"`
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		Button     string  `json:"button"`
		ClickCount int     `json:"clickCount"`
	}{Type: "mouseReleased", X: x, Y: y, Button: "left", ClickCount: 1}

	if _, err := r.sendFlat(ctx, sessionID, "Input.dispatchMouseEvent", released); err != nil {
		return fmt.Errorf("rawcdp: mouseReleased: %w", err)
	}
	return nil
}

// insertText types text into the currently focused element via CDP Input.insertText.
func (r *rawCDP) insertText(ctx context.Context, sessionID, text string) error {
	params := struct {
		Text string `json:"text"`
	}{Text: text}

	if _, err := r.sendFlat(ctx, sessionID, "Input.insertText", params); err != nil {
		return fmt.Errorf("rawcdp: insertText: %w", err)
	}
	return nil
}

// dispatchKeyEvent sends a trusted CDP Input.dispatchKeyEvent sequence
// (keyDown + keyUp) for a keyboard shortcut on a session.
// modifiers is a bitmask: 1=Alt, 2=Ctrl, 4=Meta, 8=Shift.
func (r *rawCDP) dispatchKeyEvent(ctx context.Context, sessionID string, key string, code string, keyCode int, modifiers int) error {
	down := struct {
		Type                  string `json:"type"`
		Key                   string `json:"key"`
		Code                  string `json:"code"`
		WindowsVirtualKeyCode int    `json:"windowsVirtualKeyCode"`
		Modifiers             int    `json:"modifiers"`
	}{Type: "keyDown", Key: key, Code: code, WindowsVirtualKeyCode: keyCode, Modifiers: modifiers}

	if _, err := r.sendFlat(ctx, sessionID, "Input.dispatchKeyEvent", down); err != nil {
		return fmt.Errorf("rawcdp: keyDown: %w", err)
	}

	up := struct {
		Type                  string `json:"type"`
		Key                   string `json:"key"`
		Code                  string `json:"code"`
		WindowsVirtualKeyCode int    `json:"windowsVirtualKeyCode"`
		Modifiers             int    `json:"modifiers"`
	}{Type: "keyUp", Key: key, Code: code, WindowsVirtualKeyCode: keyCode, Modifiers: modifiers}

	if _, err := r.sendFlat(ctx, sessionID, "Input.dispatchKeyEvent", up); err != nil {
		return fmt.Errorf("rawcdp: keyUp: %w", err)
	}
	return nil
}

// dispatchCharInput sends a single character using the rawKeyDown + char + keyUp
// pattern (same as Puppeteer). rawKeyDown fires without text insertion, then the
// "char" event inserts the character and fires native input events that React's
// controlled components respond to.
func (r *rawCDP) dispatchCharInput(ctx context.Context, sessionID, ch string) error {
	// Step 1: rawKeyDown — fires keydown event without inserting text.
	down := struct {
		Type                  string `json:"type"`
		Key                   string `json:"key"`
		WindowsVirtualKeyCode int    `json:"windowsVirtualKeyCode"`
	}{Type: "rawKeyDown", Key: ch}

	if _, err := r.sendFlat(ctx, sessionID, "Input.dispatchKeyEvent", down); err != nil {
		return fmt.Errorf("rawcdp: rawKeyDown: %w", err)
	}

	// Step 2: char — inserts the character and fires native input event.
	charEvt := struct {
		Type           string `json:"type"`
		Text           string `json:"text"`
		Key            string `json:"key"`
		UnmodifiedText string `json:"unmodifiedText"`
	}{Type: "char", Text: ch, Key: ch, UnmodifiedText: ch}

	if _, err := r.sendFlat(ctx, sessionID, "Input.dispatchKeyEvent", charEvt); err != nil {
		return fmt.Errorf("rawcdp: char: %w", err)
	}

	// Step 3: keyUp — completes the key sequence.
	up := struct {
		Type string `json:"type"`
		Key  string `json:"key"`
	}{Type: "keyUp", Key: ch}

	if _, err := r.sendFlat(ctx, sessionID, "Input.dispatchKeyEvent", up); err != nil {
		return fmt.Errorf("rawcdp: charUp: %w", err)
	}
	return nil
}

// registerEventHandler registers a handler for a CDP event method (e.g.
// "Page.javascriptDialogOpening"). Returns an unregister function.
func (r *rawCDP) registerEventHandler(method string, fn func(sessionID string, params json.RawMessage)) func() {
	id := r.seq.Add(1)
	r.eventMu.Lock()
	r.eventHandlers[method] = append(r.eventHandlers[method], eventHandler{id: id, fn: fn})
	r.eventMu.Unlock()
	return func() {
		r.eventMu.Lock()
		defer r.eventMu.Unlock()
		handlers := r.eventHandlers[method]
		for i, h := range handlers {
			if h.id == id {
				r.eventHandlers[method] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}

// dispatchEvent invokes all registered handlers for the given CDP event method.
func (r *rawCDP) dispatchEvent(method, sessionID string, params json.RawMessage) {
	r.eventMu.RLock()
	handlers := make([]eventHandler, len(r.eventHandlers[method]))
	copy(handlers, r.eventHandlers[method])
	r.eventMu.RUnlock()
	for _, h := range handlers {
		h.fn(sessionID, params)
	}
}

// captureScreenshot captures a screenshot of the page via CDP Page.captureScreenshot.
// Returns the raw base64-encoded image data.
func (r *rawCDP) captureScreenshot(ctx context.Context, sessionID, format string, quality int, fullPage bool) (string, error) {
	params := struct {
		Format                string `json:"format"`
		Quality               int    `json:"quality,omitempty"`
		CaptureBeyondViewport bool   `json:"captureBeyondViewport,omitempty"`
		FromSurface           bool   `json:"fromSurface"`
	}{
		Format:                format,
		FromSurface:           true,
		CaptureBeyondViewport: fullPage,
	}
	if format == "jpeg" && quality > 0 {
		params.Quality = quality
	}

	raw, err := r.sendFlat(ctx, sessionID, "Page.captureScreenshot", params)
	if err != nil {
		return "", fmt.Errorf("rawcdp: captureScreenshot: %w", err)
	}

	var resp struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", fmt.Errorf("rawcdp: unmarshal screenshot: %w", err)
	}
	return resp.Data, nil
}

// enablePageDomain sends Page.enable on a flattened session so that dialog
// events (Page.javascriptDialogOpening) are emitted.
func (r *rawCDP) enablePageDomain(ctx context.Context, sessionID string) error {
	_, err := r.sendFlat(ctx, sessionID, "Page.enable", nil)
	return err
}

// handleJavaScriptDialog accepts or dismisses a JavaScript dialog on the session.
func (r *rawCDP) handleJavaScriptDialog(ctx context.Context, sessionID string, accept bool) error {
	params := struct {
		Accept bool `json:"accept"`
	}{Accept: accept}
	_, err := r.sendFlat(ctx, sessionID, "Page.handleJavaScriptDialog", params)
	return err
}

// browserWSURL fetches the WebSocket debugger URL from /json/version.
func (r *rawCDP) browserWSURL(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.httpBase+"/json/version", nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("rawcdp: /json/version: HTTP %d", resp.StatusCode)
	}

	var info struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	if info.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("empty webSocketDebuggerUrl")
	}
	return info.WebSocketDebuggerURL, nil
}
