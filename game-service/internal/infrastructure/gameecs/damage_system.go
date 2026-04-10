package gameecs

import (
	"encoding/json"

	"game/internal/domain/models"
	"game/internal/domain/world"

	"github.com/mlange-42/ark/ecs"
)

// DamageSystem обрабатывает действия type=hit.
type DamageSystem struct {
	mapper *ecs.Map4[world.PlayerRef, world.GridPos, world.Speed, world.Health]
}

// NewDamageSystem создаёт систему урона с общим mapper мира.
func NewDamageSystem(mapper *ecs.Map4[world.PlayerRef, world.GridPos, world.Speed, world.Health]) *DamageSystem {
	return &DamageSystem{mapper: mapper}
}

func (s *DamageSystem) Update(ctx *TickContext) {
	a := ctx.CurrentAction
	if a.Type != "hit" {
		return
	}

	var hit models.HitIntent
	if err := json.Unmarshal(a.Payload, &hit); err != nil {
		return
	}
	if hit.TargetID == 0 || hit.Damage <= 0 {
		return
	}

	_ = ctx.Sink.EnsurePlayerEntity(a.PlayerID)
	targetEnt := ctx.Sink.EnsurePlayerEntity(hit.TargetID)
	s.applyDamage(targetEnt, hit.Damage)
}

func (s *DamageSystem) applyDamage(targetEnt ecs.Entity, damage int) {
	if damage <= 0 || !s.mapper.HasAll(targetEnt) {
		return
	}
	_, _, _, hp := s.mapper.Get(targetEnt)
	hp.HP -= damage
	if hp.HP < 0 {
		hp.HP = 0
	}
}
