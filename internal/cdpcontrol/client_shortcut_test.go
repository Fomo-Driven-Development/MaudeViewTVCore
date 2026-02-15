package cdpcontrol

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSendShortcutDispatchesAndSettles(t *testing.T) {
	c := &Client{}

	var got struct {
		key       string
		code      string
		keyCode   int
		modifiers int
		settle    time.Duration
	}

	origDispatch := sendShortcutDispatch
	origWait := sendShortcutWait
	t.Cleanup(func() {
		sendShortcutDispatch = origDispatch
		sendShortcutWait = origWait
	})

	sendShortcutDispatch = func(_ *Client, _ context.Context, key, code string, keyCode, modifiers int) error {
		got.key = key
		got.code = code
		got.keyCode = keyCode
		got.modifiers = modifiers
		return nil
	}
	sendShortcutWait = func(_ context.Context, d time.Duration) error {
		got.settle = d
		return nil
	}

	err := c.sendShortcut(context.Background(), "r", "KeyR", 82, 1, uiSettleLong, "failed to send Alt+R")
	if err != nil {
		t.Fatalf("sendShortcut() = %v; want nil", err)
	}
	if got.key != "r" || got.code != "KeyR" || got.keyCode != 82 || got.modifiers != 1 {
		t.Fatalf("sendShortcut dispatch = (%q, %q, %d, %d); want (\"r\", \"KeyR\", 82, 1)", got.key, got.code, got.keyCode, got.modifiers)
	}
	if got.settle != uiSettleLong {
		t.Fatalf("sendShortcut settle = %s; want %s", got.settle, uiSettleLong)
	}
}

func TestSendShortcutWrapsDispatchError(t *testing.T) {
	c := &Client{}
	dispatchErr := errors.New("dispatch failed")

	origDispatch := sendShortcutDispatch
	origWait := sendShortcutWait
	t.Cleanup(func() {
		sendShortcutDispatch = origDispatch
		sendShortcutWait = origWait
	})

	sendShortcutDispatch = func(*Client, context.Context, string, string, int, int) error {
		return dispatchErr
	}
	sendShortcutWait = func(context.Context, time.Duration) error {
		t.Fatalf("sendShortcutWait should not run when dispatch fails")
		return nil
	}

	err := c.sendShortcut(context.Background(), "z", "KeyZ", 90, 2, uiSettleMedium, "failed to send Ctrl+Z")
	if err == nil {
		t.Fatal("expected sendShortcut() to return an error")
	}

	var codedErr *CodedError
	if !errors.As(err, &codedErr) {
		t.Fatalf("expected *CodedError, got %T", err)
	}
	if codedErr.Code != CodeEvalFailure {
		t.Fatalf("error code = %s; want %s", codedErr.Code, CodeEvalFailure)
	}
	if !strings.Contains(codedErr.Message, "failed to send Ctrl+Z") {
		t.Fatalf("error message = %q; want to contain %q", codedErr.Message, "failed to send Ctrl+Z")
	}
	if !errors.Is(err, dispatchErr) {
		t.Fatalf("error should wrap dispatch error")
	}
}

func TestShortcutMethodsCallExpectedInputs(t *testing.T) {
	type call struct {
		key       string
		code      string
		keyCode   int
		modifiers int
		settle    time.Duration
	}
	c := &Client{}

	origDispatch := sendShortcutDispatch
	origWait := sendShortcutWait
	t.Cleanup(func() {
		sendShortcutDispatch = origDispatch
		sendShortcutWait = origWait
	})

	sendShortcutDispatch = func(_ *Client, _ context.Context, key, code string, keyCode, modifiers int) error {
		return nil
	}
	sendShortcutWait = func(_ context.Context, _ time.Duration) error {
		return nil
	}

	tests := []struct {
		name      string
		invoke    func(context.Context, string) error
		want      call
		wantChart string
	}{
		{
			name:      "ResetView",
			invoke:    c.ResetView,
			want:      call{key: "r", code: "KeyR", keyCode: 82, modifiers: 1, settle: uiSettleLong},
			wantChart: "chart-id",
		},
		{
			name:      "UndoChart",
			invoke:    c.UndoChart,
			want:      call{key: "z", code: "KeyZ", keyCode: 90, modifiers: 2, settle: uiSettleMedium},
			wantChart: "chart-id",
		},
		{
			name:      "RedoChart",
			invoke:    c.RedoChart,
			want:      call{key: "y", code: "KeyY", keyCode: 89, modifiers: 2, settle: uiSettleMedium},
			wantChart: "chart-id",
		},
		{
			name:      "ToggleLogScale",
			invoke:    c.ToggleLogScale,
			want:      call{key: "l", code: "KeyL", keyCode: 76, modifiers: 1, settle: uiSettleMedium},
			wantChart: "chart-id",
		},
		{
			name:      "ToggleAutoScale",
			invoke:    c.ToggleAutoScale,
			want:      call{key: "a", code: "KeyA", keyCode: 65, modifiers: 1, settle: uiSettleMedium},
			wantChart: "chart-id",
		},
		{
			name:      "ToggleExtendedHours",
			invoke:    c.ToggleExtendedHours,
			want:      call{key: "e", code: "KeyE", keyCode: 69, modifiers: 1, settle: uiSettleMedium},
			wantChart: "chart-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotCall call

			sendShortcutDispatch = func(_ *Client, _ context.Context, key, code string, keyCode, modifiers int) error {
				gotCall.key = key
				gotCall.code = code
				gotCall.keyCode = keyCode
				gotCall.modifiers = modifiers
				return nil
			}
			sendShortcutWait = func(_ context.Context, d time.Duration) error {
				gotCall.settle = d
				return nil
			}
			if err := tt.invoke(context.Background(), tt.wantChart); err != nil {
				t.Fatalf("%s() = %v; want nil", tt.name, err)
			}
			if gotCall != (call{}) {
				if gotCall.key != tt.want.key || gotCall.code != tt.want.code || gotCall.keyCode != tt.want.keyCode || gotCall.modifiers != tt.want.modifiers || gotCall.settle != tt.want.settle {
					t.Fatalf("%s call = %+v; want %+v", tt.name, gotCall, tt.want)
				}
			} else {
				t.Fatalf("did not capture call for %s", tt.name)
			}
		})
	}
}
