package gameecs

import (
	"encoding/json"

	"game/internal/domain/models"
	"game/internal/domain/world"

	"github.com/mlange-42/ark/ecs"
)

// MovementSystem обрабатывает действия type=move.
type MovementSystem struct {
	mapper *ecs.Map4[world.PlayerRef, world.GridPos, world.Speed, world.Health]
}

// NewMovementSystem создаёт систему движения с общим mapper мира.
func NewMovementSystem(mapper *ecs.Map4[world.PlayerRef, world.GridPos, world.Speed, world.Health]) *MovementSystem {
	return &MovementSystem{mapper: mapper}
}

func (s *MovementSystem) Update(ctx *TickContext) {
	a := ctx.CurrentAction
	if a.Type != "move" {
		return
	}

	var mv models.MoveIntent
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
	pos.X += dx
	pos.Y += dy
}
