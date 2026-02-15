package cdpcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/target"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func withDefaultHTTPClient(t *testing.T, transport http.RoundTripper) {
	t.Helper()
	origClient := http.DefaultClient
	t.Cleanup(func() {
		http.DefaultClient = origClient
	})
	http.DefaultClient = &http.Client{
		Transport: transport,
	}
}

func TestSyncTabsLockedWrapsListTargetsError(t *testing.T) {
	withDefaultHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/json/list" {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`oops`)),
			}, nil
		}
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(``))}, nil
	}))

	c := &Client{
		cdp:           newRawCDP("http://example.com"),
		tabs:          map[target.ID]*tabSession{},
		chartToTarget: map[string]target.ID{},
	}

	err := c.syncTabsLocked(context.Background())
	if err == nil {
		t.Fatal("expected syncTabsLocked() to fail")
	}

	var codedErr *CodedError
	if !errors.As(err, &codedErr) {
		t.Fatalf("expected *CodedError, got %T", err)
	}
	if codedErr.Code != CodeCDPUnavailable {
		t.Fatalf("error code = %s; want %s", codedErr.Code, CodeCDPUnavailable)
	}
	if !strings.Contains(codedErr.Message, "failed to list targets") {
		t.Fatalf("error message = %q; want to contain %q", codedErr.Message, "failed to list targets")
	}
}

func TestRawCDPOperationHelpersWrapErrors(t *testing.T) {
	targetID := target.ID("target-1")
	chartID := "abc"
	sessionID := "session-1"

	withDefaultHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/json/list" {
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(``))}, nil
		}

		targets := []map[string]any{
			{
				"id":    string(targetID),
				"type":  "page",
				"url":   "https://example.com/chart/" + chartID + "/",
				"title": "test",
			},
		}
		payload, marshalErr := json.Marshal(targets)
		if marshalErr != nil {
			t.Fatalf("json.Marshal() = %v", marshalErr)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(payload))),
		}, nil
	}))

	c := &Client{
		cdp: newRawCDP("http://example.com"),
		tabs: map[target.ID]*tabSession{
			targetID: {
				sessionID: sessionID,
				info: ChartInfo{
					ChartID:  chartID,
					TargetID: string(targetID),
					URL:      "https://example.com/chart/" + chartID + "/",
					Title:    "test",
				},
			},
		},
		chartToTarget: map[string]target.ID{},
	}

	tests := []struct {
		name   string
		run    func() error
		errMsg string
	}{
		{
			name:   "sendKeysOnAnyChart",
			run:    func() error { return c.sendKeysOnAnyChart(context.Background(), "k", "KeyK", 75, 0) },
			errMsg: "failed to dispatch trusted key event",
		},
		{
			name:   "clickOnAnyChart",
			run:    func() error { return c.clickOnAnyChart(context.Background(), 10, 10) },
			errMsg: "failed to dispatch trusted mouse click",
		},
		{
			name:   "insertTextOnAnyChart",
			run:    func() error { return c.insertTextOnAnyChart(context.Background(), "hello") },
			errMsg: "failed to dispatch trusted text insertion",
		},
		{
			name:   "typeTextOnAnyChart",
			run:    func() error { return c.typeTextOnAnyChart(context.Background(), "abc") },
			errMsg: "failed to dispatch trusted character input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil {
				t.Fatalf("%s = nil; want error", tt.name)
			}
			var codedErr *CodedError
			if !errors.As(err, &codedErr) {
				t.Fatalf("%s returned %T; want *CodedError", tt.name, err)
			}
			if codedErr.Code != CodeEvalFailure {
				t.Fatalf("%s code = %s; want %s", tt.name, codedErr.Code, CodeEvalFailure)
			}
			if !strings.Contains(codedErr.Message, tt.errMsg) {
				t.Fatalf("%s message = %q; want to contain %q", tt.name, codedErr.Message, tt.errMsg)
			}
		})
	}
}
