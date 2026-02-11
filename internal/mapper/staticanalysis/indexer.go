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
	"sort"
	"strings"
)

const (
	defaultDataDir      = "./research_data"
	indexRelativeOutput = "mapper/static-analysis/js-bundle-index.jsonl"
)

type jsBundleRecord struct {
	PrimaryKey string `json:"primary_key"`
	FilePath   string `json:"file_path"`
	SizeBytes  int64  `json:"size_bytes"`
	SHA256     string `json:"sha256"`
	ChunkName  string `json:"chunk_name"`
}

func indexJSBundles(ctx context.Context, dataDir string) error {
	recordsByTopLevel, err := collectJSBundleRecords(ctx, dataDir)
	if err != nil {
		return err
	}

	for topLevel, records := range recordsByTopLevel {
		if err := writeIndexFile(filepath.Join(dataDir, topLevel, indexRelativeOutput), records); err != nil {
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
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

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
