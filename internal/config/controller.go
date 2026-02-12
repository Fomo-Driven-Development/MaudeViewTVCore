package config

import (
	"strconv"
	"strings"
)

// ControllerConfig holds configuration for the Huma control API.
type ControllerConfig struct {
	CDPAddress       string
	CDPPort          int
	BindAddr         string
	TabURLFilter     string
	EvalTimeoutMS    int
	PortAutoFallback bool
	PortCandidates   []string
	LogLevel         string
	LogFile          string
	SnapshotDir      string
}

// LoadController reads controller configuration from environment variables.
func LoadController() (*ControllerConfig, error) {
	cfg := &ControllerConfig{
		CDPAddress:       getEnvOrDefault("CHROMIUM_CDP_ADDRESS", "127.0.0.1"),
		CDPPort:          getEnvIntOrDefault("CHROMIUM_CDP_PORT", 9220),
		BindAddr:         getEnvOrDefault("CONTROLLER_BIND_ADDR", "127.0.0.1:8188"),
		TabURLFilter:     getEnvOrDefault("CONTROLLER_TAB_URL_FILTER", "tradingview.com"),
		EvalTimeoutMS:    getEnvIntOrDefault("CONTROLLER_EVAL_TIMEOUT_MS", 5000),
		PortAutoFallback: getEnvBoolOrDefault("CONTROLLER_PORT_AUTO_FALLBACK", true),
		PortCandidates:   parseCSV(getEnvOrDefault("CONTROLLER_PORT_CANDIDATES", "127.0.0.1:8188,127.0.0.1:8189,127.0.0.1:8190,127.0.0.1:8191")),
		LogLevel:         strings.ToLower(getEnvOrDefault("CONTROLLER_LOG_LEVEL", "info")),
		LogFile:          getEnvOrDefault("CONTROLLER_LOG_FILE", "logs/tv_controller.log"),
		SnapshotDir:      getEnvOrDefault("SNAPSHOT_DIR", "./snapshots"),
	}
	if cfg.EvalTimeoutMS < 1000 {
		cfg.EvalTimeoutMS = 1000
	}
	if len(cfg.PortCandidates) == 0 {
		cfg.PortCandidates = []string{"127.0.0.1:8188", "127.0.0.1:8189", "127.0.0.1:8190", "127.0.0.1:8191"}
	}
	return cfg, nil
}

// ControllerCDPURL returns CDP endpoint URL for controller use.
func (c *ControllerConfig) ControllerCDPURL() string {
	return "http://" + c.CDPAddress + ":" + strconv.Itoa(c.CDPPort)
}

func parseCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}
