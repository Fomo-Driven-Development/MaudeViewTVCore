package config

import (
	"strconv"
	"strings"
)

// ControllerConfig holds configuration for the Huma control API.
type ControllerConfig struct {
	CDPAddress    string
	CDPPort       int
	BindAddr      string
	TabURLFilter  string
	EvalTimeoutMS int
	LogLevel      string
	LogFile       string
	SnapshotDir   string
}

// LoadController reads controller configuration from environment variables.
func LoadController() (*ControllerConfig, error) {
	cfg := &ControllerConfig{
		CDPAddress:    getEnvOrDefault("CHROMIUM_CDP_ADDRESS", "127.0.0.1"),
		CDPPort:       getEnvIntOrDefault("CHROMIUM_CDP_PORT", 9220),
		BindAddr:      getEnvOrDefault("CONTROLLER_BIND_ADDR", "127.0.0.1:8188"),
		TabURLFilter:  getEnvOrDefault("CONTROLLER_TAB_URL_FILTER", "tradingview.com"),
		EvalTimeoutMS: getEnvIntOrDefault("CONTROLLER_EVAL_TIMEOUT_MS", 5000),
		LogLevel:      strings.ToLower(getEnvOrDefault("CONTROLLER_LOG_LEVEL", "info")),
		LogFile:       getEnvOrDefault("CONTROLLER_LOG_FILE", "logs/tv_controller.log"),
		SnapshotDir:   getEnvOrDefault("SNAPSHOT_DIR", "./snapshots"),
	}
	if cfg.EvalTimeoutMS < 1000 {
		cfg.EvalTimeoutMS = 1000
	}
	return cfg, nil
}

// ControllerCDPURL returns CDP endpoint URL for controller use.
func (c *ControllerConfig) ControllerCDPURL() string {
	return "http://" + c.CDPAddress + ":" + strconv.Itoa(c.CDPPort)
}

