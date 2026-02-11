package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/dgnsrekt/tv_agent/internal/mapper/pipeline"
)

func main() {
	mode := flag.String("mode", pipeline.ModeFull, "pipeline mode: static-only|runtime-only|correlate|report|validate|full")
	flag.Parse()

	if err := pipeline.Run(context.Background(), os.Stdout, *mode); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "mapper failed: %v\n", err)
		os.Exit(1)
	}
}
