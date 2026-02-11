package runtimeprobes

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/dgnsrekt/tv_agent/internal/config"
)

// Run executes the runtime probes stage.
func Run(ctx context.Context, w io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	attached, injected, err := runProbeBootstrap(ctx, cfg, slog.Default())
	if err != nil {
		slog.Warn(
			"Runtime probes bootstrap unavailable",
			"error",
			err,
			"cdp_url",
			cfg.GetCDPURL(),
		)
	} else {
		slog.Info("Runtime probes bootstrap complete", "attached_tabs", attached, "injected_tabs", injected)
	}

	_, err = fmt.Fprintln(w, "runtime-probes: complete")
	return err
}

type probeTarget struct {
	ID  target.ID
	URL string
}

type probeBootstrapResult struct {
	AlreadyInjected bool     `json:"alreadyInjected"`
	URL             string   `json:"url"`
	Installed       []string `json:"installed"`
}

type probeLifecycle struct {
	Attached bool
	Result   probeBootstrapResult
	Events   []jsRuntimeTraceEvent
}

type probeRunner func(context.Context, probeTarget) (probeLifecycle, error)

const passiveProbeBootstrapJS = `(function () {
  const key = "__tvAgentPassiveProbe";
  const alreadyInjected = Object.prototype.hasOwnProperty.call(window, key) && window[key] && typeof window[key].emit === "function";
  if (!alreadyInjected) {
    const state = {
      version: "2",
      seq: 0,
      events: [],
      counters: { fetch: 0, xhr: 0, websocket: 0, event_bus: 0 },
      wrapped: [],
    };
    const mark = "__tvAgentWrapped";
    const next = function (surface) {
      state.counters[surface] = (state.counters[surface] || 0) + 1;
      return surface + "-" + state.counters[surface];
    };
    const emit = function (surface, eventType, payload, correlationId) {
      state.seq += 1;
      state.events.push({
        timestamp: new Date().toISOString(),
        surface: surface,
        eventType: eventType,
        sequence: state.seq,
        correlationId: correlationId || next(surface),
        payload: payload || {}
      });
      if (state.events.length > 1000) {
        state.events.splice(0, state.events.length - 1000);
      }
    };
    const safeString = function (value) {
      try {
        if (value === undefined || value === null) return "";
        return String(value);
      } catch (_) {
        return "";
      }
    };
    const normalizeHeaders = function (headersLike) {
      if (!headersLike) return {};
      try {
        if (headersLike instanceof Headers) {
          const out = {};
          headersLike.forEach(function (value, key) {
            out[safeString(key)] = safeString(value);
          });
          return out;
        }
      } catch (_) {}
      if (Array.isArray(headersLike)) {
        const out = {};
        for (let i = 0; i < headersLike.length; i += 1) {
          const pair = headersLike[i];
          if (Array.isArray(pair) && pair.length >= 2) {
            out[safeString(pair[0])] = safeString(pair[1]);
          }
        }
        return out;
      }
      if (typeof headersLike === "object") {
        const out = {};
        Object.keys(headersLike).forEach(function (k) {
          out[safeString(k)] = safeString(headersLike[k]);
        });
        return out;
      }
      return {};
    };

    const wrapFetch = function () {
      if (typeof window.fetch !== "function") return;
      const original = window.fetch;
      if (original[mark]) return;
      const wrapped = function (input, init) {
        const corr = next("fetch");
        let url = "";
        let method = "GET";
        let headers = {};
        try {
          if (typeof Request !== "undefined" && input instanceof Request) {
            url = safeString(input.url || "");
            method = safeString(input.method || method);
            headers = normalizeHeaders(input.headers);
          } else {
            url = safeString(input);
          }
          if (init && typeof init === "object") {
            if (init.method) method = safeString(init.method);
            if (init.headers) headers = normalizeHeaders(init.headers);
          }
        } catch (_) {}
        emit("fetch", "request", { url: url, method: method, headers: headers }, corr);
        let result;
        try {
          result = original.apply(this, arguments);
        } catch (err) {
          emit("fetch", "error", { message: safeString(err && err.message) }, corr);
          throw err;
        }
        if (result && typeof result.then === "function") {
          return result.then(function (resp) {
            emit("fetch", "response", { status: Number(resp && resp.status || 0), ok: Boolean(resp && resp.ok) }, corr);
            return resp;
          }).catch(function (err) {
            emit("fetch", "error", { message: safeString(err && err.message) }, corr);
            throw err;
          });
        }
        return result;
      };
      wrapped[mark] = true;
      window.fetch = wrapped;
      state.wrapped.push("fetch");
      emit("fetch", "hook_installed", { source: "window.fetch" }, next("fetch"));
    };

    const wrapXHR = function () {
      if (typeof XMLHttpRequest === "undefined" || !XMLHttpRequest || !XMLHttpRequest.prototype) return;
      const proto = XMLHttpRequest.prototype;
      if (proto.open && proto.open[mark]) return;
      const origOpen = proto.open;
      const origSend = proto.send;
      const origSetHeader = proto.setRequestHeader;
      proto.open = function (method, url) {
        this.__tvAgentCorr = next("xhr");
        this.__tvAgentMeta = { method: safeString(method || "GET"), url: safeString(url || ""), headers: {} };
        return origOpen.apply(this, arguments);
      };
      proto.setRequestHeader = function (k, v) {
        try {
          if (!this.__tvAgentMeta) this.__tvAgentMeta = { headers: {} };
          if (!this.__tvAgentMeta.headers) this.__tvAgentMeta.headers = {};
          this.__tvAgentMeta.headers[safeString(k)] = safeString(v);
        } catch (_) {}
        return origSetHeader.apply(this, arguments);
      };
      proto.send = function () {
        const corr = this.__tvAgentCorr || next("xhr");
        const meta = this.__tvAgentMeta || {};
        emit("xhr", "request", { url: safeString(meta.url), method: safeString(meta.method || "GET"), headers: meta.headers || {} }, corr);
        const onLoad = () => emit("xhr", "response", { status: Number(this.status || 0) }, corr);
        const onError = () => emit("xhr", "error", { status: Number(this.status || 0) }, corr);
        try {
          this.addEventListener("load", onLoad, { once: true });
          this.addEventListener("error", onError, { once: true });
        } catch (_) {}
        return origSend.apply(this, arguments);
      };
      proto.open[mark] = true;
      proto.send[mark] = true;
      proto.setRequestHeader[mark] = true;
      state.wrapped.push("xhr");
      emit("xhr", "hook_installed", { source: "XMLHttpRequest.prototype" }, next("xhr"));
    };

    const wrapWebSocket = function () {
      if (typeof WebSocket !== "function") return;
      const OriginalWS = WebSocket;
      if (OriginalWS[mark]) return;
      const WrappedWS = function (url, protocols) {
        const corr = next("websocket");
        const socket = protocols === undefined ? new OriginalWS(url) : new OriginalWS(url, protocols);
        emit("websocket", "connect", { url: safeString(url) }, corr);
        const origSend = socket.send;
        socket.send = function () {
          let size = 0;
          try {
            const data = arguments.length > 0 ? arguments[0] : "";
            size = safeString(data).length;
          } catch (_) {}
          emit("websocket", "send", { size: size }, corr);
          return origSend.apply(socket, arguments);
        };
        try {
          socket.addEventListener("open", function () { emit("websocket", "open", {}, corr); });
          socket.addEventListener("message", function (ev) { emit("websocket", "message", { size: safeString(ev && ev.data).length }, corr); });
          socket.addEventListener("close", function (ev) { emit("websocket", "close", { code: Number(ev && ev.code || 0), reason: safeString(ev && ev.reason) }, corr); });
          socket.addEventListener("error", function () { emit("websocket", "error", {}, corr); });
        } catch (_) {}
        return socket;
      };
      WrappedWS.prototype = OriginalWS.prototype;
      try {
        Object.defineProperty(WrappedWS, "name", { value: "WebSocket" });
      } catch (_) {}
      WrappedWS[mark] = true;
      window.WebSocket = WrappedWS;
      state.wrapped.push("websocket");
      emit("websocket", "hook_installed", { source: "window.WebSocket" }, next("websocket"));
    };

    const wrapEventBus = function () {
      const wrapDispatchEvent = function () {
        if (typeof EventTarget === "undefined" || !EventTarget || !EventTarget.prototype || typeof EventTarget.prototype.dispatchEvent !== "function") return;
        const original = EventTarget.prototype.dispatchEvent;
        if (original[mark]) return;
        EventTarget.prototype.dispatchEvent = function (event) {
          const corr = next("event_bus");
          let eventType = "";
          try { eventType = safeString(event && event.type); } catch (_) {}
          emit("event_bus", "dispatch_event", { eventType: eventType, target: safeString(this && this.constructor && this.constructor.name) }, corr);
          return original.apply(this, arguments);
        };
        EventTarget.prototype.dispatchEvent[mark] = true;
        state.wrapped.push("event_bus.dispatchEvent");
      };
      const wrapMethod = function (owner, methodName, ownerName) {
        try {
          if (!owner || typeof owner[methodName] !== "function") return false;
          const original = owner[methodName];
          if (original[mark]) return false;
          owner[methodName] = function () {
            const corr = next("event_bus");
            const eventName = arguments.length > 0 ? safeString(arguments[0]) : "";
            emit("event_bus", "bus_call", { owner: ownerName, method: methodName, eventName: eventName }, corr);
            return original.apply(this, arguments);
          };
          owner[methodName][mark] = true;
          state.wrapped.push(ownerName + "." + methodName);
          return true;
        } catch (_) {
          return false;
        }
      };
      wrapDispatchEvent();
      const candidateMethods = ["dispatch", "emit", "trigger", "publish"];
      const names = Object.getOwnPropertyNames(window);
      for (let i = 0; i < names.length; i += 1) {
        const name = names[i];
        if (!/(store|bus|event|emitter)/i.test(name)) continue;
        let value;
        try {
          value = window[name];
        } catch (_) {
          continue;
        }
        if (!value || (typeof value !== "object" && typeof value !== "function")) continue;
        for (let j = 0; j < candidateMethods.length; j += 1) {
          wrapMethod(value, candidateMethods[j], name);
        }
      }
      emit("event_bus", "hook_installed", { source: "dispatch/event-bus wrappers" }, next("event_bus"));
    };

    wrapFetch();
    wrapXHR();
    wrapWebSocket();
    wrapEventBus();

    Object.defineProperty(window, key, {
      value: {
        version: state.version,
        injectedAt: new Date().toISOString(),
        emit: emit,
        drain: function () {
          const snapshot = state.events.slice();
          state.events.length = 0;
          return snapshot;
        }
      },
      configurable: true
    });
  }
  return {
    alreadyInjected: alreadyInjected,
    url: String(window.location.href || ""),
    installed: ["fetch", "xhr", "websocket", "event_bus"]
  };
})();`

