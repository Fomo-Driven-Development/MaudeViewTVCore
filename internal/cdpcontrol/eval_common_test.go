package cdpcontrol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSStringAndJSONHelpers(t *testing.T) {
	if got := jsString("hello\nworld"); got != "\"hello\\nworld\"" {
		t.Fatalf("jsString = %q, want %q", got, "\"hello\\nworld\"")
	}

	got := jsJSON(map[string]any{"a": 1, "b": true})
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("jsJSON returned invalid JSON: %v", err)
	}
	if len(m) != 2 {
		t.Fatalf("jsJSON decoded map has %d fields, want 2", len(m))
	}
	if m["b"] != true {
		t.Fatalf("jsJSON decoded map = %v, want b=true", m["b"])
	}
}

func TestJSEvalWrappers(t *testing.T) {
	syncExpr := wrapJSEval("return 1;")
	if !strings.Contains(syncExpr, "(function(){\ntry {") {
		t.Fatalf("unexpected sync wrapper: %s", syncExpr)
	}
	if strings.Contains(syncExpr, "(async function") {
		t.Fatalf("sync wrapper should not be async: %s", syncExpr)
	}

	asyncExpr := wrapJSEvalAsync("await Promise.resolve(1);")
	if !strings.Contains(asyncExpr, "(async function(){\ntry {") {
		t.Fatalf("unexpected async wrapper: %s", asyncExpr)
	}
	if !strings.Contains(asyncExpr, "await Promise.resolve(1);") {
		t.Fatalf("async wrapper lost body: %s", asyncExpr)
	}
}
