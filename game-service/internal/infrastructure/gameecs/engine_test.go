package gameecs

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"game/internal/domain/models"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit/content"

	arkserde "github.com/mlange-42/ark-serde"
)

func TestEngine_MoveAndHit(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
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
	_, _, _, hp, _, _, _ := e.playerMapper.Get(ent2)
	if hp.HP != gamekit.DefaultPlayerHP-3 {
		t.Fatalf("target HP: want %d, got %d", gamekit.DefaultPlayerHP-3, hp.HP)
	}
	ent1 := e.EnsurePlayerEntity(1)
	_, pos, _, _, _, _, _ := e.playerMapper.Get(ent1)
	if pos.X != 1 || pos.Y != 0 {
		t.Fatalf("mover position: want (1,0), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestEngine_spawnTileAppearsInState(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
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
	var st gamekit.StatePayload
	if err := json.Unmarshal(body, &st); err != nil {
		t.Fatal(err)
	}
	if st.Tiles == nil || len(*st.Tiles) != 1 {
		t.Fatalf("want full tiles snapshot, got %+v", st)
	}
	t0 := (*st.Tiles)[0]
	if t0.X != 3 || t0.Y != 4 || t0.Layer != 0 || t0.Rotation != 0 ||
		t0.Texture != "wall" || !t0.Blocks {
		t.Fatalf("tiles: %+v", t0)
	}
}

func TestEngine_spawnTileInstanceArgsInState(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	inst, _ := json.Marshal(map[string]any{"door_x": 9, "door_y": 2})
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{
		X: 5, Y: 6, Layer: 0, Texture: "lever_x", Blocks: false,
		InstanceArgs: inst,
	})
	evs := e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload}})
	if len(evs) != 1 || evs[0].Type != "state" {
		t.Fatalf("events: %+v", evs)
	}
	body, _ := json.Marshal(evs[0].Payload)
	var st gamekit.StatePayload
	if err := json.Unmarshal(body, &st); err != nil {
		t.Fatal(err)
	}
	if st.Tiles == nil {
		t.Fatalf("want full tiles, got %+v", st)
	}
	var got *gamekit.Tile
	for i := range *st.Tiles {
		ti := &(*st.Tiles)[i]
		if ti.X == 5 && ti.Y == 6 && ti.Texture == "lever_x" {
			got = ti
			break
		}
	}
	if got == nil {
		t.Fatalf("no tile: %+v", st.Tiles)
	}
	var m map[string]any
	if err := json.Unmarshal(got.InstanceArgs, &m); err != nil {
		t.Fatal(err)
	}
	if int(m["door_x"].(float64)) != 9 || int(m["door_y"].(float64)) != 2 {
		t.Fatalf("instance_args: %#v", m)
	}
}

func TestEngine_spawnTileLayersAndClearTile(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
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
	var st gamekit.StatePayload
	if err := json.Unmarshal(body, &st); err != nil {
		t.Fatal(err)
	}
	if st.Tiles != nil {
		t.Fatalf("delta tick should omit tiles, got %+v", st.Tiles)
	}
	if len(st.TileUpdates) != 1 || st.TileUpdates[0].Op != gamekit.StateTileUpdateRemove ||
		st.TileUpdates[0].X != 1 || st.TileUpdates[0].Y != 1 || st.TileUpdates[0].Layer != 1 {
		t.Fatalf("tile_updates: %+v", st.TileUpdates)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	var seen []string
	q := e.tileFilter.Query()
	defer q.Close()
	for q.Next() {
		pos, lay, _, tex, _ := q.Get()
		seen = append(seen, tex.Name)
		if pos.X == 1 && pos.Y == 1 && lay.Z == 0 && tex.Name != "grass" {
			t.Fatalf("layer 0: want grass at (1,1), got %s", tex.Name)
		}
		if pos.X == 1 && pos.Y == 1 && lay.Z == 1 {
			t.Fatalf("layer 1 should be empty, still have tile %s", tex.Name)
		}
	}
	if len(seen) != 1 {
		t.Fatalf("expected one tile in world, got %v", seen)
	}
}

func TestEngine_spawnTileRotationNormalized(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{X: 0, Y: 0, Layer: 0, Rotation: 5, Texture: "arrow", Blocks: false})
	evs := e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload}})
	body, _ := json.Marshal(evs[0].Payload)
	var st gamekit.StatePayload
	_ = json.Unmarshal(body, &st)
	if st.Tiles == nil || len(*st.Tiles) != 1 || (*st.Tiles)[0].Rotation != 1 {
		t.Fatalf("rotation want 1, got %+v", st.Tiles)
	}
}

