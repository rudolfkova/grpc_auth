package gameecs

import (
	"encoding/json"
	"testing"

	"game/internal/domain/models"
	"game/internal/domain/world"
)

func TestEngine_MoveAndHit(t *testing.T) {
	e := NewEngine()

	movePayload, _ := json.Marshal(models.MoveIntent{DX: 1, DY: 0})
	hitPayload, _ := json.Marshal(models.HitIntent{TargetID: 2, Damage: 3})

	e.ProcessTick([]models.Action{
		{PlayerID: 1, Type: "move", Payload: movePayload},
		{PlayerID: 1, Type: "hit", Payload: hitPayload},
	})

	e.mu.Lock()
	defer e.mu.Unlock()

	ent2 := e.EnsurePlayerEntity(2)
	_, _, _, hp := e.playerMapper.Get(ent2)
	if hp.HP != world.DefaultPlayerHP-3 {
		t.Fatalf("target HP: want %d, got %d", world.DefaultPlayerHP-3, hp.HP)
	}
	ent1 := e.EnsurePlayerEntity(1)
	_, pos, _, _ := e.playerMapper.Get(ent1)
	if pos.X != 1 || pos.Y != 0 {
		t.Fatalf("mover position: want (1,0), got (%d,%d)", pos.X, pos.Y)
	}
}
