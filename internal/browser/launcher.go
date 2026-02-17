package browser

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

// Config holds browser launch configuration.
type Config struct {
	CDPAddress          string
	CDPPort             int
	StartURL            string
	ProfileDir          string
	LogFileDir          string
	CrashDumpDir        string
	EnableCrashReporter bool
	WindowSize          string
}

// Launcher manages the lifecycle of a browser process.
type Launcher struct {
	cfg     Config
	cmd     *exec.Cmd
	running bool
}

// NewLauncher creates a new browser launcher with the given config.
func NewLauncher(cfg Config) *Launcher {
	if cfg.WindowSize == "" {
		cfg.WindowSize = "1920,1080"
	}
	return &Launcher{cfg: cfg}
}

// detectBrowser finds an available Chrome/Chromium binary.
func detectBrowser() (string, error) {
	candidates := []string{"chromium-browser", "chromium", "google-chrome"}
	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	if runtime.GOOS == "darwin" {
		macPath := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		if _, err := os.Stat(macPath); err == nil {
			return macPath, nil
		}
	}
	return "", fmt.Errorf("no supported browser found (tried chromium-browser, chromium, google-chrome)")
}

// isPortInUse checks whether a TCP port is already listening.
func isPortInUse(address string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Launch starts the browser process unless the CDP port is already in use.
func (l *Launcher) Launch(ctx context.Context) error {
	if isPortInUse(l.cfg.CDPAddress, l.cfg.CDPPort) {
		slog.Info("browser already running, skipping launch",
			"address", l.cfg.CDPAddress, "port", l.cfg.CDPPort)
		return nil
	}

	browserPath, err := detectBrowser()
	if err != nil {
		return err
	}
	slog.Info("detected browser", "path", browserPath)

	if err := os.MkdirAll(l.cfg.ProfileDir, 0o755); err != nil {
		return fmt.Errorf("create profile dir: %w", err)
	}
	if err := os.MkdirAll(l.cfg.LogFileDir, 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	if err := os.MkdirAll(l.cfg.CrashDumpDir, 0o755); err != nil {
		return fmt.Errorf("create crash dump dir: %w", err)
	}

	args := []string{
		fmt.Sprintf("--remote-debugging-port=%d", l.cfg.CDPPort),
		fmt.Sprintf("--remote-debugging-address=%s", l.cfg.CDPAddress),
		fmt.Sprintf("--user-data-dir=%s", l.cfg.ProfileDir),
		"--no-first-run",
		"--disable-dev-shm-usage",
		"--disable-breakpad",
		"--disable-crash-reporter",
		fmt.Sprintf("--window-size=%s", l.cfg.WindowSize),
	}
	if l.cfg.EnableCrashReporter {
		args = append(args,
			"--enable-crash-reporter",
			fmt.Sprintf("--crash-dumps-dir=%s", l.cfg.CrashDumpDir),
		)
	}
	args = append(args, l.cfg.StartURL)

	l.cmd = exec.Command(browserPath, args...)
	l.cmd.Stdout = os.Stdout
	l.cmd.Stderr = os.Stderr

	if err := l.cmd.Start(); err != nil {
		return fmt.Errorf("start browser: %w", err)
	}
	l.running = true
	slog.Info("browser process started", "pid", l.cmd.Process.Pid)

	if err := l.waitForCDP(ctx); err != nil {
		l.Stop()
		return fmt.Errorf("waiting for CDP: %w", err)
	}
	slog.Info("CDP endpoint ready",
		"address", l.cfg.CDPAddress, "port", l.cfg.CDPPort)

	return nil
}

// waitForCDP polls the CDP /json/version endpoint until it responds.
func (l *Launcher) waitForCDP(ctx context.Context) error {
	url := fmt.Sprintf("http://%s:%d/json/version", l.cfg.CDPAddress, l.cfg.CDPPort)
	deadline := time.After(15 * time.Second)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	client := &http.Client{Timeout: time.Second}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("CDP did not become ready within 15s at %s", url)
		case <-ticker.C:
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

// Running reports whether this launcher spawned a browser process.
func (l *Launcher) Running() bool {
	return l.running
}

// Stop terminates the browser process with SIGTERM, falling back to SIGKILL.
func (l *Launcher) Stop() {
	if l.cmd == nil || l.cmd.Process == nil {
		return
	}
	slog.Info("stopping browser", "pid", l.cmd.Process.Pid)
	_ = l.cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		_ = l.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("browser stopped gracefully")
	case <-time.After(5 * time.Second):
		slog.Warn("browser did not exit, sending SIGKILL")
		_ = l.cmd.Process.Kill()
		<-done
	}
	l.running = false
}
