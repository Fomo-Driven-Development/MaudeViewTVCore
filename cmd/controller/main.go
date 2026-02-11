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

	"github.com/dgnsrekt/tv_agent/internal/api"
	"github.com/dgnsrekt/tv_agent/internal/cdpcontrol"
	"github.com/dgnsrekt/tv_agent/internal/config"
	"github.com/dgnsrekt/tv_agent/internal/controller"
	"github.com/dgnsrekt/tv_agent/internal/netutil"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	cfg, err := config.LoadController()
	if err != nil {
		slog.Error("failed to load controller config", "error", err)
		os.Exit(1)
	}

	if err := setupLogger(cfg.LogLevel, cfg.LogFile); err != nil {
		_, _ = io.WriteString(os.Stderr, "logger setup failed: "+err.Error()+"\n")
		os.Exit(1)
	}

	slog.Info("controller config loaded",
		"bind_addr", cfg.BindAddr,
		"tab_url_filter", cfg.TabURLFilter,
		"eval_timeout_ms", cfg.EvalTimeoutMS,
		"port_auto_fallback", cfg.PortAutoFallback,
		"port_candidates", cfg.PortCandidates,
		"log_level", cfg.LogLevel,
		"log_file", cfg.LogFile,
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
	defer func() { _ = cdpClient.Close() }()

	svc := controller.NewService(cdpClient)
	h := api.NewServer(svc)

	srv := &http.Server{Addr: bindAddr, Handler: h}

	go func() {
		slog.Info("controller listening", "addr", bindAddr, "docs", "http://"+bindAddr+"/docs")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("controller server failed", "error", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("controller shutdown failed", "error", err)
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
