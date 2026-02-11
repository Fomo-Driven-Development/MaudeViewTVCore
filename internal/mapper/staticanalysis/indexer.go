package staticanalysis

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	defaultDataDir         = "./research_data"
	indexRelativeOutput    = "mapper/static-analysis/js-bundle-index.jsonl"
	analysisRelativeOutput = "mapper/static-analysis/js-bundle-analysis.jsonl"
	errorsRelativeOutput   = "mapper/static-analysis/js-bundle-analysis-errors.jsonl"
	graphRelativeOutput    = "mapper/static-analysis/js-bundle-dependency-graph.jsonl"
)

type jsBundleRecord struct {
	PrimaryKey string `json:"primary_key"`
	FilePath   string `json:"file_path"`
	SizeBytes  int64  `json:"size_bytes"`
	SHA256     string `json:"sha256"`
	ChunkName  string `json:"chunk_name"`
}

type jsSignalAnchor struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type jsBundleAnalysisRecord struct {
	PrimaryKey    string           `json:"primary_key"`
	FilePath      string           `json:"file_path"`
	ChunkName     string           `json:"chunk_name"`
	Functions     []string         `json:"functions"`
	Classes       []string         `json:"classes"`
	Exports       []string         `json:"exports"`
	ImportEdges   []string         `json:"import_edges"`
	RequireEdges  []string         `json:"require_edges"`
	SignalAnchors []jsSignalAnchor `json:"signal_anchors"`
}

type jsBundleAnalysisErrorRecord struct {
	PrimaryKey string `json:"primary_key"`
	FilePath   string `json:"file_path"`
	ChunkName  string `json:"chunk_name"`
	Error      string `json:"error"`
}

type jsBundleGraphDependency struct {
	Type               string `json:"type"`
	Target             string `json:"target"`
	ResolvedPrimaryKey string `json:"resolved_primary_key,omitempty"`
	ResolvedChunkName  string `json:"resolved_chunk_name,omitempty"`
}

type jsBundleSourceReference struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type jsDomainHint struct {
	Domain    string `json:"domain"`
	Rationale string `json:"rationale"`
}

type jsBundleGraphNode struct {
	PrimaryKey       string                    `json:"primary_key"`
	FilePath         string                    `json:"file_path"`
	ChunkName        string                    `json:"chunk_name"`
	Dependencies     []jsBundleGraphDependency `json:"dependencies"`
	SourceReferences []jsBundleSourceReference `json:"source_references"`
	DomainHints      []jsDomainHint            `json:"domain_hints"`
}

type jsBundleExtracted struct {
	Functions    []string
	Classes      []string
	Exports      []string
	ImportEdges  []string
	RequireEdges []string
	Anchors      []jsSignalAnchor
}

var (
	actionEventAnchorRE   = regexp.MustCompile(`(?i)(?:^|[._:-])(action|event)(?:[._:-]|$)`)
	featureFlagAnchorRE   = regexp.MustCompile(`(?i)(?:^|[._:-])(feature|flag|ff)(?:[._:-]|$)|^(FEATURE|FF|ENABLE)_[A-Z0-9_]+$`)
)

func indexJSBundles(ctx context.Context, dataDir string) error {
	recordsByTopLevel, err := collectJSBundleRecords(ctx, dataDir)
	if err != nil {
		return err
	}

	for topLevel, records := range recordsByTopLevel {
		if err := writeIndexFile(filepath.Join(dataDir, topLevel, indexRelativeOutput), records); err != nil {
			return err
		}
		analysisRecords, errorRecords := extractBundleAnalysisRecords(dataDir, records)
		if err := writeAnalysisFile(filepath.Join(dataDir, topLevel, analysisRelativeOutput), analysisRecords); err != nil {
			return err
		}
		if err := writeAnalysisErrorFile(filepath.Join(dataDir, topLevel, errorsRelativeOutput), errorRecords); err != nil {
			return err
		}
		graphNodes := buildBundleDependencyGraph(records, analysisRecords)
		if err := writeDependencyGraphFile(filepath.Join(dataDir, topLevel, graphRelativeOutput), graphNodes); err != nil {
			return err
		}
	}

	return nil
}

