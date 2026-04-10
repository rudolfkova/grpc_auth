package gameecs

import (
	"game/internal/domain/models"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// SystemRegistry оркестрирует системы Ark.
type SystemRegistry struct {
	perAction []System
	postTick  []System
}

// NewSystemRegistry собирает дефолтный набор систем для одного World.
func NewSystemRegistry(
	w *ecs.World,
	playerMapper *ecs.Map4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health],
	playerFilter *ecs.Filter4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health],
	tileMapper *ecs.Map5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	tileFilter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
) *SystemRegistry {
	return &SystemRegistry{
		perAction: []System{
			NewMovementSystem(playerMapper, tileFilter),
			NewDamageSystem(playerMapper),
			NewTileSpawnSystem(w, tileMapper, tileFilter),
			NewTileClearSystem(w, tileFilter),
		},
		postTick: []System{
			NewSnapshotSystem(playerFilter, tileFilter),
		},
	}
}

// Update выполняет per-action системы для каждого действия, затем post-tick.
func (r *SystemRegistry) Update(sink PlayerEntitySink, actions []models.Action) ([]gamekit.Player, []gamekit.Tile) {
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

	return ctx.Players, ctx.Tiles
}
