package game

import (
	"encoding/json"
	"testing"

	"game/internal/domain/models"
)

func TestEngineECS_MoveAndHit(t *testing.T) {
	e := NewEngine()

	movePayload, _ := json.Marshal(models.MoveIntent{DX: 1, DY: 0})
	hitPayload, _ := json.Marshal(models.HitIntent{TargetID: 2, Damage: 3})

	e.ProcessTick([]models.Action{
		{PlayerID: 1, Type: "move", Payload: movePayload},
		{PlayerID: 1, Type: "hit", Payload: hitPayload},
	})

	e.mu.Lock()
	defer e.mu.Unlock()

	ent2 := e.ensurePlayerEntity(2)
	_, _, _, hp := e.playerMapper.Get(ent2)
	if hp.HP != defaultHP-3 {
		t.Fatalf("target HP: want %d, got %d", defaultHP-3, hp.HP)
	}
	ent1 := e.ensurePlayerEntity(1)
	_, pos, _, _ := e.playerMapper.Get(ent1)
	if pos.X != 1 || pos.Y != 0 {
		t.Fatalf("mover position: want (1,0), got (%d,%d)", pos.X, pos.Y)
	}
}
