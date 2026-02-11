package reporting

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
	defaultDataDir           = "./research_data"
	correlationInputRelative = "mapper/correlation/capability-correlations.jsonl"
	matrixOutputRelative     = "mapper/reporting/capability-matrix.jsonl"
	matrixSchemaOutputPath   = "mapper/reporting/capability-matrix.schema.json"
	summaryReportOutputPath  = "mapper/reporting/capability-matrix-summary.md"
	matrixSchemaVersion      = "1.0"
)

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
	TemporalLinkageTraceIDs       []string       `json:"temporal_linkage_trace_ids"`
	EvidenceLinks                 []evidenceLink `json:"evidence_links"`
}

type capabilityMatrixRecord struct {
	CapabilityID             string   `json:"capability_id"`
	Feature                  string   `json:"feature"`
	FeatureName              string   `json:"feature_name"`
	CandidateEntryModules    []string `json:"candidate_entry_modules"`
	ModuleRef                string   `json:"module_ref"`
	TraceID                  string   `json:"trace_id"`
	Confidence               float64  `json:"confidence"`
	Preconditions            []string `json:"preconditions"`
	RiskNotes                []string `json:"risk_notes"`
	RecommendedControlMethod string   `json:"recommended_control_method"`
	RecommendedControlPath   string   `json:"recommended_control_path"`
	EvidenceIDs              []string `json:"evidence_ids"`
}

type evidenceCatalogRecord struct {
	EvidenceID   string `json:"evidence_id"`
	CapabilityID string `json:"capability_id"`
	Source       string `json:"source"`
	Path         string `json:"path"`
	RecordID     string `json:"record_id"`
	Field        string `json:"field"`
	Value        string `json:"value"`
}

type matrixSchema struct {
	SchemaVersion string   `json:"schema_version"`
	Required      []string `json:"required"`
}

// Run executes the reporting stage.
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
	if err := generateReports(ctx, dataDir); err != nil {
		return err
	}

	_, err := fmt.Fprintln(w, "reporting: complete")
	return err
}

