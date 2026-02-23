package config

// LoadMultiController reads configuration for the multi-chart controller.
// It delegates to LoadController and overrides the bind address and log file
// so both binaries can run simultaneously.
func LoadMultiController() (*ControllerConfig, error) {
	cfg, err := LoadController()
	if err != nil {
		return nil, err
	}
	if cfg.BindAddr == "127.0.0.1:8188" {
		cfg.BindAddr = getEnvOrDefault("MULTI_CONTROLLER_BIND_ADDR", "127.0.0.1:8189")
	}
	if cfg.LogFile == "logs/tv_controller.log" {
		cfg.LogFile = getEnvOrDefault("MULTI_CONTROLLER_LOG_FILE", "logs/tv_multi_controller.log")
	}
	cfg.WindowsConfigPath = getEnvOrDefault("MULTI_CONTROLLER_WINDOWS_CONFIG", "./config/multi_windows.yaml")
	return cfg, nil
}