func TestEngine_manyMoveMessagesOneCellPerTick(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	mv, _ := json.Marshal(gamekit.MoveIntent{DX: 1, DY: 0})
	var batch []models.Action
	for range 20 {
		batch = append(batch, models.Action{PlayerID: 1, Type: "move", Payload: mv})
	}
	e.ProcessTick(batch)
	e.mu.Lock()
	defer e.mu.Unlock()
	ent1 := e.EnsurePlayerEntity(1)
	_, pos, _, _, _, _, _ := e.playerMapper.Get(ent1)
	if pos.X != 1 || pos.Y != 0 {
		t.Fatalf("many moves in one tick: want (1,0), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestEngine_moveIntentPersistsAcrossTicks(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	mv, _ := json.Marshal(gamekit.MoveIntent{DX: 1, DY: 0})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: "move", Payload: mv}})
	e.ProcessTick(nil)
	e.mu.Lock()
	defer e.mu.Unlock()
	ent1 := e.EnsurePlayerEntity(1)
	_, pos, _, _, _, _, _ := e.playerMapper.Get(ent1)
	if pos.X != 2 || pos.Y != 0 {
		t.Fatalf("intent without new messages: want (2,0), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestEngine_moveBlockedBySolidTile(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
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
	_, pos, _, _, _, _, _ := e.playerMapper.Get(ent1)
	if pos.X != 0 || pos.Y != 0 {
		t.Fatalf("player 1 blocked at wall: want (0,0), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestNewEngine_fromArkSerdeSnapshot(t *testing.T) {
	src, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	ent := src.EnsurePlayerEntity(7)
	src.mu.Lock()
	_, pos, sp, hp, _, _, _ := src.playerMapper.Get(ent)
	pos.X, pos.Y = 2, 3
	hp.HP = 5
	sp.MaxStep = 2
	src.mu.Unlock()

	snap, err := arkserde.Serialize(src.world)
	if err != nil {
		t.Fatal(err)
	}

	e, err := NewEngine(snap, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	ent2 := e.EnsurePlayerEntity(7)
	_, pos2, sp2, hp2, _, _, _ := e.playerMapper.Get(ent2)
	if pos2.X != 2 || pos2.Y != 3 || hp2.HP != 5 || sp2.MaxStep != 2 {
		t.Fatalf("loaded state: pos=(%d,%d) hp=%d max_step=%d", pos2.X, pos2.Y, hp2.HP, sp2.MaxStep)
	}
}

func TestEngine_inventoryMoveSwapHands(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	d := gamekit.NewDefaultCharacterPlayData()
	d.Inventory.HandMain = "sword"
	d.Inventory.HandOff = "axe"
	e.EnsurePlayerJoin(1, d)
	pl, _ := json.Marshal(gamekit.InventoryMoveIntent{From: gamekit.InvSlotHandMain, To: gamekit.InvSlotHandOff})
	evs := e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeInventoryMove, Payload: pl}})
	if len(evs) != 1 || evs[0].Type != "state" {
		t.Fatalf("events: %+v", evs)
	}
	body, _ := json.Marshal(evs[0].Payload)
	var st gamekit.StatePayload
	if err := json.Unmarshal(body, &st); err != nil {
		t.Fatal(err)
	}
	if len(st.Players) != 1 {
		t.Fatalf("players: %+v", st.Players)
	}
	inv := st.Players[0].Inventory
	if inv.HandMain != "axe" || inv.HandOff != "sword" {
		t.Fatalf("after swap want axe/sword, got main=%q off=%q", inv.HandMain, inv.HandOff)
	}
}

func TestEngine_stateIncludesFacing(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	evs := e.ProcessTick(nil)
	if len(evs) != 1 {
		t.Fatalf("events: %+v", evs)
	}
	body, _ := json.Marshal(evs[0].Payload)
	var st gamekit.StatePayload
	_ = json.Unmarshal(body, &st)
	if len(st.Players) != 0 {
		t.Fatalf("expected no players before join")
	}
	mv, _ := json.Marshal(gamekit.MoveIntent{DX: -1, DY: 0})
	evs = e.ProcessTick([]models.Action{{PlayerID: 9, Type: "move", Payload: mv}})
	body, _ = json.Marshal(evs[0].Payload)
	_ = json.Unmarshal(body, &st)
	if len(st.Players) != 1 || st.Players[0].FaceDX != -1 || st.Players[0].FaceDY != 0 {
		t.Fatalf("after move west: %+v", st.Players)
	}
}

func TestEngine_diagonalIntentStaircase(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	mv, _ := json.Marshal(gamekit.MoveIntent{DX: 1, DY: 1})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: "move", Payload: mv}})
	e.mu.Lock()
	_, pos, _, _, _, _, _ := e.playerMapper.Get(e.EnsurePlayerEntity(1))
	e.mu.Unlock()
	if pos.X != 1 || pos.Y != 0 {
		t.Fatalf("diag step 1: want (1,0), got (%d,%d)", pos.X, pos.Y)
	}
	e.ProcessTick(nil)
	e.mu.Lock()
	_, pos, _, _, _, _, _ = e.playerMapper.Get(e.EnsurePlayerEntity(1))
	e.mu.Unlock()
	if pos.X != 1 || pos.Y != 1 {
		t.Fatalf("diag step 2: want (1,1), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestEngine_moveApplyEveryNTicks(t *testing.T) {
	e, err := NewEngine(nil, 2, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	mv, _ := json.Marshal(gamekit.MoveIntent{DX: 1, DY: 0})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: "move", Payload: mv}})
	e.mu.Lock()
	_, pos, _, _, _, _, _ := e.playerMapper.Get(e.EnsurePlayerEntity(1))
	e.mu.Unlock()
	if pos.X != 1 || pos.Y != 0 {
		t.Fatalf("tick 1 with N=2: want (1,0), got (%d,%d)", pos.X, pos.Y)
	}
	e.ProcessTick(nil)
	e.mu.Lock()
	_, pos, _, _, _, _, _ = e.playerMapper.Get(e.EnsurePlayerEntity(1))
	e.mu.Unlock()
	if pos.X != 1 || pos.Y != 0 {
		t.Fatalf("tick 2 skip move: want (1,0), got (%d,%d)", pos.X, pos.Y)
	}
	e.ProcessTick(nil)
	e.mu.Lock()
	_, pos, _, _, _, _, _ = e.playerMapper.Get(e.EnsurePlayerEntity(1))
	e.mu.Unlock()
	if pos.X != 2 || pos.Y != 0 {
		t.Fatalf("tick 3 apply move: want (2,0), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestEngine_pickupFloorGemIntoBackpack(t *testing.T) {
	base := filepath.Join("testdata", "content")
	b, err := content.LoadBundle(filepath.Join(base, "catalog.json"), filepath.Join(base, "scripts"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(nil, 1, EngineOptions{Content: b})
	if err != nil {
		t.Fatal(err)
	}
	e.EnsurePlayerJoin(1, gamekit.NewDefaultCharacterPlayData())
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{X: 1, Y: 0, Layer: 0, Texture: "floor_gem", Blocks: false})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload}})
	cx, cy := 1, 0
	pickPayload, _ := json.Marshal(gamekit.PickupIntent{ItemDefID: "floor_gem", ClickX: &cx, ClickY: &cy})
	evs := e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypePickupItem, Payload: pickPayload}})
	if len(evs) != 1 || evs[0].Type != "state" {
		t.Fatalf("events: %+v", evs)
	}
	body, _ := json.Marshal(evs[0].Payload)
	var st gamekit.StatePayload
	if err := json.Unmarshal(body, &st); err != nil {
		t.Fatal(err)
	}
	if len(st.Players) != 1 || st.Players[0].Inventory.Backpack[0] != "floor_gem" {
		t.Fatalf("backpack: %+v", st.Players[0].Inventory)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	q := e.tileFilter.Query()
	defer q.Close()
	for q.Next() {
		pos, lay, _, tex, _ := q.Get()
		if pos.X == 1 && pos.Y == 0 && lay.Z == 0 && tex.Name == "floor_gem" {
			t.Fatal("tile should be removed after pickup")
		}
	}
}

