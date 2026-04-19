package gameecs

import (
	"encoding/json"
	"testing"
	"time"

	"game/internal/domain/models"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

func TestEngine_JoinStateSnapshotFullTilesWhenPeriodicSyncSkips(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{TileFullSyncInterval: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	e.EnsurePlayerJoin(1, gamekit.NewDefaultCharacterPlayData())
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{X: 2, Y: 3, Texture: "rock", Blocks: true})
	evs := e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload}})
	if len(evs) != 1 {
		t.Fatalf("want 1 event, got %d", len(evs))
	}
	evs = e.ProcessTick(nil)
	if len(evs) != 1 {
		t.Fatalf("want 1 event on second tick, got %d", len(evs))
	}
	body, _ := json.Marshal(evs[0].Payload)
	var tickState gamekit.StatePayload
	if err := json.Unmarshal(body, &tickState); err != nil {
		t.Fatal(err)
	}
	if tickState.Tiles != nil {
		t.Fatalf("second tick should not include full tiles (hourly interval), got %d tiles", len(*tickState.Tiles))
	}

	join := e.JoinStateSnapshot()
	if join.Tiles == nil || len(*join.Tiles) != 1 {
		t.Fatalf("join snapshot must include full tiles, got %+v", join)
	}
	t0 := (*join.Tiles)[0]
	if t0.X != 2 || t0.Y != 3 || t0.Texture != "rock" || !t0.Blocks {
		t.Fatalf("unexpected tile %+v", t0)
	}
}

func TestEngine_JoinStateSnapshotDoesNotAdvanceLastFullTileSyncAt(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{TileFullSyncInterval: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	e.EnsurePlayerJoin(1, gamekit.NewDefaultCharacterPlayData())
	_ = e.ProcessTick(nil)
	at := e.lastFullTileSyncAt
	if at.IsZero() {
		t.Fatal("expected lastFullTileSyncAt set after first tick")
	}
	_ = e.JoinStateSnapshot()
	if !e.lastFullTileSyncAt.Equal(at) {
		t.Fatalf("JoinStateSnapshot must not change lastFullTileSyncAt")
	}
}
