package gameecs

import (
	"time"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

func tileWireFromSpawnIntent(in gamekit.TileSpawnIntent) gamekit.Tile {
	rot := gamekit.NormalizeTileRotationQuarter(in.Rotation)
	inst := gamekit.NormalizeTileInstanceArgsJSON(in.InstanceArgs)
	t := gamekit.Tile{
		X:        in.X,
		Y:        in.Y,
		Layer:    in.Layer,
		Rotation: rot,
		Texture:  in.Texture,
		Blocks:   in.Blocks,
	}
	if len(inst) > 0 {
		t.InstanceArgs = inst
	}
	return t
}

func (e *Engine) resetTileDeltaForTick() {
	e.pendingTileUpdates = e.pendingTileUpdates[:0]
}

func (e *Engine) recordTileUpsertFromIntent(in gamekit.TileSpawnIntent) {
	tw := tileWireFromSpawnIntent(in)
	if len(tw.InstanceArgs) > 0 {
		b := make([]byte, len(tw.InstanceArgs))
		copy(b, tw.InstanceArgs)
		tw.InstanceArgs = b
	}
	tp := new(gamekit.Tile)
	*tp = tw
	e.pendingTileUpdates = append(e.pendingTileUpdates, gamekit.TileUpdate{
		Op:   gamekit.StateTileUpdateUpsert,
		Tile: tp,
	})
}

func (e *Engine) recordTileRemove(x, y, layer int) {
	e.pendingTileUpdates = append(e.pendingTileUpdates, gamekit.TileUpdate{
		Op:    gamekit.StateTileUpdateRemove,
		X:     x,
		Y:     y,
		Layer: layer,
	})
}

// removeTilesAtLayerRecorded удаляет все тайлы в (x,y,layer) и пишет одну дельту remove (для подбора с пола).
func (e *Engine) removeTilesAtLayerRecorded(x, y, layer int) {
	q := e.tileFilter.Query()
	var rm []ecs.Entity
	for q.Next() {
		pos, lay, _, _, _ := q.Get()
		if pos.X == x && pos.Y == y && lay.Z == layer {
			rm = append(rm, q.Entity())
		}
	}
	q.Close()
	if len(rm) == 0 {
		return
	}
	e.recordTileRemove(x, y, layer)
	for _, ent := range rm {
		e.world.RemoveEntity(ent)
	}
}

func (e *Engine) statePayloadForTick(now time.Time, players []gamekit.Player, allTiles []gamekit.Tile) gamekit.StatePayload {
	pl := gamekit.StatePayload{
		Players: players,
		TickAt:  now,
	}
	full := e.lastFullTileSyncAt.IsZero() || now.Sub(e.lastFullTileSyncAt) >= e.tileFullSyncEvery
	if full {
		e.lastFullTileSyncAt = now
		snap := append([]gamekit.Tile(nil), allTiles...)
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
	if len(e.pendingTileUpdates) > 0 {
		pl.TileUpdates = append([]gamekit.TileUpdate(nil), e.pendingTileUpdates...)
	}
	return pl
}