func TestEngine_dropItemPutsTileOnLayer3(t *testing.T) {
	base := filepath.Join("testdata", "content")
	b, err := content.LoadBundle(filepath.Join(base, "catalog.json"), filepath.Join(base, "scripts"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(nil, 1, EngineOptions{Content: b})
	if err != nil {
		t.Fatal(err)
	}
	e.EnsurePlayerJoin(1, gamekit.NewDefaultCharacterPlayData())
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{X: 1, Y: 0, Layer: 0, Texture: "floor_gem", Blocks: false})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload}})
	cx, cy := 1, 0
	pickPayload, _ := json.Marshal(gamekit.PickupIntent{ItemDefID: "floor_gem", ClickX: &cx, ClickY: &cy})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypePickupItem, Payload: pickPayload}})

	dropPayload, _ := json.Marshal(gamekit.DropItemIntent{From: "backpack_0"})
	evs := e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeDropItem, Payload: dropPayload}})
	if len(evs) != 1 || evs[0].Type != "state" {
		t.Fatalf("events: %+v", evs)
	}
	body, _ := json.Marshal(evs[0].Payload)
	var st gamekit.StatePayload
	if err := json.Unmarshal(body, &st); err != nil {
		t.Fatal(err)
	}
	if len(st.Players) != 1 || st.Players[0].Inventory.Backpack[0] != "" {
		t.Fatalf("backpack should be empty: %+v", st.Players[0].Inventory)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	_, pos, _, _, _, _, _ := e.playerMapper.Get(e.EnsurePlayerEntity(1))
	found := false
	q := e.tileFilter.Query()
	defer q.Close()
	for q.Next() {
		p, lay, _, tex, _ := q.Get()
		if p.X == pos.X && p.Y == pos.Y && lay.Z == gamekit.DroppedItemTileLayer && tex.Name == "floor_gem" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected floor_gem tile at player (%d,%d) layer %d", pos.X, pos.Y, gamekit.DroppedItemTileLayer)
	}
}

