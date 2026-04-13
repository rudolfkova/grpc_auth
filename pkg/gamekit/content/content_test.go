package content

import (
	stdctx "context"
	"path/filepath"
	"testing"
)

func TestMergeShallow(t *testing.T) {
	m := mergeShallow(
		map[string]any{"a": 1, "b": 2},
		map[string]any{"b": 3, "c": 4},
	)
	if m["a"] != 1 || m["b"] != 3 || m["c"] != 4 {
		t.Fatalf("merge: %#v", m)
	}
}

func TestLoadBundleAndRun(t *testing.T) {
	dir := filepath.Join("testdata", "scripts")
	b, err := LoadBundle(filepath.Join("testdata", "catalog.json"), dir)
	if err != nil {
		t.Fatal(err)
	}
	sc := b.Scenarios["demo.json"]
	var log string
	rcx := &RunContext{PlayerID: 7, ItemDefID: "demo_item", Log: func(s string) { log = s }}
	if err := b.Runner.Run(stdctx.Background(), rcx, sc, map[string]any{"message": "from base"}); err != nil {
		t.Fatal(err)
	}
	if log != "step" {
		t.Fatalf("log: %q", log)
	}
}
