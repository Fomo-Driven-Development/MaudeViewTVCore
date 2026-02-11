package correlation

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultDataDir              = "./research_data"
	staticAnalysisRelativePath  = "mapper/static-analysis/js-bundle-analysis.jsonl"
	staticGraphRelativePath     = "mapper/static-analysis/js-bundle-dependency-graph.jsonl"
	runtimeTraceRelativePath    = "mapper/runtime-probes/runtime-trace.jsonl"
	traceSessionsRelativePath   = "mapper/runtime-probes/trace-sessions.jsonl"
	correlationOutputRelative   = "mapper/correlation/capability-correlations.jsonl"
	temporalLinkageSlackSeconds = 2
)

type staticSignalAnchor struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type staticAnalysisRecord struct {
	PrimaryKey    string               `json:"primary_key"`
	FilePath      string               `json:"file_path"`
	ChunkName     string               `json:"chunk_name"`
	SignalAnchors []staticSignalAnchor `json:"signal_anchors"`
}

type staticDomainHint struct {
	Domain    string `json:"domain"`
	Rationale string `json:"rationale"`
}

type staticGraphNode struct {
	PrimaryKey  string             `json:"primary_key"`
	FilePath    string             `json:"file_path"`
	ChunkName   string             `json:"chunk_name"`
	DomainHints []staticDomainHint `json:"domain_hints"`
}

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

type evidenceLink struct {
	Source   string `json:"source"`
	Path     string `json:"path"`
	RecordID string `json:"record_id"`
	Field    string `json:"field"`
	Value    string `json:"value"`
}

type capabilityCorrelationRecord struct {
	Capability                    string         `json:"capability"`
	PrimaryRecommendedControlPath string         `json:"primary_recommended_control_path"`
	ControlPathRationale          string         `json:"control_path_rationale"`
	ConfidenceScore               float64        `json:"confidence_score"`
	ConfidenceRationale           []string       `json:"confidence_rationale"`
	TemporalLinkageTraceIDs       []string       `json:"temporal_linkage_trace_ids"`
	EvidenceLinks                 []evidenceLink `json:"evidence_links"`
}

type anchorCandidate struct {
	anchorType string
	value      string
}

// Run executes the correlation stage.
func Run(ctx context.Context, w io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	dataDir := os.Getenv("RESEARCHER_DATA_DIR")
	if strings.TrimSpace(dataDir) == "" {
		dataDir = defaultDataDir
	}
	if err := correlateDataDir(ctx, dataDir); err != nil {
		return err
	}

	_, err := fmt.Fprintln(w, "correlation: complete")
	return err
}