func TestEngine_dropItemIgnoredWhenLayer3Occupied(t *testing.T) {
	base := filepath.Join("testdata", "content")
	b, err := content.LoadBundle(filepath.Join(base, "catalog.json"), filepath.Join(base, "scripts"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(nil, 1, EngineOptions{Content: b})
	if err != nil {
		t.Fatal(err)
	}
	e.EnsurePlayerJoin(1, gamekit.NewDefaultCharacterPlayData())
	spawn1, _ := json.Marshal(gamekit.TileSpawnIntent{X: 1, Y: 0, Layer: 0, Texture: "floor_gem", Blocks: false})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawn1}})
	cx, cy := 1, 0
	pick1, _ := json.Marshal(gamekit.PickupIntent{ItemDefID: "floor_gem", ClickX: &cx, ClickY: &cy})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypePickupItem, Payload: pick1}})
	drop1, _ := json.Marshal(gamekit.DropItemIntent{From: "backpack_0"})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeDropItem, Payload: drop1}})

	spawn2, _ := json.Marshal(gamekit.TileSpawnIntent{X: 1, Y: 0, Layer: 0, Texture: "floor_gem", Blocks: false})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawn2}})
	pick2, _ := json.Marshal(gamekit.PickupIntent{ItemDefID: "floor_gem", ClickX: &cx, ClickY: &cy})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypePickupItem, Payload: pick2}})
	e.mu.Lock()
	invPick := e.playerGearMapper.Get(e.byUser[1])
	e.mu.Unlock()
	if invPick == nil || invPick.Backpack[0] != "floor_gem" {
		t.Fatal("expected second gem in backpack_0")
	}

	drop2, _ := json.Marshal(gamekit.DropItemIntent{From: "backpack_0"})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeDropItem, Payload: drop2}})

	e.mu.Lock()
	inv := e.playerGearMapper.Get(e.byUser[1])
	e.mu.Unlock()
	if inv == nil || inv.Backpack[0] != "floor_gem" {
		t.Fatalf("drop on occupied layer 3 must keep item; inv=%+v", inv)
	}
}

func TestEngine_InteractWorldSpawnTile(t *testing.T) {
	base := filepath.Join("testdata", "content")
	b, err := content.LoadBundle(filepath.Join(base, "catalog.json"), filepath.Join(base, "scripts"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(nil, 1, EngineOptions{Content: b})
	if err != nil {
		t.Fatal(err)
	}
	pl, _ := json.Marshal(gamekit.InteractIntent{ItemDefID: "test_lever"})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeInteract, Payload: pl}})

	e.mu.Lock()
	defer e.mu.Unlock()
	q := e.tileFilter.Query()
	defer q.Close()
	found := false
	for q.Next() {
		pos, _, _, tex, _ := q.Get()
		if pos.X == 2 && pos.Y == 3 && tex.Name == "stone" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected stone tile at (2,3)")
	}
}

