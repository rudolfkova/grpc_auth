package gameecs

import (
	"time"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// JoinStateSnapshot строит полный state с tiles для нового WS-подключения.
// Не сдвигает lastFullTileSyncAt (глобальный full-sync по таймеру не затрагивается).
func (e *Engine) JoinStateSnapshot() gamekit.StatePayload {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	players := collectPlayersFromWorld(e.playerFilter, func(ent ecs.Entity) gamekit.PlayerInventory {
		return e.playerInventorySnapshotLocked(ent)
	})
	tiles := collectTilesFromWorld(e.tileFilter)

	pl := gamekit.StatePayload{
		Players: players,
		TickAt:  now,
	}
	snap := append([]gamekit.Tile(nil), tiles...)
	for i := range snap {
		if len(snap[i].InstanceArgs) > 0 {
			b := make([]byte, len(snap[i].InstanceArgs))
			copy(b, snap[i].InstanceArgs)
			snap[i].InstanceArgs = b
		}
	}
	pl.Tiles = &snap
	return pl
}
