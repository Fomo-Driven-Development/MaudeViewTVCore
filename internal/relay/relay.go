package relay

import (
	"encoding/json"
	"log/slog"
	"strings"
	"sync"

	"context"

	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

type connectionInfo struct {
	feedName     string
	msgFilter    map[string]bool // nil means accept all
}

// Relay tracks browser WebSocket connections via CDP events and publishes
// matching frames to an SSE Broker.
type Relay struct {
	cfg    *RelayConfig
	broker *Broker

	mu          sync.Mutex
	connections map[string]connectionInfo // requestID → info

	unregisterFns []func()
}

// NewRelay creates a relay engine.
func NewRelay(cfg *RelayConfig, broker *Broker) *Relay {
	return &Relay{
		cfg:         cfg,
		broker:      broker,
		connections: make(map[string]connectionInfo),
	}
}

// Start enables the Network domain and registers CDP event handlers.
func (r *Relay) Start(ctx context.Context, client *cdpcontrol.Client) error {
	if err := client.EnableNetworkDomain(ctx); err != nil {
		return err
	}

	methods := []struct {
		name string
		fn   func(string, json.RawMessage)
	}{
		{"Network.webSocketCreated", r.onWebSocketCreated},
		{"Network.webSocketFrameReceived", r.onWebSocketFrameReceived},
		{"Network.webSocketClosed", r.onWebSocketClosed},
	}

	for _, m := range methods {
		unreg, err := client.RegisterCDPEventHandler(m.name, m.fn)
		if err != nil {
			r.Stop()
			return err
		}
		r.unregisterFns = append(r.unregisterFns, unreg)
	}

	slog.Info("relay started", "feeds", len(r.cfg.Feeds))
	return nil
}

// Stop unregisters all CDP event handlers.
func (r *Relay) Stop() {
	for _, fn := range r.unregisterFns {
		fn()
	}
	r.unregisterFns = nil
	slog.Info("relay stopped")
}

func (r *Relay) onWebSocketCreated(_ string, params json.RawMessage) {
	var evt struct {
		RequestID string `json:"requestId"`
		URL       string `json:"url"`
	}
	if err := json.Unmarshal(params, &evt); err != nil {
		return
	}

	for _, feed := range r.cfg.Feeds {
		if strings.Contains(evt.URL, feed.URLPattern) {
			info := connectionInfo{feedName: feed.Name}
			if len(feed.MessageTypes) > 0 {
				info.msgFilter = make(map[string]bool, len(feed.MessageTypes))
				for _, mt := range feed.MessageTypes {
					info.msgFilter[mt] = true
				}
			}
			r.mu.Lock()
			r.connections[evt.RequestID] = info
			r.mu.Unlock()
			slog.Debug("relay: ws matched", "feed", feed.Name, "url", evt.URL, "request_id", evt.RequestID)
			return
		}
	}
}

func (r *Relay) onWebSocketFrameReceived(_ string, params json.RawMessage) {
	var evt struct {
		RequestID string `json:"requestId"`
		Response  struct {
			PayloadData string `json:"payloadData"`
		} `json:"response"`
	}
	if err := json.Unmarshal(params, &evt); err != nil {
		return
	}

	r.mu.Lock()
	info, ok := r.connections[evt.RequestID]
	r.mu.Unlock()
	if !ok {
		return
	}

	payload := evt.Response.PayloadData
	if payload == "" {
		return
	}

	// If no message type filter, publish everything.
	if info.msgFilter == nil {
		r.broker.Publish(Event{Feed: info.feedName, Payload: payload})
		return
	}

	// Extract "m" field for filtering.
	msgType := extractMessageType(payload)
	if msgType != "" && info.msgFilter[msgType] {
		r.broker.Publish(Event{Feed: info.feedName, Payload: payload})
	}
}

func (r *Relay) onWebSocketClosed(_ string, params json.RawMessage) {
	var evt struct {
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(params, &evt); err != nil {
		return
	}
	r.mu.Lock()
	delete(r.connections, evt.RequestID)
	r.mu.Unlock()
}

// extractMessageType extracts the "m" field from a JSON payload.
// It handles both plain JSON and socket.io framed messages (~m~NNN~m~...).
func extractMessageType(payload string) string {
	data := payload

	// Strip socket.io frame prefix: ~m~NNN~m~
	if strings.HasPrefix(data, "~m~") {
		// Find second ~m~ marker
		idx := strings.Index(data[3:], "~m~")
		if idx >= 0 {
			data = data[3+idx+3:]
		}
	}

	// Quick check: must look like JSON object
	data = strings.TrimSpace(data)
	if len(data) == 0 || data[0] != '{' {
		return ""
	}

	// Lightweight extraction: unmarshal just the "m" field.
	var obj struct {
		M string `json:"m"`
	}
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return ""
	}
	return obj.M
}
