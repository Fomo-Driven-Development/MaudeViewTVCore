package smoke

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dgnsrekt/tv_agent/internal/mapper/correlation"
	"github.com/dgnsrekt/tv_agent/internal/mapper/reporting"
	"github.com/dgnsrekt/tv_agent/internal/mapper/runtimeprobes"
	"github.com/dgnsrekt/tv_agent/internal/mapper/staticanalysis"
	"github.com/dgnsrekt/tv_agent/internal/mapper/validation"
)

const (
	defaultDataDir = "./research_data"
	defaultDocsDir = "./docs"
)

type stageRunner func(context.Context, io.Writer) error

type runtimeTraceRecord struct {
	Timestamp time.Time      `json:"timestamp"`
	TraceID   string         `json:"trace_id"`
	TabID     string         `json:"tab_id"`
	TabURL    string         `json:"tab_url"`
	Surface   string         `json:"surface"`
	EventType string         `json:"event_type"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type runtimeTraceSession struct {
	SessionID   string    `json:"session_id"`
	ProfileName string    `json:"profile_name"`
	TabID       string    `json:"tab_id"`
	StartedAt   time.Time `json:"started_at"`
	EndedAt     time.Time `json:"ended_at"`
}

type httpCapture struct {
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id"`
	TabID     string    `json:"tab_id"`
	URL       string    `json:"url"`
	Method    string    `json:"method"`
}

type websocketCapture struct {
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id"`
	TabID     string    `json:"tab_id"`
	URL       string    `json:"url"`
	EventType string    `json:"event_type"`
	Direction string    `json:"direction"`
}

type correlationRecord struct {
	Capability      string  `json:"capability"`
	ConfidenceScore float64 `json:"confidence_score"`
}

type baselineMetrics struct {
	BundleCount               int
	TraceSessionCount         int
	CorrelatedCapabilityCount int
	ConfidenceDistribution    map[string]int
}

// Run executes smoke-mode mapper pipeline on existing captured data and persists a docs report.
func Run(ctx context.Context, w io.Writer) error {
	dataDir := strings.TrimSpace(os.Getenv("RESEARCHER_DATA_DIR"))
	if dataDir == "" {
		dataDir = defaultDataDir
	}
	docsDir := strings.TrimSpace(os.Getenv("RESEARCHER_DOCS_DIR"))
	if docsDir == "" {
		docsDir = defaultDocsDir
	}
	return runSmoke(ctx, w, dataDir, docsDir, time.Now().UTC(), runtimeprobes.Run)
}

func runSmoke(ctx context.Context, w io.Writer, dataDir, docsDir string, now time.Time, runtimeStage stageRunner) error {
	prevDataDir, hadDataDir := os.LookupEnv("RESEARCHER_DATA_DIR")
	if err := os.Setenv("RESEARCHER_DATA_DIR", dataDir); err != nil {
		return err
	}
	defer func() {
		if hadDataDir {
			_ = os.Setenv("RESEARCHER_DATA_DIR", prevDataDir)
			return
		}
		_ = os.Unsetenv("RESEARCHER_DATA_DIR")
	}()

	if err := staticanalysis.Run(ctx, w); err != nil {
		return err
	}
	if err := runtimeStage(ctx, w); err != nil {
		return err
	}
	if err := materializeRuntimeArtifactsFromCapturedData(ctx, dataDir, now); err != nil {
		return err
	}
	if err := correlation.Run(ctx, w); err != nil {
		return err
	}
	if err := reporting.Run(ctx, w); err != nil {
		return err
	}
	if err := validation.Run(ctx, w); err != nil {
		return err
	}

	dateRoot, err := newestDateRoot(dataDir)
	if err != nil {
		return err
	}
	metrics, err := collectBaselineMetrics(dateRoot)
	if err != nil {
		return err
	}
	reportPath, err := writeSmokeReport(docsDir, now, dateRoot, metrics)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "smoke-test-report: %s\n", filepath.ToSlash(reportPath))
	return err
}

