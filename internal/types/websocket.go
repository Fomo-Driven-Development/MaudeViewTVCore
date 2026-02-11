package types

import "time"

// WebSocketCapture represents a captured WebSocket event.
type WebSocketCapture struct {
	Timestamp    time.Time `json:"timestamp"`
	RequestID    string    `json:"request_id"`
	TabID        string    `json:"tab_id"`
	URL          string    `json:"url"`
	EventType    string    `json:"event_type"`
	Direction    string    `json:"direction,omitempty"`
	Opcode       int       `json:"opcode,omitempty"`
	PayloadData  string    `json:"payload_data,omitempty"`
	CloseCode    int       `json:"close_code,omitempty"`
	CloseReason  string    `json:"close_reason,omitempty"`
	Truncated    bool      `json:"truncated,omitempty"`
	OriginalSize int       `json:"original_size,omitempty"`
	SHA256       string    `json:"sha256,omitempty"`
}

// WebSocketConnection tracks an active WebSocket connection.
type WebSocketConnection struct {
	RequestID string
	URL       string
	TabID     string
	CreatedAt time.Time
}
