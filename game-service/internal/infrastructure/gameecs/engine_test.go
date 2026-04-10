package gameecs

import (
	"encoding/json"
	"testing"

	"game/internal/domain/models"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	arkserde "github.com/mlange-42/ark-serde"
)

func TestEngine_MoveAndHit(t *testing.T) {
	e, err := NewEngine(nil)
	if err != nil {
		t.Fatal(err)
	}

	movePayload, _ := json.Marshal(gamekit.MoveIntent{DX: 1, DY: 0})
	hitPayload, _ := json.Marshal(gamekit.HitIntent{TargetID: 2, Damage: 3})

	e.ProcessTick([]models.Action{
		{PlayerID: 1, Type: "move", Payload: movePayload},
		{PlayerID: 1, Type: "hit", Payload: hitPayload},
	})

	e.mu.Lock()
	defer e.mu.Unlock()

	ent2 := e.EnsurePlayerEntity(2)
	_, _, _, hp := e.playerMapper.Get(ent2)
	if hp.HP != gamekit.DefaultPlayerHP-3 {
		t.Fatalf("target HP: want %d, got %d", gamekit.DefaultPlayerHP-3, hp.HP)
	}
	ent1 := e.EnsurePlayerEntity(1)
	_, pos, _, _ := e.playerMapper.Get(ent1)
	if pos.X != 1 || pos.Y != 0 {
		t.Fatalf("mover position: want (1,0), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestEngine_spawnTileAppearsInState(t *testing.T) {
	e, err := NewEngine(nil)
	if err != nil {
		t.Fatal(err)
	}
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{X: 3, Y: 4, Texture: "wall", Blocks: true})
	evs := e.ProcessTick([]models.Action{
		{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload},
	})
	if len(evs) != 1 || evs[0].Type != "state" {
		t.Fatalf("events: %+v", evs)
	}
	body, err := json.Marshal(evs[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	var st struct {
		Tiles []gamekit.Tile `json:"tiles"`
	}
	if err := json.Unmarshal(body, &st); err != nil {
		t.Fatal(err)
	}
	if len(st.Tiles) != 1 || st.Tiles[0].X != 3 || st.Tiles[0].Y != 4 || st.Tiles[0].Layer != 0 || st.Tiles[0].Rotation != 0 ||
		st.Tiles[0].Texture != "wall" || !st.Tiles[0].Blocks {
		t.Fatalf("tiles: %+v", st.Tiles)
	}
}

func TestEngine_spawnTileLayersAndClearTile(t *testing.T) {
	e, err := NewEngine(nil)
	if err != nil {
		t.Fatal(err)
	}
	grass, _ := json.Marshal(gamekit.TileSpawnIntent{X: 1, Y: 1, Layer: 0, Texture: "grass", Blocks: false})
	flower, _ := json.Marshal(gamekit.TileSpawnIntent{X: 1, Y: 1, Layer: 1, Texture: "flower", Blocks: false})
	e.ProcessTick([]models.Action{
		{PlayerID: 1, Type: "spawn_tile", Payload: grass},
		{PlayerID: 1, Type: "spawn_tile", Payload: flower},
	})
	clearPayload, _ := json.Marshal(gamekit.TileClearIntent{X: 1, Y: 1, Layer: 1})
	evs := e.ProcessTick([]models.Action{{PlayerID: 1, Type: "clear_tile", Payload: clearPayload}})
	if len(evs) != 1 || evs[0].Type != "state" {
		t.Fatalf("events: %+v", evs)
	}
	body, err := json.Marshal(evs[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	var st struct {
		Tiles []gamekit.Tile `json:"tiles"`
	}
	if err := json.Unmarshal(body, &st); err != nil {
		t.Fatal(err)
	}
	if len(st.Tiles) != 1 || st.Tiles[0].Layer != 0 || st.Tiles[0].Texture != "grass" {
		t.Fatalf("after clear layer 1: %+v", st.Tiles)
	}
}

func TestEngine_spawnTileRotationNormalized(t *testing.T) {
	e, err := NewEngine(nil)
	if err != nil {
		t.Fatal(err)
	}
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{X: 0, Y: 0, Layer: 0, Rotation: 5, Texture: "arrow", Blocks: false})
	evs := e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload}})
	body, _ := json.Marshal(evs[0].Payload)
	var st struct {
		Tiles []gamekit.Tile `json:"tiles"`
	}
	_ = json.Unmarshal(body, &st)
	if len(st.Tiles) != 1 || st.Tiles[0].Rotation != 1 {
		t.Fatalf("rotation want 1, got %+v", st.Tiles)
	}
}

func TestEngine_moveBlockedBySolidTile(t *testing.T) {
	e, err := NewEngine(nil)
	if err != nil {
		t.Fatal(err)
	}
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{X: 1, Y: 0, Texture: "wall", Blocks: true})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload}})

	movePayload, _ := json.Marshal(gamekit.MoveIntent{DX: 1, DY: 0})
	hitPayload, _ := json.Marshal(gamekit.HitIntent{TargetID: 2, Damage: 3})
	e.ProcessTick([]models.Action{
		{PlayerID: 1, Type: "move", Payload: movePayload},
		{PlayerID: 1, Type: "hit", Payload: hitPayload},
	})

	e.mu.Lock()
	defer e.mu.Unlock()
	ent1 := e.EnsurePlayerEntity(1)
	_, pos, _, _ := e.playerMapper.Get(ent1)
	if pos.X != 0 || pos.Y != 0 {
		t.Fatalf("player 1 blocked at wall: want (0,0), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestNewEngine_fromArkSerdeSnapshot(t *testing.T) {
	src, err := NewEngine(nil)
	if err != nil {
		t.Fatal(err)
	}
	ent := src.EnsurePlayerEntity(7)
	src.mu.Lock()
	_, pos, sp, hp := src.playerMapper.Get(ent)
	pos.X, pos.Y = 2, 3
	hp.HP = 5
	sp.MaxStep = 2
	src.mu.Unlock()

	snap, err := arkserde.Serialize(src.world)
	if err != nil {
		t.Fatal(err)
	}

	e, err := NewEngine(snap)
	if err != nil {
		t.Fatal(err)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	ent2 := e.EnsurePlayerEntity(7)
	_, pos2, sp2, hp2 := e.playerMapper.Get(ent2)
	if pos2.X != 2 || pos2.Y != 3 || hp2.HP != 5 || sp2.MaxStep != 2 {
		t.Fatalf("loaded state: pos=(%d,%d) hp=%d max_step=%d", pos2.X, pos2.Y, hp2.HP, sp2.MaxStep)
	}
}
