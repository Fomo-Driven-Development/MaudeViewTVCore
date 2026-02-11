package staticanalysis

import (
	"context"
	"fmt"
	"io"
	"os"
)

// Run executes the static analysis stage.
func Run(ctx context.Context, w io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	dataDir := os.Getenv("RESEARCHER_DATA_DIR")
	if dataDir == "" {
		dataDir = defaultDataDir
	}
	if err := indexJSBundles(ctx, dataDir); err != nil {
		return err
	}

	_, err := fmt.Fprintln(w, "static-analysis: complete")
	return err
}
