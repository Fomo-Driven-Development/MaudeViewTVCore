package validation

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const defaultDataDir = "./research_data"

type artifactSpec struct {
	label          string
	relativePath   string
	requiredFields []string
	validateRecord func(raw map[string]json.RawMessage) error
}

// Run executes mapper artifact validation.
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

	if err := validateDataDir(ctx, dataDir); err != nil {
		return err
	}

	_, err := fmt.Fprintln(w, "validation: complete")
	return err
}

func validateDataDir(ctx context.Context, dataDir string) error {
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
		if err := validateDateRoot(dateRoot); err != nil {
			return fmt.Errorf("validate %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func validateDateRoot(dateRoot string) error {
	specs := []artifactSpec{
		{
			label:        "index",
			relativePath: "mapper/static-analysis/js-bundle-index.jsonl",
			requiredFields: []string{
				"primary_key",
				"file_path",
				"size_bytes",
				"sha256",
				"chunk_name",
			},
		},
		{
			label:        "extraction",
			relativePath: "mapper/static-analysis/js-bundle-analysis.jsonl",
			requiredFields: []string{
				"primary_key",
				"file_path",
				"chunk_name",
				"functions",
				"classes",
				"exports",
				"import_edges",
				"require_edges",
				"signal_anchors",
			},
		},
		{
			label:        "trace",
			relativePath: "mapper/runtime-probes/runtime-trace.jsonl",
			requiredFields: []string{
				"timestamp",
				"trace_id",
				"tab_id",
				"tab_url",
				"surface",
				"event_type",
			},
		},
		{
			label:        "correlation",
			relativePath: "mapper/correlation/capability-correlations.jsonl",
			requiredFields: []string{
				"capability",
				"primary_recommended_control_path",
				"control_path_rationale",
				"confidence_score",
				"temporal_linkage_trace_ids",
				"evidence_links",
			},
			validateRecord: validateCorrelationEvidence,
		},
		{
			label:        "matrix",
			relativePath: "mapper/reporting/capability-matrix.jsonl",
			requiredFields: []string{
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
			validateRecord: validateMatrixEvidence,
		},
	}

	for _, spec := range specs {
		path := filepath.Join(dateRoot, spec.relativePath)
		if err := validateArtifact(path, spec); err != nil {
			return err
		}
	}
	return nil
}

func validateArtifact(path string, spec artifactSpec) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s artifact missing: %s", spec.label, path)
		}
		return err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 10*1024*1024)

	recordCount := 0
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		recordCount++

		var raw map[string]json.RawMessage
		if err := json.Unmarshal(line, &raw); err != nil {
			return fmt.Errorf("%s artifact malformed JSON at %s:%d: %w", spec.label, path, lineNo, err)
		}

		for _, field := range spec.requiredFields {
			value, ok := raw[field]
			if !ok {
				return fmt.Errorf("%s artifact schema drift at %s:%d: missing required field %q", spec.label, path, lineNo, field)
			}
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return fmt.Errorf("%s artifact schema drift at %s:%d: required field %q is null", spec.label, path, lineNo, field)
			}
		}

		if spec.validateRecord != nil {
			if err := spec.validateRecord(raw); err != nil {
				return fmt.Errorf("%s artifact invalid at %s:%d: %w", spec.label, path, lineNo, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	if recordCount == 0 {
		return fmt.Errorf("%s artifact is empty: %s", spec.label, path)
	}
	return nil
}

func validateCorrelationEvidence(raw map[string]json.RawMessage) error {
	return ensureNonEmptyArray(raw, "evidence_links", false)
}

func validateMatrixEvidence(raw map[string]json.RawMessage) error {
	return ensureNonEmptyArray(raw, "evidence_ids", true)
}

func ensureNonEmptyArray(raw map[string]json.RawMessage, field string, requireNonEmptyStrings bool) error {
	value, ok := raw[field]
	if !ok {
		return fmt.Errorf("missing %q", field)
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(value, &arr); err != nil {
		return fmt.Errorf("%q must be an array", field)
	}
	if len(arr) == 0 {
		return fmt.Errorf("%q must include at least one evidence entry", field)
	}
	if !requireNonEmptyStrings {
		return nil
	}
	for i, item := range arr {
		var s string
		if err := json.Unmarshal(item, &s); err != nil {
			return nil
		}
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%q has empty string at index %d", field, i)
		}
	}
	return nil
}