func generateReports(ctx context.Context, dataDir string) error {
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
		if err := generateDateReport(dateRoot); err != nil {
			return fmt.Errorf("report %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func generateDateReport(dateRoot string) error {
	correlationPath := filepath.Join(dateRoot, correlationInputRelative)
	correlations, err := readJSONL[capabilityCorrelationRecord](correlationPath)
	if err != nil {
		return err
	}

	matrixRows := make([]capabilityMatrixRecord, 0, len(correlations))
	evidenceCatalog := make([]evidenceCatalogRecord, 0)
	for _, rec := range correlations {
		row, catalogRows := buildMatrixRow(rec)
		matrixRows = append(matrixRows, row)
		evidenceCatalog = append(evidenceCatalog, catalogRows...)
	}

	matrixPath := filepath.Join(dateRoot, matrixOutputRelative)
	if err := writeJSONL(matrixPath, matrixRows); err != nil {
		return err
	}

	schemaPath := filepath.Join(dateRoot, matrixSchemaOutputPath)
	schema := matrixSchema{
		SchemaVersion: matrixSchemaVersion,
		Required: []string{
			"capability_id",
			"feature",
			"feature_name",
			"candidate_entry_modules",
			"module_ref",
			"trace_id",
			"confidence",
			"preconditions",
			"risk_notes",
			"recommended_control_method",
			"recommended_control_path",
			"evidence_ids",
		},
	}
	if err := writeJSON(schemaPath, schema); err != nil {
		return err
	}

	reportPath := filepath.Join(dateRoot, summaryReportOutputPath)
	return writeMarkdownReport(reportPath, dateRoot, matrixRows, evidenceCatalog)
}

func buildMatrixRow(rec capabilityCorrelationRecord) (capabilityMatrixRecord, []evidenceCatalogRecord) {
	capabilityID := strings.ToLower(strings.TrimSpace(rec.Capability))
	featureName := toFeatureName(capabilityID)
	entryModules := candidateEntryModules(rec.EvidenceLinks)
	preconditions := buildPreconditions(rec.PrimaryRecommendedControlPath)
	riskNotes := buildRiskNotes(rec)
	controlMethod := mapControlMethod(rec.PrimaryRecommendedControlPath)

	evidenceIDs := make([]string, 0, len(rec.EvidenceLinks))
	evidenceCatalog := make([]evidenceCatalogRecord, 0, len(rec.EvidenceLinks))
	prefix := normalizeIDPrefix(capabilityID)
	for i, link := range rec.EvidenceLinks {
		evidenceID := fmt.Sprintf("%s-E%03d", prefix, i+1)
		evidenceIDs = append(evidenceIDs, evidenceID)
		evidenceCatalog = append(evidenceCatalog, evidenceCatalogRecord{
			EvidenceID:   evidenceID,
			CapabilityID: capabilityID,
			Source:       link.Source,
			Path:         link.Path,
			RecordID:     link.RecordID,
			Field:        link.Field,
			Value:        link.Value,
		})
	}

	moduleRef := ""
	if len(entryModules) > 0 {
		moduleRef = entryModules[0]
	}

	traceID := ""
	if len(rec.TemporalLinkageTraceIDs) > 0 {
		traceID = rec.TemporalLinkageTraceIDs[0]
	}

	return capabilityMatrixRecord{
		CapabilityID:             capabilityID,
		Feature:                  featureName,
		FeatureName:              featureName,
		CandidateEntryModules:    entryModules,
		ModuleRef:                moduleRef,
		TraceID:                  traceID,
		Confidence:               rec.ConfidenceScore,
		Preconditions:            preconditions,
		RiskNotes:                riskNotes,
		RecommendedControlMethod: controlMethod,
		RecommendedControlPath:   rec.PrimaryRecommendedControlPath,
		EvidenceIDs:              evidenceIDs,
	}, evidenceCatalog
}

func toFeatureName(capability string) string {
	if capability == "" {
		return "Unknown"
	}
	return strings.ToUpper(capability[:1]) + capability[1:]
}

func normalizeIDPrefix(capabilityID string) string {
	if capabilityID == "" {
		return "UNKNOWN"
	}
	var b strings.Builder
	for _, r := range strings.ToUpper(capabilityID) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

func candidateEntryModules(links []evidenceLink) []string {
	moduleSet := make(map[string]struct{})
	for _, link := range links {
		if link.Source != "static-analysis" && link.Source != "static-graph" {
			continue
		}
		if link.Field == "chunk_name" || link.Field == "stack_resolved_chunk" {
			if value := strings.TrimSpace(link.Value); value != "" {
				moduleSet[value] = struct{}{}
			}
		}
		if strings.HasSuffix(strings.ToLower(link.RecordID), ".js") {
			moduleSet[link.RecordID] = struct{}{}
		}
	}
	out := make([]string, 0, len(moduleSet))
	for module := range moduleSet {
		out = append(out, module)
	}
	sort.Strings(out)
	return out
}

func mapControlMethod(path string) string {
	switch path {
	case "runtime.websocket_flow":
		return "network_drivable.websocket"
	case "runtime.api_request_path":
		return "network_drivable.http_api"
	case "runtime.event_bus_path":
		return "internal_api_drivable.event_bus"
	case "static.stack_resolved_chunk_path":
		return "internal_api_drivable.static_stack"
	case "static.domain_chunk_entry_path":
		return "dom_drivable.entry_module"
	default:
		return "undetermined"
	}
}

func buildPreconditions(controlPath string) []string {
	base := []string{"Authenticated TradingView session captured in a profiled tab."}
	switch controlPath {
	case "runtime.websocket_flow":
		return append(base, "WebSocket probe instrumentation collected runtime flow evidence.")
	case "runtime.api_request_path":
		return append(base, "Fetch/XHR probe instrumentation captured API request traces.")
	case "runtime.event_bus_path":
		return append(base, "Event bus instrumentation captured dispatch or bus-call traces.")
	case "static.stack_resolved_chunk_path":
		return append(base, "Runtime stack hints and static chunk graph were both available.")
	case "static.domain_chunk_entry_path":
		return append(base, "Static domain hints identified entry modules for this capability.")
	default:
		return append(base, "Additional runtime/static evidence collection is required.")
	}
}

func buildRiskNotes(rec capabilityCorrelationRecord) []string {
	notes := make([]string, 0, 4)
	if rec.ConfidenceScore < 0.5 {
		notes = append(notes, "Low confidence score increases risk of false-positive control selection.")
	}
	if len(rec.TemporalLinkageTraceIDs) == 0 {
		notes = append(notes, "No temporal trace linkage was found for this capability.")
	}
	if len(rec.EvidenceLinks) < 2 {
		notes = append(notes, "Evidence volume is limited; collect more traces before applying controls.")
	}
	if strings.TrimSpace(rec.ControlPathRationale) != "" {
		notes = append(notes, rec.ControlPathRationale)
	}
	if len(notes) == 0 {
		notes = append(notes, "Control recommendation appears stable with current evidence coverage.")
	}
	return notes
}

func writeMarkdownReport(path, dateRoot string, rows []capabilityMatrixRecord, evidenceCatalog []evidenceCatalogRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	_, _ = fmt.Fprintf(w, "# Capability Matrix Summary\n\nGenerated: %s\n\n", time.Now().UTC().Format(time.RFC3339))
	_, _ = fmt.Fprintf(w, "Date root: `%s`\n\n", filepath.ToSlash(dateRoot))

	if len(rows) == 0 {
		_, _ = fmt.Fprintln(w, "No capability correlations were found.")
		return w.Flush()
	}

	for _, row := range rows {
		_, _ = fmt.Fprintf(w, "## %s (`%s`)\n\n", row.FeatureName, row.CapabilityID)
		_, _ = fmt.Fprintf(w, "- Recommended control method: `%s` (`%s`)\n", row.RecommendedControlMethod, row.RecommendedControlPath)
		_, _ = fmt.Fprintf(w, "- Confidence: `%.2f`\n", row.Confidence)
		_, _ = fmt.Fprintf(w, "- Candidate entry modules: %s\n", formatInlineList(row.CandidateEntryModules))
		_, _ = fmt.Fprintf(w, "- Preconditions: %s\n", formatSentenceList(row.Preconditions))
		_, _ = fmt.Fprintf(w, "- Risk notes: %s\n", formatSentenceList(row.RiskNotes))
		_, _ = fmt.Fprintf(w, "- Evidence IDs: %s\n\n", formatEvidenceLinks(row.EvidenceIDs))
	}

	_, _ = fmt.Fprintln(w, "## Evidence Index")
	_, _ = fmt.Fprintln(w)
	for _, evidence := range evidenceCatalog {
		anchorID := strings.ToLower("evidence-" + evidence.EvidenceID)
		_, _ = fmt.Fprintf(w, "### <a id=\"%s\"></a>`%s`\n\n", anchorID, evidence.EvidenceID)
		_, _ = fmt.Fprintf(w, "- Capability: `%s`\n", evidence.CapabilityID)
		_, _ = fmt.Fprintf(w, "- Source: `%s`\n", evidence.Source)
		_, _ = fmt.Fprintf(w, "- Path: `%s`\n", evidence.Path)
		_, _ = fmt.Fprintf(w, "- Record ID: `%s`\n", evidence.RecordID)
		_, _ = fmt.Fprintf(w, "- Field: `%s`\n", evidence.Field)
		_, _ = fmt.Fprintf(w, "- Value: `%s`\n\n", evidence.Value)
	}

	return w.Flush()
}

func formatInlineList(values []string) string {
	if len(values) == 0 {
		return "`none`"
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, "`"+value+"`")
	}
	return strings.Join(out, ", ")
}

func formatSentenceList(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, " ")
}

func formatEvidenceLinks(evidenceIDs []string) string {
	if len(evidenceIDs) == 0 {
		return "`none`"
	}
	parts := make([]string, 0, len(evidenceIDs))
	for _, id := range evidenceIDs {
		anchor := strings.ToLower("evidence-" + id)
		parts = append(parts, fmt.Sprintf("[%s](#%s)", id, anchor))
	}
	return strings.Join(parts, ", ")
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

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
