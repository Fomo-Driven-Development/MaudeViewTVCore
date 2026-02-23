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

	"github.com/dgnsrekt/MaudeViewTVCore/internal/apimulti"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/browser"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/config"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/multicontroller"
	"github.com/dgnsrekt/MaudeViewTVCore/internal/snapshot"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	cfg, err := config.LoadMultiController()
	if err != nil {
		slog.Error("failed to load multi controller config", "error", err)
		os.Exit(1)
	}

	if err := setupLogger(cfg.LogLevel, cfg.LogFile); err != nil {
		if _, writeErr := io.WriteString(os.Stderr, "logger setup failed: "+err.Error()+"\n"); writeErr != nil {
			slog.Debug("logger setup stderr write failed", "error", writeErr)
		}
		os.Exit(1)
	}

	slog.Info("tv_multi_controller config loaded",
		"bind_addr", cfg.BindAddr,
		"tab_url_filter", cfg.TabURLFilter,
		"eval_timeout_ms", cfg.EvalTimeoutMS,
		"log_level", cfg.LogLevel,
		"log_file", cfg.LogFile,
		"snapshot_dir", cfg.SnapshotDir,
		"launch_browser", cfg.LaunchBrowser,
	)

	winCfg, winCfgErr := config.LoadMultiWindows(cfg.WindowsConfigPath)
	if winCfgErr == nil && len(winCfg.Windows) > 0 && cfg.LaunchBrowser {
		cfg.StartURL = winCfg.Windows[0].URL
	}

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

	if winCfgErr == nil {
		slog.Info("opening windows from config", "count", len(winCfg.Windows), "path", cfg.WindowsConfigPath)
		for i := 1; i < len(winCfg.Windows); i++ {
			if err := cdpClient.OpenWindow(context.Background(), winCfg.Windows[i].URL); err != nil {
				slog.Warn("failed to open window", "index", i, "url", winCfg.Windows[i].URL, "error", err)
			} else {
				slog.Info("opened window", "index", i, "url", winCfg.Windows[i].URL)
			}
		}
		slog.Info("waiting for windows to load", "delay", "5s")
		time.Sleep(5 * time.Second)
	} else if !os.IsNotExist(winCfgErr) {
		slog.Warn("windows config not loaded", "path", cfg.WindowsConfigPath, "error", winCfgErr)
	}

	snapStore, err := snapshot.NewStore(cfg.SnapshotDir)
	if err != nil {
		slog.Error("failed to create snapshot store", "dir", cfg.SnapshotDir, "error", err)
		if launcher != nil && launcher.Running() {
			launcher.Stop()
		}
		os.Exit(1)
	}

	svc := multicontroller.NewService(cdpClient, snapStore)

	h := apimulti.NewServer(svc)

	srv := &http.Server{Addr: bindAddr, Handler: h}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("tv_multi_controller listening", "addr", bindAddr, "docs", "http://"+bindAddr+"/docs")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("tv_multi_controller server failed", "error", err)
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
		slog.Error("tv_multi_controller shutdown failed", "error", err)
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
