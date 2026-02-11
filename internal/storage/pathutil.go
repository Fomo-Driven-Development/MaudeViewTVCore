package storage

import (
	"net/url"
	stdpath "path"
	"strings"
)

// TransformURLToPathSegment transforms a URL path into a filesystem-safe path segment.
func TransformURLToPathSegment(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	path := strings.TrimPrefix(parsed.Path, "/")
	if path == "" {
		return "root", nil
	}
	path = strings.TrimSuffix(path, "/")
	path = strings.ReplaceAll(path, "/", "_")
	return path, nil
}

// BrowserIDFromTargetID returns the first 8 chars of a CDP target ID.
func BrowserIDFromTargetID(targetID string) string {
	if len(targetID) >= 8 {
		return targetID[:8]
	}
	return targetID
}

// MapResourceType maps CDP ResourceType to a static resource directory.
// Empty means treat as API/event traffic and keep in JSONL only.
func MapResourceType(resourceType string) string {
	switch resourceType {
	case "XHR", "Fetch", "WebSocket", "EventSource", "Ping":
		return ""
	case "Script":
		return "js"
	case "Stylesheet":
		return "css"
	case "Image":
		return "img"
	case "Font":
		return "font"
	case "Media":
		return "media"
	case "Document":
		return "docs"
	case "Manifest":
		return "manifest"
	default:
		return "other"
	}
}

// FilenameFromURL extracts a filename from URL path.
func FilenameFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "resource"
	}
	filename := stdpath.Base(parsed.Path)
	if filename == "" || filename == "." || filename == "/" {
		return "resource"
	}
	return filename
}
