package gameecs

import (
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// spawnTileAt заменяет тайлы в клетке на слое и создаёт одну сущность тайла (как TileSpawnSystem).
// rec при не-nil получает upsert в дельту state (тот же снимок, что уходит в wire).
func spawnTileAt(
	w *ecs.World,
	tiles *ecs.Map5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	filter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	in gamekit.TileSpawnIntent,
	rec *Engine,
) {
	rot := gamekit.NormalizeTileRotationQuarter(in.Rotation)
	inst := gamekit.NormalizeTileInstanceArgsJSON(in.InstanceArgs)
	removeTilesAtLayer(w, filter, in.X, in.Y, in.Layer)
	tiles.NewEntity(
		&gamekit.GridPos{X: in.X, Y: in.Y},
		&gamekit.TileLayer{Z: in.Layer},
		&gamekit.TileFacing{RotationQuarter: rot},
		&gamekit.TileTexture{Name: in.Texture, InstanceArgs: inst},
		&gamekit.TileSolid{Blocks: in.Blocks},
	)
	if rec != nil {
		in.Rotation = rot
		in.InstanceArgs = inst
		rec.recordTileUpsertFromIntent(in)
	}
}

func removeTilesAtLayer(w *ecs.World, filter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid], x, y, layer int) {
	q := filter.Query()
	var rm []ecs.Entity
	for q.Next() {
		pos, lay, _, _, _ := q.Get()
		if pos.X == x && pos.Y == y && lay.Z == layer {
			rm = append(rm, q.Entity())
		}
	}
	q.Close()
	for _, e := range rm {
		w.RemoveEntity(e)
	}
}
