package storage

import (
	"log/slog"
	"sync"
)

// WriterRegistry manages multiple JSONLWriter instances, one per path+type combination.
// This allows each tab's data to be written to its own directory based on the URL path.
type WriterRegistry struct {
	baseDir    string
	maxSizeMB  int
	bufferSize int

	// writers maps pathSegment -> dataType -> writer
	// e.g., "flow_overview" -> "http" -> *JSONLWriter
	writers map[string]map[string]*JSONLWriter
	mu      sync.RWMutex
}

// NewWriterRegistry creates a new WriterRegistry for managing multiple JSONL writers.
func NewWriterRegistry(baseDir string, bufferSize int, maxSizeMB int) *WriterRegistry {
	return &WriterRegistry{
		baseDir:    baseDir,
		maxSizeMB:  maxSizeMB,
		bufferSize: bufferSize,
		writers:    make(map[string]map[string]*JSONLWriter),
	}
}

// GetWriter returns (or creates) a JSONLWriter for the given path segment and data type.
// pathSegment is the transformed URL path (e.g., "flow_overview")
// dataType is "http" or "websocket"
// browserID is the short identifier for the browser tab (first 8 chars of target ID)
func (r *WriterRegistry) GetWriter(pathSegment, dataType, browserID string) *JSONLWriter {
	r.mu.RLock()
	if typeMap, ok := r.writers[pathSegment]; ok {
		if writer, ok := typeMap[dataType]; ok {
			r.mu.RUnlock()
			return writer
		}
	}
	r.mu.RUnlock()

	// Need to create a new writer
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if typeMap, ok := r.writers[pathSegment]; ok {
		if writer, ok := typeMap[dataType]; ok {
			return writer
		}
	}

	// Create the type map if needed
	if r.writers[pathSegment] == nil {
		r.writers[pathSegment] = make(map[string]*JSONLWriter)
	}

	// Create the subdirectory path: pathSegment/dataType
	// e.g., "flow_overview/http"
	subDir := pathSegment + "/" + dataType

	// Create new writer with browserID
	writer := NewJSONLWriterWithBrowserID(
		r.baseDir,
		subDir,
		r.bufferSize,
		r.maxSizeMB,
		browserID,
	)

	r.writers[pathSegment][dataType] = writer

	slog.Info("Created new JSONL writer",
		"path_segment", pathSegment,
		"data_type", dataType,
		"browser_id", browserID)

	return writer
}

// Close closes all managed writers.
func (r *WriterRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for pathSeg, typeMap := range r.writers {
		for dataType, writer := range typeMap {
			if err := writer.Close(); err != nil {
				slog.Error("Failed to close writer",
					"path_segment", pathSeg,
					"data_type", dataType,
					"error", err)
				lastErr = err
			}
		}
	}

	// Clear the map
	r.writers = make(map[string]map[string]*JSONLWriter)

	return lastErr
}
