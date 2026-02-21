package relay

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// FeedConfig describes a single WebSocket feed to relay.
type FeedConfig struct {
	Name         string   `yaml:"name"`
	URLPattern   string   `yaml:"url_pattern"`
	MessageTypes []string `yaml:"message_types,omitempty"`
}

// RelayConfig is the top-level YAML configuration.
type RelayConfig struct {
	Feeds []FeedConfig `yaml:"feeds"`
}

// LoadConfig reads and validates a relay YAML config file.
func LoadConfig(path string) (*RelayConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("relay config: %w", err)
	}
	var cfg RelayConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("relay config: %w", err)
	}
	for i, f := range cfg.Feeds {
		if f.Name == "" {
			return nil, fmt.Errorf("relay config: feed[%d] missing name", i)
		}
		if f.URLPattern == "" {
			return nil, fmt.Errorf("relay config: feed[%d] (%s) missing url_pattern", i, f.Name)
		}
	}
	return &cfg, nil
}
