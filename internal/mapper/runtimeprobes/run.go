package runtimeprobes

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/dgnsrekt/tv_agent/internal/config"
)

// Run executes the runtime probes stage.
func Run(ctx context.Context, w io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	attached, injected, err := runProbeBootstrap(ctx, cfg, slog.Default())
	if err != nil {
		slog.Warn(
			"Runtime probes bootstrap unavailable",
			"error",
			err,
			"cdp_url",
			cfg.GetCDPURL(),
		)
	} else {
		slog.Info("Runtime probes bootstrap complete", "attached_tabs", attached, "injected_tabs", injected)
	}

	_, err = fmt.Fprintln(w, "runtime-probes: complete")
	return err
}

type probeTarget struct {
	ID  target.ID
	URL string
}

type probeBootstrapResult struct {
	AlreadyInjected bool   `json:"alreadyInjected"`
	URL             string `json:"url"`
}

type probeLifecycle struct {
	Attached bool
	Result   probeBootstrapResult
}

type probeRunner func(context.Context, probeTarget) (probeLifecycle, error)

const passiveProbeBootstrapJS = `(function () {
  const key = "__tvAgentPassiveProbe";
  const alreadyInjected = Object.prototype.hasOwnProperty.call(window, key);
  if (!alreadyInjected) {
    Object.defineProperty(window, key, {
      value: Object.freeze({
        version: "1",
        injectedAt: new Date().toISOString(),
      }),
      configurable: true
    });
  }
  return {
    alreadyInjected: alreadyInjected,
    url: String(window.location.href || "")
  };
})();`

func runProbeBootstrap(ctx context.Context, cfg *config.Config, logger *slog.Logger) (int, int, error) {
	cdpURL := cfg.GetCDPURL()
	logger.Info("Runtime probes bootstrap start", "cdp_url", cdpURL, "tab_url_filter", cfg.TabURLFilter)

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, cdpURL)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	if err := chromedp.Run(browserCtx); err != nil {
		return 0, 0, fmt.Errorf("connect to browser: %w", err)
	}

	targets, err := chromedp.Targets(browserCtx)
	if err != nil {
		return 0, 0, fmt.Errorf("enumerate targets: %w", err)
	}

	matches := filterProbeTargets(targets, cfg.TabURLFilter)
	if len(matches) == 0 {
		logger.Warn("Runtime probes found no matching tabs", "tab_url_filter", cfg.TabURLFilter)
		return 0, 0, nil
	}

	return bootstrapTargets(ctx, logger, matches, makeCDPRunner(allocCtx))
}

func filterProbeTargets(targets []*target.Info, urlFilter string) []probeTarget {
	filter := strings.ToLower(urlFilter)
	matches := make([]probeTarget, 0, len(targets))
	for _, t := range targets {
		if t.Type != "page" {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(t.URL), filter) {
			continue
		}
		matches = append(matches, probeTarget{ID: t.TargetID, URL: t.URL})
	}
	return matches
}

func bootstrapTargets(ctx context.Context, logger *slog.Logger, targets []probeTarget, runner probeRunner) (int, int, error) {
	attached := 0
	injected := 0

	for _, tab := range targets {
		lifecycle, err := runner(ctx, tab)

		if lifecycle.Attached {
			attached++
			logger.Info("Runtime probe attach success", "tab_id", tab.ID, "url", truncateURL(tab.URL))
		}

		if err != nil {
			stage := "attach"
			if lifecycle.Attached {
				stage = "inject"
			}
			logger.Warn("Runtime probe lifecycle failed", "stage", stage, "tab_id", tab.ID, "error", err)
			continue
		}

		injected++
		logger.Info(
			"Runtime probe inject success",
			"tab_id",
			tab.ID,
			"already_injected",
			lifecycle.Result.AlreadyInjected,
			"url",
			truncateURL(lifecycle.Result.URL),
		)
	}

	return attached, injected, nil
}

func makeCDPRunner(allocCtx context.Context) probeRunner {
	return func(_ context.Context, tab probeTarget) (probeLifecycle, error) {
		tabCtx, tabCancel := chromedp.NewContext(allocCtx, chromedp.WithTargetID(tab.ID))
		defer tabCancel()

		attachCtx, attachCancel := context.WithTimeout(tabCtx, 10*time.Second)
		defer attachCancel()
		if err := chromedp.Run(attachCtx); err != nil {
			return probeLifecycle{}, fmt.Errorf("attach to tab: %w", err)
		}

		probeCtx, probeCancel := context.WithTimeout(tabCtx, 10*time.Second)
		defer probeCancel()

		var result probeBootstrapResult
		if err := chromedp.Run(probeCtx, chromedp.Evaluate(passiveProbeBootstrapJS, &result)); err != nil {
			return probeLifecycle{Attached: true}, fmt.Errorf("inject probe: %w", err)
		}

		return probeLifecycle{Attached: true, Result: result}, nil
	}
}

func truncateURL(url string) string {
	if len(url) > 120 {
		return url[:120] + "..."
	}
	return url
}
