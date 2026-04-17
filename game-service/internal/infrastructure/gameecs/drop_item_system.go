package gameecs

import (
	"encoding/json"
	"strings"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// DropItemSystem — type=drop_item: выложить предмет из слота тайлом на клетку игрока (слой DroppedItemTileLayer).
type DropItemSystem struct {
	engine *Engine
}

func NewDropItemSystem(e *Engine) *DropItemSystem {
	return &DropItemSystem{engine: e}
}

func (s *DropItemSystem) Update(ctx *TickContext) {
	a := ctx.CurrentAction
	if a.Type != gamekit.TypeDropItem {
		return
	}
	var in gamekit.DropItemIntent
	if err := json.Unmarshal(a.Payload, &in); err != nil {
		return
	}
	from := strings.TrimSpace(in.From)
	if from == "" {
		return
	}
	ent, ok := s.engine.byUser[a.PlayerID]
	if !ok || !s.engine.playerMapper.HasAll(ent) || !s.engine.playerGearMapper.HasAll(ent) {
		return
	}
	invPtr := s.engine.playerGearMapper.Get(ent)
	if invPtr == nil {
		return
	}
	itemID, slotOK := invPtr.ItemAtSlot(from)
	if !slotOK || itemID == "" {
		return
	}
	_, pos, _, _, _, _, _ := s.engine.playerMapper.Get(ent)
	if s.cellHasTileOnDroppedLayer(pos.X, pos.Y) {
		return
	}
	spawn := gamekit.TileSpawnIntent{
		X: pos.X, Y: pos.Y, Layer: gamekit.DroppedItemTileLayer,
		Texture: itemID, Blocks: false,
	}
	spawnTileAt(s.engine.world, s.engine.tileMapper, s.engine.tileFilter, spawn, s.engine)
	invPtr.PutItemAtSlot(from, "")
	gamekit.NormalizeInventory(invPtr)
}

func (s *DropItemSystem) cellHasTileOnDroppedLayer(x, y int) bool {
	q := s.engine.tileFilter.Query()
	defer q.Close()
	for q.Next() {
		p, lay, _, _, _ := q.Get()
		if p.X == x && p.Y == y && lay.Z == gamekit.DroppedItemTileLayer {
			return true
		}
	}
	return false
}