func materializeRuntimeArtifactsFromCapturedData(ctx context.Context, dataDir string, now time.Time) error {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if !entry.IsDir() {
			continue
		}
		dateRoot := filepath.Join(dataDir, entry.Name())
		tracePath := filepath.Join(dateRoot, "mapper/runtime-probes/runtime-trace.jsonl")
		sessionPath := filepath.Join(dateRoot, "mapper/runtime-probes/trace-sessions.jsonl")

		traceCount, err := countJSONLLines(tracePath)
		if err != nil {
			return err
		}
		sessionCount, err := countJSONLLines(sessionPath)
		if err != nil {
			return err
		}

		var traces []runtimeTraceRecord
		if traceCount == 0 {
			var buildErr error
			traces, buildErr = buildRuntimeArtifactsFromCapturedDate(dateRoot, now)
			if buildErr != nil {
				return buildErr
			}
			if len(traces) == 0 {
				continue
			}
			if err := writeJSONL(tracePath, traces); err != nil {
				return err
			}
		} else if sessionCount == 0 {
			var readErr error
			traces, readErr = readJSONL[runtimeTraceRecord](tracePath)
			if readErr != nil {
				return readErr
			}
		}

		if sessionCount == 0 {
			sessions := buildSessionsFromTraces(traces)
			if len(sessions) == 0 {
				continue
			}
			if err := writeJSONL(sessionPath, sessions); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildRuntimeArtifactsFromCapturedDate(dateRoot string, now time.Time) ([]runtimeTraceRecord, error) {
	traces := make([]runtimeTraceRecord, 0, 64)
	seq := 0

	err := filepath.WalkDir(dateRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path == filepath.Join(dateRoot, "mapper") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".jsonl" {
			return nil
		}
		norm := filepath.ToSlash(path)
		switch {
		case strings.Contains(norm, "/http/"):
			records, err := parseHTTPCaptures(path, now, &seq)
			if err != nil {
				return err
			}
			traces = append(traces, records...)
		case strings.Contains(norm, "/websocket/"):
			records, err := parseWebSocketCaptures(path, now, &seq)
			if err != nil {
				return err
			}
			traces = append(traces, records...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(traces, func(i, j int) bool {
		if traces[i].Timestamp.Equal(traces[j].Timestamp) {
			return traces[i].TraceID < traces[j].TraceID
		}
		return traces[i].Timestamp.Before(traces[j].Timestamp)
	})

	return traces, nil
}

func parseHTTPCaptures(path string, fallbackNow time.Time, seq *int) ([]runtimeTraceRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	out := make([]runtimeTraceRecord, 0, 32)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var capture httpCapture
		if err := json.Unmarshal([]byte(line), &capture); err != nil {
			return nil, fmt.Errorf("decode http capture %s: %w", path, err)
		}
		if isStaticResourceURL(capture.URL) {
			continue
		}
		*seq = *seq + 1
		tabID := strings.TrimSpace(capture.TabID)
		if tabID == "" {
			tabID = "unknown-tab"
		}
		reqID := strings.TrimSpace(capture.RequestID)
		if reqID == "" {
			reqID = fmt.Sprintf("http-%d", *seq)
		}
		ts := capture.Timestamp.UTC()
		if ts.IsZero() {
			ts = fallbackNow
		}
		out = append(out, runtimeTraceRecord{
			Timestamp: ts,
			TraceID:   tabID + ":fetch:" + reqID,
			TabID:     tabID,
			TabURL:    capture.URL,
			Surface:   "fetch",
			EventType: "request",
			Payload: map[string]any{
				"method": capture.Method,
				"source": "captured_http",
			},
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func parseWebSocketCaptures(path string, fallbackNow time.Time, seq *int) ([]runtimeTraceRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	out := make([]runtimeTraceRecord, 0, 16)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var capture websocketCapture
		if err := json.Unmarshal([]byte(line), &capture); err != nil {
			return nil, fmt.Errorf("decode websocket capture %s: %w", path, err)
		}
		*seq = *seq + 1
		tabID := strings.TrimSpace(capture.TabID)
		if tabID == "" {
			tabID = "unknown-tab"
		}
		reqID := strings.TrimSpace(capture.RequestID)
		if reqID == "" {
			reqID = fmt.Sprintf("ws-%d", *seq)
		}
		ts := capture.Timestamp.UTC()
		if ts.IsZero() {
			ts = fallbackNow
		}
		eventType := strings.TrimSpace(capture.EventType)
		if eventType == "" {
			eventType = "event"
		}
		out = append(out, runtimeTraceRecord{
			Timestamp: ts,
			TraceID:   tabID + ":websocket:" + reqID,
			TabID:     tabID,
			TabURL:    capture.URL,
			Surface:   "websocket",
			EventType: eventType,
			Payload: map[string]any{
				"direction": capture.Direction,
				"source":    "captured_websocket",
			},
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func buildSessionsFromTraces(traces []runtimeTraceRecord) []runtimeTraceSession {
	if len(traces) == 0 {
		return nil
	}

	type acc struct {
		profile string
		start   time.Time
		end     time.Time
	}
	groups := make(map[string]acc)

	for _, rec := range traces {
		tabID := strings.TrimSpace(rec.TabID)
		if tabID == "" {
			tabID = "unknown-tab"
		}
		ts := rec.Timestamp.UTC()
		if ts.IsZero() {
			continue
		}
		profile := profileForURL(rec.TabURL)
		current, ok := groups[tabID]
		if !ok {
			groups[tabID] = acc{profile: profile, start: ts, end: ts}
			continue
		}
		if ts.Before(current.start) {
			current.start = ts
		}
		if ts.After(current.end) {
			current.end = ts
		}
		groups[tabID] = current
	}

	out := make([]runtimeTraceSession, 0, len(groups))
	for tabID, session := range groups {
		out = append(out, runtimeTraceSession{
			SessionID:   fmt.Sprintf("%s:%s:%d", tabID, session.profile, session.start.UnixNano()),
			ProfileName: session.profile,
			TabID:       tabID,
			StartedAt:   session.start,
			EndedAt:     session.end,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].TabID == out[j].TabID {
			return out[i].ProfileName < out[j].ProfileName
		}
		return out[i].TabID < out[j].TabID
	})
	return out
}

func profileForURL(rawURL string) string {
	lower := strings.ToLower(rawURL)
	switch {
	case strings.Contains(lower, "watchlist"):
		return "watchlist_edits"
	case strings.Contains(lower, "replay"):
		return "replay_actions"
	case strings.Contains(lower, "trading") && strings.Contains(lower, "panel"):
		return "trading_panel_interactions"
	default:
		return "chart_interaction"
	}
}

func isStaticResourceURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	if strings.Contains(lower, "/static/bundles/") || strings.Contains(lower, "/static/images/") {
		return true
	}
	switch {
	case strings.HasSuffix(lower, ".js"),
		strings.HasSuffix(lower, ".css"),
		strings.HasSuffix(lower, ".png"),
		strings.HasSuffix(lower, ".svg"),
		strings.HasSuffix(lower, ".jpg"),
		strings.HasSuffix(lower, ".jpeg"),
		strings.HasSuffix(lower, ".gif"),
		strings.HasSuffix(lower, ".webp"),
		strings.HasSuffix(lower, ".ico"),
		strings.HasSuffix(lower, ".woff"),
		strings.HasSuffix(lower, ".woff2"),
		strings.HasSuffix(lower, ".ttf"),
		strings.HasSuffix(lower, ".map"):
		return true
	default:
		return false
	}
}

func newestDateRoot(dataDir string) (string, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return "", err
	}
	dirs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	if len(dirs) == 0 {
		return "", fmt.Errorf("no captured date directories in %s", dataDir)
	}
	sort.Strings(dirs)
	return filepath.Join(dataDir, dirs[len(dirs)-1]), nil
}

func collectBaselineMetrics(dateRoot string) (baselineMetrics, error) {
	bundleCount, err := countJSONLLines(filepath.Join(dateRoot, "mapper/static-analysis/js-bundle-index.jsonl"))
	if err != nil {
		return baselineMetrics{}, err
	}
	traceSessionCount, err := countJSONLLines(filepath.Join(dateRoot, "mapper/runtime-probes/trace-sessions.jsonl"))
	if err != nil {
		return baselineMetrics{}, err
	}
	correlatedCount, err := countJSONLLines(filepath.Join(dateRoot, "mapper/correlation/capability-correlations.jsonl"))
	if err != nil {
		return baselineMetrics{}, err
	}

	correlations, err := readJSONL[correlationRecord](filepath.Join(dateRoot, "mapper/correlation/capability-correlations.jsonl"))
	if err != nil {
		return baselineMetrics{}, err
	}
	dist := map[string]int{}
	for _, rec := range correlations {
		dist[confidenceBucket(rec.ConfidenceScore)]++
	}

	return baselineMetrics{
		BundleCount:               bundleCount,
		TraceSessionCount:         traceSessionCount,
		CorrelatedCapabilityCount: correlatedCount,
		ConfidenceDistribution:    dist,
	}, nil
}

func confidenceBucket(score float64) string {
	switch {
	case score >= 0.85:
		return "high (>=0.85)"
	case score >= 0.60:
		return "medium (0.60-0.84)"
	case score >= 0.35:
		return "low (0.35-0.59)"
	default:
		return "very-low (<0.35)"
	}
}

func writeSmokeReport(docsDir string, runAt time.Time, dateRoot string, metrics baselineMetrics) (string, error) {
	if docsDir == "" {
		docsDir = defaultDocsDir
	}
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		return "", err
	}

	reportName := fmt.Sprintf("mapper-smoke-test-%s.md", runAt.UTC().Format("20060102T150405Z"))
	reportPath := filepath.Join(docsDir, reportName)

	relDateRoot := filepath.ToSlash(filepath.Clean(dateRoot))
	artifacts := []string{
		relDateRoot + "/mapper/static-analysis/js-bundle-index.jsonl",
		relDateRoot + "/mapper/runtime-probes/runtime-trace.jsonl",
		relDateRoot + "/mapper/runtime-probes/trace-sessions.jsonl",
		relDateRoot + "/mapper/correlation/capability-correlations.jsonl",
		relDateRoot + "/mapper/reporting/capability-matrix.jsonl",
		relDateRoot + "/mapper/reporting/capability-matrix-summary.md",
	}

	keys := make([]string, 0, len(metrics.ConfidenceDistribution))
	for key := range metrics.ConfidenceDistribution {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	f, err := os.Create(reportPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	_, _ = fmt.Fprintln(w, "# Mapper Smoke Test Report")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "- Run timestamp (UTC): `%s`\n", runAt.UTC().Format(time.RFC3339))
	_, _ = fmt.Fprintf(w, "- Captured date root: `%s`\n", relDateRoot)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "## Baseline Metrics")
	_, _ = fmt.Fprintf(w, "- Bundle count: `%d`\n", metrics.BundleCount)
	_, _ = fmt.Fprintf(w, "- Trace session count: `%d`\n", metrics.TraceSessionCount)
	_, _ = fmt.Fprintf(w, "- Correlated capability count: `%d`\n", metrics.CorrelatedCapabilityCount)
	_, _ = fmt.Fprintln(w, "- Confidence distribution:")
	if len(keys) == 0 {
		_, _ = fmt.Fprintln(w, "  - `none`: `0`")
	} else {
		for _, key := range keys {
			_, _ = fmt.Fprintf(w, "  - `%s`: `%d`\n", key, metrics.ConfidenceDistribution[key])
		}
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "## Artifact Paths")
	for _, artifact := range artifacts {
		_, _ = fmt.Fprintf(w, "- `%s`\n", artifact)
	}
	if err := w.Flush(); err != nil {
		return "", err
	}
	return reportPath, nil
}

func countJSONLLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 10*1024*1024)
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func writeJSONL[T any](path string, records []T) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	enc := json.NewEncoder(w)
	for _, rec := range records {
		if err := enc.Encode(rec); err != nil {
			return err
		}
	}
	return w.Flush()
}

func readJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 10*1024*1024)
	out := make([]T, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec T
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
