package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgnsrekt/MaudeViewTVCore/internal/api"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/config"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/controller"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/netutil"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/snapshot"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	cfg, err := config.LoadController()
	if err != nil {
		slog.Error("failed to load controller config", "error", err)
		os.Exit(1)
	}

	if err := setupLogger(cfg.LogLevel, cfg.LogFile); err != nil {
		if _, writeErr := io.WriteString(os.Stderr, "logger setup failed: "+err.Error()+"\n"); writeErr != nil {
			slog.Debug("logger setup stderr write failed", "error", writeErr)
		}
		os.Exit(1)
	}

	slog.Info("tv_controller config loaded",
		"bind_addr", cfg.BindAddr,
		"tab_url_filter", cfg.TabURLFilter,
		"eval_timeout_ms", cfg.EvalTimeoutMS,
		"port_auto_fallback", cfg.PortAutoFallback,
		"port_candidates", cfg.PortCandidates,
		"log_level", cfg.LogLevel,
		"log_file", cfg.LogFile,
		"snapshot_dir", cfg.SnapshotDir,
	)

	bindAddr, err := netutil.SelectBindAddr(cfg.BindAddr, cfg.PortCandidates, cfg.PortAutoFallback)
	if err != nil {
		slog.Error("failed to select bind address", "preferred", cfg.BindAddr, "error", err)
		os.Exit(1)
	}

	cdpClient := cdpcontrol.NewClient(cfg.ControllerCDPURL(), cfg.TabURLFilter, time.Duration(cfg.EvalTimeoutMS)*time.Millisecond)
	if err := cdpClient.Connect(context.Background()); err != nil {
		slog.Error("failed to connect CDP controller", "cdp_url", cfg.ControllerCDPURL(), "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := cdpClient.Close(); err != nil {
			slog.Debug("CDP client close failed", "error", err)
		}
	}()

	snapStore, err := snapshot.NewStore(cfg.SnapshotDir)
	if err != nil {
		slog.Error("failed to create snapshot store", "dir", cfg.SnapshotDir, "error", err)
		os.Exit(1)
	}

	svc := controller.NewService(cdpClient, snapStore)
	h := api.NewServer(svc)

	srv := &http.Server{Addr: bindAddr, Handler: h}

	go func() {
		slog.Info("tv_controller listening", "addr", bindAddr, "docs", "http://"+bindAddr+"/docs")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("tv_controller server failed", "error", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("tv_controller shutdown failed", "error", err)
	}
}

func setupLogger(level, filename string) error {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		return err
	}

	logWriter := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    25,
		MaxBackups: 10,
		MaxAge:     14,
		Compress:   true,
	}

	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	h := slog.NewTextHandler(io.MultiWriter(os.Stdout, logWriter), &slog.HandlerOptions{Level: slogLevel})
	slog.SetDefault(slog.New(h))
	return nil
}
