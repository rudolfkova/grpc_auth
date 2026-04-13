package gamekit

import (
	"encoding/json"
	"testing"
)

func TestNormalizeTileInstanceArgsJSON(t *testing.T) {
	if got := NormalizeTileInstanceArgsJSON(json.RawMessage(`[1,2]`)); got != nil {
		t.Fatalf("array: %q", got)
	}
	if got := NormalizeTileInstanceArgsJSON(json.RawMessage(`{}`)); got != nil {
		t.Fatalf("empty object: %q", got)
	}
	if got := NormalizeTileInstanceArgsJSON(json.RawMessage(`{"a":1}`)); string(got) != `{"a":1}` {
		t.Fatalf("object: %q", got)
	}
	oversize := make([]byte, MaxTileInstanceArgsJSONBytes+1)
	if got := NormalizeTileInstanceArgsJSON(json.RawMessage(oversize)); got != nil {
		t.Fatal("expected oversize rejected")
	}
}
