package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// WindowEntry describes a single browser window to open at startup.
type WindowEntry struct {
	URL string `yaml:"url"`
}

// MultiWindowsConfig is the top-level YAML configuration for startup windows.
type MultiWindowsConfig struct {
	Windows []WindowEntry `yaml:"windows"`
}

// LoadMultiWindows reads and validates a windows YAML config file.
// Returns an os.ErrNotExist-wrapped error if the file is absent (caller
// silently skips in that case).
func LoadMultiWindows(path string) (*MultiWindowsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("multi_windows config: %w", err)
	}
	var cfg MultiWindowsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("multi_windows config: %w", err)
	}
	if len(cfg.Windows) < 1 {
		return nil, fmt.Errorf("multi_windows config: at least one window entry is required")
	}
	for i, w := range cfg.Windows {
		if w.URL == "" {
			return nil, fmt.Errorf("multi_windows config: windows[%d] missing url", i)
		}
	}
	return &cfg, nil
}
