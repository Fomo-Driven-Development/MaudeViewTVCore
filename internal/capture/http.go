package capture

import (
	"encoding/base64"
	"log/slog"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/chromedp/cdproto/network"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/storage"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/types"
)

// HTTPCapture handles capturing and correlating HTTP traffic.
type HTTPCapture struct {
	registry       *storage.WriterRegistry
	resourceWriter *storage.ResourceWriter
	tabRegistry    types.TabInfoProvider

	captureHTTP   bool
	captureStatic bool
	maxBodyBytes  int
	maxResBytes   int

	pending   map[string]*types.PendingRequest
	pendingMu sync.RWMutex

	done chan struct{}
}

func NewHTTPCapture(
	registry *storage.WriterRegistry,
	resourceWriter *storage.ResourceWriter,
	tabRegistry types.TabInfoProvider,
	captureHTTP bool,
	captureStatic bool,
	maxBodyBytes int,
	maxResBytes int,
) *HTTPCapture {
	h := &HTTPCapture{
		registry:       registry,
		resourceWriter: resourceWriter,
		tabRegistry:    tabRegistry,
		captureHTTP:    captureHTTP,
		captureStatic:  captureStatic,
		maxBodyBytes:   maxBodyBytes,
		maxResBytes:    maxResBytes,
		pending:        make(map[string]*types.PendingRequest),
		done:           make(chan struct{}),
	}
	go h.cleanupLoop()
	return h
}

func (h *HTTPCapture) Close() {
	close(h.done)
}

func (h *HTTPCapture) OnRequestWillBeSent(tabID string, ev *network.EventRequestWillBeSent) {
	var postData string
	if ev.Request.HasPostData && len(ev.Request.PostDataEntries) > 0 {
		var decodedParts []byte
		for _, entry := range ev.Request.PostDataEntries {
			if entry.Bytes == "" {
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(entry.Bytes)
			if err != nil {
				decodedParts = append(decodedParts, []byte(entry.Bytes)...)
			} else {
				decodedParts = append(decodedParts, decoded...)
			}
		}
		postData = string(decodedParts)
	}

	capture := &types.HTTPCapture{
		Timestamp: time.Now().UTC(),
		RequestID: string(ev.RequestID),
		TabID:     tabID,
		URL:       ev.Request.URL,
		Method:    ev.Request.Method,
		Request: types.HTTPRequest{
			Headers:  headerMapToStringMap(ev.Request.Headers),
			PostData: postData,
		},
	}

	h.pendingMu.Lock()
	h.pending[string(ev.RequestID)] = &types.PendingRequest{
		Capture:   capture,
		Timestamp: time.Now(),
	}
	h.pendingMu.Unlock()
}

func (h *HTTPCapture) OnResponseReceived(tabID string, ev *network.EventResponseReceived) {
	h.pendingMu.Lock()
	pending, ok := h.pending[string(ev.RequestID)]
	if ok {
		pending.Capture.Response = &types.HTTPResponse{
			Status:     int(ev.Response.Status),
			StatusText: ev.Response.StatusText,
			Headers:    headerMapToStringMap(ev.Response.Headers),
		}
		pending.ResourceType = string(ev.Type)
	}
	h.pendingMu.Unlock()

	if !ok {
		return
	}
}

func (h *HTTPCapture) OnLoadingFinished(tabID string, ev *network.EventLoadingFinished, getBody func() ([]byte, bool, error)) {
	h.pendingMu.Lock()
	pending, ok := h.pending[string(ev.RequestID)]
	if ok {
		delete(h.pending, string(ev.RequestID))
	}
	h.pendingMu.Unlock()

	if !ok {
		return
	}

	tabInfo, ok := h.tabRegistry.GetByStringID(tabID)
	if !ok {
		tabInfo = &types.TabInfo{PathSegment: "unknown", BrowserID: "unknown"}
	}

	pathSegment := tabInfo.PathSegment
	browserID := tabInfo.BrowserID
	resourceDir := storage.MapResourceType(pending.ResourceType)
	requestURL := pending.Capture.URL

	go func() {
		var body []byte
		if pending.Capture.Response != nil && getBody != nil {
			fetchedBody, _, err := getBody()
			if err != nil {
				slog.Debug("Failed to get response body", "request_id", ev.RequestID, "error", err)
			} else {
				body = fetchedBody
			}
		}

		if h.captureStatic && resourceDir != "" && len(body) > 0 {
			resourceBody, truncated, originalSize, bodyHash := truncateBytes(body, h.maxResBytes)
			filename := storage.FilenameFromURL(requestURL)
			if err := h.resourceWriter.WriteRaw(pathSegment, resourceDir, filename, resourceBody); err != nil {
				slog.Error("Failed to write resource file", "request_id", ev.RequestID, "error", err)
			} else if truncated {
				slog.Warn("Resource truncated due to max size", "request_id", ev.RequestID, "original_size", originalSize, "kept_size", len(resourceBody), "sha256", bodyHash)
			}
		}

		if !h.captureHTTP {
			return
		}

		if pending.Capture.Response != nil && len(body) > 0 {
			bodyForJSONL, truncated, originalSize, bodyHash := truncateBytes(body, h.maxBodyBytes)
			if utf8.Valid(bodyForJSONL) {
				pending.Capture.Response.Body = string(bodyForJSONL)
			} else {
				pending.Capture.Response.BodyBase64 = base64.StdEncoding.EncodeToString(bodyForJSONL)
			}
			if truncated {
				pending.Capture.Response.Truncated = true
				pending.Capture.Response.OriginalSize = originalSize
				pending.Capture.Response.SHA256 = bodyHash
			}
		}

		writer := h.registry.GetWriter(pathSegment, "http", browserID)
		if err := writer.Write(pending.Capture); err != nil {
			slog.Error("Failed to write HTTP capture", "request_id", ev.RequestID, "error", err)
		}
	}()
}

func (h *HTTPCapture) OnLoadingFailed(tabID string, ev *network.EventLoadingFailed) {
	h.pendingMu.Lock()
	delete(h.pending, string(ev.RequestID))
	h.pendingMu.Unlock()
}

func (h *HTTPCapture) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.cleanupStale()
		case <-h.done:
			return
		}
	}
}

func (h *HTTPCapture) cleanupStale() {
	threshold := time.Now().Add(-5 * time.Minute)

	h.pendingMu.Lock()
	defer h.pendingMu.Unlock()

	for id, pending := range h.pending {
		if pending.Timestamp.Before(threshold) {
			delete(h.pending, id)
		}
	}
}

func headerMapToStringMap(headers map[string]any) map[string]string {
	result := make(map[string]string, len(headers))
	for k, v := range headers {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result
}
