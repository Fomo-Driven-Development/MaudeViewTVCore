package storage

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// ResourceWriter writes raw file content for static resources.
type ResourceWriter struct {
	baseDir string
}

func NewResourceWriter(baseDir string) *ResourceWriter {
	return &ResourceWriter{baseDir: baseDir}
}

// WriteRaw saves bytes to: baseDir/date/pathSegment/resources/resourceType/filename.
func (w *ResourceWriter) WriteRaw(pathSegment, resourceType, filename string, data []byte) error {
	date := time.Now().UTC().Format("2006-01-02")
	dir := filepath.Join(w.baseDir, date, pathSegment, "resources", resourceType)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return err
	}
	slog.Debug("Resource file written", "path", filePath, "size", len(data))
	return nil
}
