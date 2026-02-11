package cdp

import (
	"sync"

	"github.com/chromedp/cdproto/target"
	"github.com/dgnsrekt/tv_agent/internal/storage"
	"github.com/dgnsrekt/tv_agent/internal/types"
)

// TabRegistry maps CDP target IDs to tab metadata.
type TabRegistry struct {
	tabs map[target.ID]*types.TabInfo
	mu   sync.RWMutex
}

func NewTabRegistry() *TabRegistry {
	return &TabRegistry{tabs: make(map[target.ID]*types.TabInfo)}
}

func (r *TabRegistry) Register(targetID target.ID, url string) (*types.TabInfo, error) {
	pathSegment, err := storage.TransformURLToPathSegment(url)
	if err != nil {
		return nil, err
	}

	info := &types.TabInfo{
		TargetID:    string(targetID),
		URL:         url,
		PathSegment: pathSegment,
		BrowserID:   storage.BrowserIDFromTargetID(string(targetID)),
	}

	r.mu.Lock()
	r.tabs[targetID] = info
	r.mu.Unlock()

	return info, nil
}

func (r *TabRegistry) Get(targetID target.ID) (*types.TabInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.tabs[targetID]
	return info, ok
}

func (r *TabRegistry) GetByStringID(tabID string) (*types.TabInfo, bool) {
	return r.Get(target.ID(tabID))
}

func (r *TabRegistry) Remove(targetID target.ID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tabs, targetID)
}

func (r *TabRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tabs)
}
