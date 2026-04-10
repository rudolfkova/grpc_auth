package gameecs

import (
	"game/internal/domain/models"
	"game/internal/domain/world"

	"github.com/mlange-42/ark/ecs"
)

// SystemRegistry оркестрирует системы Ark.
type SystemRegistry struct {
	perAction []System
	postTick  []System
}

// NewSystemRegistry собирает дефолтный набор систем для одного World.
func NewSystemRegistry(
	mapper *ecs.Map4[world.PlayerRef, world.GridPos, world.Speed, world.Health],
	filter *ecs.Filter4[world.PlayerRef, world.GridPos, world.Speed, world.Health],
) *SystemRegistry {
	return &SystemRegistry{
		perAction: []System{
			NewMovementSystem(mapper),
			NewDamageSystem(mapper),
		},
		postTick: []System{
			NewSnapshotSystem(filter),
		},
	}
}

// Update выполняет per-action системы для каждого действия, затем post-tick.
func (r *SystemRegistry) Update(sink PlayerEntitySink, actions []models.Action) []models.Player {
	ctx := &TickContext{Sink: sink}

	for _, a := range actions {
		if a.PlayerID == 0 {
			continue
		}
		ctx.CurrentAction = a
		for _, sys := range r.perAction {
			sys.Update(ctx)
		}
	}

	ctx.CurrentAction = models.Action{}
	for _, sys := range r.postTick {
		sys.Update(ctx)
	}

	return ctx.Players
}
