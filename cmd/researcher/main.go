package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgnsrekt/tv_agent/internal/capture"
	"github.com/dgnsrekt/tv_agent/internal/cdp"
	"github.com/dgnsrekt/tv_agent/internal/config"
	"github.com/dgnsrekt/tv_agent/internal/storage"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		slog.Debug("log directory creation failed", "error", err)
	}

	logWriter := &lumberjack.Logger{
		Filename:   "logs/researcher.log",
		MaxSize:    25,
		MaxBackups: 10,
		MaxAge:     14,
		Compress:   true,
	}

	handler := slog.NewTextHandler(io.MultiWriter(os.Stdout, logWriter), &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting TradingView passive researcher")

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("Configuration loaded",
		"cdp_address", cfg.CDPAddress,
		"cdp_port", cfg.CDPPort,
		"data_dir", cfg.DataDir,
		"tab_url_filter", cfg.TabURLFilter,
		"reload_on_attach", cfg.ReloadOnAttach,
		"capture_http", cfg.CaptureHTTP,
		"capture_ws", cfg.CaptureWS,
		"capture_static", cfg.CaptureStatic,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	writerRegistry := storage.NewWriterRegistry(cfg.DataDir, cfg.BufferSize, cfg.MaxFileSizeMB)
	defer func() {
		if err := writerRegistry.Close(); err != nil {
			slog.Warn("Writer close failed", "error", err)
		}
	}()

	resourceWriter := storage.NewResourceWriter(cfg.DataDir)
	tabRegistry := cdp.NewTabRegistry()

	httpCapture := capture.NewHTTPCapture(writerRegistry, resourceWriter, tabRegistry,
		cfg.CaptureHTTP, cfg.CaptureStatic, cfg.HTTPMaxBodyBytes, cfg.ResourceMaxBytes)
	defer httpCapture.Close()

	wsCapture := capture.NewWebSocketCapture(writerRegistry, tabRegistry, cfg.CaptureWS, cfg.WSMaxFrameBytes)

	cdpClient := cdp.NewClient(cfg, httpCapture, wsCapture, tabRegistry)
	if err := cdpClient.Connect(ctx); err != nil {
		slog.Error("Failed to connect to browser", "error", err)
		slog.Info("Make sure Chromium is running with remote debugging enabled")
		slog.Info("Run: just start-browser")
		os.Exit(1)
	}
	defer func() {
		if err := cdpClient.Close(); err != nil {
			slog.Warn("CDP close failed", "error", err)
		}
	}()

	slog.Info("Researcher running", "tabs", cdpClient.GetTabCount(), "output_dir", cfg.DataDir)
	slog.Info("Press Ctrl+C to stop")

	<-sigCh
	slog.Info("Shutdown signal received")
	cancel()
	slog.Info("Researcher stopped")
}
