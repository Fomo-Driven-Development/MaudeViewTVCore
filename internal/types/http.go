package types

import "time"

// HTTPCapture represents a captured HTTP request/response pair.
type HTTPCapture struct {
	Timestamp time.Time     `json:"timestamp"`
	RequestID string        `json:"request_id"`
	TabID     string        `json:"tab_id"`
	URL       string        `json:"url"`
	Method    string        `json:"method"`
	Request   HTTPRequest   `json:"request"`
	Response  *HTTPResponse `json:"response,omitempty"`
}

// HTTPRequest represents the request portion of an HTTP capture.
type HTTPRequest struct {
	Headers  map[string]string `json:"headers,omitempty"`
	PostData string            `json:"post_data,omitempty"`
}

// HTTPResponse represents the response portion of an HTTP capture.
type HTTPResponse struct {
	Status       int               `json:"status"`
	StatusText   string            `json:"status_text"`
	Headers      map[string]string `json:"headers,omitempty"`
	Body         string            `json:"body,omitempty"`
	BodyBase64   string            `json:"body_base64,omitempty"`
	Truncated    bool              `json:"truncated,omitempty"`
	OriginalSize int               `json:"original_size,omitempty"`
	SHA256       string            `json:"sha256,omitempty"`
}

// PendingRequest tracks an in-flight HTTP request waiting for response.
type PendingRequest struct {
	Capture      *HTTPCapture
	Timestamp    time.Time
	ResourceType string
}
