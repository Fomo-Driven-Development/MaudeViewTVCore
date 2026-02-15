package capture

import (
	"log/slog"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/storage"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/types"
)

// WebSocketConnectionInfo extends types.WebSocketConnection with tab routing info.
type WebSocketConnectionInfo struct {
	*types.WebSocketConnection
	PathSegment string
	BrowserID   string
}

// WebSocketCapture handles capturing WebSocket traffic.
type WebSocketCapture struct {
	registry      *storage.WriterRegistry
	tabRegistry   types.TabInfoProvider
	captureWS     bool
	maxFrameBytes int

	connections   map[string]*WebSocketConnectionInfo
	connectionsMu sync.RWMutex
}

func NewWebSocketCapture(registry *storage.WriterRegistry, tabRegistry types.TabInfoProvider, captureWS bool, maxFrameBytes int) *WebSocketCapture {
	return &WebSocketCapture{
		registry:      registry,
		tabRegistry:   tabRegistry,
		captureWS:     captureWS,
		maxFrameBytes: maxFrameBytes,
		connections:   make(map[string]*WebSocketConnectionInfo),
	}
}

func (w *WebSocketCapture) OnWebSocketCreated(tabID string, ev *network.EventWebSocketCreated) {
	if !w.captureWS {
		return
	}

	tabInfo, ok := w.tabRegistry.GetByStringID(tabID)
	pathSegment := "unknown"
	browserID := "unknown"
	if ok {
		pathSegment = tabInfo.PathSegment
		browserID = tabInfo.BrowserID
	}

	conn := &WebSocketConnectionInfo{
		WebSocketConnection: &types.WebSocketConnection{
			RequestID: string(ev.RequestID),
			URL:       ev.URL,
			TabID:     tabID,
			CreatedAt: time.Now().UTC(),
		},
		PathSegment: pathSegment,
		BrowserID:   browserID,
	}

	w.connectionsMu.Lock()
	w.connections[string(ev.RequestID)] = conn
	w.connectionsMu.Unlock()

	capture := &types.WebSocketCapture{
		Timestamp: time.Now().UTC(),
		RequestID: string(ev.RequestID),
		TabID:     tabID,
		URL:       ev.URL,
		EventType: "created",
	}

	writer := w.registry.GetWriter(pathSegment, "websocket", browserID)
	if err := writer.Write(capture); err != nil {
		slog.Error("Failed to write WebSocket created event", "request_id", ev.RequestID, "error", err)
	}
}

func (w *WebSocketCapture) OnWebSocketFrameReceived(tabID string, ev *network.EventWebSocketFrameReceived) {
	if !w.captureWS {
		return
	}

	w.connectionsMu.RLock()
	conn, ok := w.connections[string(ev.RequestID)]
	w.connectionsMu.RUnlock()
	if !ok {
		return
	}

	payload, truncated, originalSize, payloadHash := truncateStringBytes(ev.Response.PayloadData, w.maxFrameBytes)
	capture := &types.WebSocketCapture{
		Timestamp:    time.Now().UTC(),
		RequestID:    string(ev.RequestID),
		TabID:        tabID,
		URL:          conn.URL,
		EventType:    "frame_received",
		Direction:    "incoming",
		Opcode:       int(ev.Response.Opcode),
		PayloadData:  payload,
		Truncated:    truncated,
		OriginalSize: originalSize,
		SHA256:       payloadHash,
	}

	writer := w.registry.GetWriter(conn.PathSegment, "websocket", conn.BrowserID)
	if err := writer.Write(capture); err != nil {
		slog.Error("Failed to write WebSocket frame", "request_id", ev.RequestID, "error", err)
	}
}

func (w *WebSocketCapture) OnWebSocketFrameSent(tabID string, ev *network.EventWebSocketFrameSent) {
	if !w.captureWS {
		return
	}

	w.connectionsMu.RLock()
	conn, ok := w.connections[string(ev.RequestID)]
	w.connectionsMu.RUnlock()
	if !ok {
		return
	}

	payload, truncated, originalSize, payloadHash := truncateStringBytes(ev.Response.PayloadData, w.maxFrameBytes)
	capture := &types.WebSocketCapture{
		Timestamp:    time.Now().UTC(),
		RequestID:    string(ev.RequestID),
		TabID:        tabID,
		URL:          conn.URL,
		EventType:    "frame_sent",
		Direction:    "outgoing",
		Opcode:       int(ev.Response.Opcode),
		PayloadData:  payload,
		Truncated:    truncated,
		OriginalSize: originalSize,
		SHA256:       payloadHash,
	}

	writer := w.registry.GetWriter(conn.PathSegment, "websocket", conn.BrowserID)
	if err := writer.Write(capture); err != nil {
		slog.Error("Failed to write WebSocket frame", "request_id", ev.RequestID, "error", err)
	}
}

func (w *WebSocketCapture) OnWebSocketClosed(tabID string, ev *network.EventWebSocketClosed) {
	if !w.captureWS {
		return
	}

	w.connectionsMu.Lock()
	conn, ok := w.connections[string(ev.RequestID)]
	if ok {
		delete(w.connections, string(ev.RequestID))
	}
	w.connectionsMu.Unlock()
	if !ok {
		return
	}

	capture := &types.WebSocketCapture{
		Timestamp: time.Now().UTC(),
		RequestID: string(ev.RequestID),
		TabID:     tabID,
		URL:       conn.URL,
		EventType: "closed",
	}

	writer := w.registry.GetWriter(conn.PathSegment, "websocket", conn.BrowserID)
	if err := writer.Write(capture); err != nil {
		slog.Error("Failed to write WebSocket closed event", "request_id", ev.RequestID, "error", err)
	}
}

func (w *WebSocketCapture) GetActiveConnections() int {
	w.connectionsMu.RLock()
	defer w.connectionsMu.RUnlock()
	return len(w.connections)
}

func truncateStringBytes(in string, maxBytes int) (string, bool, int, string) {
	raw := []byte(in)
	out, truncated, origLen, hash := truncateBytes(raw, maxBytes)
	return string(out), truncated, origLen, hash
}