func collectJSBundleRecords(ctx context.Context, dataDir string) (map[string][]jsBundleRecord, error) {
	recordsByTopLevel := make(map[string][]jsBundleRecord)

	err := filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if d.IsDir() {
			return nil
		}
		if !isJSBundlePath(path) {
			return nil
		}

		record, topLevel, err := buildRecord(dataDir, path, d)
		if err != nil {
			return err
		}
		recordsByTopLevel[topLevel] = append(recordsByTopLevel[topLevel], record)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return map[string][]jsBundleRecord{}, nil
		}
		return nil, err
	}

	for topLevel := range recordsByTopLevel {
		sort.Slice(recordsByTopLevel[topLevel], func(i, j int) bool {
			return recordsByTopLevel[topLevel][i].PrimaryKey < recordsByTopLevel[topLevel][j].PrimaryKey
		})
		if err := validateNoDuplicatePrimaryKeys(recordsByTopLevel[topLevel]); err != nil {
			return nil, fmt.Errorf("top-level %q: %w", topLevel, err)
		}
	}

	return recordsByTopLevel, nil
}

func buildRecord(dataDir, path string, d fs.DirEntry) (jsBundleRecord, string, error) {
	relPath, err := filepath.Rel(dataDir, path)
	if err != nil {
		return jsBundleRecord{}, "", err
	}
	relPath = filepath.ToSlash(relPath)

	topLevel := topLevelFromRelativePath(relPath)
	info, err := d.Info()
	if err != nil {
		return jsBundleRecord{}, "", err
	}

	hashHex, err := fileSHA256(path)
	if err != nil {
		return jsBundleRecord{}, "", err
	}

	return jsBundleRecord{
		PrimaryKey: relPath,
		FilePath:   relPath,
		SizeBytes:  info.Size(),
		SHA256:     hashHex,
		ChunkName:  inferChunkName(path),
	}, topLevel, nil
}

func writeIndexFile(outputPath string, records []jsBundleRecord) error {
	return writeJSONLFile(outputPath, records)
}

func writeAnalysisFile(outputPath string, records []jsBundleAnalysisRecord) error {
	return writeJSONLFile(outputPath, records)
}

func writeAnalysisErrorFile(outputPath string, records []jsBundleAnalysisErrorRecord) error {
	return writeJSONLFile(outputPath, records)
}

func writeDependencyGraphFile(outputPath string, records []jsBundleGraphNode) error {
	return writeJSONLFile(outputPath, records)
}

func writeJSONLFile[T any](outputPath string, records []T) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			return err
		}
		if _, err := w.Write(line); err != nil {
			return err
		}
		if err := w.WriteByte('\n'); err != nil {
			return err
		}
	}

	return w.Flush()
}

func extractBundleAnalysisRecords(dataDir string, records []jsBundleRecord) ([]jsBundleAnalysisRecord, []jsBundleAnalysisErrorRecord) {
	analysisRecords := make([]jsBundleAnalysisRecord, 0, len(records))
	errorRecords := make([]jsBundleAnalysisErrorRecord, 0)

	for _, record := range records {
		sourcePath := filepath.Join(dataDir, filepath.FromSlash(record.FilePath))
		sourceBytes, err := os.ReadFile(sourcePath)
		if err != nil {
			errorRecords = append(errorRecords, jsBundleAnalysisErrorRecord{
				PrimaryKey: record.PrimaryKey,
				FilePath:   record.FilePath,
				ChunkName:  record.ChunkName,
				Error:      err.Error(),
			})
			continue
		}

		extracted, err := parseIndexedBundleSource(string(sourceBytes))
		if err != nil {
			errorRecords = append(errorRecords, jsBundleAnalysisErrorRecord{
				PrimaryKey: record.PrimaryKey,
				FilePath:   record.FilePath,
				ChunkName:  record.ChunkName,
				Error:      err.Error(),
			})
			continue
		}

		analysisRecords = append(analysisRecords, jsBundleAnalysisRecord{
			PrimaryKey:    record.PrimaryKey,
			FilePath:      record.FilePath,
			ChunkName:     record.ChunkName,
			Functions:     extracted.Functions,
			Classes:       extracted.Classes,
			Exports:       extracted.Exports,
			ImportEdges:   extracted.ImportEdges,
			RequireEdges:  extracted.RequireEdges,
			SignalAnchors: extracted.Anchors,
		})
	}

	return analysisRecords, errorRecords
}

