package reporting

import (
	"context"
	"fmt"
	"io"
)

// Run executes the reporting stage.
func Run(ctx context.Context, w io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := fmt.Fprintln(w, "reporting: complete")
	return err
}
