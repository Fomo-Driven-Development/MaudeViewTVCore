package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dgnsrekt/MaudeViewTVCore/internal/api"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/browser"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/config"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/controller"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/relay"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/snapshot"
	"gopkg.in/natefinch/lumberjack.v2"
)

var version = "dev"

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

	slog.Info("tv_controller starting", "version", version)
	slog.Info("tv_controller config loaded",
		"bind_addr", cfg.BindAddr,
		"tab_url_filter", cfg.TabURLFilter,
		"eval_timeout_ms", cfg.EvalTimeoutMS,
		"log_level", cfg.LogLevel,
		"log_file", cfg.LogFile,
		"snapshot_dir", cfg.SnapshotDir,
		"launch_browser", cfg.LaunchBrowser,
	)

	var launcher *browser.Launcher
	if cfg.LaunchBrowser {
		launcher = browser.NewLauncher(browser.Config{
			CDPAddress:          cfg.CDPAddress,
			CDPPort:             cfg.CDPPort,
			StartURL:            cfg.StartURL,
			ProfileDir:          cfg.ProfileDir,
			LogFileDir:          cfg.LogFileDir,
			CrashDumpDir:        cfg.CrashDumpDir,
			EnableCrashReporter: cfg.EnableCrashReporter,
		})
		if err := launcher.Launch(context.Background()); err != nil {
			slog.Error("failed to launch browser", "error", err)
			os.Exit(1)
		}
	}

	bindAddr := cfg.BindAddr

	cdpClient := cdpcontrol.NewClient(cfg.ControllerCDPURL(), cfg.TabURLFilter, time.Duration(cfg.EvalTimeoutMS)*time.Millisecond)
	if err := cdpClient.Connect(context.Background()); err != nil {
		slog.Error("failed to connect CDP controller", "cdp_url", cfg.ControllerCDPURL(), "error", err)
		if launcher != nil && launcher.Running() {
			launcher.Stop()
		}
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
		if launcher != nil && launcher.Running() {
			launcher.Stop()
		}
		os.Exit(1)
	}

	svc := controller.NewService(cdpClient, snapStore)

	var serverOpts []api.ServerOption
	var wsRelay *relay.Relay
	if cfg.RelayEnabled {
		relayCfg, err := relay.LoadConfig(cfg.RelayConfigPath)
		if err != nil {
			slog.Error("failed to load relay config", "path", cfg.RelayConfigPath, "error", err)
			os.Exit(1)
		}
		broker := relay.NewBroker()
		wsRelay = relay.NewRelay(relayCfg, broker)
		if err := wsRelay.Start(context.Background(), cdpClient); err != nil {
			slog.Error("failed to start relay", "error", err)
			os.Exit(1)
		}
		serverOpts = append(serverOpts, api.WithRelayHandler(relay.SSEHandler(broker)))
		slog.Info("ws relay enabled", "config", cfg.RelayConfigPath, "feeds", len(relayCfg.Feeds))
	}

	h := api.NewServer(svc, serverOpts...)

	srv := &http.Server{Addr: bindAddr, Handler: h}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("tv_controller listening", "addr", bindAddr, "docs", "http://"+bindAddr+"/docs")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("tv_controller server failed", "error", err)
			slog.Error("hint: kill the process holding the port with: lsof -ti:"+bindAddr[strings.LastIndex(bindAddr, ":")+1:]+" | xargs kill -9")
			serverErr <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	exitCode := 0
	select {
	case <-sigCh:
	case <-serverErr:
		exitCode = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("tv_controller shutdown failed", "error", err)
	}

	if wsRelay != nil {
		wsRelay.Stop()
	}

	if launcher != nil && launcher.Running() {
		launcher.Stop()
	}

	os.Exit(exitCode)
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
