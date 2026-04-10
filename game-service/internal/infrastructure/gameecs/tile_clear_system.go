package gameecs

import (
	"encoding/json"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// TileClearSystem обрабатывает type=clear_tile: удаляет все тайлы в (x,y) на заданном слое.
type TileClearSystem struct {
	world  *ecs.World
	filter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid]
}

// NewTileClearSystem ...
func NewTileClearSystem(
	w *ecs.World,
	filter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
) *TileClearSystem {
	return &TileClearSystem{world: w, filter: filter}
}

func (s *TileClearSystem) Update(ctx *TickContext) {
	a := ctx.CurrentAction
	if a.Type != "clear_tile" {
		return
	}

	var in gamekit.TileClearIntent
	if err := json.Unmarshal(a.Payload, &in); err != nil {
		return
	}

	q := s.filter.Query()
	defer q.Close()

	var rm []ecs.Entity
	for q.Next() {
		pos, lay, _, _, _ := q.Get()
		if pos.X == in.X && pos.Y == in.Y && lay.Z == in.Layer {
			rm = append(rm, q.Entity())
		}
	}
	for _, e := range rm {
		s.world.RemoveEntity(e)
	}
}
