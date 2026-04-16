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
	engine *Engine
}

// NewTileClearSystem ...
func NewTileClearSystem(
	w *ecs.World,
	filter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	engine *Engine,
) *TileClearSystem {
	return &TileClearSystem{world: w, filter: filter, engine: engine}
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
	type rmEnt struct {
		ent ecs.Entity
		x   int
		y   int
		z   int
	}
	var rm []rmEnt
	for q.Next() {
		pos, lay, _, _, _ := q.Get()
		if pos.X == in.X && pos.Y == in.Y && lay.Z == in.Layer {
			rm = append(rm, rmEnt{ent: q.Entity(), x: pos.X, y: pos.Y, z: lay.Z})
		}
	}
	q.Close()
	for _, r := range rm {
		if s.engine != nil {
			s.engine.recordTileRemove(r.x, r.y, r.z)
		}
		s.world.RemoveEntity(r.ent)
	}
}
