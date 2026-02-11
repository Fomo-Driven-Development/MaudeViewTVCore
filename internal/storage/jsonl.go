package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// JSONLWriter handles async writing of JSON lines to date-organized files
type JSONLWriter struct {
	baseDir     string
	subDir      string // e.g., "flow_overview/http" or "http"
	maxSizeMB   int
	browserID   string // Short ID for filename (optional, uses timestamp if empty)
	writeCh     chan any
	done        chan struct{}
	wg          sync.WaitGroup
	currentDate string
	logger      *lumberjack.Logger
	mu          sync.Mutex
}

// NewJSONLWriter creates a new async JSONL writer (uses timestamp-based filenames)
func NewJSONLWriter(baseDir, subDir string, bufferSize int, maxSizeMB int) *JSONLWriter {
	return newJSONLWriter(baseDir, subDir, bufferSize, maxSizeMB, "")
}

// NewJSONLWriterWithBrowserID creates a new async JSONL writer with a specific browser ID for filenames.
// The browserID is used as the filename base (e.g., "B0D5A8E8.jsonl") instead of a timestamp.
func NewJSONLWriterWithBrowserID(baseDir, subDir string, bufferSize int, maxSizeMB int, browserID string) *JSONLWriter {
	return newJSONLWriter(baseDir, subDir, bufferSize, maxSizeMB, browserID)
}

func newJSONLWriter(baseDir, subDir string, bufferSize int, maxSizeMB int, browserID string) *JSONLWriter {
	w := &JSONLWriter{
		baseDir:   baseDir,
		subDir:    subDir,
		maxSizeMB: maxSizeMB,
		browserID: browserID,
		writeCh:   make(chan any, bufferSize),
		done:      make(chan struct{}),
	}

	w.wg.Add(1)
	go w.writeLoop()

	return w
}

// Write queues a record for async writing
func (w *JSONLWriter) Write(record any) error {
	slog.Debug("JSONL write queued",
		"subdir", w.subDir)
	select {
	case w.writeCh <- record:
		return nil
	case <-w.done:
		return fmt.Errorf("writer is closed")
	default:
		// Channel full, log warning but don't block
		slog.Warn("JSONL write buffer full, dropping record",
			"subdir", w.subDir)
		return fmt.Errorf("buffer full")
	}
}

// Close shuts down the writer and flushes pending data
func (w *JSONLWriter) Close() error {
	close(w.done)

	// Drain remaining items with timeout
	timeout := time.After(5 * time.Second)
	for {
		select {
		case record := <-w.writeCh:
			w.writeRecord(record)
		case <-timeout:
			slog.Warn("JSONL writer close timeout, some records may be lost",
				"subdir", w.subDir)
			goto done
		default:
			goto done
		}
	}

done:
	w.wg.Wait()

	w.mu.Lock()
	defer w.mu.Unlock()
	if w.logger != nil {
		return w.logger.Close()
	}
	return nil
}

func (w *JSONLWriter) writeLoop() {
	defer w.wg.Done()

	for {
		select {
		case record := <-w.writeCh:
			w.writeRecord(record)
		case <-w.done:
			return
		}
	}
}

func (w *JSONLWriter) writeRecord(record any) {
	data, err := json.Marshal(record)
	if err != nil {
		slog.Error("Failed to marshal record",
			"error", err,
			"subdir", w.subDir)
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if we need to rotate to a new date directory
	currentDate := time.Now().UTC().Format("2006-01-02")
	if currentDate != w.currentDate {
		w.rotateForDate(currentDate)
	}

	if w.logger == nil {
		w.rotateForDate(currentDate)
	}

	// Write the JSON line
	_, err = w.logger.Write(append(data, '\n'))
	if err != nil {
		slog.Error("Failed to write record",
			"error", err,
			"subdir", w.subDir)
	}
}

func (w *JSONLWriter) rotateForDate(date string) {
	// Close existing logger
	if w.logger != nil {
		w.logger.Close()
	}

	// Create new directory for date
	dir := filepath.Join(w.baseDir, date, w.subDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("Failed to create output directory",
			"error", err,
			"dir", dir)
		return
	}

	// Create filename: use browserID if set, otherwise use timestamp
	var filename string
	if w.browserID != "" {
		filename = filepath.Join(dir, w.browserID+".jsonl")
	} else {
		filename = filepath.Join(dir, fmt.Sprintf("%d.jsonl", time.Now().Unix()))
	}

	w.logger = &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    w.maxSizeMB,
		MaxBackups: 100, // Keep many backups
		MaxAge:     30,  // 30 days
		Compress:   false,
		LocalTime:  false, // Use UTC
	}

	w.currentDate = date
	slog.Info("Opened new JSONL file",
		"file", filename,
		"subdir", w.subDir)
}
