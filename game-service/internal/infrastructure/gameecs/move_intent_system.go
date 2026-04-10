package gameecs

import (
	"encoding/json"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// MoveIntentCaptureSystem обрабатывает type=move: только обновляет MoveIntentStore (без шага по клеткам).
type MoveIntentCaptureSystem struct {
	mapper *ecs.Map4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health]
}

// NewMoveIntentCaptureSystem ...
func NewMoveIntentCaptureSystem(
	mapper *ecs.Map4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health],
) *MoveIntentCaptureSystem {
	return &MoveIntentCaptureSystem{mapper: mapper}
}

func (s *MoveIntentCaptureSystem) Update(ctx *TickContext) {
	if ctx.Intents == nil {
		return
	}
	a := ctx.CurrentAction
	if a.Type != "move" {
		return
	}

	var mv gamekit.MoveIntent
	if err := json.Unmarshal(a.Payload, &mv); err != nil {
		return
	}

	ent := ctx.Sink.EnsurePlayerEntity(a.PlayerID)
	if !s.mapper.HasAll(ent) {
		return
	}
	_, _, speed, _ := s.mapper.Get(ent)
	step := speed.MaxStep
	if step <= 0 {
		ctx.Intents.Set(a.PlayerID, 0, 0)
		return
	}

	dx, dy := mv.DX, mv.DY
	dx = clampInt(dx, -step, step)
	dy = clampInt(dy, -step, step)
	ctx.Intents.Set(a.PlayerID, dx, dy)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
