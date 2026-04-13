package content

import (
	"encoding/json"
	"testing"
)

func TestMergeInteractBase(t *testing.T) {
	cat := map[string]any{"a": 1, "b": 2}
	tileRaw := json.RawMessage(`{"b":3,"c":4}`)
	m := MergeInteractBase(cat, tileRaw)
	if m["a"] != 1 || m["b"] != float64(3) || m["c"] != float64(4) {
		t.Fatalf("merge: %#v", m)
	}
}
