package gameecs

import "testing"

func TestSystemRegistryPipelineNamesStable(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}

	gotPerAction := e.systems.PerActionPipelineNames()
	wantPerAction := []string{
		PipelineMoveIntentCapture,
		PipelineDamage,
		PipelineInventoryMove,
		PipelinePickup,
		PipelineDropItem,
		PipelineTileSpawn,
		PipelineTileClear,
		PipelineInteract,
	}
	if len(gotPerAction) != len(wantPerAction) {
		t.Fatalf("per-action length mismatch: got %d want %d", len(gotPerAction), len(wantPerAction))
	}
	for i := range wantPerAction {
		if gotPerAction[i] != wantPerAction[i] {
			t.Fatalf("per-action pipeline changed at %d: got %q want %q", i, gotPerAction[i], wantPerAction[i])
		}
	}

	gotPostTick := e.systems.PostTickPipelineNames()
	wantPostTick := []string{
		PipelineMovementApply,
		PipelineSnapshot,
	}
	if len(gotPostTick) != len(wantPostTick) {
		t.Fatalf("post-tick length mismatch: got %d want %d", len(gotPostTick), len(wantPostTick))
	}
	for i := range wantPostTick {
		if gotPostTick[i] != wantPostTick[i] {
			t.Fatalf("post-tick pipeline changed at %d: got %q want %q", i, gotPostTick[i], wantPostTick[i])
		}
	}
}
