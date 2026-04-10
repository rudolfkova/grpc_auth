package gameecs

import (
	"encoding/json"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// MovementSystem обрабатывает действия type=move.
type MovementSystem struct {
	mapper     *ecs.Map4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health]
	tileFilter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid]
}

// NewMovementSystem создаёт систему движения с общим mapper мира.
func NewMovementSystem(
	mapper *ecs.Map4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health],
	tileFilter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
) *MovementSystem {
	return &MovementSystem{mapper: mapper, tileFilter: tileFilter}
}

func (s *MovementSystem) Update(ctx *TickContext) {
	a := ctx.CurrentAction
	if a.Type != "move" {
		return
	}

	var mv gamekit.MoveIntent
	if err := json.Unmarshal(a.Payload, &mv); err != nil {
		return
	}

	ent := ctx.Sink.EnsurePlayerEntity(a.PlayerID)
	s.applyStep(ent, mv.DX, mv.DY)
}

func (s *MovementSystem) applyStep(ent ecs.Entity, dx, dy int) {
	if !s.mapper.HasAll(ent) {
		return
	}
	_, pos, speed, _ := s.mapper.Get(ent)
	step := speed.MaxStep
	if step <= 0 {
		return
	}
	if dx < -step || dx > step || dy < -step || dy > step {
		return
	}
	tx, ty := pos.X+dx, pos.Y+dy
	if s.blockedCell(tx, ty) {
		return
	}
	pos.X = tx
	pos.Y = ty
}

func (s *MovementSystem) blockedCell(x, y int) bool {
	if s.tileFilter == nil {
		return false
	}
	q := s.tileFilter.Query()
	defer q.Close()
	for q.Next() {
		pos, _, _, _, solid := q.Get()
		if pos.X == x && pos.Y == y && solid.Blocks {
			return true
		}
	}
	return false
}
