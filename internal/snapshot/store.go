package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// SnapshotMeta describes stored snapshot metadata.
type SnapshotMeta struct {
	ID          string    `json:"id"`
	ChartID     string    `json:"chart_id"`
	Format      string    `json:"format"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	SizeBytes   int       `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
	Symbol      string    `json:"symbol,omitempty"`
	Exchange    string    `json:"exchange,omitempty"`
	Resolution  string    `json:"resolution,omitempty"`
	Description string    `json:"description,omitempty"`
	Theme       string    `json:"theme,omitempty"`
	Layout      string    `json:"layout,omitempty"`
	Notes       string    `json:"notes,omitempty"`
}

// Store manages snapshot files on disk.
type Store struct {
	dir string
	mu  sync.RWMutex
}

// NewStore creates a Store and ensures the directory exists.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("snapshot store: mkdir %s: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

func (s *Store) validateID(id string) error {
	if !uuidRe.MatchString(id) {
		return fmt.Errorf("invalid snapshot id: %q", id)
	}
	return nil
}

// Save writes both the image file and metadata sidecar.
func (s *Store) Save(meta SnapshotMeta, imageData []byte) error {
	if err := s.validateID(meta.ID); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	imgPath := filepath.Join(s.dir, meta.ID+"."+meta.Format)
	jsonPath := filepath.Join(s.dir, meta.ID+".json")

	if err := os.WriteFile(imgPath, imageData, 0o644); err != nil {
		return fmt.Errorf("snapshot store: write image: %w", err)
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		_ = os.Remove(imgPath)
		return fmt.Errorf("snapshot store: marshal meta: %w", err)
	}

	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		_ = os.Remove(imgPath)
		return fmt.Errorf("snapshot store: write meta: %w", err)
	}

	return nil
}

// Get reads snapshot metadata by ID.
func (s *Store) Get(id string) (SnapshotMeta, error) {
	if err := s.validateID(id); err != nil {
		return SnapshotMeta{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	jsonPath := filepath.Join(s.dir, id+".json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return SnapshotMeta{}, fmt.Errorf("snapshot not found: %s", id)
		}
		return SnapshotMeta{}, fmt.Errorf("snapshot store: read meta: %w", err)
	}

	var meta SnapshotMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return SnapshotMeta{}, fmt.Errorf("snapshot store: unmarshal meta: %w", err)
	}
	return meta, nil
}

// List returns all snapshots sorted by creation time (newest first).
func (s *Store) List() ([]SnapshotMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	matches, err := filepath.Glob(filepath.Join(s.dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("snapshot store: glob: %w", err)
	}

	metas := make([]SnapshotMeta, 0, len(matches))
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var meta SnapshotMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		metas = append(metas, meta)
	}

	sort.Slice(metas, func(i, j int) bool {
		return metas[i].CreatedAt.After(metas[j].CreatedAt)
	})

	return metas, nil
}

// ReadImage reads the raw image bytes and returns the format.
func (s *Store) ReadImage(id string) ([]byte, string, error) {
	meta, err := s.Get(id)
	if err != nil {
		return nil, "", err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	imgPath := filepath.Join(s.dir, id+"."+meta.Format)
	data, err := os.ReadFile(imgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("snapshot image not found: %s", id)
		}
		return nil, "", fmt.Errorf("snapshot store: read image: %w", err)
	}
	return data, meta.Format, nil
}

// Delete removes both the image and metadata files.
func (s *Store) Delete(id string) error {
	if err := s.validateID(id); err != nil {
		return err
	}

	// Read meta first to know the format.
	meta, err := s.Get(id)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	imgPath := filepath.Join(s.dir, id+"."+meta.Format)
	jsonPath := filepath.Join(s.dir, id+".json")

	_ = os.Remove(imgPath)
	_ = os.Remove(jsonPath)
	return nil
}
