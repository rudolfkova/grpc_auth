package gameecs

import (
	"encoding/json"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// TileSpawnSystem обрабатывает type=spawn_tile: в клетке (x,y) на слое layer заменяет тайл и создаёт сущность.
type TileSpawnSystem struct {
	world  *ecs.World
	tiles  *ecs.Map5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid]
	filter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid]
	engine *Engine
}

// NewTileSpawnSystem ...
func NewTileSpawnSystem(
	w *ecs.World,
	tiles *ecs.Map5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	filter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	engine *Engine,
) *TileSpawnSystem {
	return &TileSpawnSystem{world: w, tiles: tiles, filter: filter, engine: engine}
}

func (s *TileSpawnSystem) Update(ctx *TickContext) {
	a := ctx.CurrentAction
	if a.Type != "spawn_tile" {
		return
	}

	var in gamekit.TileSpawnIntent
	if err := json.Unmarshal(a.Payload, &in); err != nil {
		return
	}
	in.InstanceArgs = gamekit.NormalizeTileInstanceArgsJSON(in.InstanceArgs)

	spawnTileAt(s.world, s.tiles, s.filter, in, s.engine)
}