func correlateDataDir(ctx context.Context, dataDir string) error {
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
		if err := correlateDate(ctx, dateRoot); err != nil {
			return fmt.Errorf("correlate %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func correlateDate(ctx context.Context, dateRoot string) error {
	staticAnalysisPath := filepath.Join(dateRoot, staticAnalysisRelativePath)
	staticGraphPath := filepath.Join(dateRoot, staticGraphRelativePath)
	runtimeTracePath := filepath.Join(dateRoot, runtimeTraceRelativePath)
	traceSessionsPath := filepath.Join(dateRoot, traceSessionsRelativePath)

	staticAnalysisRecords, err := readJSONL[staticAnalysisRecord](staticAnalysisPath)
	if err != nil {
		return err
	}
	staticGraphNodes, err := readJSONL[staticGraphNode](staticGraphPath)
	if err != nil {
		return err
	}
	runtimeTraceRecords, err := readJSONL[runtimeTraceRecord](runtimeTracePath)
	if err != nil {
		return err
	}
	traceSessions, err := readJSONL[runtimeTraceSession](traceSessionsPath)
	if err != nil {
		return err
	}

	records := buildCapabilityCorrelations(
		ctx,
		dateRoot,
		staticAnalysisPath,
		staticGraphPath,
		runtimeTracePath,
		traceSessionsPath,
		staticAnalysisRecords,
		staticGraphNodes,
		runtimeTraceRecords,
		traceSessions,
	)

	outputPath := filepath.Join(dateRoot, correlationOutputRelative)
	return writeJSONL(outputPath, records)
}

func buildCapabilityCorrelations(
	ctx context.Context,
	dateRoot, staticAnalysisPath, staticGraphPath, runtimeTracePath, traceSessionsPath string,
	staticAnalysisRecords []staticAnalysisRecord,
	staticGraphNodes []staticGraphNode,
	runtimeTraceRecords []runtimeTraceRecord,
	traceSessions []runtimeTraceSession,
) []capabilityCorrelationRecord {
	analysisByPrimaryKey := make(map[string]staticAnalysisRecord, len(staticAnalysisRecords))
	for _, rec := range staticAnalysisRecords {
		analysisByPrimaryKey[rec.PrimaryKey] = rec
	}

	capabilitySet := make(map[string]struct{})
	for _, session := range traceSessions {
		capabilitySet[mapProfileToCapability(session.ProfileName)] = struct{}{}
	}
	for _, node := range staticGraphNodes {
		for _, hint := range node.DomainHints {
			capabilitySet[strings.ToLower(strings.TrimSpace(hint.Domain))] = struct{}{}
		}
	}

	capabilities := make([]string, 0, len(capabilitySet))
	for capability := range capabilitySet {
		if capability == "" {
			continue
		}
		capabilities = append(capabilities, capability)
	}
	sort.Strings(capabilities)

	records := make([]capabilityCorrelationRecord, 0, len(capabilities))
	for _, capability := range capabilities {
		select {
		case <-ctx.Done():
			return records
		default:
		}

		sessions := sessionsForCapability(traceSessions, capability)
		linkedTraces := tracesFromTemporalLinkage(runtimeTraceRecords, sessions)
		evidence := make([]evidenceLink, 0, len(linkedTraces)*2)
		evidenceSet := map[string]struct{}{}

		appendEvidence := func(link evidenceLink) {
			key := link.Source + "\x00" + link.RecordID + "\x00" + link.Field + "\x00" + link.Value
			if _, ok := evidenceSet[key]; ok {
				return
			}
			evidenceSet[key] = struct{}{}
			evidence = append(evidence, link)
		}

		for _, session := range sessions {
			appendEvidence(evidenceLink{
				Source:   "runtime-session",
				Path:     filepath.ToSlash(strings.TrimPrefix(traceSessionsPath, dateRoot+"/")),
				RecordID: session.SessionID,
				Field:    "profile_name",
				Value:    session.ProfileName,
			})
		}

		anchorMatches := make(map[string]struct{})
		stackMatches := make(map[string]struct{})
		surfaceCounts := map[string]int{}
		temporalTraceIDs := make([]string, 0, len(linkedTraces))

		for _, trace := range linkedTraces {
			temporalTraceIDs = append(temporalTraceIDs, trace.TraceID)
			surfaceCounts[trace.Surface]++

			appendEvidence(evidenceLink{
				Source:   "runtime-trace",
				Path:     filepath.ToSlash(strings.TrimPrefix(runtimeTracePath, dateRoot+"/")),
				RecordID: trace.TraceID,
				Field:    "surface",
				Value:    trace.Surface,
			})
			appendEvidence(evidenceLink{
				Source:   "runtime-trace",
				Path:     filepath.ToSlash(strings.TrimPrefix(runtimeTracePath, dateRoot+"/")),
				RecordID: trace.TraceID,
				Field:    "event_type",
				Value:    trace.EventType,
			})
		}
		sort.Strings(temporalTraceIDs)

		candidateNodes := staticNodesForCapability(staticGraphNodes, capability)
		for _, node := range candidateNodes {
			appendEvidence(evidenceLink{
				Source:   "static-graph",
				Path:     filepath.ToSlash(strings.TrimPrefix(staticGraphPath, dateRoot+"/")),
				RecordID: node.PrimaryKey,
				Field:    "chunk_name",
				Value:    node.ChunkName,
			})
			for _, hint := range node.DomainHints {
				if strings.EqualFold(hint.Domain, capability) {
					appendEvidence(evidenceLink{
						Source:   "static-graph",
						Path:     filepath.ToSlash(strings.TrimPrefix(staticGraphPath, dateRoot+"/")),
						RecordID: node.PrimaryKey,
						Field:    "domain_hint",
						Value:    hint.Rationale,
					})
				}
			}
		}

		for _, trace := range linkedTraces {
			stackHints := extractStackHints(trace.Payload)
			for _, node := range staticGraphNodes {
				if matchesNodeFromStackHints(node, stackHints) {
					stackMatches[node.PrimaryKey] = struct{}{}
					appendEvidence(evidenceLink{
						Source:   "runtime-trace",
						Path:     filepath.ToSlash(strings.TrimPrefix(runtimeTracePath, dateRoot+"/")),
						RecordID: trace.TraceID,
						Field:    "stack_hint",
						Value:    summarizeStackHints(stackHints),
					})
					appendEvidence(evidenceLink{
						Source:   "static-graph",
						Path:     filepath.ToSlash(strings.TrimPrefix(staticGraphPath, dateRoot+"/")),
						RecordID: node.PrimaryKey,
						Field:    "stack_resolved_chunk",
						Value:    node.ChunkName,
					})
				}
			}

			runtimeAnchors := extractRuntimeAnchors(trace.Payload)
			for _, anchor := range runtimeAnchors {
				for _, staticRec := range staticAnalysisRecords {
					for _, staticAnchor := range staticRec.SignalAnchors {
						if strings.EqualFold(anchor.anchorType, staticAnchor.Type) && strings.EqualFold(anchor.value, staticAnchor.Value) {
							key := staticRec.PrimaryKey + "\x00" + staticAnchor.Type + "\x00" + staticAnchor.Value
							anchorMatches[key] = struct{}{}
							appendEvidence(evidenceLink{
								Source:   "runtime-trace",
								Path:     filepath.ToSlash(strings.TrimPrefix(runtimeTracePath, dateRoot+"/")),
								RecordID: trace.TraceID,
								Field:    "anchor_match",
								Value:    anchor.anchorType + ":" + anchor.value,
							})
							appendEvidence(evidenceLink{
								Source:   "static-analysis",
								Path:     filepath.ToSlash(strings.TrimPrefix(staticAnalysisPath, dateRoot+"/")),
								RecordID: staticRec.PrimaryKey,
								Field:    "signal_anchor",
								Value:    staticAnchor.Type + ":" + staticAnchor.Value,
							})
						}
					}
				}
			}
		}

		controlPath, controlRationale := choosePrimaryControlPath(surfaceCounts, len(anchorMatches), len(stackMatches), candidateNodes)
		score, rationale := scoreConfidence(len(sessions), len(candidateNodes), len(temporalTraceIDs), len(anchorMatches), len(stackMatches))
		sort.Slice(evidence, func(i, j int) bool {
			if evidence[i].Source == evidence[j].Source {
				if evidence[i].RecordID == evidence[j].RecordID {
					if evidence[i].Field == evidence[j].Field {
						return evidence[i].Value < evidence[j].Value
					}
					return evidence[i].Field < evidence[j].Field
				}
				return evidence[i].RecordID < evidence[j].RecordID
			}
			return evidence[i].Source < evidence[j].Source
		})

		record := capabilityCorrelationRecord{
			Capability:                    capability,
			PrimaryRecommendedControlPath: controlPath,
			ControlPathRationale:          controlRationale,
			ConfidenceScore:               score,
			ConfidenceRationale:           rationale,
			TemporalLinkageTraceIDs:       temporalTraceIDs,
			EvidenceLinks:                 evidence,
		}

		if len(record.EvidenceLinks) == 0 {
			for _, node := range candidateNodes {
				analysis, ok := analysisByPrimaryKey[node.PrimaryKey]
				if !ok {
					continue
				}
				record.EvidenceLinks = append(record.EvidenceLinks, evidenceLink{
					Source:   "static-analysis",
					Path:     filepath.ToSlash(strings.TrimPrefix(staticAnalysisPath, dateRoot+"/")),
					RecordID: analysis.PrimaryKey,
					Field:    "chunk_name",
					Value:    analysis.ChunkName,
				})
			}
		}

		records = append(records, record)
	}

	return records
}

func readJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	out := make([]T, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec T
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		out = append(out, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func writeJSONL(path string, records []capabilityCorrelationRecord) error {
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

func mapProfileToCapability(profile string) string {
	switch profile {
	case "chart_interaction":
		return "chart"
	case "watchlist_edits":
		return "watchlist"
	case "study_add_remove":
		return "studies"
	case "replay_actions":
		return "replay"
	case "trading_panel_interactions":
		return "trading"
	default:
		p := strings.TrimSpace(strings.ToLower(profile))
		if p == "" {
			return "unknown"
		}
		return p
	}
}

func sessionsForCapability(sessions []runtimeTraceSession, capability string) []runtimeTraceSession {
	out := make([]runtimeTraceSession, 0)
	for _, session := range sessions {
		if mapProfileToCapability(session.ProfileName) == capability {
			out = append(out, session)
		}
	}
	return out
}

func tracesFromTemporalLinkage(traces []runtimeTraceRecord, sessions []runtimeTraceSession) []runtimeTraceRecord {
	if len(traces) == 0 || len(sessions) == 0 {
		return nil
	}
	out := make([]runtimeTraceRecord, 0)
	seen := map[string]struct{}{}
	for _, trace := range traces {
		for _, session := range sessions {
			start := session.StartedAt.Add(-temporalLinkageSlackSeconds * time.Second)
			end := session.EndedAt.Add(temporalLinkageSlackSeconds * time.Second)
			if session.TabID != "" && trace.TabID != "" && session.TabID != trace.TabID {
				continue
			}
			if !trace.Timestamp.IsZero() && (trace.Timestamp.Before(start) || trace.Timestamp.After(end)) {
				continue
			}
			if _, ok := seen[trace.TraceID]; ok {
				continue
			}
			seen[trace.TraceID] = struct{}{}
			out = append(out, trace)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].TraceID < out[j].TraceID
	})
	return out
}

func staticNodesForCapability(nodes []staticGraphNode, capability string) []staticGraphNode {
	out := make([]staticGraphNode, 0)
	for _, node := range nodes {
		matched := false
		for _, hint := range node.DomainHints {
			if strings.EqualFold(hint.Domain, capability) {
				matched = true
				break
			}
		}
		if !matched {
			lowerPath := strings.ToLower(node.FilePath + " " + node.ChunkName)
			if strings.Contains(lowerPath, strings.ToLower(capability)) {
				matched = true
			}
		}
		if matched {
			out = append(out, node)
		}
	}
	return out
}

func extractRuntimeAnchors(payload map[string]any) []anchorCandidate {
	if len(payload) == 0 {
		return nil
	}
	stringsInPayload := make([]string, 0, 8)
	collectPayloadStrings(payload, &stringsInPayload)
	anchors := make([]anchorCandidate, 0, len(stringsInPayload))
	seen := map[string]struct{}{}
	for _, s := range stringsInPayload {
		candidate := strings.TrimSpace(s)
		if candidate == "" {
			continue
		}
		lower := strings.ToLower(candidate)
		switch {
		case strings.HasPrefix(candidate, "/api"), strings.Contains(lower, "/api/"), strings.Contains(lower, "/graphql"):
			key := "api_route\x00" + candidate
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			anchors = append(anchors, anchorCandidate{anchorType: "api_route", value: candidate})
		case strings.HasPrefix(lower, "ws://"), strings.HasPrefix(lower, "wss://"), strings.Contains(lower, "/ws/"), strings.Contains(lower, "websocket"):
			key := "websocket_channel\x00" + candidate
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			anchors = append(anchors, anchorCandidate{anchorType: "websocket_channel", value: candidate})
		case looksLikeActionEvent(candidate):
			key := "action_event\x00" + candidate
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			anchors = append(anchors, anchorCandidate{anchorType: "action_event", value: candidate})
		case looksLikeFeatureFlag(candidate):
			key := "feature_flag\x00" + candidate
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			anchors = append(anchors, anchorCandidate{anchorType: "feature_flag", value: candidate})
		}
	}
	return anchors
}

func collectPayloadStrings(value any, out *[]string) {
	switch v := value.(type) {
	case map[string]any:
		for k, nested := range v {
			*out = append(*out, k)
			collectPayloadStrings(nested, out)
		}
	case []any:
		for _, nested := range v {
			collectPayloadStrings(nested, out)
		}
	case string:
		*out = append(*out, v)
	}
}

func looksLikeActionEvent(s string) bool {
	lower := strings.ToLower(s)
	return strings.Contains(lower, "action") || strings.Contains(lower, "event") || strings.Contains(s, ":")
}

func looksLikeFeatureFlag(s string) bool {
	upper := strings.ToUpper(s)
	return strings.HasPrefix(upper, "FEATURE_") || strings.HasPrefix(upper, "FF_") || strings.HasPrefix(upper, "ENABLE_")
}

func extractStackHints(payload map[string]any) []string {
	if len(payload) == 0 {
		return nil
	}
	keys := []string{"stack", "stack_trace", "source", "file", "module", "chunk", "url"}
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
			out = append(out, text)
		}
	}
	return out
}

func matchesNodeFromStackHints(node staticGraphNode, stackHints []string) bool {
	if len(stackHints) == 0 {
		return false
	}
	needleParts := []string{
		strings.ToLower(filepath.Base(node.FilePath)),
		strings.ToLower(node.ChunkName),
		strings.ToLower(node.PrimaryKey),
	}
	for _, hint := range stackHints {
		lowerHint := strings.ToLower(hint)
		for _, needle := range needleParts {
			if needle == "" {
				continue
			}
			if strings.Contains(lowerHint, needle) {
				return true
			}
		}
	}
	return false
}

func summarizeStackHints(hints []string) string {
	if len(hints) == 0 {
		return ""
	}
	if len(hints) == 1 {
		return hints[0]
	}
	return hints[0] + " | " + hints[1]
}

func choosePrimaryControlPath(surfaceCounts map[string]int, anchorMatches, stackMatches int, staticNodes []staticGraphNode) (string, string) {
	websocketCount := surfaceCounts["websocket"]
	eventBusCount := surfaceCounts["event_bus"]
	apiCount := surfaceCounts["fetch"] + surfaceCounts["xhr"]

	switch {
	case websocketCount > apiCount && websocketCount > eventBusCount:
		return "runtime.websocket_flow", "websocket runtime traces dominate temporal linkage"
	case apiCount > 0 && apiCount >= eventBusCount:
		return "runtime.api_request_path", "fetch/xhr runtime traces dominate and align with API/event anchors"
	case eventBusCount > 0:
		return "runtime.event_bus_path", "event bus runtime traces provide the strongest control-surface signal"
	case stackMatches > 0:
		return "static.stack_resolved_chunk_path", "no runtime dominance; stack hints resolved into static chunks"
	case len(staticNodes) > 0 || anchorMatches > 0:
		return "static.domain_chunk_entry_path", "static chunk/domain evidence available without stronger runtime ordering"
	default:
		return "unresolved", "insufficient correlated runtime/static evidence"
	}
}

func scoreConfidence(sessionCount, staticNodesCount, temporalCount, anchorMatches, stackMatches int) (float64, []string) {
	score := 0.0
	rationale := make([]string, 0, 6)

	if sessionCount > 0 {
		score += 0.2
		rationale = append(rationale, fmt.Sprintf("runtime session alignment present (%d session records)", sessionCount))
	}
	if staticNodesCount > 0 {
		score += 0.15
		rationale = append(rationale, fmt.Sprintf("static domain/chunk evidence present (%d nodes)", staticNodesCount))
	}
	if temporalCount > 0 {
		score += 0.25
		rationale = append(rationale, fmt.Sprintf("temporal linkage connected %d trace events to capability sessions", temporalCount))
	}
	if anchorMatches > 0 {
		score += 0.2
		rationale = append(rationale, fmt.Sprintf("anchor correlation matched %d runtime-to-static anchors", anchorMatches))
	}
	if stackMatches > 0 {
		score += 0.2
		rationale = append(rationale, fmt.Sprintf("stack hints resolved to %d static module/chunk records", stackMatches))
	}
	if anchorMatches >= 2 {
		score += 0.05
		rationale = append(rationale, "multiple anchor matches increased cross-source reliability")
	}
	if temporalCount >= 5 {
		score += 0.05
		rationale = append(rationale, "high temporal trace volume improved confidence")
	}
	if score > 1 {
		score = 1
	}
	if len(rationale) == 0 {
		rationale = append(rationale, "confidence is low because cross-source evidence was not found")
	}
	return score, rationale
}
