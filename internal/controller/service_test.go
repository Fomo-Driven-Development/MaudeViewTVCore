package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/dgnsrekt/MaudeViewTVCore/internal/cdpcontrol"
)

func TestRequireNonEmpty(t *testing.T) {
	s := &Service{}
	if err := s.requireNonEmpty("AAPL", "symbol"); err != nil {
		t.Fatalf("requireNonEmpty() = %v; want nil", err)
	}

	if err := s.requireNonEmpty("   ", "symbol"); err == nil {
		t.Fatalf("requireNonEmpty() = nil; want validation error")
	} else if got, ok := err.(*cdpcontrol.CodedError); !ok {
		t.Fatalf("requireNonEmpty() = %T; want *cdpcontrol.CodedError", err)
	} else if got.Code != cdpcontrol.CodeValidation {
		t.Fatalf("requireNonEmpty() code = %q; want %q", got.Code, cdpcontrol.CodeValidation)
	} else if got.Message != "symbol is required" {
		t.Fatalf("requireNonEmpty() message = %q; want %q", got.Message, "symbol is required")
	}
}

func TestSetSymbol_RequiresNonEmptySymbol(t *testing.T) {
	s := &Service{}
	_, err := s.SetSymbol(context.Background(), "chart-id", "   ", 0)
	if err == nil {
		t.Fatalf("SetSymbol() = nil; want validation error")
	}
	var got *cdpcontrol.CodedError
	if !errors.As(err, &got) {
		t.Fatalf("SetSymbol() error type = %T; want *cdpcontrol.CodedError", err)
	}
	if got.Code != cdpcontrol.CodeValidation {
		t.Fatalf("SetSymbol() code = %q; want %q", got.Code, cdpcontrol.CodeValidation)
	}
	if got.Message != "symbol is required" {
		t.Fatalf("SetSymbol() message = %q; want %q", got.Message, "symbol is required")
	}
}
