package gameecs

import (
	"encoding/json"
	"testing"

	"game/internal/domain/models"
	"game/internal/domain/world"

	arkserde "github.com/mlange-42/ark-serde"
)

func TestEngine_MoveAndHit(t *testing.T) {
	e, err := NewEngine(nil)
	if err != nil {
		t.Fatal(err)
	}

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

func TestNewEngine_fromArkSerdeSnapshot(t *testing.T) {
	src, err := NewEngine(nil)
	if err != nil {
		t.Fatal(err)
	}
	ent := src.EnsurePlayerEntity(7)
	src.mu.Lock()
	_, pos, sp, hp := src.playerMapper.Get(ent)
	pos.X, pos.Y = 2, 3
	hp.HP = 5
	sp.MaxStep = 2
	src.mu.Unlock()

	snap, err := arkserde.Serialize(src.world)
	if err != nil {
		t.Fatal(err)
	}

	e, err := NewEngine(snap)
	if err != nil {
		t.Fatal(err)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	ent2 := e.EnsurePlayerEntity(7)
	_, pos2, sp2, hp2 := e.playerMapper.Get(ent2)
	if pos2.X != 2 || pos2.Y != 3 || hp2.HP != 5 || sp2.MaxStep != 2 {
		t.Fatalf("loaded state: pos=(%d,%d) hp=%d max_step=%d", pos2.X, pos2.Y, hp2.HP, sp2.MaxStep)
	}
}