func TestEngine_InteractClickUsesTileInstanceArgs(t *testing.T) {
	base := filepath.Join("testdata", "content")
	b, err := content.LoadBundle(filepath.Join(base, "catalog.json"), filepath.Join(base, "scripts"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(nil, 1, EngineOptions{Content: b})
	if err != nil {
		t.Fatal(err)
	}
	inst, _ := json.Marshal(map[string]any{
		"item_def_id": "test_lever",
		"x":           9,
		"y":           8,
		"layer":       0,
		"rotation":    0,
		"texture":     "grass",
		"blocks":      false,
	})
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{
		X: 5, Y: 5, Layer: 0, Texture: "tent_1", Blocks: false,
		InstanceArgs: inst,
	})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload}})
	cx, cy := 5, 5
	pl, _ := json.Marshal(gamekit.InteractIntent{ItemDefID: "test_lever", ClickX: &cx, ClickY: &cy})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeInteract, Payload: pl}})

	e.mu.Lock()
	defer e.mu.Unlock()
	q := e.tileFilter.Query()
	defer q.Close()
	found := false
	for q.Next() {
		pos, _, _, tex, _ := q.Get()
		if pos.X == 9 && pos.Y == 8 && tex.Name == "grass" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected grass tile at (9,8) from merged instance_args")
	}
}

func TestEngine_pickupUsesInstanceArgsItemDefIDFallbackChain(t *testing.T) {
	base := filepath.Join("testdata", "content")
	b, err := content.LoadBundle(filepath.Join(base, "catalog.json"), filepath.Join(base, "scripts"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(nil, 1, EngineOptions{Content: b})
	if err != nil {
		t.Fatal(err)
	}
	e.EnsurePlayerJoin(1, gamekit.NewDefaultCharacterPlayData())
	inst, _ := json.Marshal(map[string]any{"item_def_id": "floor_gem"})
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{
		X:            1,
		Y:            0,
		Layer:        0,
		Texture:      "tent_1",
		Blocks:       false,
		InstanceArgs: inst,
	})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload}})

	cx, cy := 1, 0
	pickPayload, _ := json.Marshal(gamekit.PickupIntent{
		ItemDefID: "floor_gem",
		ClickX:    &cx,
		ClickY:    &cy,
	})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypePickupItem, Payload: pickPayload}})

	e.mu.Lock()
	defer e.mu.Unlock()
	ent, ok := e.byUser[1]
	if !ok || !e.playerGearMapper.HasAll(ent) {
		t.Fatal("player entity not found")
	}
	inv := e.playerGearMapper.Get(ent)
	if inv == nil || inv.Backpack[0] != "floor_gem" {
		t.Fatalf("pickup by instance_args item_def_id failed, inv=%+v", inv)
	}
}

