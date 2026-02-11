package tv_agent_test

import (
	"os"
	"strings"
	"testing"
)

func TestMapperRunbookCoverage(t *testing.T) {
	path := "docs/mapper-runbook.md"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	content := string(data)

	required := []string{
		"## Prerequisites",
		"## Environment Variables",
		"CHROMIUM_CDP_ADDRESS",
		"CHROMIUM_CDP_PORT",
		"CHROMIUM_START_URL",
		"RESEARCHER_DATA_DIR",
		"## Command Order",
		"just start-browser",
		"just run-researcher",
		"just mapper-full",
		"just mapper-validate",
		"## Expected `research_data/` Layout",
		"js-bundle-index.jsonl",
		"runtime-trace.jsonl",
		"capability-correlations.jsonl",
		"capability-matrix.jsonl",
		"## Passive-Only Guardrails",
		"## Known Limitations (Minified Bundle Analysis)",
	}

	for _, needle := range required {
		if !strings.Contains(content, needle) {
			t.Fatalf("runbook missing required content %q", needle)
		}
	}
}
