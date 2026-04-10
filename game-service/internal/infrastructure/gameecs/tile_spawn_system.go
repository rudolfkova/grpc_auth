package gameecs

import (
	"encoding/json"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// TileSpawnSystem обрабатывает type=spawn_tile: заменяет тайл в клетке (x,y) и создаёт сущность.
type TileSpawnSystem struct {
	world  *ecs.World
	tiles  *ecs.Map3[gamekit.GridPos, gamekit.TileTexture, gamekit.TileSolid]
	filter *ecs.Filter3[gamekit.GridPos, gamekit.TileTexture, gamekit.TileSolid]
}

// NewTileSpawnSystem ...
func NewTileSpawnSystem(
	w *ecs.World,
	tiles *ecs.Map3[gamekit.GridPos, gamekit.TileTexture, gamekit.TileSolid],
	filter *ecs.Filter3[gamekit.GridPos, gamekit.TileTexture, gamekit.TileSolid],
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

	s.removeTilesAt(in.X, in.Y)
	s.tiles.NewEntity(
		&gamekit.GridPos{X: in.X, Y: in.Y},
		&gamekit.TileTexture{Name: in.Texture},
		&gamekit.TileSolid{Blocks: in.Blocks},
	)
}

func (s *TileSpawnSystem) removeTilesAt(x, y int) {
	q := s.filter.Query()
	defer q.Close()

	var rm []ecs.Entity
	for q.Next() {
		pos, _, _ := q.Get()
		if pos.X == x && pos.Y == y {
			rm = append(rm, q.Entity())
		}
	}
	for _, e := range rm {
		s.world.RemoveEntity(e)
	}
}