func TestEngine_TentUnfoldAndFoldAnchorOnlySeparateIDs(t *testing.T) {
	base := filepath.Join("testdata", "content")
	b, err := content.LoadBundle(filepath.Join(base, "catalog.json"), filepath.Join(base, "scripts"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(nil, 1, EngineOptions{Content: b})
	if err != nil {
		t.Fatal(err)
	}
	foldedArgs, _ := json.Marshal(map[string]any{"item_def_id": "tent_folded"})
	spawnFoldedPayload, _ := json.Marshal(gamekit.TileSpawnIntent{
		X:            10,
		Y:            10,
		Layer:        gamekit.DroppedItemTileLayer,
		Texture:      "tent_1",
		Blocks:       false,
		InstanceArgs: foldedArgs,
	})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeSpawnTile, Payload: spawnFoldedPayload}})

	cx, cy, cl := 10, 10, gamekit.DroppedItemTileLayer
	unfoldPayload, _ := json.Marshal(gamekit.InteractIntent{
		ItemDefID:  "tent_folded",
		ClickX:     &cx,
		ClickY:     &cy,
		ClickLayer: &cl,
	})
	evs := e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeInteract, Payload: unfoldPayload}})
	if len(evs) != 1 || evs[0].Type != gamekit.TypeState {
		t.Fatalf("unexpected unfold events: %+v", evs)
	}

	e.mu.Lock()
	tiles := make(map[[3]int]gamekit.Tile)
	q := e.tileFilter.Query()
	for q.Next() {
		pos, lay, facing, tex, solid := q.Get()
		tiles[[3]int{pos.X, pos.Y, lay.Z}] = gamekit.Tile{
			X:            pos.X,
			Y:            pos.Y,
			Layer:        lay.Z,
			Rotation:     facing.RotationQuarter,
			Texture:      tex.Name,
			Blocks:       solid.Blocks,
			InstanceArgs: append(json.RawMessage(nil), tex.InstanceArgs...),
		}
	}
	q.Close()
	e.mu.Unlock()

	if len(tiles) != 4 {
		t.Fatalf("unfold expected 4 tent tiles, got %d", len(tiles))
	}
	anchor, ok := tiles[[3]int{10, 10, gamekit.DroppedItemTileLayer}]
	if !ok {
		t.Fatalf("missing anchor tile at (10,10,%d)", gamekit.DroppedItemTileLayer)
	}
	if anchor.Texture != "tent_3" {
		t.Fatalf("anchor texture mismatch: %+v", anchor)
	}
	if topLeft, ok := tiles[[3]int{10, 9, gamekit.DroppedItemTileLayer}]; !ok || topLeft.Texture != "tent_1" {
		t.Fatalf("top-left texture mismatch: %+v", topLeft)
	}
	if topRight, ok := tiles[[3]int{11, 9, gamekit.DroppedItemTileLayer}]; !ok || topRight.Texture != "tent_2" {
		t.Fatalf("top-right texture mismatch: %+v", topRight)
	}
	if bottomRight, ok := tiles[[3]int{11, 10, gamekit.DroppedItemTileLayer}]; !ok || bottomRight.Texture != "tent_4" {
		t.Fatalf("bottom-right texture mismatch: %+v", bottomRight)
	}
	var anchorInst struct {
		ItemDefID string `json:"item_def_id"`
	}
	if err := json.Unmarshal(anchor.InstanceArgs, &anchorInst); err != nil {
		t.Fatalf("anchor instance_args parse: %v", err)
	}
	if anchorInst.ItemDefID != "tent_anchor" {
		t.Fatalf("anchor item_def_id mismatch: %+v", anchorInst)
	}

	foldPayload, _ := json.Marshal(gamekit.InteractIntent{
		ItemDefID:  "tent_anchor",
		ClickX:     &cx,
		ClickY:     &cy,
		ClickLayer: &cl,
	})
	evs = e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeInteract, Payload: foldPayload}})
	if len(evs) != 1 || evs[0].Type != gamekit.TypeState {
		t.Fatalf("unexpected fold events: %+v", evs)
	}
	body, err := json.Marshal(evs[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	var st gamekit.StatePayload
	if err := json.Unmarshal(body, &st); err != nil {
		t.Fatal(err)
	}
	if len(st.TileUpdates) < 5 {
		t.Fatalf("fold must emit remove+upsert tile updates, got %+v", st.TileUpdates)
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	q = e.tileFilter.Query()
	defer q.Close()
	count := 0
	var folded gamekit.Tile
	for q.Next() {
		pos, lay, facing, tex, solid := q.Get()
		count++
		folded = gamekit.Tile{
			X:            pos.X,
			Y:            pos.Y,
			Layer:        lay.Z,
			Rotation:     facing.RotationQuarter,
			Texture:      tex.Name,
			Blocks:       solid.Blocks,
			InstanceArgs: append(json.RawMessage(nil), tex.InstanceArgs...),
		}
	}
	if count != 1 {
		t.Fatalf("after fold expected single folded tile, got %d", count)
	}
	if folded.X != 10 || folded.Y != 10 || folded.Layer != gamekit.DroppedItemTileLayer {
		t.Fatalf("folded tile position mismatch: %+v", folded)
	}
	var foldedInst struct {
		ItemDefID string `json:"item_def_id"`
	}
	if err := json.Unmarshal(folded.InstanceArgs, &foldedInst); err != nil {
		t.Fatalf("folded instance_args parse: %v", err)
	}
	if foldedInst.ItemDefID != "tent_folded" {
		t.Fatalf("folded item_def_id mismatch: %+v", foldedInst)
	}
}

func TestEngine_TentUnfoldIgnoresWhenFootprintOccupiedAtLayer3(t *testing.T) {
	base := filepath.Join("testdata", "content")
	b, err := content.LoadBundle(filepath.Join(base, "catalog.json"), filepath.Join(base, "scripts"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(nil, 1, EngineOptions{Content: b})
	if err != nil {
		t.Fatal(err)
	}
	foldedArgs, _ := json.Marshal(map[string]any{"item_def_id": "tent_folded"})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeSpawnTile, Payload: mustJSON(t, gamekit.TileSpawnIntent{
		X:            10,
		Y:            10,
		Layer:        gamekit.DroppedItemTileLayer,
		Texture:      "tent_1",
		Blocks:       false,
		InstanceArgs: foldedArgs,
	})}})
	// Occupy future top-right footprint cell (x+1, y-1) on layer 3.
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeSpawnTile, Payload: mustJSON(t, gamekit.TileSpawnIntent{
		X:       11,
		Y:       9,
		Layer:   gamekit.DroppedItemTileLayer,
		Texture: "floor_gem",
		Blocks:  false,
	})}})

	cx, cy, cl := 10, 10, gamekit.DroppedItemTileLayer
	unfoldPayload := mustJSON(t, gamekit.InteractIntent{
		ItemDefID:  "tent_folded",
		ClickX:     &cx,
		ClickY:     &cy,
		ClickLayer: &cl,
	})
	e.ProcessTick([]models.Action{{PlayerID: 1, Type: gamekit.TypeInteract, Payload: unfoldPayload}})

	e.mu.Lock()
	defer e.mu.Unlock()
	q := e.tileFilter.Query()
	defer q.Close()

	byPos := make(map[[3]int]string)
	for q.Next() {
		pos, lay, _, tex, _ := q.Get()
		byPos[[3]int{pos.X, pos.Y, lay.Z}] = tex.Name
	}
	if got := byPos[[3]int{10, 10, gamekit.DroppedItemTileLayer}]; got != "tent_1" {
		t.Fatalf("folded tile should remain at anchor, got %q", got)
	}
	if got := byPos[[3]int{11, 9, gamekit.DroppedItemTileLayer}]; got != "floor_gem" {
		t.Fatalf("occupied footprint tile should remain, got %q", got)
	}
	if _, ok := byPos[[3]int{11, 10, gamekit.DroppedItemTileLayer}]; ok {
		t.Fatalf("unexpected deployed tile at (11,10,%d)", gamekit.DroppedItemTileLayer)
	}
	if _, ok := byPos[[3]int{10, 9, gamekit.DroppedItemTileLayer}]; ok {
		t.Fatalf("unexpected deployed tile at (10,9,%d)", gamekit.DroppedItemTileLayer)
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return b
}

