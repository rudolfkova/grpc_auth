package game

import (
	"encoding/json"

	"game/internal/domain/models"
)

// dispatchAction routes one action to a handler by type.
// Add new action types here only; implement logic in apply* methods.
func (e *Engine) dispatchAction(a models.Action, emit *emitter) {
	switch a.Type {
	case "move":
		e.applyMove(a, emit)
	case "hit":
		e.applyHit(a, emit)
	default:
		// Unknown action type: ignore.
	}
}

func (e *Engine) applyMove(a models.Action, _ *emitter) {
	var mv models.MoveIntent
	if err := json.Unmarshal(a.Payload, &mv); err != nil {
		return
	}

	ent := e.ensurePlayerEntity(a.PlayerID)
	e.runMovementStep(ent, mv.DX, mv.DY)
}

func (e *Engine) applyHit(a models.Action, _ *emitter) {
	var hit models.HitIntent
	if err := json.Unmarshal(a.Payload, &hit); err != nil {
		return
	}
	if hit.TargetID == 0 || hit.Damage <= 0 {
		return
	}

	e.ensurePlayerEntity(a.PlayerID)
	targetEnt := e.ensurePlayerEntity(hit.TargetID)
	e.runDamageStep(targetEnt, hit.Damage)
}