func parseIndexedBundleSource(source string) (jsBundleExtracted, error) {
	return parseIndexedBundleSourceAST(source)
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func decodeQuotedContent(raw string) string {
	unquoted, err := strconv.Unquote(`"` + raw + `"`)
	if err != nil {
		return raw
	}
	return unquoted
}

func classifySignalAnchor(candidate string) (string, bool) {
	value := strings.TrimSpace(candidate)
	if len(value) < 3 || len(value) > 180 {
		return "", false
	}

	lower := strings.ToLower(value)
	switch {
	case strings.HasPrefix(value, "/api"), strings.Contains(lower, "/api/"), strings.Contains(lower, "/graphql"):
		return "api_route", true
	case strings.HasPrefix(lower, "ws://"), strings.HasPrefix(lower, "wss://"), strings.Contains(lower, "/ws/"), strings.HasSuffix(lower, "/ws"), strings.Contains(lower, "websocket"):
		return "websocket_channel", true
	case featureFlagAnchorRE.MatchString(value):
		return "feature_flag", true
	case !strings.ContainsAny(value, " \t\r\n/") && actionEventAnchorRE.MatchString(value):
		return "action_event", true
	default:
		return "", false
	}
}

func validateBalancedSyntax(source string) error {
	stack := make([]rune, 0, 16)
	mode := byte(0) // 0 normal, 1 single quote, 2 double quote, 3 template, 4 line comment, 5 block comment

	for i := 0; i < len(source); i++ {
		ch := source[i]
		next := byte(0)
		if i+1 < len(source) {
			next = source[i+1]
		}

		switch mode {
		case 4:
			if ch == '\n' {
				mode = 0
			}
			continue
		case 5:
			if ch == '*' && next == '/' {
				mode = 0
				i++
			}
			continue
		case 1:
			if ch == '\\' {
				i++
				continue
			}
			if ch == '\'' {
				mode = 0
			}
			continue
		case 2:
			if ch == '\\' {
				i++
				continue
			}
			if ch == '"' {
				mode = 0
			}
			continue
		case 3:
			if ch == '\\' {
				i++
				continue
			}
			if ch == '`' {
				mode = 0
			}
			continue
		}

		if ch == '/' && next == '/' {
			mode = 4
			i++
			continue
		}
		if ch == '/' && next == '*' {
			mode = 5
			i++
			continue
		}
		if ch == '\'' {
			mode = 1
			continue
		}
		if ch == '"' {
			mode = 2
			continue
		}
		if ch == '`' {
			mode = 3
			continue
		}

		switch ch {
		case '(', '[', '{':
			stack = append(stack, rune(ch))
		case ')', ']', '}':
			if len(stack) == 0 {
				return fmt.Errorf("unbalanced delimiter %q", string(ch))
			}
			top := stack[len(stack)-1]
			if !isDelimiterPair(top, rune(ch)) {
				return fmt.Errorf("mismatched delimiter %q closing %q", string(ch), string(top))
			}
			stack = stack[:len(stack)-1]
		}
	}

	if mode != 0 {
		return fmt.Errorf("unterminated token sequence")
	}
	if len(stack) > 0 {
		return fmt.Errorf("unclosed delimiter %q", string(stack[len(stack)-1]))
	}
	return nil
}

func isDelimiterPair(open, close rune) bool {
	return (open == '(' && close == ')') || (open == '[' && close == ']') || (open == '{' && close == '}')
}

func validateNoDuplicatePrimaryKeys(records []jsBundleRecord) error {
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if _, ok := seen[record.PrimaryKey]; ok {
			return fmt.Errorf("duplicate primary key %q", record.PrimaryKey)
		}
		seen[record.PrimaryKey] = struct{}{}
	}
	return nil
}

