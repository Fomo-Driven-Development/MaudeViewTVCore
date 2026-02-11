package tv_agent_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMapperScriptsExist(t *testing.T) {
	scripts := []string{
		"mapper-static-only.sh",
		"mapper-runtime-only.sh",
		"mapper-correlate.sh",
		"mapper-report.sh",
		"mapper-full.sh",
	}

	for _, script := range scripts {
		path := filepath.Join("scripts", script)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("os.Stat(%q) error = %v", path, err)
		}
		if info.IsDir() {
			t.Fatalf("%q is a directory, want file", path)
		}
		if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
			t.Fatalf("%q is not executable", path)
		}
	}
}
