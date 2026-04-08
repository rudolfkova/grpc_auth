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
	if mv.DX < -1 || mv.DX > 1 || mv.DY < -1 || mv.DY > 1 {
		return
	}

	st := e.ensureActor(a.PlayerID)
	st.Pos.X += mv.DX
	st.Pos.Y += mv.DY
	e.state[a.PlayerID] = st
}

func (e *Engine) applyHit(a models.Action, _ *emitter) {
	var hit models.HitIntent
	if err := json.Unmarshal(a.Payload, &hit); err != nil {
		return
	}
	if hit.TargetID == 0 || hit.Damage <= 0 {
		return
	}

	// Ensure attacker and target exist in world state.
	e.ensureActor(a.PlayerID)
	target := e.ensureActor(hit.TargetID)

	target.HP -= hit.Damage
	if target.HP < 0 {
		target.HP = 0
	}
	e.state[hit.TargetID] = target
}
