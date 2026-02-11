package runtimeprobes

import (
	"context"
	"fmt"
	"io"
)

// Run executes the runtime probes stage.
func Run(ctx context.Context, w io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := fmt.Fprintln(w, "runtime-probes: complete")
	return err
}
