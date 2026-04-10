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
}

// NewTileSpawnSystem ...
func NewTileSpawnSystem(
	w *ecs.World,
	tiles *ecs.Map5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	filter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
) *TileSpawnSystem {
	return &TileSpawnSystem{world: w, tiles: tiles, filter: filter}
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

	rot := gamekit.NormalizeTileRotationQuarter(in.Rotation)
	s.removeTilesAtLayer(in.X, in.Y, in.Layer)
	s.tiles.NewEntity(
		&gamekit.GridPos{X: in.X, Y: in.Y},
		&gamekit.TileLayer{Z: in.Layer},
		&gamekit.TileFacing{RotationQuarter: rot},
		&gamekit.TileTexture{Name: in.Texture},
		&gamekit.TileSolid{Blocks: in.Blocks},
	)
}

func (s *TileSpawnSystem) removeTilesAtLayer(x, y, layer int) {
	q := s.filter.Query()
	defer q.Close()

	var rm []ecs.Entity
	for q.Next() {
		pos, lay, _, _, _ := q.Get()
		if pos.X == x && pos.Y == y && lay.Z == layer {
			rm = append(rm, q.Entity())
		}
	}
	for _, e := range rm {
		s.world.RemoveEntity(e)
	}
}