const passiveProbeDrainJS = `(function () {
  const key = "__tvAgentPassiveProbe";
  if (!Object.prototype.hasOwnProperty.call(window, key)) {
    return [];
  }
  const probe = window[key];
  if (!probe || typeof probe.drain !== "function") {
    return [];
  }
  return probe.drain();
})();`

const runtimeTraceRelativeOutput = "mapper/runtime-probes/runtime-trace.jsonl"

type jsRuntimeTraceEvent struct {
	Timestamp     string         `json:"timestamp"`
	Surface       string         `json:"surface"`
	EventType     string         `json:"eventType"`
	Sequence      int            `json:"sequence"`
	CorrelationID string         `json:"correlationId"`
	Payload       map[string]any `json:"payload"`
}

type runtimeTraceRecord struct {
	Timestamp     time.Time      `json:"timestamp"`
	TraceID       string         `json:"trace_id"`
	TabID         string         `json:"tab_id"`
	TabURL        string         `json:"tab_url"`
	Surface       string         `json:"surface"`
	EventType     string         `json:"event_type"`
	Sequence      int            `json:"sequence"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
}

var (
	secretKVPattern        = regexp.MustCompile(`(?i)(token|session|auth|authorization|cookie|secret|api[_-]?key|jwt|bearer)=([^&\s]+)`)
	bearerTokenPattern     = regexp.MustCompile(`(?i)bearer\s+[a-z0-9\-._~+/]+=*`)
	jwtLikePattern         = regexp.MustCompile(`eyJ[A-Za-z0-9_-]{6,}\.[A-Za-z0-9_-]{6,}\.[A-Za-z0-9_-]{6,}`)
	sensitiveKeyName       = regexp.MustCompile(`(?i)(token|session|auth|authorization|cookie|secret|password|pass|api[_-]?key|jwt|bearer|sid)`)
	runtimeProbeSurfaces   = []string{"fetch", "xhr", "websocket", "event_bus"}
	defaultRuntimeProbeDir = "./research_data"
)

func runProbeBootstrap(ctx context.Context, cfg *config.Config, logger *slog.Logger) (int, int, error) {
	cdpURL := cfg.GetCDPURL()
	logger.Info("Runtime probes bootstrap start", "cdp_url", cdpURL, "tab_url_filter", cfg.TabURLFilter)

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, cdpURL)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	if err := chromedp.Run(browserCtx); err != nil {
		return 0, 0, fmt.Errorf("connect to browser: %w", err)
	}

	targets, err := chromedp.Targets(browserCtx)
	if err != nil {
		return 0, 0, fmt.Errorf("enumerate targets: %w", err)
	}

	matches := filterProbeTargets(targets, cfg.TabURLFilter)
	if len(matches) == 0 {
		logger.Warn("Runtime probes found no matching tabs", "tab_url_filter", cfg.TabURLFilter)
		return 0, 0, nil
	}

	attached, injected, records, err := bootstrapTargets(ctx, logger, matches, makeCDPRunner(allocCtx))
	if err != nil {
		return 0, 0, err
	}

	if err := persistRuntimeTraceRecords(cfg.DataDir, time.Now().UTC(), records); err != nil {
		return 0, 0, fmt.Errorf("persist runtime trace artifacts: %w", err)
	}

	return attached, injected, nil
}

func filterProbeTargets(targets []*target.Info, urlFilter string) []probeTarget {
	filter := strings.ToLower(urlFilter)
	matches := make([]probeTarget, 0, len(targets))
	for _, t := range targets {
		if t.Type != "page" {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(t.URL), filter) {
			continue
		}
		matches = append(matches, probeTarget{ID: t.TargetID, URL: t.URL})
	}
	return matches
}

func bootstrapTargets(ctx context.Context, logger *slog.Logger, targets []probeTarget, runner probeRunner) (int, int, []runtimeTraceRecord, error) {
	attached := 0
	injected := 0
	records := make([]runtimeTraceRecord, 0, len(targets)*8)

	for _, tab := range targets {
		lifecycle, err := runner(ctx, tab)

		if lifecycle.Attached {
			attached++
			logger.Info("Runtime probe attach success", "tab_id", tab.ID, "url", truncateURL(tab.URL))
		}

		if err != nil {
			stage := "attach"
			if lifecycle.Attached {
				stage = "inject"
			}
			logger.Warn("Runtime probe lifecycle failed", "stage", stage, "tab_id", tab.ID, "error", err)
			continue
		}

		injected++
		records = append(records, buildHookInstalledRecords(tab, lifecycle.Result.Installed)...)
		records = append(records, normalizeRuntimeTraceRecords(tab, lifecycle.Events)...)
		logger.Info(
			"Runtime probe inject success",
			"tab_id",
			tab.ID,
			"already_injected",
			lifecycle.Result.AlreadyInjected,
			"url",
			truncateURL(lifecycle.Result.URL),
		)
	}

	return attached, injected, records, nil
}

func makeCDPRunner(allocCtx context.Context) probeRunner {
	return func(_ context.Context, tab probeTarget) (probeLifecycle, error) {
		tabCtx, tabCancel := chromedp.NewContext(allocCtx, chromedp.WithTargetID(tab.ID))
		defer tabCancel()

		attachCtx, attachCancel := context.WithTimeout(tabCtx, 10*time.Second)
		defer attachCancel()
		if err := chromedp.Run(attachCtx); err != nil {
			return probeLifecycle{}, fmt.Errorf("attach to tab: %w", err)
		}

		probeCtx, probeCancel := context.WithTimeout(tabCtx, 10*time.Second)
		defer probeCancel()

		var result probeBootstrapResult
		if err := chromedp.Run(probeCtx, chromedp.Evaluate(passiveProbeBootstrapJS, &result)); err != nil {
			return probeLifecycle{Attached: true}, fmt.Errorf("inject probe: %w", err)
		}

		var events []jsRuntimeTraceEvent
		if err := chromedp.Run(probeCtx, chromedp.Evaluate(passiveProbeDrainJS, &events)); err != nil {
			return probeLifecycle{Attached: true, Result: result}, fmt.Errorf("drain probe events: %w", err)
		}

		return probeLifecycle{Attached: true, Result: result, Events: events}, nil
	}
}

func buildHookInstalledRecords(tab probeTarget, surfaces []string) []runtimeTraceRecord {
	if len(surfaces) == 0 {
		surfaces = runtimeProbeSurfaces
	}
	now := time.Now().UTC()
	out := make([]runtimeTraceRecord, 0, len(surfaces))
	for idx, surface := range surfaces {
		sequence := idx + 1
		corr := surface + "-hook-" + strconv.Itoa(sequence)
		out = append(out, runtimeTraceRecord{
			Timestamp:     now,
			TraceID:       buildTraceID(tab.ID, surface, corr, sequence),
			TabID:         string(tab.ID),
			TabURL:        tab.URL,
			Surface:       surface,
			EventType:     "hook_installed",
			Sequence:      sequence,
			CorrelationID: corr,
			Payload: map[string]any{
				"source": "runtime_probe",
			},
		})
	}
	return out
}

func normalizeRuntimeTraceRecords(tab probeTarget, events []jsRuntimeTraceEvent) []runtimeTraceRecord {
	records := make([]runtimeTraceRecord, 0, len(events))
	for idx, ev := range events {
		sequence := ev.Sequence
		if sequence == 0 {
			sequence = idx + 1
		}
		ts := parseEventTimestamp(ev.Timestamp)
		corr := ev.CorrelationID
		if corr == "" {
			corr = ev.Surface + "-event-" + strconv.Itoa(sequence)
		}
		payload := redactSecrets(ev.Payload)
		records = append(records, runtimeTraceRecord{
			Timestamp:     ts,
			TraceID:       buildTraceID(tab.ID, ev.Surface, corr, sequence),
			TabID:         string(tab.ID),
			TabURL:        tab.URL,
			Surface:       ev.Surface,
			EventType:     ev.EventType,
			Sequence:      sequence,
			CorrelationID: corr,
			Payload:       payload,
		})
	}
	return records
}

func parseEventTimestamp(raw string) time.Time {
	if raw == "" {
		return time.Now().UTC()
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed.UTC()
}

func buildTraceID(tabID target.ID, surface, corr string, sequence int) string {
	safeSurface := strings.TrimSpace(surface)
	if safeSurface == "" {
		safeSurface = "unknown"
	}
	safeCorr := strings.TrimSpace(corr)
	if safeCorr == "" {
		safeCorr = strconv.Itoa(sequence)
	}
	return string(tabID) + ":" + safeSurface + ":" + safeCorr
}

func redactSecrets(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		if sensitiveKeyName.MatchString(k) {
			out[k] = "[REDACTED]"
			continue
		}
		out[k] = redactValue(v)
	}
	return out
}

func redactValue(v any) any {
	switch tv := v.(type) {
	case string:
		return redactString(tv)
	case map[string]any:
		return redactSecrets(tv)
	case []any:
		out := make([]any, len(tv))
		for i := range tv {
			out[i] = redactValue(tv[i])
		}
		return out
	default:
		return v
	}
}

func redactString(s string) string {
	s = secretKVPattern.ReplaceAllString(s, "$1=[REDACTED]")
	s = bearerTokenPattern.ReplaceAllString(s, "Bearer [REDACTED]")
	s = jwtLikePattern.ReplaceAllString(s, "[REDACTED_JWT]")
	return s
}

func persistRuntimeTraceRecords(dataDir string, now time.Time, records []runtimeTraceRecord) error {
	if dataDir == "" {
		dataDir = defaultRuntimeProbeDir
	}
	datePart := now.UTC().Format("2006-01-02")
	outputPath := filepath.Join(dataDir, datePart, runtimeTraceRelativeOutput)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	enc := json.NewEncoder(w)
	for i := range records {
		records[i].Payload = redactSecrets(records[i].Payload)
		if err := enc.Encode(records[i]); err != nil {
			return err
		}
	}
	return w.Flush()
}

func truncateURL(url string) string {
	if len(url) > 120 {
		return url[:120] + "..."
	}
	return url
}
