package notify

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSendPostsCompletionMessage(t *testing.T) {
	ctx := context.Background()

	var receivedMethod string
	var receivedPath string
	var receivedBody string
	var receivedContentType string

	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			receivedMethod = r.Method
			receivedPath = r.URL.Path
			receivedContentType = r.Header.Get("Content-Type")
			rawBody, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			receivedBody = string(rawBody)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	if err := Send(ctx, client, "http://example.com/notifications", completionMessage); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if got, want := receivedMethod, http.MethodPost; got != want {
		t.Fatalf("method = %q; want %q", got, want)
	}
	if got, want := receivedPath, "/notifications"; got != want {
		t.Fatalf("path = %q; want %q", got, want)
	}
	if got, want := receivedContentType, "text/plain"; got != want {
		t.Fatalf("content-type = %q; want %q", got, want)
	}
	if got, want := receivedBody, completionMessage; got != want {
		t.Fatalf("body = %q; want %q", got, want)
	}
}

func TestSendReturnsErrorForServerError(t *testing.T) {
	ctx := context.Background()

	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("server failure")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	err := Send(ctx, client, "http://example.com/notifications", completionMessage)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "ntfy notification failed") {
		t.Fatalf("error = %q; want to contain %q", err, "ntfy notification failed")
	}
}

func TestSendDisallowsMissingEndpoint(t *testing.T) {
	ctx := context.Background()
	err := Send(ctx, http.DefaultClient, "", completionMessage)
	if err == nil {
		t.Fatal("expected error for missing endpoint")
	}
}
