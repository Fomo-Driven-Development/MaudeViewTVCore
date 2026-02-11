package pipeline

import (
	"context"
	"fmt"
	"io"

	"github.com/dgnsrekt/tv_agent/internal/mapper/correlation"
	"github.com/dgnsrekt/tv_agent/internal/mapper/reporting"
	"github.com/dgnsrekt/tv_agent/internal/mapper/runtimeprobes"
	"github.com/dgnsrekt/tv_agent/internal/mapper/staticanalysis"
)

const (
	ModeStaticOnly  = "static-only"
	ModeRuntimeOnly = "runtime-only"
	ModeCorrelate   = "correlate"
	ModeReport      = "report"
	ModeFull        = "full"
)

// Run executes one mapper mode.
func Run(ctx context.Context, w io.Writer, mode string) error {
	switch mode {
	case ModeStaticOnly:
		return staticanalysis.Run(ctx, w)
	case ModeRuntimeOnly:
		return runtimeprobes.Run(ctx, w)
	case ModeCorrelate:
		return correlation.Run(ctx, w)
	case ModeReport:
		return reporting.Run(ctx, w)
	case ModeFull:
		if err := staticanalysis.Run(ctx, w); err != nil {
			return err
		}
		if err := runtimeprobes.Run(ctx, w); err != nil {
			return err
		}
		if err := correlation.Run(ctx, w); err != nil {
			return err
		}
		return reporting.Run(ctx, w)
	default:
		return fmt.Errorf("unknown mode %q", mode)
	}
}
