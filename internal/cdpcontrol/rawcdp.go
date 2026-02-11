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
}

func newRawCDP(httpBase string) *rawCDP {
	return &rawCDP{
		httpBase: strings.TrimRight(httpBase, "/"),
		pending:  make(map[int64]chan json.RawMessage),
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
			ID int64 `json:"id"`
		}
		if json.Unmarshal(data, &msg) == nil && msg.ID > 0 {
			r.pendingMu.Lock()
			ch, ok := r.pending[msg.ID]
			if ok {
				delete(r.pending, msg.ID)
			}
			r.pendingMu.Unlock()
			if ok {
				ch <- json.RawMessage(data)
			}
		}
		// Events (no id) are ignored — we don't need them.
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

// send sends a CDP command and waits for the matching response.
func (r *rawCDP) send(ctx context.Context, method string, params any) (json.RawMessage, error) {
	r.mu.Lock()
	conn := r.conn
	r.mu.Unlock()
	if conn == nil {
		return nil, fmt.Errorf("rawcdp: not connected")
	}

	id := r.seq.Add(1)

	req := struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params any    `json:"params,omitempty"`
	}{ID: id, Method: method, Params: params}

	ch := make(chan json.RawMessage, 1)
	r.pendingMu.Lock()
	r.pending[id] = ch
	r.pendingMu.Unlock()

	data, err := json.Marshal(req)
	if err != nil {
		r.pendingMu.Lock()
		delete(r.pending, id)
		r.pendingMu.Unlock()
		return nil, fmt.Errorf("rawcdp: marshal: %w", err)
	}

	r.mu.Lock()
	err = wsutil.WriteClientText(conn, data)
	r.mu.Unlock()
	if err != nil {
		r.pendingMu.Lock()
		delete(r.pending, id)
		r.pendingMu.Unlock()
		return nil, fmt.Errorf("rawcdp: send: %w", err)
	}

	select {
	case resp, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("rawcdp: connection closed")
		}
		return resp, nil
	case <-ctx.Done():
		r.pendingMu.Lock()
		delete(r.pending, id)
		r.pendingMu.Unlock()
		return nil, ctx.Err()
	}
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
	r.mu.Lock()
	conn := r.conn
	r.mu.Unlock()
	if conn == nil {
		return nil, fmt.Errorf("rawcdp: not connected")
	}

	id := r.seq.Add(1)

	req := struct {
		ID        int64  `json:"id"`
		Method    string `json:"method"`
		SessionID string `json:"sessionId,omitempty"`
		Params    any    `json:"params,omitempty"`
	}{ID: id, Method: method, SessionID: sessionID, Params: params}

	ch := make(chan json.RawMessage, 1)
	r.pendingMu.Lock()
	r.pending[id] = ch
	r.pendingMu.Unlock()

	data, err := json.Marshal(req)
	if err != nil {
		r.pendingMu.Lock()
		delete(r.pending, id)
		r.pendingMu.Unlock()
		return nil, fmt.Errorf("rawcdp: marshal: %w", err)
	}

	r.mu.Lock()
	err = wsutil.WriteClientText(conn, data)
	r.mu.Unlock()
	if err != nil {
		r.pendingMu.Lock()
		delete(r.pending, id)
		r.pendingMu.Unlock()
		return nil, fmt.Errorf("rawcdp: send: %w", err)
	}

	select {
	case resp, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("rawcdp: connection closed")
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
	case <-ctx.Done():
		r.pendingMu.Lock()
		delete(r.pending, id)
		r.pendingMu.Unlock()
		return nil, ctx.Err()
	}
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