func buildBundleDependencyGraph(indexRecords []jsBundleRecord, analysisRecords []jsBundleAnalysisRecord) []jsBundleGraphNode {
	recordByKey := make(map[string]jsBundleRecord, len(indexRecords))
	for _, record := range indexRecords {
		recordByKey[record.PrimaryKey] = record
	}

	nodes := make([]jsBundleGraphNode, 0, len(analysisRecords))
	for _, analysis := range analysisRecords {
		record, ok := recordByKey[analysis.PrimaryKey]
		if !ok {
			continue
		}
		dependencies := buildGraphDependencies(analysis, recordByKey)
		nodes = append(nodes, jsBundleGraphNode{
			PrimaryKey:   analysis.PrimaryKey,
			FilePath:     analysis.FilePath,
			ChunkName:    analysis.ChunkName,
			Dependencies: dependencies,
			SourceReferences: []jsBundleSourceReference{
				{Type: "index_record_primary_key", Value: record.PrimaryKey},
				{Type: "analysis_record_primary_key", Value: analysis.PrimaryKey},
			},
			DomainHints: buildDomainHints(record, analysis, dependencies),
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].PrimaryKey < nodes[j].PrimaryKey
	})
	return nodes
}

func buildGraphDependencies(analysis jsBundleAnalysisRecord, recordByKey map[string]jsBundleRecord) []jsBundleGraphDependency {
	deps := make([]jsBundleGraphDependency, 0, len(analysis.ImportEdges)+len(analysis.RequireEdges))
	seen := make(map[string]struct{}, len(analysis.ImportEdges)+len(analysis.RequireEdges))

	appendEdges := func(depType string, edges []string) {
		for _, edge := range edges {
			dep := jsBundleGraphDependency{Type: depType, Target: edge}
			if resolved, ok := resolveDependencyPrimaryKey(analysis.FilePath, edge, recordByKey); ok {
				dep.ResolvedPrimaryKey = resolved.PrimaryKey
				dep.ResolvedChunkName = resolved.ChunkName
			}
			key := dep.Type + "\x00" + dep.Target + "\x00" + dep.ResolvedPrimaryKey
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			deps = append(deps, dep)
		}
	}

	appendEdges("import", analysis.ImportEdges)
	appendEdges("require", analysis.RequireEdges)

	sort.Slice(deps, func(i, j int) bool {
		if deps[i].Type == deps[j].Type {
			if deps[i].Target == deps[j].Target {
				return deps[i].ResolvedPrimaryKey < deps[j].ResolvedPrimaryKey
			}
			return deps[i].Target < deps[j].Target
		}
		return deps[i].Type < deps[j].Type
	})

	return deps
}

func resolveDependencyPrimaryKey(filePath, edge string, recordByKey map[string]jsBundleRecord) (jsBundleRecord, bool) {
	if edge == "" {
		return jsBundleRecord{}, false
	}
	if !strings.HasPrefix(edge, "./") && !strings.HasPrefix(edge, "../") {
		return jsBundleRecord{}, false
	}

	sourceDir := filepath.Dir(filepath.FromSlash(filePath))
	basePath := filepath.Clean(filepath.Join(sourceDir, edge))
	candidates := make([]string, 0, 6)
	if strings.HasSuffix(basePath, ".js") {
		candidates = append(candidates, basePath)
	} else {
		candidates = append(candidates, basePath+".js", filepath.Join(basePath, "index.js"))
	}
	candidates = append(candidates, basePath)

	for _, candidate := range candidates {
		key := filepath.ToSlash(candidate)
		if record, ok := recordByKey[key]; ok {
			return record, true
		}
	}
	return jsBundleRecord{}, false
}

