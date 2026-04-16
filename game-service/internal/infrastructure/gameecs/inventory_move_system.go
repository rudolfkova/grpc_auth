package gameecs

import (
	"encoding/json"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// InventoryMoveSystem обрабатывает type=inventory_move (swap двух слотов).
type InventoryMoveSystem struct {
	engine *Engine
}

func NewInventoryMoveSystem(e *Engine) *InventoryMoveSystem {
	return &InventoryMoveSystem{engine: e}
}

func (s *InventoryMoveSystem) Update(ctx *TickContext) {
	a := ctx.CurrentAction
	if a.Type != gamekit.TypeInventoryMove {
		return
	}
	var in gamekit.InventoryMoveIntent
	if err := json.Unmarshal(a.Payload, &in); err != nil {
		return
	}
	s.engine.applyInventoryMoveLocked(a.PlayerID, in.From, in.To)
}
