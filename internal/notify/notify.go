package notify

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	defaultNTFYEndpoint = "http://nas1-oryx:2586/notifications"
	completionMessage   = "All tasks for the tv agent refactoring are complete. The eval god file was split into nine domain files, keyboard shortcut boilerplate was extracted, validation helpers were added, truncation logic was consolidated, input types were renamed, server handlers were split into domain files, and silent error discards now have debug logging."
)

// SendCompletion sends the predefined completion notification to the NTFY endpoint.
func SendCompletion(ctx context.Context, client *http.Client) error {
	return Send(ctx, client, defaultNTFYEndpoint, completionMessage)
}

// Send sends a message to the requested endpoint using HTTP POST.
func Send(ctx context.Context, client *http.Client, endpoint, message string) error {
	c := client
	if c == nil {
		c = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(message))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy notification failed: status=%d", resp.StatusCode)
	}
	return nil
}