func buildDomainHints(record jsBundleRecord, analysis jsBundleAnalysisRecord, dependencies []jsBundleGraphDependency) []jsDomainHint {
	evidence := collectDomainEvidence(record, analysis, dependencies)
	rules := []struct {
		domain   string
		keywords []string
	}{
		{domain: "chart", keywords: []string{"chart", "candl", "ohlc", "symbol", "price-scale"}},
		{domain: "studies", keywords: []string{"study", "studies", "indicator", "overlay", "rsi", "macd"}},
		{domain: "trading", keywords: []string{"trade", "trading", "order", "position", "broker", "portfolio"}},
		{domain: "watchlist", keywords: []string{"watchlist", "watch-list", "symbol-list"}},
		{domain: "replay", keywords: []string{"replay", "playback", "bar-replay"}},
		{domain: "widget", keywords: []string{"widget", "embed", "iframe", "mini-chart"}},
	}

	hints := make([]jsDomainHint, 0, len(rules))
	for _, rule := range rules {
		if rationale, ok := firstMatchingEvidence(rule.keywords, evidence); ok {
			hints = append(hints, jsDomainHint{
				Domain:    rule.domain,
				Rationale: rationale,
			})
		}
	}
	return hints
}

type domainEvidence struct {
	label string
	value string
}

func collectDomainEvidence(record jsBundleRecord, analysis jsBundleAnalysisRecord, dependencies []jsBundleGraphDependency) []domainEvidence {
	out := []domainEvidence{
		{label: "chunk_name", value: record.ChunkName},
	}

	for _, value := range analysis.Functions {
		out = append(out, domainEvidence{label: "function", value: value})
	}
	for _, value := range analysis.Classes {
		out = append(out, domainEvidence{label: "class", value: value})
	}
	for _, value := range analysis.Exports {
		out = append(out, domainEvidence{label: "export", value: value})
	}
	for _, anchor := range analysis.SignalAnchors {
		out = append(out, domainEvidence{label: "signal_anchor_" + anchor.Type, value: anchor.Value})
	}
	for _, dep := range dependencies {
		out = append(out, domainEvidence{label: dep.Type + "_edge", value: dep.Target})
		if dep.ResolvedPrimaryKey != "" {
			out = append(out, domainEvidence{label: dep.Type + "_resolved", value: dep.ResolvedPrimaryKey})
		}
		if dep.ResolvedChunkName != "" {
			out = append(out, domainEvidence{label: dep.Type + "_resolved_chunk", value: dep.ResolvedChunkName})
		}
	}
	return out
}

func firstMatchingEvidence(keywords []string, evidence []domainEvidence) (string, bool) {
	for _, candidate := range evidence {
		lower := strings.ToLower(candidate.value)
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				return fmt.Sprintf("matched keyword %q in %s %q", keyword, candidate.label, candidate.value), true
			}
		}
	}
	return "", false
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func topLevelFromRelativePath(relPath string) string {
	if relPath == "" {
		return "root"
	}
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 || parts[0] == "" || parts[0] == "." {
		return "root"
	}
	return parts[0]
}

func isJSBundlePath(path string) bool {
	if filepath.Ext(path) != ".js" {
		return false
	}
	slashed := filepath.ToSlash(path)
	return strings.Contains(slashed, "/resources/js/")
}

func inferChunkName(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" {
		return "unknown"
	}
	parts := strings.Split(base, ".")
	for _, part := range parts {
		if part != "" {
			return part
		}
	}
	return base
}
