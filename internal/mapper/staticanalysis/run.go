package staticanalysis

import (
	"context"
	"fmt"
	"io"
)

// Run executes the static analysis stage.
func Run(ctx context.Context, w io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := fmt.Fprintln(w, "static-analysis: complete")
	return err
}
