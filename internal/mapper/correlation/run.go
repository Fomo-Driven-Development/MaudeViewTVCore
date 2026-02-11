package correlation

import (
	"context"
	"fmt"
	"io"
)

// Run executes the correlation stage.
func Run(ctx context.Context, w io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := fmt.Fprintln(w, "correlation: complete")
	return err
}
