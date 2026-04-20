package gameecs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit/content"
)

// GiveItemToPlayerBackpack adds itemDefID to the first free backpack slot for userID.
// Fails with no inventory mutation if the backpack is full, the item cannot go in the backpack per catalog, or the player is missing.
func (e *Engine) GiveItemToPlayerBackpack(userID int64, itemDefID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.tryGiveItemToPlayerBackpackUnlocked(userID, itemDefID)
}

// tryGiveItemToPlayerBackpackUnlocked is like GiveItemToPlayerBackpack but caller must hold e.mu (e.g. ProcessTick / content ops).
func (e *Engine) tryGiveItemToPlayerBackpackUnlocked(userID int64, itemDefID string) error {
	id := strings.TrimSpace(itemDefID)
	if id == "" {
		return fmt.Errorf("empty item_def_id")
	}
	ent, ok := e.byUser[userID]
	if !ok || !e.playerGearMapper.HasAll(ent) {
		return fmt.Errorf("player not found")
	}
	invPtr := e.playerGearMapper.Get(ent)
	if invPtr == nil {
		return fmt.Errorf("player not found")
	}
	return content.TryPutItemInFirstBackpack(invPtr, e.inventoryCatalog, id)
}

// SpawnPickableItemAt spawns a pickable-style tile at (x,y,layer): texture and instance_args.item_def_id set to itemDefID.
// Fails if a tile already exists at that cell and layer (nothing is spawned).
func (e *Engine) SpawnPickableItemAt(x, y, layer int, itemDefID string) error {
	id := strings.TrimSpace(itemDefID)
	if id == "" {
		return fmt.Errorf("empty item_def_id")
	}
	if layer < 0 {
		return fmt.Errorf("invalid layer %d", layer)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.tileExistsAtLayer(x, y, layer) {
		return fmt.Errorf("tile already at (%d,%d) layer %d", x, y, layer)
	}
	raw, err := json.Marshal(map[string]any{"item_def_id": id})
	if err != nil {
		return err
	}
	in := gamekit.TileSpawnIntent{
		X:            x,
		Y:            y,
		Layer:        layer,
		Texture:      id,
		Blocks:       false,
		InstanceArgs: gamekit.NormalizeTileInstanceArgsJSON(raw),
	}
	spawnTileAt(e.world, e.tileMapper, e.tileFilter, in, e)
	return nil
}
