package types

// TabInfo holds metadata about a browser tab for routing captured data.
// This is used by capture handlers to determine which writer to use.
type TabInfo struct {
	TargetID    string
	URL         string
	PathSegment string // Transformed URL path, e.g., "flow_overview"
	BrowserID   string // Short ID from target ID, e.g., "B0D5A8E8"
}

// TabInfoProvider is an interface for looking up tab information by ID.
// This breaks the import cycle between capture and cdp packages.
type TabInfoProvider interface {
	GetByStringID(tabID string) (*TabInfo, bool)
}
