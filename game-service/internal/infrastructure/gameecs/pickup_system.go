package gameecs

import (
	"encoding/json"
	"strings"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit/content"
)

// PickupSystem — type=pickup_item: pickable предмет с пола (тайл texture == item_def_id) в рюкзак, если есть место и игрок рядом.
type PickupSystem struct {
	bundle *content.Bundle
	engine *Engine
}

func NewPickupSystem(bundle *content.Bundle, engine *Engine) *PickupSystem {
	return &PickupSystem{bundle: bundle, engine: engine}
}

func (s *PickupSystem) Update(ctx *TickContext) {
	if s.bundle == nil || s.bundle.Catalog == nil {
		return
	}
	a := ctx.CurrentAction
	if a.Type != gamekit.TypePickupItem {
		return
	}
	var in gamekit.PickupIntent
	if err := json.Unmarshal(a.Payload, &in); err != nil {
		return
	}
	id := strings.TrimSpace(in.ItemDefID)
	if id == "" {
		return
	}
	it, ok := s.bundle.Catalog.Items[id]
	if !ok || !it.Pickable {
		return
	}
	clickIn := gamekit.InteractIntent{
		ItemDefID:  in.ItemDefID,
		ClickX:     in.ClickX,
		ClickY:     in.ClickY,
		ClickLayer: in.ClickLayer,
	}
	tx, ty, layer, _, ok := s.engine.resolveCatalogTileAtClick(clickIn)
	if !ok {
		return
	}
	ent, ok := s.engine.byUser[a.PlayerID]
	if !ok || !s.engine.playerMapper.HasAll(ent) || !s.engine.playerGearMapper.HasAll(ent) {
		return
	}
	_, pos, _, _, _, _, _ := s.engine.playerMapper.Get(ent)
	if !gamekit.ChebyshevDist1(pos.X, pos.Y, tx, ty) {
		return
	}
	invPtr := s.engine.playerGearMapper.Get(ent)
	if invPtr == nil {
		return
	}
	slot := gamekit.FirstEmptyBackpackSlot(invPtr)
	if slot < 0 {
		return
	}
	s.engine.removeTilesAtLayerRecorded(tx, ty, layer)
	invPtr.Backpack[slot] = id
	gamekit.NormalizeInventory(invPtr)
}
