package gameecs

import (
	"strings"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit/content"

	"github.com/mlange-42/ark/ecs"
)

func (e *Engine) applyInventoryMoveLocked(userID int64, from, to string) {
	ent, ok := e.byUser[userID]
	if !ok || !e.playerMapper.HasAll(ent) || !e.playerGearMapper.HasAll(ent) {
		return
	}
	ptr := e.playerGearMapper.Get(ent)
	if ptr == nil {
		return
	}
	canPlace := func(slot, itemDefID string) bool {
		itemDefID = strings.TrimSpace(itemDefID)
		if itemDefID == "" {
			return true
		}
		if e.inventoryCatalog == nil {
			return true
		}
		return content.ItemFitsInventorySlot(e.inventoryCatalog, itemDefID, slot)
	}
	gamekit.TrySwapInventorySlots(ptr, from, to, canPlace)
}

func (e *Engine) playerInventorySnapshotLocked(ent ecs.Entity) gamekit.PlayerInventory {
	if !e.playerGearMapper.HasAll(ent) {
		return gamekit.DefaultPlayerInventory()
	}
	ptr := e.playerGearMapper.Get(ent)
	if ptr == nil {
		return gamekit.DefaultPlayerInventory()
	}
	inv := *ptr
	gamekit.NormalizeInventory(&inv)
	return inv
}
