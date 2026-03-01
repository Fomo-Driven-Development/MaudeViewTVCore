package cdpcontrol

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// screencastFrame is queued between the event handler and the writer goroutine.
type screencastFrame struct {
	raw   string // raw base64 string from event (decode happens in writer)
	ackID int    // frame number for Page.screencastFrameAck sessionId param
}

// screencastSession owns the frame-writing goroutine and ack loop for one screencast.
type screencastSession struct {
	id           string
	chartID      string
	cdpSessionID string // CDP flattened session string — for routing acks
	rawCDP       *rawCDP
	dir          string
	format       string

	frameCount atomic.Int64
	mu         sync.Mutex // protects: status, stoppedAt, unregister
	status     string     // "active" | "stopped"
	startedAt  time.Time
	stoppedAt  time.Time
	unregister func()

	frameCh chan screencastFrame // buffered 128 — handler→writer
	done    chan struct{}
	wg      sync.WaitGroup
}

// handleFrame is called from rawCDP's readLoop — must never block.
func (ss *screencastSession) handleFrame(sessionID string, params json.RawMessage) {
	if sessionID != ss.cdpSessionID {
		return
	}
	var evt struct {
		Data      string `json:"data"`
		SessionID int    `json:"sessionId"`
	}
	_ = json.Unmarshal(params, &evt)

	select {
	case ss.frameCh <- screencastFrame{raw: evt.Data, ackID: evt.SessionID}:
	default:
		// Channel full — spin a goroutine to ack only; frame write is dropped.
		ackID, cdp, sid := evt.SessionID, ss.rawCDP, ss.cdpSessionID
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, _ = cdp.sendFlat(ctx, sid, "Page.screencastFrameAck",
				struct {
					SessionID int `json:"sessionId"`
				}{SessionID: ackID})
		}()
	}
}

// writerLoop drains frameCh, acking first and writing second.
func (ss *screencastSession) writerLoop() {
	defer ss.wg.Done()
	for {
		select {
		case frame := <-ss.frameCh:
			ss.processFrame(frame)
		case <-ss.done:
			// Drain remaining frames (up to 5 s).
			timeout := time.After(5 * time.Second)
			for {
				select {
				case frame := <-ss.frameCh:
					ss.processFrame(frame)
				case <-timeout:
					return
				default:
					return
				}
			}
		}
	}
}

// processFrame acks the frame first, then decodes and writes to disk.
func (ss *screencastSession) processFrame(frame screencastFrame) {
	// 1. Ack first — unblocks browser regardless of disk outcome.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _ = ss.rawCDP.sendFlat(ctx, ss.cdpSessionID, "Page.screencastFrameAck",
		struct {
			SessionID int `json:"sessionId"`
		}{SessionID: frame.ackID})

	// 2. Decode and write — failure logs but doesn't stall.
	decoded, err := base64.StdEncoding.DecodeString(frame.raw)
	if err != nil {
		slog.Warn("screencast: base64 decode failed", "id", ss.id, "error", err)
		return
	}
	n := ss.frameCount.Add(1)
	fname := filepath.Join(ss.dir, fmt.Sprintf("frame_%06d.%s", n, ss.format))
	if err := os.WriteFile(fname, decoded, 0o644); err != nil {
		slog.Warn("screencast: write frame failed", "id", ss.id, "error", err)
		ss.frameCount.Add(-1) // revert so numbering stays contiguous
	}
}

// info returns a thread-safe snapshot of the session state.
func (ss *screencastSession) info() ScreencastInfo {
	ss.mu.Lock()
	status, stopped := ss.status, ss.stoppedAt
	ss.mu.Unlock()
	out := ScreencastInfo{
		ID:         ss.id,
		ChartID:    ss.chartID,
		Status:     status,
		Format:     ss.format,
		Dir:        ss.dir,
		FrameCount: ss.frameCount.Load(),
		StartedAt:  ss.startedAt,
	}
	if !stopped.IsZero() {
		t := stopped
		out.StoppedAt = &t
	}
	return out
}

// abort signals the session to stop without sending CDP commands.
// Safe to call while holding c.mu (no CDP network I/O, no c.mu acquisition).
func (ss *screencastSession) abort() {
	ss.mu.Lock()
	if ss.status != "active" {
		ss.mu.Unlock()
		return
	}
	ss.status = "stopped"
	ss.stoppedAt = time.Now()
	unreg := ss.unregister
	ss.unregister = nil
	ss.mu.Unlock()

	if unreg != nil {
		unreg()
	}
	select {
	case <-ss.done:
	default:
		close(ss.done)
	}
}

// stop sends Page.stopScreencast, then aborts and waits for the writer to drain.
func (ss *screencastSession) stop(ctx context.Context) ScreencastInfo {
	_, _ = ss.rawCDP.sendFlat(ctx, ss.cdpSessionID, "Page.stopScreencast", nil)
	ss.abort()
	ss.wg.Wait()
	return ss.info()
}
