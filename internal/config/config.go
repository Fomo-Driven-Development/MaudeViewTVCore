package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the passive researcher.
type Config struct {
	// CDP connection settings
	CDPAddress string
	CDPPort    int

	// Storage settings
	DataDir       string
	MaxFileSizeMB int
	BufferSize    int

	// Tab matching and behavior
	TabURLFilter   string
	ReloadOnAttach bool

	// Capture behavior
	CaptureHTTP   bool
	CaptureWS     bool
	CaptureStatic bool

	// Payload safety limits
	HTTPMaxBodyBytes int
	WSMaxFrameBytes  int
	ResourceMaxBytes int
}

// Load reads configuration from environment variables and optional .env file.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Debug("failed to load .env file", "error", err)
	}

	cfg := &Config{
		CDPAddress:       getEnvOrDefault("CHROMIUM_CDP_ADDRESS", "127.0.0.1"),
		CDPPort:          getEnvIntOrDefault("CHROMIUM_CDP_PORT", 9220),
		DataDir:          getEnvOrDefault("RESEARCHER_DATA_DIR", "./research_data"),
		MaxFileSizeMB:    getEnvIntOrDefault("RESEARCHER_MAX_FILE_SIZE_MB", 200),
		BufferSize:       getEnvIntOrDefault("RESEARCHER_BUFFER_SIZE", 5000),
		TabURLFilter:     getEnvOrDefault("RESEARCHER_TAB_URL_FILTER", "tradingview.com"),
		ReloadOnAttach:   getEnvBoolOrDefault("RESEARCHER_RELOAD_ON_ATTACH", true),
		CaptureHTTP:      getEnvBoolOrDefault("RESEARCHER_CAPTURE_HTTP", true),
		CaptureWS:        getEnvBoolOrDefault("RESEARCHER_CAPTURE_WS", true),
		CaptureStatic:    getEnvBoolOrDefault("RESEARCHER_CAPTURE_STATIC", true),
		HTTPMaxBodyBytes: getEnvIntOrDefault("RESEARCHER_HTTP_MAX_BODY_BYTES", 50*1024*1024),
		WSMaxFrameBytes:  getEnvIntOrDefault("RESEARCHER_WS_MAX_FRAME_BYTES", 20*1024*1024),
		ResourceMaxBytes: getEnvIntOrDefault("RESEARCHER_RESOURCE_MAX_BYTES", 100*1024*1024),
	}

	return cfg, nil
}

// GetCDPURL returns the full CDP HTTP endpoint used by chromedp remote allocator.
func (c *Config) GetCDPURL() string {
	return fmt.Sprintf("http://%s:%d", c.CDPAddress, c.CDPPort)
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvIntOrDefault(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvBoolOrDefault(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}
