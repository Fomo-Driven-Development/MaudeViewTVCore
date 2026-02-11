package cdp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/dgnsrekt/tv_agent/internal/capture"
	"github.com/dgnsrekt/tv_agent/internal/config"
)

// Client manages CDP connections to browser tabs.
type Client struct {
	cfg         *config.Config
	httpCapture *capture.HTTPCapture
	wsCapture   *capture.WebSocketCapture
	tabRegistry *TabRegistry
	allocCtx    context.Context
	allocCancel context.CancelFunc
	tabs        map[target.ID]*TabContext
	tabsMu      sync.RWMutex
	done        chan struct{}
}

type TabContext struct {
	ID     target.ID
	URL    string
	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient(cfg *config.Config, httpCapture *capture.HTTPCapture, wsCapture *capture.WebSocketCapture, tabRegistry *TabRegistry) *Client {
	return &Client{
		cfg:         cfg,
		httpCapture: httpCapture,
		wsCapture:   wsCapture,
		tabRegistry: tabRegistry,
		tabs:        make(map[target.ID]*TabContext),
		done:        make(chan struct{}),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	_ = ctx
	cdpURL := c.cfg.GetCDPURL()
	slog.Info("Connecting to Chromium", "url", cdpURL)

	c.allocCtx, c.allocCancel = chromedp.NewRemoteAllocator(context.Background(), cdpURL)

	tempCtx, tempCancel := chromedp.NewContext(c.allocCtx)
	defer tempCancel()

	if err := chromedp.Run(tempCtx); err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}

	targets, err := chromedp.Targets(tempCtx)
	if err != nil {
		return fmt.Errorf("failed to enumerate targets: %w", err)
	}

	slog.Info("Found browser targets", "count", len(targets))

	attachedCount := 0
	for _, t := range targets {
		if t.Type != "page" {
			continue
		}
		if !c.matchesTabURL(t.URL) {
			slog.Debug("Skipping tab (url filter)", "url", t.URL)
			continue
		}
		if err := c.attachToTab(t.TargetID, t.URL); err != nil {
			slog.Error("Failed to attach to tab", "target_id", t.TargetID, "url", t.URL, "error", err)
			continue
		}
		attachedCount++
	}

	if attachedCount == 0 {
		return fmt.Errorf("no tabs found matching RESEARCHER_TAB_URL_FILTER=%q", c.cfg.TabURLFilter)
	}

	slog.Info("Attached to tabs", "count", attachedCount, "tab_url_filter", c.cfg.TabURLFilter)
	return nil
}

func (c *Client) attachToTab(targetID target.ID, url string) error {
	tabInfo, err := c.tabRegistry.Register(targetID, url)
	if err != nil {
		return fmt.Errorf("failed to register tab: %w", err)
	}

	tabCtx, tabCancel := chromedp.NewContext(c.allocCtx, chromedp.WithTargetID(targetID))
	tab := &TabContext{ID: targetID, URL: url, ctx: tabCtx, cancel: tabCancel}

	c.tabsMu.Lock()
	c.tabs[targetID] = tab
	c.tabsMu.Unlock()

	if err := chromedp.Run(tabCtx, network.Enable(), network.SetCacheDisabled(true), page.Enable()); err != nil {
		tabCancel()
		c.tabRegistry.Remove(targetID)
		return fmt.Errorf("failed to enable network/page domains: %w", err)
	}

	slog.Info("Attached to tab", "target_id", targetID, "path_segment", tabInfo.PathSegment, "browser_id", tabInfo.BrowserID, "url", truncateURL(url))
	chromedp.ListenTarget(tabCtx, c.createEventHandler(string(targetID)))

	if c.cfg.ReloadOnAttach {
		reloadCtx, reloadCancel := context.WithTimeout(tabCtx, 30*time.Second)
		defer reloadCancel()
		if err := chromedp.Run(reloadCtx, chromedp.Reload()); err != nil {
			slog.Warn("Failed to reload tab (continuing)", "target_id", targetID, "error", err)
		} else {
			slog.Info("Reloaded tab after attach", "target_id", targetID, "url", truncateURL(url))
		}
	}

	return nil
}

func (c *Client) createEventHandler(tabID string) func(ev interface{}) {
	return func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventFrameNavigated:
			if e.Frame.ParentID == "" {
				if info, err := c.tabRegistry.Register(target.ID(tabID), e.Frame.URL); err == nil {
					slog.Info("Tab navigated (full)", "tab_id", tabID, "path_segment", info.PathSegment, "url", truncateURL(e.Frame.URL))
				}
			}
		case *page.EventNavigatedWithinDocument:
			if info, err := c.tabRegistry.Register(target.ID(tabID), e.URL); err == nil {
				slog.Info("Tab navigated (SPA)", "tab_id", tabID, "path_segment", info.PathSegment, "url", truncateURL(e.URL))
			}
		case *network.EventRequestWillBeSent:
			c.httpCapture.OnRequestWillBeSent(tabID, e)
		case *network.EventResponseReceived:
			c.httpCapture.OnResponseReceived(tabID, e)
		case *network.EventLoadingFinished:
			c.tabsMu.RLock()
			tab, ok := c.tabs[target.ID(tabID)]
			c.tabsMu.RUnlock()

			var getBody func() ([]byte, bool, error)
			if ok {
				tabCtx := tab.ctx
				getBody = func() ([]byte, bool, error) {
					bodyCtx, bodyCancel := context.WithTimeout(tabCtx, 10*time.Second)
					defer bodyCancel()

					var body []byte
					err := chromedp.Run(bodyCtx, chromedp.ActionFunc(func(ctx context.Context) error {
						var err error
						body, err = network.GetResponseBody(e.RequestID).Do(ctx)
						return err
					}))
					return body, false, err
				}
			}
			c.httpCapture.OnLoadingFinished(tabID, e, getBody)
		case *network.EventLoadingFailed:
			c.httpCapture.OnLoadingFailed(tabID, e)
		case *network.EventWebSocketCreated:
			c.wsCapture.OnWebSocketCreated(tabID, e)
		case *network.EventWebSocketFrameReceived:
			c.wsCapture.OnWebSocketFrameReceived(tabID, e)
		case *network.EventWebSocketFrameSent:
			c.wsCapture.OnWebSocketFrameSent(tabID, e)
		case *network.EventWebSocketClosed:
			c.wsCapture.OnWebSocketClosed(tabID, e)
		}
	}
}

func (c *Client) Close() error {
	close(c.done)

	c.tabsMu.Lock()
	defer c.tabsMu.Unlock()
	c.tabs = make(map[target.ID]*TabContext)

	if c.allocCancel != nil {
		c.allocCancel()
	}

	slog.Info("CDP client closed")
	return nil
}

func (c *Client) GetTabCount() int {
	c.tabsMu.RLock()
	defer c.tabsMu.RUnlock()
	return len(c.tabs)
}

func (c *Client) matchesTabURL(url string) bool {
	if c.cfg.TabURLFilter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(url), strings.ToLower(c.cfg.TabURLFilter))
}

func truncateURL(url string) string {
	if len(url) > 120 {
		return url[:120] + "..."
	}
	return url
}
