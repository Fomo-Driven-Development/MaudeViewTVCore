package relay

import (
	"fmt"
	"net/http"
	"strings"
)

// SSEHandler returns an http.HandlerFunc that streams relay events as SSE.
// Clients may filter feeds via ?feeds=name1,name2 query parameter.
func SSEHandler(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		// Parse optional feed filter.
		var feedFilter map[string]bool
		if q := r.URL.Query().Get("feeds"); q != "" {
			feedFilter = make(map[string]bool)
			for _, f := range strings.Split(q, ",") {
				if f = strings.TrimSpace(f); f != "" {
					feedFilter[f] = true
				}
			}
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher.Flush()

		id, ch := broker.Subscribe()
		defer broker.Unsubscribe(id)

		for {
			select {
			case <-r.Context().Done():
				return
			case evt, ok := <-ch:
				if !ok {
					return
				}
				if feedFilter != nil && !feedFilter[evt.Feed] {
					continue
				}
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Feed, evt.Payload)
				flusher.Flush()
			}
		}
	}
}