func TestEngine_quietSecondTickOmitsTileKeysInJSON(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	e.ProcessTick(nil)
	evs := e.ProcessTick(nil)
	if len(evs) != 1 {
		t.Fatalf("events: %+v", evs)
	}
	body, err := json.Marshal(evs[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["tiles"]; ok {
		t.Fatalf("expected no tiles key on quiet delta tick, got keys: %v", raw)
	}
	if _, ok := raw["tile_updates"]; ok {
		t.Fatalf("expected no tile_updates key, got keys: %v", raw)
	}
}

func TestEngine_spawnAfterFullSyncUsesTileUpdates(t *testing.T) {
	e, err := NewEngine(nil, 1, EngineOptions{})
	if err != nil {
		t.Fatal(err)
	}
	e.ProcessTick(nil)
	spawnPayload, _ := json.Marshal(gamekit.TileSpawnIntent{X: 0, Y: 0, Texture: "grass", Blocks: false})
	evs := e.ProcessTick([]models.Action{{PlayerID: 1, Type: "spawn_tile", Payload: spawnPayload}})
	if len(evs) != 1 {
		t.Fatalf("events: %+v", evs)
	}
	body, err := json.Marshal(evs[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	var st gamekit.StatePayload
	if err := json.Unmarshal(body, &st); err != nil {
		t.Fatal(err)
	}
	if st.Tiles != nil {
		t.Fatalf("expected delta tick without full tiles, got %+v", st)
	}
	if len(st.TileUpdates) != 1 || st.TileUpdates[0].Op != gamekit.StateTileUpdateUpsert ||
		st.TileUpdates[0].Tile == nil || st.TileUpdates[0].Tile.Texture != "grass" {
		t.Fatalf("tile_updates: %+v", st.TileUpdates)
	}
}
